package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func Downloadpackage(url, filepath string) (message string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("cannot fetch remote package: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("remote server returned status: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("cannot create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot write to file: %w", err)
	}

	return "Download completed successfully", nil
}

func Extractpackage(targzpath, destdir string) (message string, err error) {
	if err := os.MkdirAll(destdir, 0755); err != nil {
		return "", fmt.Errorf("cannot create destination directory: %w", err)
	}

	cmd := exec.Command("tar", "-xzf", targzpath, "-C", destdir)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cannot extract tar.gz: %w", err)
	}
	return "Extraction completed successfully", nil
}
