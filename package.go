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

func NextStepSetup(cfg config, resultversion *versionCheck, plj *packageLocalJson) error {
	var message string
	var err error
	var filename string

	// Nstep lock check
	lockfile, err := LockNstep(cfg)
	if err != nil {
		return fmt.Errorf("Error update.lock %w", err)
	}
	defer lockfile.Close()
	defer os.Remove(cfg.GetLockFilePath())

	// format filepath to store download
	downloadpath := cfg.GetDownloadPath()
	filename = fmt.Sprintf("nextstep_%s.tar.gz", resultversion.GetLatestVersion())
	downloadfilepath := fmt.Sprintf("%s/%s", downloadpath, filename)

	message, err = Downloadpackage(resultversion.GetDownloadURL(), downloadfilepath)

	if err != nil {
		return fmt.Errorf("Error downloading package %w", err)
	}
	println(message)

	// Verifying package integrity

	err = VerifyChecksum(downloadfilepath, resultversion.GetChecksum())

	if err != nil {
		return fmt.Errorf("Verification failed %w", err)
	} else {
		fmt.Println("Package verified successfully")
	}

	// Extract the downloaded package, function from package.go
	versionpath := cfg.GetVersionPath()
	filename = fmt.Sprintf("nextstep_%s", resultversion.LatestVersion) // also used in currentfilepath
	versionfilepath := fmt.Sprintf("%s/%s", versionpath, filename)

	message, err = Extractpackage(downloadfilepath, versionfilepath)
	if err != nil {
		return fmt.Errorf("Error extracting package %w: ", err)
	}
	println(message)

	// Symlink the new version to the current one
	err = EmtyDir(cfg.GetCurrentPath())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	currentfilepath := fmt.Sprintf("%s/%s", cfg.GetCurrentPath(), filename)

	err = os.Symlink(versionfilepath, currentfilepath)
	if err != nil {
		return fmt.Errorf("Error symlinking %w", err)
	}

	// Safty check to see if this is a install or update
	var setupStatus string
	optionsSetup := []string{"update", "install"}
	_, err = os.ReadDir(plj.GetLocalWebpath())

	if err != nil {
		if os.IsExist(err) {
			// This is a install
			setupStatus = optionsSetup[1]
		}
	} else {
		// This is a update
		setupStatus = optionsSetup[0]

	}

	switch setupStatus {
	case optionsSetup[0]: // update
		// backup the current nextstep installation

	case optionsSetup[1]: // install
	default:
		return fmt.Errorf("no option setup")
	}

	//err = updatemove(resultversion, plj, cfg)
	// get the current code and move to web portal
	// Update backs up the db (so db is seperate between update)
	// New db gets updated by scripts if needed
	return nil
}

func updatemove(resultversion *versionCheck, plj *packageLocalJson, cfg config) error {
	var err error

	// First Backup the current install
	backfilepath := fmt.Sprintf("%s/%s", cfg.GetBackupPath(), resultversion.GetCurrentVersion())
	err = os.Rename(plj.GetLocalWebpath(), backfilepath)
	if err != nil {
		return fmt.Errorf("could not move current Nextstep version to backup %w", err)
	}

	//currentversion := resultversion.GetCurrentVersion()

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

func loadlocalpackage(cfg config) (*packageLocalJson, error) {
	packagepath := cfg.GetPackagePath()

	jsonLocalFile, err := os.Open(packagepath)
	if err != nil {
		return nil, fmt.Errorf("cannot open package.json: %w", err)
	}
	defer jsonLocalFile.Close()

	packageLocalItem := &packageLocalJson{}
	decoder := json.NewDecoder(jsonLocalFile)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(packageLocalItem); err != nil {
		return nil, fmt.Errorf("cannot decode local package.json: %w", err)
	}
	return packageLocalItem, nil
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
