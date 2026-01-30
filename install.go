package main

import (
	"fmt"
	"os"
	"os/exec"
)

func installNextStep(plj *packageLocalJson, cfg config, settings settingsConfig, state *state, status *status) error {
	var err error

	fmt.Printf("%s %s installation setup...\n", green("===>"), plj.getName())

	// The same process as in update.go but the local version is just v0.0.0
	resultversion, err := versionchecker(plj, state, cfg)
	if err != nil {
		return err
	}
	err = nextStepSetup(cfg, resultversion, plj, settings, state, status, nil)
	if err != nil {
		return err
	}

	// Run the setup_nextstep.sh script
	// This is for manipulating files and starting services
	installScript := plj.getNextStepInstallScript()
	cmd := exec.Command("bash", installScript)

	if settings.getOutputStatus() {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("%s -  cannot execute nextstep install script", red("ERROR"))
	}

	fmt.Printf("%s %s installation completed successfully\n", green("===>"), plj.getName())

	return nil
}
