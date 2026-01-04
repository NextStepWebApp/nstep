package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
)

func InstallNextStep(plj *packageLocalJson, cfg config) error {
	var err error
	// Run the setup_nextstep.sh script
	// This is for manipulating files and starting services
	installScript := plj.GetNextStepInstallScript()
	cmd := exec.Command("bash", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("error executing nextstep install script %w\n", err)
	}

	dirs := plj.GetRequiredDirs()

	// Make the required directories with the write permissions and ownerships
	for _, dir := range dirs {
		if dir == "/var/lib/nextstepwebapp" {
			err = os.MkdirAll(dir, 0775)
			// Get the uid, gid for the chown function
			uid, gid, err := getUidGid("http")
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

	// Download and place the Nextstep code correctly

	// The same process as in update.go but the local version is just v0.0.0
	resultversion, err := Versionchecker(cfg, plj)
	err = NextStepSetup(cfg, resultversion)

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

	fmt.Println("Nextstep installation completed successfully!")

	return nil
}

// a function that moves a file from old path to new path
func moveFile(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("failed to move %s to %s: %w", oldPath, newPath, err)
	}
	return nil
}

// function that removes a directory
func removeDir(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", path, err)
	}
	return nil
}

// This function gets the uid, gid from the group you give it
// Usefull to use with chown
func getUidGid(group string) (uid int, gid int, err error) {
	groupuser, err := user.Lookup(group)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot find group %w\n", err)
	}

	uid, err = strconv.Atoi(groupuser.Uid)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get uid %w\n", err)
	}
	gid, err = strconv.Atoi(groupuser.Gid)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get gid %w\n", err)
	}

	return uid, gid, nil
}
