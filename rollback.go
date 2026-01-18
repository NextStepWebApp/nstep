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

	versions := make([]string, 0, 5)

	for _, entry := range entries {
		versions = append(versions, entry.Name())
	}

	for i, version := range versions {
		cleanName := regexVersion(version)
		fmt.Printf("%d  nextstep/%s\n", i, cleanName)
	}

	fmt.Println(":: Select version to rollback:")
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
	cleanName := regexVersion(versions[num])
	tempDir := fmt.Sprintf("%s/%s", os.TempDir(), cleanName)
	fmt.Println(tempDir)

	// rollback does not need the versioncheck like install and update, so that's why
	// I justed pased a nil, the part where the versioncheck is used is also skipped in
	// the function for rollback, so it is not needed
	err = nextStepSetup(cfg, nil, plj, status, &tempDir)

	return nil
}

func regexVersion(version string) string {
	pattern := `v\d+\.\d+\.\d+`
	r := regexp.MustCompile(pattern)
	cleanName := r.FindString(version)
	return cleanName
}
