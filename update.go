package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// For online json
type packageOnlineJson struct {
	Version      string `json:"version"`
	Url          string `json:"download_url"`
	Checksum     string `json:"checksum"`
	ReleaseNotes string `json:"release_notes"`
}

// For local json
type packageLocalJson struct {
	NextStep nextStep `json:"nextstep"`
}

type nextStep struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Remote  string `json:"remote_project"`
}

// To store information for version checker
type versionCheck struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	Message         string
	DownloadURL     string
	ReleaseNotes    string
	Checksum        string
}

func UpdateNextStep(cfg config) error {
	resultversion, err := versionchecker(cfg)
	if err != nil {
		return fmt.Errorf("Error checking version %w", err)
	}

	fmt.Println(resultversion.Message)
	if resultversion.UpdateAvailable {
		fmt.Printf("New version available: %s\n", resultversion.LatestVersion)
		fmt.Printf("Download: %s\n", resultversion.DownloadURL)
		fmt.Printf("Release notes: %s\n", resultversion.ReleaseNotes)
	}

	// confirmation part if there is a update
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Proceed with installation? [Y/n] ")

	response, err := reader.ReadString('\n')

	if err != nil {
		return fmt.Errorf("Error reading input %w", err)
	}

	response = strings.TrimSpace(response)

	if response == "Y" || response == "y" || response == "" ||
		response == "Yes" || response == "yes" {
		var message string
		var err error

		// Nstep lock check
		lockfile, err := LockNstep(cfg)
		if err != nil {
			return fmt.Errorf("Error update.lock %w", err)
		}
		defer lockfile.Close()
		defer os.Remove(cfg.GetLockFilePath())

		// format filepath to store download
		downloadpath := cfg.GetDownloadPath()
		filename := fmt.Sprintf("nextstep_%s.tar.gz", resultversion.LatestVersion)
		filepath := fmt.Sprintf("%s/%s", downloadpath, filename)

		message, err = Downloadpackage(resultversion.DownloadURL, filepath)

		if err != nil {
			return fmt.Errorf("Error downloading package %w", err)
		}
		println(message)

		// Verifying package integrity

		err = VerifyChecksum(filepath, resultversion.Checksum)

		if err != nil {
			return fmt.Errorf("Verification failed %w", err)
		} else {
			fmt.Println("Package verified successfully")
		}

		// Extract the downloaded package, function from package.go
		versionpath := cfg.GetVersionPath()
		message, err = Extractpackage(filepath, versionpath)
		if err != nil {
			return fmt.Errorf("Error extracting package %w: ", err)
		}
		println(message)

		// Symlink the new version to the current one

	} else {
		fmt.Println("Installation cancelled")
	}

	return nil
}

// This function gets the local version and remote project version
// And then compares them to see if a new version came out
func versionchecker(cfg config) (*versionCheck, error) {
	// Get local version

	packagepath := cfg.GetPackagePath()

	jsonLocalFile, err := os.Open(packagepath)
	if err != nil {
		return nil, fmt.Errorf("cannot open package.json: %w", err)
	}
	defer jsonLocalFile.Close()

	packageLocalItem := packageLocalJson{}
	decoder := json.NewDecoder(jsonLocalFile)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&packageLocalItem); err != nil {
		return nil, fmt.Errorf("cannot decode local package.json: %w", err)
	}

	// Get the url to see the version of the project
	remotePackageUrl := packageLocalItem.NextStep.Remote
	response, err := http.Get(remotePackageUrl)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch remote version: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote server returned status: %d", response.StatusCode)
	}

	packageOnlineItem := packageOnlineJson{}
	decoder = json.NewDecoder(response.Body)
	if err := decoder.Decode(&packageOnlineItem); err != nil {
		return nil, fmt.Errorf("cannot decode remote package.json: %w", err)
	}

	localVersion := packageLocalItem.NextStep.Version
	onlineVersion := packageOnlineItem.Version

	// Create result struct
	result := &versionCheck{
		CurrentVersion: localVersion,
		LatestVersion:  onlineVersion,
		DownloadURL:    packageOnlineItem.Url,
		ReleaseNotes:   packageOnlineItem.ReleaseNotes,
		Checksum:       packageOnlineItem.Checksum,
	}

	// Compare the versions to see if an update is needed
	if localVersion == onlineVersion {
		result.UpdateAvailable = false
		result.Message = fmt.Sprintf("Already up to date (%s)", localVersion)
	} else {
		result.UpdateAvailable = true
		result.Message = fmt.Sprintf("Update available: %s -> %s", localVersion, onlineVersion)
	}

	return result, nil
}
