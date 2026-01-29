package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type config struct {
	Packages packageconfig `json:"packages"`
	Nstep    nstepjson     `json:"nstep"`
}

type packageconfig struct {
	FilePath  string `json:"file_path"`
	SetupFile string `json:"setup_file"`
}

type nstepjson struct {
	NstepSettingsFile string `json:"settingsfile"`
	NstepDownloadPath string `json:"nstepdownloadpath"`
	NstepVersionPath  string `json:"nstepversionpath"`
	NstepCurrentPath  string `json:"nstepcurrentpath"`
	NstepBackupPath   string `json:"nstepbackuppath"`
	Lockfile          string `json:"lockfile"`
	StateFile         string `json:"statefile"`
}

// Functions to get specific info from the json

func (c config) getSettingsFile() string {
	return c.Nstep.NstepSettingsFile
}

func (c config) getSetupFile() string {
	return c.Packages.SetupFile
}

func (c config) getStateFile() string {
	return c.Nstep.StateFile
}

func (c config) getBackupPath() string {
	return c.Nstep.NstepBackupPath
}

func (c config) getPackagePath() string {
	return c.Packages.FilePath
}

func (c config) getLockFilePath() string {
	return c.Nstep.Lockfile
}

func (c config) getDownloadPath() string {
	return c.Nstep.NstepDownloadPath
}

func (c config) getVersionPath() string {
	return c.Nstep.NstepVersionPath
}

func (c config) getCurrentPath() string {
	return c.Nstep.NstepCurrentPath
}

// function to see if the dirs are available,
// if not create them
func (c config) diravailable(settings settingsConfig) error {
	paths := []string{
		c.getDownloadPath(),
		c.getVersionPath(),
		c.getBackupPath(),
		c.getCurrentPath(),
	}
	fmt.Printf("%s Required directory setup...\n", green("===>"))
	for _, path := range paths {
		err := os.MkdirAll(path, os.FileMode(settings.getSettingPermissionDir()))
		if err != nil {
			return fmt.Errorf("%s - cannot create %s", red("ERROR"), path)
		} else {
			if settings.getOutputStatus() {
				message := fmt.Sprintf("%s created %s", yellow(" ->"), path)
				verbosePrint(message, settings)
			}
		}
	}
	fmt.Printf("%s Required directory setup completed successfully\n", green("===>"))
	return nil
}

func loadconfig(nstepconfigfile string) (config, error) {

	configfile, err := os.Open(nstepconfigfile)
	if err != nil {
		return config{}, fmt.Errorf("%s - cannot load the config.json file", red("ERROR"))
	}
	defer configfile.Close()

	configItem := config{}
	decoder := json.NewDecoder(configfile)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&configItem); err != nil {
		return config{}, fmt.Errorf("%s - cannot decode the config.json file", red("ERROR"))
	}

	return configItem, nil
}
