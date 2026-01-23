package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type state struct {
	InstalledVersion        string    `json:"installed_version"`
	InstalledPackageVersion int       `json:"installed_package_version"`
	LastUpdate              time.Time `json:"last_update"`
	InstallationDate        time.Time `json:"installation_date"`
}

// Methods for the state struct

func (s *state) getInstalledVersion() string {
	return s.InstalledVersion
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
				InstalledVersion:        "v0.0.0",
				InstalledPackageVersion: 0,
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

func saveState(cfg config, state *state) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal state: %w", err)
	}

	err = os.WriteFile(cfg.getStateFile(), data, 0644)
	if err != nil {
		return fmt.Errorf("cannot write state file: %w", err)
	}

	return nil
}
