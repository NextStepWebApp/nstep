package main

import (
	"fmt"
	"os"
	"os/exec"
)

func InstallNextStep(plj *packageLocalJson) error {

	installScript := plj.GetNextStepInstallScript()

	cmd := exec.Command("bash", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error executing nextstep install script %w\n", err)
	}

	return nil
}
