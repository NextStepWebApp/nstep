package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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
	Web     string `json:"webpath"`
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

func (vc *versionCheck) GetCurrentVersion() string {
	return vc.CurrentVersion
}

func (vc *versionCheck) GetLatestVersion() string {
	return vc.LatestVersion
}

func (vc *versionCheck) IsUpdateAvailable() bool {
	return vc.UpdateAvailable
}

func (vc *versionCheck) GetMessage() string {
	return vc.Message
}

func (vc *versionCheck) GetDownloadURL() string {
	return vc.DownloadURL
}

func (vc *versionCheck) GetReleaseNotes() string {
	return vc.ReleaseNotes
}

func (vc *versionCheck) GetChecksum() string {
	return vc.Checksum
}

// This function gets the local version and remote project version
// And then compares them to see if a new version came out
func Versionchecker(cfg config) (*versionCheck, error) {
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
