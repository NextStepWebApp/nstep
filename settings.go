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
func (s settingsConfig) getSettingsPermissionsAll() []int {
	return []int{
		s.Permissions.Defaults.Files,
		s.Permissions.Defaults.Dirs,
		s.Permissions.Defaults.Exec,
		s.Permissions.Defaults.Sensitive,
	}
}

func loadSettings(cfg config) error {
	settingItem := settingsConfig{}

	tomlFile := cfg.getSettingsFile()
	if _, err := toml.DecodeFile(tomlFile, settingItem); err != nil {
		fmt.Errorf("%s - cannot decode %s", red("ERROR"), tomlFile)
	}

	for _, permission := range settingItem.getSettingsPermissionsAll() {
		if !isValidOctalPermission(permission) {
			fmt.Errorf("%s - wrong permission number in %s", red("ERROR"), tomlFile)
		}
	}

	return nil
}

func isValidOctalPermission(perm int) bool {
	if perm < 0 || perm > 0777 {
		return false
	}

	var temp, digit int
	temp = perm
	for temp > 0 {
		digit = temp % 10
		if digit > 7 {
			return false
		}
		temp = temp / 10
	}

	return true
}
