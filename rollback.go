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

func RollbackNextStep(cfg config) error {
	var err error

	entries, err := os.ReadDir(cfg.GetBackupPath())
	if err != nil {
		return fmt.Errorf("cannot read %s %w", cfg.GetBackupPath(), err)
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

	// Version to backup
	restorePath := fmt.Sprintf("%s/%s", cfg.GetBackupPath(), versions[num])

	_, err = Extractpackage(restorePath, os.TempDir())
	if err != nil {
		return fmt.Errorf("cannot extract %s %w\n", restorePath, err)
	}

	// Name of the extracted directory
	cleanName := regexVersion(versions[num])
	tempDir := fmt.Sprintf("%s/%s", os.TempDir(), cleanName)
	fmt.Println(tempDir)

	return nil
}

func regexVersion(version string) string {
	pattern := `v\d+\.\d+\.\d+`
	r := regexp.MustCompile(pattern)
	cleanName := r.FindString(version)
	return cleanName
}
