package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

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
