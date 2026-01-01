package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func UpdateNextStep(cfg config) error {
	// function from package.go uses methods to get information
	resultversion, err := Versionchecker(cfg)
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
		var message string
		var err error
		var filename string

		// Nstep lock check
		lockfile, err := LockNstep(cfg)
		if err != nil {
			return fmt.Errorf("Error update.lock %w", err)
		}
		defer lockfile.Close()
		defer os.Remove(cfg.GetLockFilePath())

		// format filepath to store download
		downloadpath := cfg.GetDownloadPath()
		filename = fmt.Sprintf("nextstep_%s.tar.gz", resultversion.GetLatestVersion())
		downloadfilepath := fmt.Sprintf("%s/%s", downloadpath, filename)

		message, err = Downloadpackage(resultversion.GetDownloadURL(), downloadfilepath)

		if err != nil {
			return fmt.Errorf("Error downloading package %w", err)
		}
		println(message)

		// Verifying package integrity

		err = VerifyChecksum(downloadfilepath, resultversion.GetChecksum())

		if err != nil {
			return fmt.Errorf("Verification failed %w", err)
		} else {
			fmt.Println("Package verified successfully")
		}

		// Extract the downloaded package, function from package.go
		versionpath := cfg.GetVersionPath()
		filename = fmt.Sprintf("nextstep_%s", resultversion.LatestVersion) // also used in currentfilepath
		versionfilepath := fmt.Sprintf("%s/%s", versionpath, filename)

		message, err = Extractpackage(downloadfilepath, versionfilepath)
		if err != nil {
			return fmt.Errorf("Error extracting package %w: ", err)
		}
		println(message)

		// Symlink the new version to the current one
		err = EmtyDir(cfg.GetCurrentPath())
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		currentfilepath := fmt.Sprintf("%s/%s", cfg.GetCurrentPath(), filename)

		err = os.Symlink(versionfilepath, currentfilepath)
		if err != nil {
			return fmt.Errorf("Error symlinking %w", err)
		}

	} else {
		fmt.Println("Installation cancelled")
	}

	return nil
}
