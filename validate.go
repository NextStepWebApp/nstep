package main

import (
	"fmt"
	"strconv"
)

func validatePermissionManager(plj *packageLocalJson, settings settingsConfig) error {
	var err error
	warnings := make([]string, 0, 10)

	// The required dirs
	for _, dir := range plj.getRequiredDirInfo() {
		err = validatePermissionNumbers(&dir.Permission, "dir", settings, &warnings)
		if err != nil {
			return err
		}
	}

	// The update move dirs
	for _, dir := range plj.getUpdateMoveActions() {
		err = validatePermissionNumbers(&dir.Permissions, "file", settings, &warnings)
		if err != nil {
			return err
		}
	}

	// The install move dirs
	for _, dir := range plj.getInstallMoveActions() {
		err = validatePermissionNumbers(&dir.Permissions, "file", settings, &warnings)
		if err != nil {
			return err
		}
	}

	if len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Println(warning)
		}
		/*data, err := json.MarshalIndent(plj, "", "  ")
		if err != nil {
			return fmt.Errorf("cannot marshal plj: %w", err)
		}
		err = os.WriteFile(cfg.getPackagePath(), data, os.FileMode(settings.getSettingPermissionFile()))
		if err != nil {
			return fmt.Errorf("cannot write plj file: %w", err)
		}
		*/
	}
	return nil
}

func validatePermissionNumbers(permission *int, defaultFileType string, setting settingsConfig, warnings *[]string) error {
	// File type check
	var defaultPermission int
	var err error

	switch defaultFileType {
	case "file":
		defaultPermission = setting.getSettingPermissionFile()
	case "dir":
		defaultPermission = setting.getSettingPermissionDir()
	case "exec":
		defaultPermission = setting.getSettingsPermissionExec()
	case "sensitive":
		defaultPermission = setting.getSettingsPermissionSensitive()
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
			message := fmt.Sprintf("%s - wrong permission number, defaulting to %d", yellow("Warning"), defaultPermission)
			*warnings = append(*warnings, message)
			*permission, err = convertDecimalToOctalPermission(defaultPermission)
			if err != nil {
				return err
			}
			return nil
		}
		nums = append(nums, digit)
		temp /= 10
	}

	// Validate length and format
	if len(nums) == 3 {
		nums = append(nums, 0)
		reverse(nums)
		*permission, err = convertDecimalToOctalPermission(arrayToInt(nums))
		return nil
	} else if len(nums) == 4 {
		// Check the special permission digit (should be 0 for basic permissions)
		if nums[len(nums)-1] != 0 {
			message := fmt.Sprintf("%s - special permission bits not supported, defaulting to %d", yellow("Warning"), defaultPermission)
			*warnings = append(*warnings, message)
			*permission, err = convertDecimalToOctalPermission(defaultPermission)
		}
		return nil
	}

	// Invalid length
	message := fmt.Sprintf("%s - invalid permission format, defaulting to %d", yellow("Warning"), defaultPermission)
	*warnings = append(*warnings, message)
	*permission, err = convertDecimalToOctalPermission(defaultPermission)
	return nil
}

func convertDecimalToOctalPermission(permission int) (int, error) {
	permissionString := fmt.Sprintf("%d", permission)
	octalValue, err := strconv.ParseUint(permissionString, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("%s - cannot convert decimal number (%d) to octal (%d)", red("ERROR"), permission, int(octalValue))
	}
	return int(octalValue), nil
}
