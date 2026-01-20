package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func updateNextStep(cfg config, plj *packageLocalJson, status *status) error {
	// function from package.go uses methods to get information
	resultversion, err := versionchecker(plj)
	if err != nil {
		return fmt.Errorf("Error checking version %w\n", err)
	}

	if resultversion.isUpdateAvailable() {
		fmt.Println(resultversion.getMessage())
		fmt.Printf("New version available: %s\n", resultversion.getLatestVersion())
		fmt.Printf("Download: %s\n", resultversion.getDownloadURL())
		fmt.Printf("Release notes: %s\n", resultversion.getReleaseNotes())
	} else {
		return errors.New(resultversion.getMessage())
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

		err := nextStepSetup(cfg, resultversion, plj, status, nil)
		if err != nil {
			return fmt.Errorf("Error NextStepWebApp setup %w\n", err)
		}

	} else {
		fmt.Println("Installation cancelled")
	}

	return nil
}
