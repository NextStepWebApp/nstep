package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func rollbackNextStep(cfg config, plj *packageLocalJson, status *status) error {
	var err error

	entries, err := os.ReadDir(cfg.getBackupPath())
	if err != nil {
		return fmt.Errorf("cannot read %s %w", cfg.getBackupPath(), err)
	}
	if len(entries) < 1 {
		return fmt.Errorf("no backups to rollback to")
	}

	versions := make([]string, 0, 5)

	for _, entry := range entries {
		versions = append(versions, entry.Name())
	}

	// ui printing starts here

	fmt.Printf(":: Current version: %s\n\n", plj.getVersion())

	for i, version := range versions {
		cleanName := regexVersion(version)
		fmt.Printf("%d  nextstep/%s\n", i, cleanName)
	}

	fmt.Println("\n:: Select version to rollback:")
	fmt.Print(":: ")

	reader := bufio.NewReader(os.Stdin)

	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("Error reading input %w", err)
	}
	response = strings.TrimSpace(response)

	num, err := strconv.Atoi(response)
	if err != nil {
		return errors.New("Invalid number")
	}
	versionlen := len(versions)

	if num >= versionlen {
		return errors.New("Not a valid option")
	}

	// Version to restore
	restorePath := fmt.Sprintf("%s/%s", cfg.getBackupPath(), versions[num])

	_, err = extractpackage(restorePath, os.TempDir(), 0)
	if err != nil {
		return fmt.Errorf("cannot extract %s %w\n", restorePath, err)
	}

	// Name of the extracted directory
	cleanChosenName := regexVersion(versions[num])
	tempDir := fmt.Sprintf("%s/%s", os.TempDir(), cleanChosenName)
	fmt.Println(tempDir)

	// rollback does not need the all the information in versioncheck so make a custom one
	// The nextstepsetup function requires the resultversion.getCurrentVersion()
	// So the currentversion

	resultversion := &versionCheck{
		CurrentVersion: plj.getVersion(),
		LatestVersion:  cleanChosenName,
	}

	err = nextStepSetup(cfg, resultversion, plj, status, &tempDir)
	if err != nil {
		return fmt.Errorf("Error NextStepWebApp setup %w\n", err)
	}

	return nil
}

func regexVersion(version string) string {
	pattern := `v\d+\.\d+\.\d+`
	r := regexp.MustCompile(pattern)
	cleanName := r.FindString(version)
	return cleanName
}
