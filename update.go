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
}

func UpdateNextStep(cfg config) {
	resultversion, err := versionchecker(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error checking version: ", err)
		os.Exit(1)
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
		fmt.Fprintln(os.Stderr, "Error reading input: ", err)
		os.Exit(1)
	}

	response = strings.TrimSpace(response)

	if response == "Y" || response == "y" || response == "" {
		var message string
		var err error

		// Check to see if the directories exist
		err = cfg.Diravailable()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating filepath", err)
			os.Exit(1)
		}

		// format filepath to store download
		downloadpath := cfg.GetDownloadPath()
		filename := fmt.Sprintf("nextstep_%s.tar.gz", resultversion.LatestVersion)
		filepath := fmt.Sprintf("%s/%s", downloadpath, filename)

		message, err = Downloadpackage(resultversion.DownloadURL, filepath)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error downloading package: ", err)
			os.Exit(1)
		}
		println(message)

		// Extract the downloaded package, function from download.go
		versionpath := cfg.GetVersionPath()
		message, err = Extractpackage(filepath, versionpath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error extracting package: ", err)
			os.Exit(1)
		}
		println(message)

		// Symlink the new version to the current one

		os.Exit(0)

	} else {
		fmt.Println("Installation cancelled")
	}

	os.Exit(0)
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
