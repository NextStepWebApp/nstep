package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// This files is for all the functions that are spessifically for nextstep.go
// To make the lines of code of nextstep.go smaller and better to read

func setupMovesRollback(currentfilepath string) error {

	entries, err := os.ReadDir(currentfilepath)
	if err != nil {
		return fmt.Errorf("cannot read %s %w", currentfilepath, err)
	}

	for _, entry := range entries {

		dirName := fmt.Sprintf("%s/%s", currentfilepath, entry.Name())

		fmt.Printf("dirname: %s\n", dirName)

		realName := fmt.Sprintf("/%s", strings.ReplaceAll(entry.Name(), "-", "/"))

		fmt.Printf("realname: %s\n", realName)

		err = copyDir(dirName, realName)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

	}

	return nil
}

// Manager function that orchestrates the concurrency
func updateAllComponents(cfg config, plj *packageLocalJson, resultversion *versionCheck) (currentwebpath string, err error) {
	startNow := time.Now()

	ch := make(chan string)
	errCh := make(chan error, 2)
	var wg sync.WaitGroup

	if resultversion.isUpdatePackageAvailable() {
		wg.Add(1)
		go func() {
			if err := onlineToLocalPackage(cfg, resultversion, ch, &wg); err != nil {
				errCh <- err
			}
		}()
	}
	if resultversion.isUpdateWebAppAvailable() {
		wg.Add(1)
		go func() {
			currentwebpath, err = onlineToLocalWebApp(cfg, plj, resultversion, ch, &wg)

			if err != nil {
				errCh <- err
			}
		}()
	}

	// Close channel when all goroutines are done
	go func() {
		wg.Wait()
		close(ch)
		close(errCh)
	}()

	// Print results as they come in
	for result := range ch {
		fmt.Println(result)
	}

	// Check for errors
	for err := range errCh {
		if err != nil {
			return "", fmt.Errorf("update components %w", err)
		}
	}

	fmt.Printf("This operation took: %v\n", time.Since(startNow))

	return currentwebpath, nil
}

func onlineToLocalPackage(cfg config, resultversion *versionCheck, ch chan<- string, wg *sync.WaitGroup) error {
	defer wg.Done()

	var err error
	packageUrl := resultversion.getPackageURL()
	packagePath := cfg.getPackagePath()

	err = downloadpackage(packageUrl, packagePath)
	if err != nil {
		return fmt.Errorf("cannot download %s to %s %w", packageUrl, packagePath, err)
	}

	ch <- fmt.Sprintf("%s downloaded successfully", getPackageName(cfg))

	return nil
}

func onlineToLocalWebApp(cfg config, plj *packageLocalJson, resultversion *versionCheck, ch chan<- string, wg *sync.WaitGroup) (string, error) {
	defer wg.Done()

	var err error
	var filename string
	// format filepath to store download
	downloadpath := cfg.getDownloadPath()
	filename = fmt.Sprintf("nextstep_%s.tar.gz", resultversion.getLatestWebAppVersion())
	downloadfilepath := fmt.Sprintf("%s/%s", downloadpath, filename)

	err = downloadpackage(resultversion.getDownloadURL(), downloadfilepath)
	if err != nil {
		return "", fmt.Errorf("Error downloading package %w", err)
	}

	// Verifying package integrity

	err = verifyChecksum(downloadfilepath, resultversion.getChecksum())
	if err != nil {
		return "", fmt.Errorf("Verification failed %w", err)
	}
	ch <- fmt.Sprintf("%s verified successfully", plj.getName())

	// Extract the downloaded package, function from package.go
	versionpath := cfg.getVersionPath()
	filename = fmt.Sprintf("nextstep_%s", resultversion.getLatestWebAppVersion()) // also used in currentfilepath
	versionfilepath := fmt.Sprintf("%s/%s", versionpath, filename)

	err = extractpackage(downloadfilepath, versionfilepath, 1)
	if err != nil {
		return "", fmt.Errorf("Error extracting package %w: ", err)
	}

	ch <- fmt.Sprintf("%s extracted successfully", plj.getName())

	// Symlink the new version to the current one
	err = emptyDir(cfg.getCurrentPath())
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	currentfilepath := fmt.Sprintf("%s/%s", cfg.getCurrentPath(), filename)

	err = os.Symlink(versionfilepath, currentfilepath)
	if err != nil {
		return "", fmt.Errorf("Error symlinking %w", err)
	}

	ch <- fmt.Sprintf("%s downloaded successfully", plj.getName())

	return currentfilepath, nil
}

func setupMovesInstallUpdate(commandStatus string, plj *packageLocalJson) error {
	// Safty check
	if commandStatus != "install" && commandStatus != "update" {
		return fmt.Errorf("internal code error, wrong use of function")
	}

	var moveActions []moveAction
	var dirsToRemove []string

	switch commandStatus {
	case "install":
		moveActions = plj.getInstallMoveActions()
		dirsToRemove = plj.getInstallRemoves()
	case "update":
		moveActions = plj.getUpdateMoveActions()
		dirsToRemove = plj.getUpdateRemoves()
	}

	// Execute all moves
	for _, move := range moveActions {
		err := moveFile(move.From, move.To)
		if err != nil {
			return fmt.Errorf("Error moving file %w\n", err)
		}

		// Set up correct permissions

		err = nextstepPermissionManager(move)
		if err != nil {
			return fmt.Errorf("cannot set up permission for %s %w", move.To, err)
		}

		fmt.Printf("Moved: %s -> %s\n", move.From, move.To)
	}

	// Remove directories
	for _, dir := range dirsToRemove {
		err := removeDir(dir)
		if err != nil {
			return fmt.Errorf("Error removing directory %s %w", dir, err)
		}
		fmt.Printf("Removed: %s\n", dir)
	}

	return nil
}

