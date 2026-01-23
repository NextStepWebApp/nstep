package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// For online json
type packageOnlineJson struct {
	WebAppVersion  string `json:"version"`
	PackageVersion int    `json:"package_version"`
	DownloadUrl    string `json:"download_url"`
	PackageUrl     string `json:"package_url"`
	Checksum       string `json:"checksum"`
	ReleaseNotes   string `json:"release_notes"`
}

// For local json
type packageLocalJson struct {
	NextStep nextStep `json:"nextstep"`
}

type nextStep struct {
	Name           string     `json:"name"`
	PackageVersion int        `json:"package_version"`
	Remote         string     `json:"remote_project"`
	InstallScript  string     `json:"install_script"`
	Web            string     `json:"webpath"`
	RequiredDirs   []string   `json:"required_dirs"`
	Operations     operations `json:"operations"`
}

type operations struct {
	Update  operationDetails `json:"update"`
	Install operationDetails `json:"install"`
}

type operationDetails struct {
	Moves   []moveAction `json:"moves"`
	Removes []string     `json:"removes"`
}

type moveAction struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Methods for local package json calling

func (plj packageLocalJson) getRequiredDirs() []string {
	return plj.NextStep.RequiredDirs
}

func (plj packageLocalJson) getRemote() string {
	return plj.NextStep.Remote
}

func (plj packageLocalJson) getname() string {
	return plj.NextStep.Name
}

func (plj packageLocalJson) getLocalWebpath() string {
	return plj.NextStep.Web
}

func (plj packageLocalJson) getNextStepInstallScript() string {
	return plj.NextStep.InstallScript
}

// To store information for version checker
type versionCheck struct {
	CurrentWebAppVersion   string
	LatestWebAppVersion    string
	CurrentPackageVersion  int
	LatestPackageVersion   int
	UpdateWebAppAvailable  bool
	UpdatePackageAvailable bool
	Message                []string
	DownloadURL            string
	PackageURL             string
	ReleaseNotes           string
	Checksum               string
}

// Method calls for versionstuct

func (vc versionCheck) getCurrentPackageVersion() int {
	return vc.CurrentPackageVersion
}

func (vc versionCheck) getLatestPackageVersion() int {
	return vc.LatestPackageVersion
}

func (vc versionCheck) isUpdatePackageAvailable() bool {
	return vc.UpdatePackageAvailable
}

func (vc versionCheck) getPackageURL() string {
	return vc.PackageURL
}

func (vc versionCheck) getCurrentWebAppVersion() string {
	return vc.CurrentWebAppVersion
}

func (vc versionCheck) getLatestWebAppVersion() string {
	return vc.LatestWebAppVersion
}

func (vc versionCheck) isUpdateWebAppAvailable() bool {
	return vc.UpdateWebAppAvailable
}

func (vc versionCheck) getMessageWebApp() string {
	return vc.Message[1]
}

func (vc versionCheck) getMessagePackage() string {
	return vc.Message[0]
}

func (vc versionCheck) getDownloadURL() string {
	return vc.DownloadURL
}

func (vc versionCheck) getReleaseNotes() string {
	return vc.ReleaseNotes
}

func (vc versionCheck) getChecksum() string {
	return vc.Checksum
}

// This function gets the local version and remote project version
// And then compares them to see if a new version came out
func versionchecker(plj *packageLocalJson, state *state, cfg config) (*versionCheck, error) {
	// The local package json is loaded in main by loadlocal package function in this file
	// Get the url to see the version of the project
	remotePackageUrl := plj.getRemote()
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
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&packageOnlineItem); err != nil {
		return nil, fmt.Errorf("cannot decode remote package.json: %w", err)
	}

	localWebAppVersion := state.getInstalledWebAppVersion()
	onlineWebAppVersion := packageOnlineItem.WebAppVersion
	localPackageVersion := state.getInstalledPackageVersion()
	onlinePackageVersion := packageOnlineItem.PackageVersion

	// Create result struct
	result := &versionCheck{
		CurrentWebAppVersion:  localWebAppVersion,
		LatestWebAppVersion:   onlineWebAppVersion,
		CurrentPackageVersion: localPackageVersion,
		LatestPackageVersion:  onlinePackageVersion,
		DownloadURL:           packageOnlineItem.DownloadUrl,
		PackageURL:            packageOnlineItem.PackageUrl,
		ReleaseNotes:          packageOnlineItem.ReleaseNotes,
		Checksum:              packageOnlineItem.Checksum,
	}

	// Compare the versions of the package.json to see if there is a update needed

	namePackage := getPackageName(cfg)

	if localPackageVersion == onlinePackageVersion {
		result.UpdatePackageAvailable = false

		result.Message[0] = fmt.Sprintf("%s is already up to date (%d)", namePackage, localPackageVersion)
	} else {
		result.UpdatePackageAvailable = true
		result.Message[0] = fmt.Sprintf("Update available for %s: %d -> %d", namePackage, localPackageVersion, onlinePackageVersion)
	}

	// Compare the versions to see if an update is needed for the webapp
	if localWebAppVersion == onlineWebAppVersion {
		result.UpdateWebAppAvailable = false

		result.Message[1] = fmt.Sprintf("%s is already up to date (%s)", plj.getname(), localWebAppVersion)
	} else {
		result.UpdateWebAppAvailable = true
		result.Message[1] = fmt.Sprintf("Update available: %s -> %s", localWebAppVersion, onlineWebAppVersion)
	}

	return result, nil
}
