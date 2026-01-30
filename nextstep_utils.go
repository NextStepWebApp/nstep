package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// This files is for all the functions that are spessifically for nextstep.go
// To make the lines of code of nextstep.go smaller and better to read

func setupMovesRollback(currentfilepath string, settings settingsConfig) error {

	entries, err := os.ReadDir(currentfilepath)
	if err != nil {
		return fmt.Errorf("cannot read %s %w", currentfilepath, err)
	}

	for _, entry := range entries {

		dirName := fmt.Sprintf("%s/%s", currentfilepath, entry.Name())

		fmt.Printf("dirname: %s\n", dirName)

		realName := fmt.Sprintf("/%s", strings.ReplaceAll(entry.Name(), "-", "/"))

		fmt.Printf("realname: %s\n", realName)

		err = copyDir(dirName, realName, settings)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

	}

	return nil
}

func updateAllComponents(cfg config, settings settingsConfig, plj *packageLocalJson, resultversion *versionCheck) (currentwebpath string, err error) {
	startNow := time.Now()

	fmt.Printf("%s Component setup...\n", green("===>"))

	// Update package if available
	if resultversion.isUpdatePackageAvailable() {
		err := onlineToLocalPackage(cfg, resultversion, settings)
		if err != nil {
			return "", fmt.Errorf("%s - cannot update package component: %w", red("ERROR"), err)
		}
	}

	// Update webapp if available
	if resultversion.isUpdateWebAppAvailable() {
		currentwebpath, err = onlineToLocalWebApp(cfg, plj, resultversion, settings)
		if err != nil {
			return "", fmt.Errorf("%s - cannot update webapp component: %w", red("ERROR"), err)
		}
	}

	message := fmt.Sprintf("%s This operation took: %v", yellow(" ->"), time.Since(startNow))
	verbosePrint(message, settings)

	fmt.Printf("%s Component setup finished successfully\n", green("===>"))

	return currentwebpath, nil
}

func onlineToLocalPackage(cfg config, resultversion *versionCheck, settings settingsConfig) error {
	packageUrl := resultversion.getPackageURL()
	packagePath := cfg.getPackagePath()

	err := downloadpackage(packageUrl, packagePath)
	if err != nil {
		return fmt.Errorf("cannot download %s to %s", packageUrl, packagePath)
	}

	message := fmt.Sprintf("%s %s downloaded successfully", yellow(" ->"), getPackageName(cfg))
	verbosePrint(message, settings)
	return nil
}

func onlineToLocalWebApp(cfg config, plj *packageLocalJson, resultversion *versionCheck, settings settingsConfig) (string, error) {
	var filename, message string

	// Format filepath to store download
	downloadpath := cfg.getDownloadPath()
	filename = fmt.Sprintf("nextstep_%s.tar.gz", resultversion.getLatestWebAppVersion())
	downloadfilepath := fmt.Sprintf("%s/%s", downloadpath, filename)

	err := downloadpackage(resultversion.getDownloadURL(), downloadfilepath)
	if err != nil {
		return "", fmt.Errorf("cannot download %s", plj.getName())
	}

	message = fmt.Sprintf("%s %s downloaded successfully", yellow(" ->"), plj.getName())
	verbosePrint(message, settings)

	// Verifying package integrity
	err = verifyChecksum(downloadfilepath, resultversion.getChecksum())
	if err != nil {
		return "", fmt.Errorf("%s verification failed", plj.getName())
	}

	message = fmt.Sprintf("%s %s verified successfully", yellow(" ->"), plj.getName())
	verbosePrint(message, settings)

	// Extract the downloaded package
	versionpath := cfg.getVersionPath()
	filename = fmt.Sprintf("nextstep_%s", resultversion.getLatestWebAppVersion())
	versionfilepath := fmt.Sprintf("%s/%s", versionpath, filename)

	err = extractpackage(downloadfilepath, versionfilepath, 1)
	if err != nil {
		return "", fmt.Errorf("error extracting %s", plj.getName())
	}
	message = fmt.Sprintf("%s %s extracted successfully", yellow(" ->"), plj.getName())
	verbosePrint(message, settings)

	// Symlink the new version to the current one
	err = emptyDir(cfg.getCurrentPath())
	if err != nil {
		return "", fmt.Errorf("cannot empty directory: %w", err)
	}

	currentfilepath := fmt.Sprintf("%s/%s", cfg.getCurrentPath(), filename)
	err = os.Symlink(versionfilepath, currentfilepath)
	if err != nil {
		return "", fmt.Errorf("cannot create symlink")
	}

	return currentfilepath, nil
}

