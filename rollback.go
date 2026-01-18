package main

import (
	"bufio"
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
		pattern := `v\d+\.\d+\.\d+`
		r := regexp.MustCompile(pattern)
		cleanName := r.FindString(version)
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
		return fmt.Errorf("invalid number: %w", err)
	}

	fmt.Println(versions[num])

	return nil
}
