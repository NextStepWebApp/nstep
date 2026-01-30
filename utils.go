package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
)

func verbosePrint(message string, settings settingsConfig) {
	if settings.getOutputStatus() {
		fmt.Println(message)
	}
}

// This function gets the uid, gid from the group you give it
// Usefull to use with chown
func getUidGid(group string) (uid int, gid int, err error) {
	groupuser, err := user.Lookup(group)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot find group %w\n", err)
	}

	uid, err = strconv.Atoi(groupuser.Uid)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get uid %w\n", err)
	}
	gid, err = strconv.Atoi(groupuser.Gid)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get gid %w\n", err)
	}

	return uid, gid, nil
}

func copyDir(src, dst string, settings settingsConfig) error {
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
			if err := copyDir(srcPath, dstPath, settings); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath, settings); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string, settings settingsConfig) error {
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
	message := fmt.Sprintf("%s %s -> %s", cyan("  - copy"), srcFile.Name(), dstFile.Name())
	verbosePrint(message, settings)

	return err
}

func emptyDir(dirpath string) error {
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

func sudoPowerChecker() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("%s - this command must be run as root %s", yellow("Warning"), blue("(use sudo)"))
	}
	return nil
}

func powerHandler(err error) {
	err = sudoPowerChecker()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// a function that moves a file from old path to new path
func moveFile(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}
	return nil
}

// function that removes a directory
func removeDir(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", path, err)
	}
	return nil
}

// These functions, reverse and arrayToInt are used in validate.go
func reverse[T any](arr []T) {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
}

func arrayToInt(arr []int) int {
	var result int
	for _, digit := range arr {
		result = result*10 + digit
	}
	return result
}
