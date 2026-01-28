package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func validatePermissionManager(cfg config, plj *packageLocalJson, settings settingsConfig) error {
	var err error
	warnings := make([]string, 0, 10)

	// The required dirs
	for _, dir := range plj.getRequiredDirInfo() {
		err = validatePermissionNumbers(&dir.Permission, "dir", settings, &warnings)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	// The update move dirs
	for _, dir := range plj.getUpdateMoveActions() {
		err = validatePermissionNumbers(&dir.Permissions, "file", settings, &warnings)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	// The install move dirs
	for _, dir := range plj.getInstallMoveActions() {
		err = validatePermissionNumbers(&dir.Permissions, "file", settings, &warnings)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	if len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Println(warning)
		}
		data, err := json.MarshalIndent(plj, "", "  ")
		if err != nil {
			return fmt.Errorf("cannot marshal plj: %w", err)
		}
		err = os.WriteFile(cfg.getPackagePath(), data, os.FileMode(settings.getSettingPermissionFile()))
		if err != nil {
			return fmt.Errorf("cannot write plj file: %w", err)
		}
	}
	return nil
}

func validatePermissionNumbers(permission *int, defaultFileType string, setting settingsConfig, warnings *[]string) error {
	// File type check
	var defaultPerm int
	switch defaultFileType {
	case "file":
		defaultPerm = setting.getSettingPermissionFile()
	case "dir":
		defaultPerm = setting.getSettingPermissionDir()
	case "exec":
		defaultPerm = setting.getSettingsPermissionExec()
	case "sensitive":
		defaultPerm = setting.getSettingsPermissionSensitive()
	default:
		return fmt.Errorf("%s - internal code error, wrong use of function", red("ERROR"))
	}

	if *permission < 0 {
		*permission = -*permission
	}

	// Extract digits
	nums := make([]int, 0, 4)
	temp := *permission
	for temp > 0 {
		digit := temp % 10
		if digit < 0 || digit > 7 {
			message := fmt.Sprintf("%s - wrong permission number, defaulting to %d", yellow("Warning"), defaultPerm)
			*warnings = append(*warnings, message)
			*permission = defaultPerm
			return nil
		}
		nums = append(nums, digit)
		temp /= 10
	}

	// Validate length and format
	if len(nums) == 3 {
		nums = append(nums, 0)
		reverse(nums)
		*permission = arrayToInt(nums)
		return nil
	} else if len(nums) == 4 {
		// Check the special permission digit (should be 0 for basic permissions)
		if nums[len(nums)-1] != 0 {
			message := fmt.Sprintf("%s - special permission bits not supported, defaulting to %d", yellow("Warning"), defaultPerm)
			*warnings = append(*warnings, message)
			*permission = defaultPerm
		}
		return nil
	}

	// Invalid length
	message := fmt.Sprintf("%s - invalid permission format, defaulting to %d", yellow("Warning"), defaultPerm)
	*warnings = append(*warnings, message)
	*permission = defaultPerm
	return nil
}
