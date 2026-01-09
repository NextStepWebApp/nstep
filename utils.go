package main

import (
	"fmt"
	"os"
)

func EmtyDir(dirpath string) error {
	entries, err := os.ReadDir(dirpath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Error reading current directory %w", err)
	}

	for _, entry := range entries {
		path := fmt.Sprintf("%s/%s", dirpath, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("Error removing %s: %w", path, err)
		}
	}
	return nil
}

func SudoPowerChecker() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("this command must be run as root (use sudo)")
	}
	return nil
}

func PowerHandler(err error) {
	err = SudoPowerChecker()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