/*func nextstepPermissionHelper(dir string, uid, gid, permission int) error {

entries, err := os.ReadDir(dir)
if err != nil {
	return fmt.Errorf("cannot read directory %s: %w", dir, err)
}

for _, entry := range entries {

	path := fmt.Sprintf("%s/%s", dir, entry.Name())

	if entry.IsDir() {
		// recurssion to also give chown in dirs
		if err := nextstepPermissionHelper(path, uid, gid, permission); err != nil {
			return err
		}
	} else {
		err = os.Chmod(path, os.FileMode(permission))
		if err != nil {
			fmt.Errorf("cannot change permission of %s %w\n", dir, err)
		}
		err = os.Chown(path, uid, gid)
		if err != nil {
			return fmt.Errorf("Error changing ownership of %s %w\n", dir, err)
		}
	}
}

return nil
}*/

func nextstepPermissionManager(moveAction moveAction) error {
	// The nextstepCreate function has to run first to make sure
	// that the required dirs are created
	var err error

	owner := moveAction.Owner
	group := moveAction.Group
	dir := moveAction.To
	permission := moveAction.Permissions

	uid, gid, err := getUidGid(owner)
	if err != nil {
		return fmt.Errorf("cannot get uid gid for %s %w\n", owner, err)
	}

	if group != owner {
		_, gid, err = getUidGid(group)
		if err != nil {
			return fmt.Errorf("cannot get uid gid for %s %w\n", group, err)
		}
	}

	// Directory permissions setup
	err = os.Chmod(dir, os.FileMode(permission))
	if err != nil {
		return fmt.Errorf("cannot change permission of %s %w\n", dir, err)
	}
	err = os.Chown(dir, uid, gid)
	if err != nil {
		return fmt.Errorf("cannot change ownership of directory %s %w\n", dir, err)

	}

	/*err = nextstepPermissionHelper(dir, uid, gid, permission)
	if err != nil {
		return fmt.Errorf("%w", err)
		}*/

	return nil
}

func nextStepCreate(plj *packageLocalJson) error {
	var err error
	// Make the required directories with the write permissions and ownerships
	dirs := plj.getRequiredDirInfo()

	for _, dir := range dirs {
		err = os.MkdirAll(dir.Dir, os.FileMode(dir.Permission))
		if err != nil {
			return fmt.Errorf("cannot create directory %s %w\n", dir.Dir, err)
		}
	}

	return nil
}

func nextStepBackup(cfg config, resultversion *versionCheck, settings settingsConfig, plj *packageLocalJson) error {
	// run the extra update stuff: bakup the nstep instance
	var err error
	var name string

	// make the version directory
	versionbackup := fmt.Sprintf("%s/%s", cfg.getBackupPath(), resultversion.getCurrentWebAppVersion())
	err = os.MkdirAll(versionbackup, os.FileMode(settings.getSettingPermissionDir()))
	if err != nil {
		return fmt.Errorf("cannot make %s %w", versionbackup, err)
	}

	dirs := plj.getRequiredDirInfo()

	for _, dir := range dirs {
		// make the dir name
		cleanPath := filepath.Clean(dir.Dir)
		safeName := strings.ReplaceAll(strings.Trim(cleanPath, "/"), "/", "-")
		name = fmt.Sprintf("%s/%s", versionbackup, safeName)

		// So what it does is:
		// var dir is like essential dirs in the config where config files are
		// var name is the location where to save the config files
		// Like /var/lib/backup/v0.1.12/etc-nextstepwebapp

		// Before I had rename, but this in a way resets the web app
		// So it needs to be copy
		err = copyDir(dir.Dir, name)
		if err != nil {
			return fmt.Errorf("cannot backup %s %w", dir.Dir, err)
		}
	}

	// Now need to move the web app source code itself
	cleanPath := filepath.Clean(plj.getLocalWebpath())
	safeName := strings.ReplaceAll(strings.Trim(cleanPath, "/"), "/", "-")
	name = fmt.Sprintf("%s/%s", versionbackup, safeName)
	err = os.Rename(plj.getLocalWebpath(), name)
	if err != nil {
		return fmt.Errorf("cannot backup web path %s: %w", plj.getLocalWebpath(), err)
	}

	// Now compress it to a compressed file (.tar.gz)
	fmt.Println("Compressing backup...")

	tarballPath := fmt.Sprintf("%s.tar.gz", versionbackup)

	// Create tarball
	cmd := exec.Command("tar", "-czf", tarballPath, "-C", cfg.getBackupPath(), resultversion.getCurrentWebAppVersion())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tarball: %w\n", err)
	}

	// Now remove the normal backup folder
	// So the leftover uncompressed folder
	err = os.RemoveAll(versionbackup)
	if err != nil {
		return fmt.Errorf("cannot remove %s %w\n", versionbackup, err)
	}

	return nil
}
