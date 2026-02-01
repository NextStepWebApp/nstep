package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type settingsConfig struct {
	Output struct {
		Verbose bool `toml:"verbose"`
	} `toml:"output"`

	Permissions struct {
		Defaults struct {
			Files     int `toml:"files"`
			Dirs      int `toml:"directories"`
			Exec      int `toml:"executables"`
			Sensitive int `toml:"sensitive"`
		} `toml:"defaults"`
	} `toml:"permissions"`

	Shell struct {
		Shell string `toml:"shell"`
	} `toml:"shell"`
}

func (s settingsConfig) validateSettingShell() error {
	validShells := []string{"bash", "zsh", "fish"}
	shell := s.getSettingShell()

	isValid := false
	for _, validShell := range validShells {
		if shell == validShell {
			isValid = true
		}
	}

	if !isValid {
		return fmt.Errorf("%s - invalid shell %s in settings", red("ERROR"), yellow(shell))
	}

	return nil
}

func (s settingsConfig) getSettingShell() string {
	return s.Shell.Shell
}

func (s settingsConfig) getOutputStatus() bool {
	return s.Output.Verbose
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
