package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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

/*func Extractpackage(filepath string) (message string, err error) {
out := tar.NewWriter(filepath)
}*/
