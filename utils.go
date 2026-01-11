package main

import (
	"fmt"
	"os"
)

func CopyDir(reader, writer string) error {
	var err error

	dircontents, err := os.ReadDir(reader)
	if err != nil {
		return fmt.Errorf("cannot read the contents of %s %w\n", reader, err)
	}

	// Loop to copy all the contents over

	for _, dircontent := range dircontents {
		fmt.Println(dircontent)
	}

	return nil
}

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

// a function that moves a file from old path to new path
func MoveFile(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("failed to move %s to %s: %w", oldPath, newPath, err)
	}
	return nil
}

// function that removes a directory
func RemoveDir(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", path, err)
	}
	return nil
}
