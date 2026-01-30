package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type state struct {
	InstalledWebAppVersion  string    `json:"installed_version"`
	InstalledPackageVersion int       `json:"installed_package_version"`
	LastUpdate              time.Time `json:"last_update"`
	InstallationDate        time.Time `json:"installation_date"`
}

// Methods for the state struct

func (s *state) getInstalledWebAppVersion() string {
	return s.InstalledWebAppVersion
}

func (s *state) getInstalledPackageVersion() int {
	return s.InstalledPackageVersion
}

func (s *state) lastUpdate() time.Time {
	return s.LastUpdate
}

func (s *state) installationDate() time.Time {
	return s.InstallationDate
}

func loadState(cfg config) (*state, error) {

	stateFile, err := os.Open(cfg.getStateFile())
	if err != nil {
		if os.IsNotExist(err) {
			// First install - return default state
			return &state{
				InstalledWebAppVersion:  "v0.0.0",
				InstalledPackageVersion: 0,
				LastUpdate:              time.Now(),
				InstallationDate:        time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("cannot open %s %w", cfg.getStateFile(), err)
	}
	defer stateFile.Close()

	stateItem := &state{}
	decoder := json.NewDecoder(stateFile)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(stateItem); err != nil {
		return nil, fmt.Errorf("cannot decode state file: %w", err)
	}

	return stateItem, nil
}

func saveState(plj *packageLocalJson, cfg config, resultversion *versionCheck, state *state) error {
	var err error

	if resultversion.isUpdatePackageAvailable() {

		state.InstalledPackageVersion = resultversion.getLatestPackageVersion()
		fmt.Printf("%s Updating local %s system %d -> %d\n", green("===>"), getPackageName(cfg),
			resultversion.getCurrentPackageVersion(), resultversion.getLatestPackageVersion())
	}

	if resultversion.isUpdateWebAppAvailable() {

		state.InstalledWebAppVersion = resultversion.getLatestWebAppVersion()
		fmt.Printf("%s Updating local %s system %s -> %s\n", green("===>"), plj.getName(),
			resultversion.getCurrentWebAppVersion(), resultversion.getLatestWebAppVersion())
	}

	// update latest update state
	state.LastUpdate = time.Now()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("%s - cannot marshal state", red("ERROR"))
	}

	err = os.WriteFile(cfg.getStateFile(), data, 0644)
	if err != nil {
		return fmt.Errorf("%s - cannot write state file", red("ERROR"))
	}

	return nil
}
