package main

import (
	"fmt"
	"os"
)

// This is function that would be called from for example update.go
func LockNstep(cfg config) error {
	lockfilepath := cfg.GetLockFilePath()
	lockfile, err := os.OpenFile(lockfilepath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)

	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("another update is already running\nRun 'nstep unlock' if the previous update crashed")
		} else {
			return fmt.Errorf("Error: Failed to create lock file: %w", err)
		}
	}

	lockfile.Close()
	defer os.Remove(lockfilepath)
	return nil
}
