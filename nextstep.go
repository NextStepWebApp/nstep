package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func NextStepSetup(cfg config, resultversion *versionCheck, plj *packageLocalJson, status *Status) error {
	var message string
	var err error
	var filename string

	// Safty check to see if this is a install or update

	commandStatus, err := status.GetStatus()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	setupStatus := false
	_, err = os.ReadDir(plj.GetLocalWebpath())
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
		return fmt.Errorf("command given is update system says install, run 'sudo nstep install'\n")

	// Working installation
	case setupStatus == false && commandStatus == "install":
		fmt.Println("Installation setup...")

	// Alreaddy installed
	case setupStatus == true && commandStatus == "install":
		return fmt.Errorf("NextStep is already installed, run 'sudo nstep update'\n")

	// Working update setup
	case setupStatus == true && commandStatus == "update":
		fmt.Println("Update setup...")
	}

	// Nstep lock check
	lockfile, err := LockNstep(cfg)
	if err != nil {
		return fmt.Errorf("Error update.lock %w", err)
	}
	defer lockfile.Close()
	defer os.Remove(cfg.GetLockFilePath())

	// format filepath to store download
	downloadpath := cfg.GetDownloadPath()
	filename = fmt.Sprintf("nextstep_%s.tar.gz", resultversion.GetLatestVersion())
	downloadfilepath := fmt.Sprintf("%s/%s", downloadpath, filename)

	message, err = Downloadpackage(resultversion.GetDownloadURL(), downloadfilepath)

	if err != nil {
		return fmt.Errorf("Error downloading package %w", err)
	}
	println(message)

	// Verifying package integrity

	err = VerifyChecksum(downloadfilepath, resultversion.GetChecksum())

	if err != nil {
		return fmt.Errorf("Verification failed %w", err)
	} else {
		fmt.Println("Package verified successfully")
	}

	// Extract the downloaded package, function from package.go
	versionpath := cfg.GetVersionPath()
	filename = fmt.Sprintf("nextstep_%s", resultversion.LatestVersion) // also used in currentfilepath
	versionfilepath := fmt.Sprintf("%s/%s", versionpath, filename)

	message, err = Extractpackage(downloadfilepath, versionfilepath)
	if err != nil {
		return fmt.Errorf("Error extracting package %w: ", err)
	}
	println(message)

	// Symlink the new version to the current one
	err = EmtyDir(cfg.GetCurrentPath())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	currentfilepath := fmt.Sprintf("%s/%s", cfg.GetCurrentPath(), filename)

	err = os.Symlink(versionfilepath, currentfilepath)
	if err != nil {
		return fmt.Errorf("Error symlinking %w", err)
	}

	// Run extra update stuff
	if commandStatus == "update" {
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

	// get the current version to the web portal
	err = CopyDir(currentfilepath, plj.GetLocalWebpath())
	if err != nil {
		return fmt.Errorf("Error copy current to webpath %w\n", err)
	}

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
		err := MoveFile(move[0], move[1])
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
		err := RemoveDir(dir)
		if err != nil {
			return fmt.Errorf("Error removing directory %s %w", dir, err)
		}
		fmt.Printf("Removed: %s\n", dir)
	}

	//err = updatemove(resultversion, plj, cfg)
	// get the current code and move to web portal
	// Update backs up the db (so db is seperate between update)
	// New db gets updated by scripts if needed
	return nil
}

func nextStepCreate(plj packageLocalJson) error {
	var err error
	// Make the required directories with the write permissions and ownerships
	dirs := plj.GetRequiredDirs()

	for _, dir := range dirs {
		if dir == "/var/lib/nextstepwebapp" {
			err = os.MkdirAll(dir, 0775)
			// Get the uid, gid for the chown function
			uid, gid, err := GetUidGid("http")
			if err != nil {
				return fmt.Errorf("Error get uid gid %w\n", err)
			}
			err = os.Chown(dir, uid, gid)
			if err != nil {
				return fmt.Errorf("Error changing ownership of %s %w\n", dir, err)
			}

		} else {
			err = os.MkdirAll(dir, 0755)
		}
		if err != nil {
			return fmt.Errorf("cannot create directory %s %w\n", dir, err)
		}
	}
	return nil
}

func nextStepBackup(cfg config, resultversion *versionCheck, plj packageLocalJson) error {
	// run the extra update stuff: bakup the nstep instance
	var err error
	var name string

	// make the version directory
	versionbackup := fmt.Sprintf("%s/%s", cfg.GetBackupPath(), resultversion.GetCurrentVersion())
	err = os.MkdirAll(versionbackup, 0755)
	if err != nil {
		return fmt.Errorf("cannot make %s %w", versionbackup, err)
	}

	dirs := plj.GetRequiredDirs()

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
	cleanPath := filepath.Clean(plj.GetLocalWebpath())
	safeName := strings.ReplaceAll(strings.Trim(cleanPath, "/"), "/", "-")
	name = fmt.Sprintf("%s/%s", versionbackup, safeName)
	err = os.Rename(plj.GetLocalWebpath(), name)
	if err != nil {
		return fmt.Errorf("cannot backup web path %s: %w", plj.GetLocalWebpath(), err)
	}

	// Now compress it to a compressed file (.tar.gz)
	fmt.Println("Compressing backup...")

	tarballPath := fmt.Sprintf("%s.tar.gz", versionbackup)

	// Create tarball
	cmd := exec.Command("tar", "-czf", tarballPath, "-C", cfg.GetBackupPath(), resultversion.GetCurrentVersion())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tarball: %w\n", err)
	}

	return nil
}
