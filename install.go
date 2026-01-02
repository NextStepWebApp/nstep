package main

import (
	"fmt"
	"os/exec"
)

func InstallNextStep(plj *packageLocalJson) error {

	installScript := plj.GetNextStepInstallScript()

	cmd := exec.Command("./", installScript)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error executing nextstep install script %w\n", err)
	}

	return nil
}
