package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func RollbackNextStep(cfg config) error {
	var err error

	entries, err := os.ReadDir(cfg.GetBackupPath())
	if err != nil {
		return fmt.Errorf("cannot read %s %w", cfg.GetBackupPath(), err)
	}

	for i, entry := range entries {
		pattern := `v\d+\.\d+\.\d+`
		r, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid regex %s %w", entry.Name(), err)
		}

		cleanName := r.FindString(entry.Name())

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

	fmt.Println(response)

	return nil
}
