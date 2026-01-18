package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	Name          string   `json:"name"`
	Version       string   `json:"version"`
	Remote        string   `json:"remote_project"`
	InstallScript string   `json:"install_script"`
	Web           string   `json:"webpath"`
	RequiredDirs  []string `json:"required_dirs"`
}

// Methods for local package json calling

func (plj packageLocalJson) GetRequiredDirs() []string {
	return plj.NextStep.RequiredDirs
}

func (plj packageLocalJson) GetRemote() string {
	return plj.NextStep.Remote
}

func (plj packageLocalJson) GetVersion() string {
	return plj.NextStep.Version
}
func (plj packageLocalJson) Getname() string {
	return plj.NextStep.Name
}

func (plj packageLocalJson) GetLocalWebpath() string {
	return plj.NextStep.Web
}

func (plj packageLocalJson) GetNextStepInstallScript() string {
	return plj.NextStep.InstallScript
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

// Method calls for versionstuct

func (vc versionCheck) GetCurrentVersion() string {
	return vc.CurrentVersion
}

func (vc versionCheck) GetLatestVersion() string {
	return vc.LatestVersion
}

func (vc versionCheck) IsUpdateAvailable() bool {
	return vc.UpdateAvailable
}

func (vc versionCheck) GetMessage() string {
	return vc.Message
}

func (vc versionCheck) GetDownloadURL() string {
	return vc.DownloadURL
}

func (vc versionCheck) GetReleaseNotes() string {
	return vc.ReleaseNotes
}

func (vc versionCheck) GetChecksum() string {
	return vc.Checksum
}

func Localpackageupdater(plj *packageLocalJson, resultversion *versionCheck, cfg config) error {
	var err error
	plj.NextStep.Version = resultversion.GetLatestVersion()
	fmt.Printf("==> Updating local system %s -> %s", resultversion.GetCurrentVersion(), resultversion.GetLatestVersion())

	updatedLocalPackage, err := json.MarshalIndent(plj, "", "\t")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = os.WriteFile(cfg.GetPackagePath(), updatedLocalPackage, 0664)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// This function gets the local version and remote project version
// And then compares them to see if a new version came out
func Versionchecker(cfg config, plj *packageLocalJson) (*versionCheck, error) {
	// The local package json is loaded in main by loadlocal package function in this file
	// Get the url to see the version of the project
	remotePackageUrl := plj.GetRemote()
	response, err := http.Get(remotePackageUrl)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch remote version: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote server returned status: %d", response.StatusCode)
	}

	packageOnlineItem := packageOnlineJson{}
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&packageOnlineItem); err != nil {
		return nil, fmt.Errorf("cannot decode remote package.json: %w", err)
	}

	localVersion := plj.GetVersion()
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
