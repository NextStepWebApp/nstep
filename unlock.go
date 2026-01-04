package main

import (
	"fmt"
	"os"
)

// function that is called in main to unlock nstep
func UnlockNstep(cfg config) error {
	lockfilepath := cfg.GetLockFilePath()

	err := os.Remove(lockfilepath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No lock file found")
			return nil
		}
		return fmt.Errorf("failed to remove lock: %w", err)
	}

	fmt.Println("Lock removed successfully")
	return nil
}

// This is function that would be called from for example update.go
func LockNstep(cfg config) (*os.File, error) {
	lockfilepath := cfg.GetLockFilePath()

	// This line is a 2 in 1, if it does not exist the file will be created
	// If it does exist it gives a error
	lockfile, err := os.OpenFile(lockfilepath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)

	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("another process is already running\nTry again later or run 'nstep unlock' to remove a stale lock\n")
		}
		return nil, fmt.Errorf("failed to create lock file %w\n", err)
	}

	return lockfile, nil
}
