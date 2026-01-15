package main

import (
	"fmt"
	"os"
	"os/exec"
)

func InstallNextStep(plj *packageLocalJson, cfg config, status *Status) error {
	var err error

	// The same process as in update.go but the local version is just v0.0.0
	resultversion, err := Versionchecker(cfg, plj)
	if err != nil {
		return fmt.Errorf("Error checking version %w\n", err)
	}
	err = NextStepSetup(cfg, resultversion, plj, status)
	if err != nil {
		return fmt.Errorf("Error NextStepWebApp setup %w\n", err)
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

	fmt.Println("Nextstep installation completed successfully!")

	return nil
}
