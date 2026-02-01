package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func validatePermissionManager(plj *packageLocalJson, settings settingsConfig) error {
	var err error
	warnings := make([]string, 0, 10)

	// The required dirs
	for i := range plj.NextStep.RequiredDirs {
		err = validatePermissionNumbers(&plj.NextStep.RequiredDirs[i].Permission, "dir", plj.NextStep.RequiredDirs[i].Dir, settings, &warnings)
		if err != nil {
			return err
		}
	}

	// The update move dirs
	for i := range plj.NextStep.Operations.Update.Moves {
		err = validatePermissionNumbers(&plj.NextStep.Operations.Update.Moves[i].Permissions, "file", plj.NextStep.Operations.Update.Moves[i].To, settings, &warnings)
		if err != nil {
			return err
		}
	}

	// The install move dirs
	for i := range plj.NextStep.Operations.Install.Moves {
		err = validatePermissionNumbers(&plj.NextStep.Operations.Install.Moves[i].Permissions, "file", plj.NextStep.Operations.Install.Moves[i].To, settings, &warnings)
		if err != nil {
			return err
		}
	}

	if len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Println(warning)
		}
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Proceed with default permission numbers? [Y/n] ")
		response, err := reader.ReadString('\n')

		if err != nil {
			return fmt.Errorf("%s -  reading input", red("ERROR"))
		}

		response = strings.TrimSpace(response)

		if response == "N" || response == "n" || response == "" ||
			response == "No" || response == "no" {
			return fmt.Errorf("%s - program aborted by user due to invalid permissions", yellow("Warning"))

		} else if response != "Y" && response != "y" &&
			response != "Yes" && response != "yes" {
			return fmt.Errorf("%s - wrong user input", red("ERROR"))
		}

		/*
			data, err := json.MarshalIndent(plj, "", "  ")
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

func validatePermissionNumbers(permission *int, defaultFileType, filePath string, setting settingsConfig, warnings *[]string) error {
	// File type check
	var defaultPermission, decimalPermission int
	var err error

	switch defaultFileType {
	case "file":
		defaultFileType = "file"
		decimalPermission = setting.getSettingPermissionFile()
		defaultPermission, err = convertDecimalToOctalPermission(decimalPermission)
		if err != nil {
			return err
		}
	case "dir":
		defaultFileType = "directory"
		decimalPermission = setting.getSettingPermissionDir()
		defaultPermission, err = convertDecimalToOctalPermission(decimalPermission)
		if err != nil {
			return err
		}
	case "exec":
		defaultFileType = "executable"
		decimalPermission = setting.getSettingsPermissionExec()
		defaultPermission, err = convertDecimalToOctalPermission(decimalPermission)
		if err != nil {
			return err
		}
	case "sensitive":
		defaultFileType = "sensative file"
		decimalPermission = setting.getSettingsPermissionSensitive()
		defaultPermission, err = convertDecimalToOctalPermission(decimalPermission)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s - internal code error, wrong use of validation function", red("ERROR"))
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
			message := fmt.Sprintf("%s - wrong permission number %s, defaulting to %s for %s %s", yellow("Warning"), red(*permission), blue(decimalPermission), defaultFileType, filePath)
			*warnings = append(*warnings, message)
			*permission = defaultPermission
			return nil
		}
		nums = append(nums, digit)
		temp /= 10
	}

	// Validate length and format

	// This is a good permission
	if len(nums) == 3 {
		reverse(nums)
		*permission, err = convertDecimalToOctalPermission(arrayToInt(nums))
		if err != nil {
			return err
		}
		return nil
	} else {
		// Invalid length
		message := fmt.Sprintf("%s - invalid permission format %s, defaulting to %s for %s %s", yellow("Warning"), red(*permission), blue(decimalPermission), defaultFileType, filePath)
		*warnings = append(*warnings, message)
		*permission = defaultPermission
		return nil
	}

}

func convertDecimalToOctalPermission(permission int) (int, error) {
	permissionString := fmt.Sprintf("%d", permission)
	octalValue, err := strconv.ParseUint(permissionString, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("%s - cannot convert decimal number (%d) to octal (%d)", red("ERROR"), permission, int(octalValue))
	}
	return int(octalValue), nil
}

// Cannot decode json with a number starting with 0
/*else if len(nums) == 4 {
	// Check the special permission digit (should be 0 for basic permissions)
	if nums[len(nums)-1] != 0 {
		message := fmt.Sprintf("%s - special permission bits not supported %s, defaulting to %d for %s %s", yellow("Warning"), red(*permission), defaultPermission, defaultFileType, filePath)
		*warnings = append(*warnings, message)
		*permission = defaultPermission
	}

	*permission, err = convertDecimalToOctalPermission(*permission)
	if err != nil {
		return err
	}

	return nil
}
*/
