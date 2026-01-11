package main

import (
	"fmt"
	"os"
	"os/exec"
)

func InstallNextStep(plj *packageLocalJson, cfg config, status *Status) error {
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

	// The same process as in update.go but the local version is just v0.0.0
	resultversion, err := Versionchecker(cfg, plj)
	err = NextStepSetup(cfg, resultversion, plj, status)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Run the setup_nextstep.sh script
	// This is for manipulating files and starting services
	installScript := plj.GetNextStepInstallScript()
	cmd := exec.Command("bash", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("error executing nextstep install script %w\n", err)
	}

	// Make hiddenfiles to say that the install is done

	fmt.Println("Nextstep installation completed successfully!")

	return nil
}
