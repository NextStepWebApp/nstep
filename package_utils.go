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
	"path/filepath"
)

// These are the package utils

// Load the local package.json, called/used in main.go
func loadlocalpackage(cfg config) (*packageLocalJson, error) {
	packagepath := cfg.getPackagePath()

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

func downloadpackage(url, filepath string) (message string, err error) {
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

func extractpackage(targzpath, destdir string, striplevel int) (message string, err error) {
	if err := os.MkdirAll(destdir, 0755); err != nil {
		return "", fmt.Errorf("cannot create destination directory %w\n", err)
	}

	//cmd := exec.Command("tar", "-xzf", targzpath, "-C", destdir)
	strip := fmt.Sprintf("--strip-components=%d", striplevel)

	cmd := exec.Command("tar", "-xzf", targzpath, "-C", destdir, strip)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cannot extract tar.gz %w\n", err)
	}
	return "Extraction completed successfully", nil
}

// VerifyChecksum calculates the SHA256 checksum of a file
// and compares it with the expected checksum
func verifyChecksum(filepathdownload, expectedChecksum string) error {
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

func getPackageName(cfg config) string {
	namePackage := filepath.Base(cfg.getPackagePath())
	return namePackage
}