func setupMovesInstallUpdate(commandStatus string, plj *packageLocalJson, settings settingsConfig) error {
	// Safty check
	if commandStatus != "install" && commandStatus != "update" {
		return fmt.Errorf("%s - internal code error, wrong use of function", red("ERROR"))
	}

	var moveActions []moveAction
	var dirsToRemove []string
	var message string

	switch commandStatus {
	case "install":
		moveActions = plj.getInstallMoveActions()
		dirsToRemove = plj.getInstallRemoves()
	case "update":
		moveActions = plj.getUpdateMoveActions()
		dirsToRemove = plj.getUpdateRemoves()
	}

	fmt.Printf("%s install/update setup...\n", green("===>"))

	message = fmt.Sprintf("%s moving files", yellow(" ->"))
	verbosePrint(message, settings)

	// Execute all moves
	for _, move := range moveActions {
		err := moveFile(move.From, move.To)
		if err != nil {
			return fmt.Errorf("%s - cannot move file %s -> %s", red("ERROR"), move.From, move.To)
		}

		// Set up correct permissions

		err = nextstepPermissionManager(move)
		if err != nil {
			return fmt.Errorf("%s - cannot set up permission for %s", red("ERROR"), move.To)
		}

		message = fmt.Sprintf("%s %s -> %s", cyan("  - move"), move.From, move.To)
		verbosePrint(message, settings)
	}

	message = fmt.Sprintf("%s removing directories", yellow(" ->"))
	verbosePrint(message, settings)

	// Remove directories
	for _, dir := range dirsToRemove {
		err := removeDir(dir)
		if err != nil {
			return fmt.Errorf("%s - removing %s", red("ERROR"), dir)
		}
		fmt.Printf("%s %s\n", red("  - remove"), dir)
	}

	fmt.Printf("%s install/update setup finished successfully\n", green("===>"))

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
		//return fmt.Errorf("cannot get uid gid for %s %w", owner, err)
		return err
	}

	if group != owner {
		_, gid, err = getUidGid(group)
		if err != nil {
			//return fmt.Errorf("cannot get uid gid for %s %w", group, err)
			return err
		}
	}

	// Directory permissions setup
	err = os.Chmod(dir, os.FileMode(permission))
	if err != nil {
		//return fmt.Errorf("cannot change permission of %s %w", dir, err)
		return err
	}
	err = os.Chown(dir, uid, gid)
	if err != nil {
		//return fmt.Errorf("cannot change ownership of directory %s %w", dir, err)
		return err

	}

	/*err = nextstepPermissionHelper(dir, uid, gid, permission)
	if err != nil {
		return fmt.Errorf("%w", err)
		}*/

	return nil
}

func nextStepCreate(plj *packageLocalJson, settings settingsConfig) error {
	var err error

	fmt.Printf("%s %s required directory setup\n", green("===>"), plj.getName())

	// Make the required directories with the write permissions and ownerships
	dirs := plj.getRequiredDirInfo()

	for _, dir := range dirs {
		err = os.MkdirAll(dir.Dir, os.FileMode(dir.Permission))
		if err != nil {
			return fmt.Errorf("%s - cannot create directory %s", red("ERROR"), dir.Dir)
		} else {
			message := fmt.Sprintf("%s created %s", yellow(" ->"), dir.Dir)
			verbosePrint(message, settings)
		}
	}

	fmt.Printf("%s %s required directory setup completed successfully\n", green("===>"), plj.getName())

	return nil
}

func nextStepBackup(cfg config, resultversion *versionCheck, settings settingsConfig, plj *packageLocalJson) error {
	// run the extra update stuff: bakup the nstep instance
	var err error
	var name, message string

	fmt.Printf("%s %s %s backup...\n", green("===>"), plj.getName(), resultversion.getCurrentWebAppVersion())

	// make the version directory
	versionbackup := fmt.Sprintf("%s/%s", cfg.getBackupPath(), resultversion.getCurrentWebAppVersion())
	err = os.MkdirAll(versionbackup, os.FileMode(settings.getSettingPermissionDir()))
	if err != nil {
		return fmt.Errorf("%s - cannot make %s", red("ERROR"), versionbackup)
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

		message = fmt.Sprintf("%s copying directories", yellow(" ->"))
		verbosePrint(message, settings)

		err = copyDir(dir.Dir, name, settings)
		if err != nil {
			return fmt.Errorf("%s - cannot backup %s", red("ERROR"), dir.Dir)
		}
	}

	// Now need to move the web app source code itself
	message = fmt.Sprintf("%s moving web app source code...", yellow(" ->"))
	verbosePrint(message, settings)

	cleanPath := filepath.Clean(plj.getLocalWebpath())
	safeName := strings.ReplaceAll(strings.Trim(cleanPath, "/"), "/", "-")
	name = fmt.Sprintf("%s/%s", versionbackup, safeName)
	err = os.Rename(plj.getLocalWebpath(), name)
	if err != nil {
		return fmt.Errorf("%s - cannot backup %s", red("ERROR"), plj.getLocalWebpath())
	}

	// Now compress it to a compressed file (.tar.gz)
	message = fmt.Sprintf("%s compressing backup...", yellow(" ->"))
	verbosePrint(message, settings)

	tarballPath := fmt.Sprintf("%s.tar.gz", versionbackup)

	// Create tarball

	message = fmt.Sprintf("%s creating tarball...", yellow(" ->"))
	verbosePrint(message, settings)

	cmd := exec.Command("tar", "-czf", tarballPath, "-C", cfg.getBackupPath(), resultversion.getCurrentWebAppVersion())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s - failed to create tarball", red("ERROR"))
	}

	// Now remove the normal backup folder
	// So the leftover uncompressed folder
	message = fmt.Sprintf("%s cleaning up...", yellow(" ->"))
	verbosePrint(message, settings)

	err = os.RemoveAll(versionbackup)
	if err != nil {
		return fmt.Errorf("%s cannot remove %s", red("ERROR"), versionbackup)
	}

	fmt.Printf("%s %s %s backup completed successfully\n", green("===>"), plj.getName(), resultversion.getCurrentWebAppVersion())

	return nil
}
