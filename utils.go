package main

import (
	"fmt"
	"io"
	"os"
)

func CopyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("cannot create destination directory: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("cannot read directory %s: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := fmt.Sprintf("%s/%s", src, entry.Name())
		dstPath := fmt.Sprintf("%s/%s", dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	fmt.Printf("Copy %s -> %s", srcFile.Name(), dstFile.Name())

	return err
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
