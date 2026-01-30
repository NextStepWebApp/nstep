package main

import (
	"fmt"
	"os"
)

// The sourceDir is only meant for the rollback functionality
func nextStepSetup(cfg config, resultversion *versionCheck, plj *packageLocalJson, settings settingsConfig, state *state, status *status, sourceDir *string) error {
	var err error

	// Nstep lock check
	lockfile, err := lockNstep(cfg)
	if err != nil {
		return fmt.Errorf("Error update.lock %w", err)
	}
	defer lockfile.Close()
	defer os.Remove(cfg.getLockFilePath())

	// Safty check to see if this is a install or update
	commandStatus, err := status.getStatus()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// These are environment checks

	setupStatus := false
	_, err = os.Stat(plj.getLocalWebpath())
	if err != nil {
		if os.IsNotExist(err) {
			setupStatus = false //install
		} else {
			return fmt.Errorf("%s - cannot check installation status", red("ERROR"))
		}
	} else {
		setupStatus = true //update
	}

	switch {

	// So a broken update setup
	case setupStatus == false && commandStatus == "update":
		return fmt.Errorf("command given is update system says install, run 'sudo nstep install'")

	// Working installation
	case setupStatus == false && commandStatus == "install":
		if sourceDir != nil {
			return fmt.Errorf("internal error: install command, but source directory is not equal to nil")
		}

	// Alreaddy installed
	case setupStatus == true && commandStatus == "install":
		return fmt.Errorf("nextstep is already installed, run 'sudo nstep update'")

	// Working update setup
	case setupStatus == true && commandStatus == "update":
		if sourceDir != nil {
			return fmt.Errorf("internal error: update command, but source directory is not equal to nil")
		}

	// Invalid rollback use, needs to install first
	case setupStatus == false && commandStatus == "rollback":
		return fmt.Errorf("nextstep is not installed, run 'sudo nstep install'")

	// Working rollback setup
	case setupStatus == true && commandStatus == "rollback":
		// this is a saftey check for code usage
		if sourceDir == nil {
			return fmt.Errorf("internal error: rollback command, but source directory is nil")
		}

	}
	// End of checks

	// End of preperation
	//
	//
	// Start CORE
	switch commandStatus {
	case "install":
		currentfilepath, err := updateAllComponents(cfg, settings, plj, resultversion)
		if err != nil {
			return err
		}

		// Create or recreate the nextstep structure
		err = nextStepCreate(plj, settings)
		if err != nil {
			return err
		}

		// get the current version to the web portal
		err = copyDir(currentfilepath, plj.getLocalWebpath(), settings)
		if err != nil {
			return err
		}

		// put the config files etc in the right place and remove unused files in the web portal
		err = setupMovesInstallUpdate("install", plj, settings)
		if err != nil {
			return err
		}
	case "update":
		currentfilepath, err := updateAllComponents(cfg, settings, plj, resultversion)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		err = nextStepBackup(cfg, resultversion, settings, plj)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		// get the current version to the web portal
		err = copyDir(currentfilepath, plj.getLocalWebpath(), settings)
		if err != nil {
			return fmt.Errorf("cannot copy current to webpath %w", err)
		}
		err = setupMovesInstallUpdate("update", plj, settings)
		if err != nil {
			return fmt.Errorf("cannot do the setup moves %w", err)
		}

	case "rollback":
		err = nextStepBackup(cfg, resultversion, settings, plj)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		// Name the currentfilepath for rollback
		// sourceDir is passed throug the function
		currentfilepath := *sourceDir
		err = setupMovesRollback(currentfilepath, settings)
		if err != nil {
			return fmt.Errorf("cannot do the setup moves %w", err)
		}
	default:
		return fmt.Errorf("internal code error")

	}
	// END CORE

	// Finishing touches

	/* 	// Now give all the stuff the correct permssion and ownership
	   	err = nextstepPermissionManager(plj.getr)
	   	if err != nil {
	   		return fmt.Errorf("%w", err)
	   	}
	*/
	// Now update the state
	err = saveState(plj, cfg, resultversion, state)
	if err != nil {
		return err
	}

	return nil

}
