package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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

// These are the package utils

// Load the local package.json, called/used in main.go
func Loadlocalpackage(cfg config) (*packageLocalJson, error) {
	packagepath := cfg.GetPackagePath()

	jsonLocalFile, err := os.Open(packagepath)
	if err != nil {
		return nil, fmt.Errorf("cannot open package.json %w", err)
	}
	defer jsonLocalFile.Close()

	packageLocalItem := &packageLocalJson{}
	decoder := json.NewDecoder(jsonLocalFile)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(packageLocalItem); err != nil {
		return nil, fmt.Errorf("cannot decode local package.json: %w", err)
	}
	return packageLocalItem, nil
}

func Downloadpackage(url, filepath string) (message string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("cannot fetch remote package %w\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("remote server returned status %d\n", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("cannot create file %w\n", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot write to file %w\n", err)
	}

	return "Download completed successfully", nil
}

func Extractpackage(targzpath, destdir string) (message string, err error) {
	if err := os.MkdirAll(destdir, 0755); err != nil {
		return "", fmt.Errorf("cannot create destination directory %w\n", err)
	}

	//cmd := exec.Command("tar", "-xzf", targzpath, "-C", destdir)
	cmd := exec.Command("tar", "-xzf", targzpath, "-C", destdir, "--strip-components=1")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cannot extract tar.gz %w\n", err)
	}
	return "Extraction completed successfully", nil
}

// VerifyChecksum calculates the SHA256 checksum of a file
// and compares it with the expected checksum
func VerifyChecksum(filepathdownload, expectedChecksum string) error {
	//expectedChecksum is from the online json
	file, err := os.Open(filepathdownload)
	if err != nil {
		return fmt.Errorf("cannot open file for verification %w\n", err)
	}
	defer file.Close()

	hasher := sha256.New()

	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("cannot read file for hashing %w\n", err)
	}

	calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))

	if calculatedChecksum != expectedChecksum {
		return fmt.Errorf(
			"checksum mismatch:\n  expected: %s\n  got: %s\n",
			expectedChecksum,
			calculatedChecksum,
		)
	}

	return nil
}
