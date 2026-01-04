package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func UpdateNextStep(cfg config, plj *packageLocalJson) error {
	// function from package.go uses methods to get information
	resultversion, err := Versionchecker(cfg, plj)
	if err != nil {
		return fmt.Errorf("Error checking version %w", err)
	}

	fmt.Println(resultversion.GetMessage())
	if resultversion.IsUpdateAvailable() {
		fmt.Printf("New version available: %s\n", resultversion.GetLatestVersion())
		fmt.Printf("Download: %s\n", resultversion.GetDownloadURL())
		fmt.Printf("Release notes: %s\n", resultversion.GetReleaseNotes())
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

		err := NextStepSetup(cfg, resultversion)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

	} else {
		fmt.Println("Installation cancelled")
	}

	return nil
}

func updatemove(resultversion *versionCheck, plj *packageLocalJson, cfg config) error {
	var err error

	// First Backup the current install
	backfilepath := fmt.Sprintf("%s/%s", cfg.GetBackupPath(), resultversion.GetCurrentVersion())
	err = os.Rename(plj.GetLocalWebpath(), backfilepath)
	if err != nil {
		return fmt.Errorf("could not move current Nextstep version to backup %w", err)
	}

	//currentversion := resultversion.GetCurrentVersion()

	return nil
}
