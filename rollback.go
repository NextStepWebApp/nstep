package main

import (
	"fmt"
	"os"
)

func RollbackNextStep(cfg config) error {
	var err error

	entries, err := os.ReadDir(cfg.GetBackupPath())
	if err != nil {
		return fmt.Errorf("cannot read %s %w", cfg.GetBackupPath(), err)
	}

	for i, entry := range entries {
		fmt.Printf("%d\t%s", i, entry.Name())
	}

	return nil
}
