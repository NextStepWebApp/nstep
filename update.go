package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func UpdateNextStep(cfg config, plj *packageLocalJson, status *Status) error {
	// function from package.go uses methods to get information
	resultversion, err := Versionchecker(cfg, plj)
	if err != nil {
		return fmt.Errorf("Error checking version %w\n", err)
	}

	if resultversion.IsUpdateAvailable() {
		fmt.Println(resultversion.GetMessage())
		fmt.Printf("New version available: %s\n", resultversion.GetLatestVersion())
		fmt.Printf("Download: %s\n", resultversion.GetDownloadURL())
		fmt.Printf("Release notes: %s\n", resultversion.GetReleaseNotes())
	} else {
		return errors.New(resultversion.GetMessage())
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

		err := NextStepSetup(cfg, resultversion, plj, status)
		if err != nil {
			return fmt.Errorf("Error NextStepWebApp setup %w\n", err)
		}

	} else {
		fmt.Println("Installation cancelled")
	}

	return nil
}
