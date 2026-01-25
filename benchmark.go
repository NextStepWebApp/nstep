package main

import (
	"fmt"
	"os"
	"time"
)

// Sequential version - no concurrency
func updateAllComponentsSequential(cfg config, resultversion *versionCheck) (currentwebpath string, err error) {
	startNow := time.Now()

	// Update package if available
	if resultversion.isUpdatePackageAvailable() {
		err := onlineToLocalPackageSequential(cfg, resultversion)
		if err != nil {
			return "", err
		}
	}

	// Update webapp if available
	if resultversion.isUpdateWebAppAvailable() {
		currentwebpath, err = onlineToLocalWebAppSequential(cfg, resultversion)
		if err != nil {
			return "", err
		}
	}

	fmt.Printf("Sequential operation took: %v\n", time.Since(startNow))
	return currentwebpath, nil
}

func onlineToLocalPackageSequential(cfg config, resultversion *versionCheck) error {
	var err error
	packageUrl := resultversion.getPackageURL()
	packagePath := cfg.getPackagePath()
	err = downloadpackage(packageUrl, packagePath)
	if err != nil {
		return fmt.Errorf("cannot download %s to %s: %w", packageUrl, packagePath, err)
	}
	fmt.Printf("%s downloaded successfully\n", getPackageName(cfg))
	return nil
}

func onlineToLocalWebAppSequential(cfg config, resultversion *versionCheck) (string, error) {
	var err error
	var filename string

	// format filepath to store download
	downloadpath := cfg.getDownloadPath()
	filename = fmt.Sprintf("nextstep_%s.tar.gz", resultversion.getLatestWebAppVersion())
	downloadfilepath := fmt.Sprintf("%s/%s", downloadpath, filename)
	err = downloadpackage(resultversion.getDownloadURL(), downloadfilepath)
	if err != nil {
		return "", fmt.Errorf("error downloading package: %w", err)
	}

	// Verifying package integrity
	err = verifyChecksum(downloadfilepath, resultversion.getChecksum())
	if err != nil {
		return "", fmt.Errorf("verification failed: %w", err)
	}
	fmt.Println("Package verified successfully")

	// Extract the downloaded package
	versionpath := cfg.getVersionPath()
	filename = fmt.Sprintf("nextstep_%s", resultversion.getLatestWebAppVersion())
	versionfilepath := fmt.Sprintf("%s/%s", versionpath, filename)
	err = extractpackage(downloadfilepath, versionfilepath, 1)
	if err != nil {
		return "", fmt.Errorf("error extracting package: %w", err)
	}

	// Symlink the new version to the current one
	err = emptyDir(cfg.getCurrentPath())
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	currentfilepath := fmt.Sprintf("%s/%s", cfg.getCurrentPath(), filename)
	err = os.Symlink(versionfilepath, currentfilepath)
	if err != nil {
		return "", fmt.Errorf("error symlinking: %w", err)
	}

	return currentfilepath, nil
}
