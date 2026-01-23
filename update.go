package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func updateNextStep(cfg config, plj *packageLocalJson, state *state, status *status) error {
	// function from package.go uses methods to get information
	resultversion, err := versionchecker(plj, state, cfg)
	if err != nil {
		return fmt.Errorf("Error checking version %w\n", err)
	}

	// Check for package.json update
	if resultversion.isUpdatePackageAvailable() {
		fmt.Println(resultversion.getMessagePackage())
		fmt.Printf("New %s version available: %d\n", getPackageName(cfg), resultversion.getLatestPackageVersion())
		fmt.Printf("Download: %s\n", resultversion.getPackageURL())
	} else {
		return fmt.Errorf("%s", resultversion.getMessagePackage())
	}

	if resultversion.isUpdateWebAppAvailable() {
		fmt.Println(resultversion.getMessageWebApp())
		fmt.Printf("New %s version available: %s\n", plj.getname(), resultversion.getLatestWebAppVersion())
		fmt.Printf("Download: %s\n", resultversion.getDownloadURL())
		fmt.Printf("Release notes: %s\n", resultversion.getReleaseNotes())
	} else {
		return errors.New(resultversion.getMessageWebApp())
	}

	// confirmation part if there is a update
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Proceed with installation? [Y/n] ")

	response, err := reader.ReadString('\n')

	if err != nil {
		return fmt.Errorf("Error reading input %w", err)
	}

	response = strings.TrimSpace(response)

	if response == "Y" || response == "y" || response == "" ||
		response == "Yes" || response == "yes" {

		err := nextStepSetup(cfg, resultversion, plj, state, status, nil)
		if err != nil {
			return fmt.Errorf("Error NextStepWebApp setup %w\n", err)
		}

	} else {
		fmt.Println("Installation cancelled")
	}

	return nil
}
