package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type settingsConfig struct {
	Permissions struct {
		Defaults struct {
			Files     int `toml:"files"`
			Dirs      int `toml:"directories"`
			Exec      int `toml:"executables"`
			Sensitive int `toml:"sensitive"`
		} `toml:"defaults"`
	} `toml:"permissions"`
}

func (s settingsConfig) getSettingPermissionFile() int {
	return s.Permissions.Defaults.Files
}
func (s settingsConfig) getSettingPermissionDir() int {
	return s.Permissions.Defaults.Dirs
}
func (s settingsConfig) getSettingsPermissionExec() int {
	return s.Permissions.Defaults.Exec
}
func (s settingsConfig) getSettingsPermissionSensitive() int {
	return s.Permissions.Defaults.Sensitive
}
func loadSettings(cfg config) (settingsConfig, error) {
	settingItem := settingsConfig{}

	tomlFile := cfg.getSettingsFile()
	if _, err := toml.DecodeFile(tomlFile, &settingItem); err != nil {
		return settingsConfig{}, fmt.Errorf("%s - cannot decode %s", red("ERROR"), tomlFile)
	}
	return settingItem, nil
}
