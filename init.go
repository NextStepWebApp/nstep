package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// struct for the setup.json
type initPackage struct {
	ProjectName string `json:"project_name"`
	VersionUrl  string `json:"version_url"`
}

// This function gets the version url of the remote project
// At the beginning that does not exits so get it from that
// and after delete the install.json (no longer needed)
func initLocalPackage(cfg config) error {
	var err error

	// Check to see if the command whas already executed
	if _, err = os.Stat(cfg.getPackagePath()); err == nil {
		return fmt.Errorf("Already initialized")
	}

	setupFile, err := os.Open(cfg.getSetupFile())
	if err != nil {
		return fmt.Errorf("cannot open setup file %w", err)
	}
	defer setupFile.Close()

	// Get the information
	setupItem := initPackage{}
	decoder := json.NewDecoder(setupFile)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&setupItem); err != nil {
		return fmt.Errorf("cannot decode %w", err)
	}

	projectName := setupItem.ProjectName
	projectUrl := setupItem.VersionUrl

	fmt.Println("===> Package build setup...")
	fmt.Printf("Downloading %s package.json from %s\n", projectName, projectUrl)

	message, err := downloadpackage(projectUrl, cfg.getPackagePath())
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	fmt.Println(message)

	fmt.Println("===> Package build setup completed successfully")
	return nil
}

/*
// Function called to make sure the package.json is set up
func packageGuard(cfg config) error {
	var err error

	// Check to see if init was already executed
	if _, err = os.Stat(cfg.getPackagePath()); err != nil {
		return fmt.Errorf("not initialized. Run 'sudo nstep init'")
	}
	return nil
}
*/
