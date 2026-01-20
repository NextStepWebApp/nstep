package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// The sourceDir is only meant for the rollback functionality
func nextStepSetup(cfg config, resultversion *versionCheck, plj *packageLocalJson, status *status, sourceDir *string) error {
	var err error

	// Safty check to see if this is a install or update
	commandStatus, err := status.getStatus()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// These are environment checks

	setupStatus := false
	_, err = os.ReadDir(plj.getLocalWebpath())
	if err != nil {
		if os.IsNotExist(err) {
			setupStatus = false //install
		} else {
			return fmt.Errorf("cannot check installation status: %w", err)
		}
	} else {
		setupStatus = true //update
	}

	switch {

	// So a broken update setup
	case setupStatus == false && commandStatus == "update":
		return fmt.Errorf("command given is update system says install, run 'sudo nstep install'")

	// Working installation
	case setupStatus == false && commandStatus == "install":
		if sourceDir != nil {
			return fmt.Errorf("internal error: install command, but source directory is not equal to nil")
		}
		fmt.Println("==> Installation setup...")

	// Alreaddy installed
	case setupStatus == true && commandStatus == "install":
		return fmt.Errorf("nextstep is already installed, run 'sudo nstep update'")

	// Working update setup
	case setupStatus == true && commandStatus == "update":
		if sourceDir != nil {
			return fmt.Errorf("internal error: update command, but source directory is not equal to nil")
		}
		fmt.Println("==> Update setup...")

	// Invalid rollback use, needs to install first
	case setupStatus == false && commandStatus == "rollback":
		return fmt.Errorf("nextstep is not installed, run 'sudo nstep install'")

	// Working rollback setup
	case setupStatus == true && commandStatus == "rollback":
		// this is a saftey check for code usage
		if sourceDir == nil {
			return fmt.Errorf("internal error: rollback command, but source directory is nil")
		}

		fmt.Println("==> Rollback setup..")
	}
	// End of checks

	// Nstep lock check
	lockfile, err := lockNstep(cfg)
	if err != nil {
		return fmt.Errorf("Error update.lock %w", err)
	}
	defer lockfile.Close()
	defer os.Remove(cfg.getLockFilePath())

	var currentfilepath string
	if commandStatus == "install" || commandStatus == "update" {
		currentfilepath, err = onlineToLocal(cfg, resultversion)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	// Run extra update and rollback stuff
	if commandStatus == "update" || commandStatus == "rollback" {
		err = nextStepBackup(cfg, resultversion, *plj)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	// Create or recreate the nextstep structure (destroyed by the renames!)
	err = nextStepCreate(*plj)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Move changes between update/install and rollback

	// Setup moves for install and update
	switch commandStatus {
	case "update", "install":
		// get the current version to the web portal
		err = copyDir(currentfilepath, plj.getLocalWebpath())
		if err != nil {
			return fmt.Errorf("cannot copy current to webpath %w", err)
		}
		// put the config files etc in the right place and remove unused files in the web portal
		err = setupMovesUpdateInstall(plj)
		if err != nil {
			return fmt.Errorf("cannot do the setup moves %w", err)
		}
	case "rollback":
		// Name the currentfilepath for rollback
		// sourceDir is passed throug the function
		currentfilepath = *sourceDir
		err = setupMovesRollback(currentfilepath)
		if err != nil {
			return fmt.Errorf("cannot do the setup moves %w", err)
		}

	}

	// Now update the version in the local package.json
	err = localpackageupdater(plj, resultversion, cfg)
	if err != nil {
		return fmt.Errorf("Error updating local package %w", err)
	}

	return nil
}

func setupMovesRollback(currentfilepath string) error {

	entries, err := os.ReadDir(currentfilepath)
	if err != nil {
		return fmt.Errorf("cannot read %s %w", currentfilepath, err)
	}

	for _, entry := range entries {

		dirName := fmt.Sprintf("%s/%s", currentfilepath, entry.Name())

		fmt.Printf("dirname: %s\n", dirName)

		realName := strings.ReplaceAll(entry.Name(), "-", "/")

		fmt.Printf("realname: %s\n", realName)

		err = copyDir(dirName, realName)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

	}

	return nil
}

func onlineToLocal(cfg config, resultversion *versionCheck) (string, error) {
	var err error
	var filename, message string
	// format filepath to store download
	downloadpath := cfg.getDownloadPath()
	filename = fmt.Sprintf("nextstep_%s.tar.gz", resultversion.getLatestVersion())
	downloadfilepath := fmt.Sprintf("%s/%s", downloadpath, filename)

	message, err = downloadpackage(resultversion.getDownloadURL(), downloadfilepath)

	if err != nil {
		return "", fmt.Errorf("Error downloading package %w", err)
	}
	println(message)

	// Verifying package integrity

	err = verifyChecksum(downloadfilepath, resultversion.getChecksum())

	if err != nil {
		return "", fmt.Errorf("Verification failed %w", err)
	} else {
		fmt.Println("Package verified successfully")
	}

	// Extract the downloaded package, function from package.go
	versionpath := cfg.getVersionPath()
	filename = fmt.Sprintf("nextstep_%s", resultversion.LatestVersion) // also used in currentfilepath
	versionfilepath := fmt.Sprintf("%s/%s", versionpath, filename)

	message, err = extractpackage(downloadfilepath, versionfilepath, 1)
	if err != nil {
		return "", fmt.Errorf("Error extracting package %w: ", err)
	}
	println(message)

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
	return currentfilepath, nil
}

func setupMovesUpdateInstall(plj *packageLocalJson) error {
	var err error

	// Move all the files to there places
	moves := [][2]string{
		{"/srv/http/NextStep/config/nextstep_config.json", "/etc/nextstepwebapp/nextstep_config.json"},
		{"/srv/http/NextStep/config/branding.json", "/var/lib/nextstepwebapp/branding.json"},
		{"/srv/http/NextStep/config/config.json", "/var/lib/nextstepwebapp/config.json"},
		{"/srv/http/NextStep/config/theme.json", "/var/lib/nextstepwebapp/theme.json"},
		{"/srv/http/NextStep/config/errors.json", "/var/lib/nextstepwebapp/errors.json"},
		{"/srv/http/NextStep/config/setup.json", "/var/lib/nextstepwebapp/setup.json"},
		{"/srv/http/NextStep/data/import.py", "/opt/nextstepwebapp/import.py"},
	}

	// Execute all moves
	for _, move := range moves {
		err := moveFile(move[0], move[1])
		if err != nil {
			return fmt.Errorf("Error moving file %w\n", err)
		}
		fmt.Printf("Moved: %s -> %s\n", move[0], move[1])
	}

	// Remove some dirs
	dirsToRemove := []string{
		"/srv/http/NextStep/config",
		"/srv/http/NextStep/data",
	}

	// Remove directories
	for _, dir := range dirsToRemove {
		err := removeDir(dir)
		if err != nil {
			return fmt.Errorf("Error removing directory %s %w", dir, err)
		}
		fmt.Printf("Removed: %s\n", dir)
	}

	// Now give all the stuff the correct permssion and ownership
	err = nextstepPermissionManager(plj)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func nextstepPermissionHelper(dir string, uid, gid int) error {

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("cannot read directory %s: %w", dir, err)
	}

	for _, entry := range entries {

		path := fmt.Sprintf("%s/%s", dir, entry.Name())

		if entry.IsDir() {
			// recurssion to also give chown in dirs
			if err := nextstepPermissionHelper(path, uid, gid); err != nil {
				return err
			}
		} else {
			err = os.Chown(path, uid, gid)
			if err != nil {
				return fmt.Errorf("Error changing ownership of %s %w\n", dir, err)
			}
		}
	}

	return nil
}

func nextstepPermissionManager(plj *packageLocalJson) error {
	// The nextstepCreate function has to run first to make sure
	// that the required dirs are created
	var err error
	dirs := plj.getRequiredDirs()

	// Get the uid and gid of http for chown
	uid, gid, err := getUidGid("http")
	fmt.Printf("uid: %d\ngid: %d\n", uid, gid)
	if err != nil {
		return fmt.Errorf("Error get uid gid %w\n", err)
	}

	for _, dir := range dirs {
		// Chown the directory itself first
		err = os.Chown(dir, uid, gid)
		if err != nil {
			return fmt.Errorf("Error changing ownership of directory %s: %w", dir, err)
		}
		err = nextstepPermissionHelper(dir, uid, gid)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

func nextStepCreate(plj packageLocalJson) error {
	var err error
	// Make the required directories with the write permissions and ownerships
	dirs := plj.getRequiredDirs()

	for _, dir := range dirs {
		if dir == "/var/lib/nextstepwebapp" {
			err = os.MkdirAll(dir, 0775)
			if err != nil {
				return fmt.Errorf("cannot create directory %s %w\n", dir, err)
			}

		} else {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return fmt.Errorf("cannot create directory %s %w\n", dir, err)
			}
		}

	}
	return nil
}

func nextStepBackup(cfg config, resultversion *versionCheck, plj packageLocalJson) error {
	// run the extra update stuff: bakup the nstep instance
	var err error
	var name string

	// make the version directory
	versionbackup := fmt.Sprintf("%s/%s", cfg.getBackupPath(), resultversion.getCurrentVersion())
	err = os.MkdirAll(versionbackup, 0755)
	if err != nil {
		return fmt.Errorf("cannot make %s %w", versionbackup, err)
	}

	dirs := plj.getRequiredDirs()

	for _, dir := range dirs {
		// make the dir name
		cleanPath := filepath.Clean(dir)
		safeName := strings.ReplaceAll(strings.Trim(cleanPath, "/"), "/", "-")
		name = fmt.Sprintf("%s/%s", versionbackup, safeName)
		err = os.Rename(dir, name)
		if err != nil {
			return fmt.Errorf("cannot backup %s %w", dir, err)
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
	cmd := exec.Command("tar", "-czf", tarballPath, "-C", cfg.getBackupPath(), resultversion.getCurrentVersion())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tarball: %w\n", err)
	}

	// Now remove the normal backup folder
	err = os.RemoveAll(versionbackup)
	if err != nil {
		return fmt.Errorf("cannot remove %s %w\n", versionbackup, err)
	}

	return nil
}
