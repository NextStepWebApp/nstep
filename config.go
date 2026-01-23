package main

import (
	"encoding/json"
	"errors"
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
	NstepDownloadPath string `json:"nstepdownloadpath"`
	NstepVersionPath  string `json:"nstepversionpath"`
	NstepCurrentPath  string `json:"nstepcurrentpath"`
	NstepBackupPath   string `json:"nstepbackuppath"`
	Lockfile          string `json:"lockfile"`
	StateFile         string `json:"statefile"`
}

// Functions to get specific info from the json

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
func (c config) diravailable() error {
	paths := []string{
		c.getDownloadPath(),
		c.getVersionPath(),
		c.getBackupPath(),
		c.getCurrentPath(),
	}

	for _, path := range paths {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("cannot create %s: %w", path, err)
		}
	}
	return nil
}

// Create a struct print called in main
func loadconfig(nstepconfigfile string) (config, error) {

	configfile, err := os.Open(nstepconfigfile)
	if err != nil {
		return config{}, errors.New("Cannot load the config.json file")
	}
	defer configfile.Close()

	configItem := config{}
	decoder := json.NewDecoder(configfile)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&configItem); err != nil {
		return config{}, errors.New("Decode error")
	}

	return configItem, nil
}
