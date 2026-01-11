package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func NextStepSetup(cfg config, resultversion *versionCheck, plj *packageLocalJson) error {
	var message string
	var err error
	var filename string

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

	// Safty check to see if this is a install or update
	setupStatus := false
	_, err = os.ReadDir(plj.GetLocalWebpath())
	if err != nil {
		if os.IsNotExist(err) {
			setupStatus = false
		} else {
			return fmt.Errorf("cannot check installation status: %w", err)
		}
	} else {
		setupStatus = true //update
	}
	if setupStatus == true {
		var err error
		var name string

		// run the extra update stuff: bakup the nstep instance

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
		name = fmt.Sprintf("%s/%s", versionbackup, plj.GetLocalWebpath())
		err = os.Rename(plj.GetLocalWebpath(), name)

		// Now compress it to a compressed file (.tar.gz)
		fmt.Println("Compress part")

	}

	// get the current version to the web portal
	err = CopyDir(currentfilepath, plj.GetLocalWebpath())

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
