package main

import (
	"fmt"
	"os"
)

// The sourceDir is only meant for the rollback functionality
func nextStepSetup(cfg config, resultversion *versionCheck, plj *packageLocalJson, state *state, status *status, sourceDir *string) error {
	var err error

	// Safty check to see if this is a install or update
	commandStatus, err := status.getStatus()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// These are environment checks

	setupStatus := false
	_, err = os.ReadDir(plj.getLocalWebpath())
	if err != nil {
		if os.IsNotExist(err) {
			setupStatus = false //install
		} else {
			return fmt.Errorf("cannot check installation status: %w", err)
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
		fmt.Println("==> Installation setup...")

	// Alreaddy installed
	case setupStatus == true && commandStatus == "install":
		return fmt.Errorf("nextstep is already installed, run 'sudo nstep update'")

	// Working update setup
	case setupStatus == true && commandStatus == "update":
		if sourceDir != nil {
			return fmt.Errorf("internal error: update command, but source directory is not equal to nil")
		}
		fmt.Println("==> Update setup...")

	// Invalid rollback use, needs to install first
	case setupStatus == false && commandStatus == "rollback":
		return fmt.Errorf("nextstep is not installed, run 'sudo nstep install'")

	// Working rollback setup
	case setupStatus == true && commandStatus == "rollback":
		// this is a saftey check for code usage
		if sourceDir == nil {
			return fmt.Errorf("internal error: rollback command, but source directory is nil")
		}

		fmt.Println("==> Rollback setup..")
	}
	// End of checks

	// Nstep lock check
	lockfile, err := lockNstep(cfg)
	if err != nil {
		return fmt.Errorf("Error update.lock %w", err)
	}
	defer lockfile.Close()
	defer os.Remove(cfg.getLockFilePath())

	// End of preperation
	//
	//
	// Start CORE
	switch commandStatus {
	case "install":
		currentfilepath, err := updateAllComponents(cfg, resultversion)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		// Create or recreate the nextstep structure
		err = nextStepCreate(*plj)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		// get the current version to the web portal
		err = copyDir(currentfilepath, plj.getLocalWebpath())
		if err != nil {
			return fmt.Errorf("cannot copy current to webpath %w", err)
		}

		// put the config files etc in the right place and remove unused files in the web portal
		err = setupMovesInstallUpdate("install")
		if err != nil {
			return fmt.Errorf("cannot do the setup moves %w", err)
		}
	case "update":
		currentfilepath, err := updateAllComponents(cfg, resultversion)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		err = nextStepBackup(cfg, resultversion, *plj)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		// get the current version to the web portal
		err = copyDir(currentfilepath, plj.getLocalWebpath())
		if err != nil {
			return fmt.Errorf("cannot copy current to webpath %w", err)
		}
		err = setupMovesInstallUpdate("update")
		if err != nil {
			return fmt.Errorf("cannot do the setup moves %w", err)
		}

	case "rollback":
		err = nextStepBackup(cfg, resultversion, *plj)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		// Name the currentfilepath for rollback
		// sourceDir is passed throug the function
		currentfilepath := *sourceDir
		err = setupMovesRollback(currentfilepath)
		if err != nil {
			return fmt.Errorf("cannot do the setup moves %w", err)
		}
	default:
		return fmt.Errorf("internal code error")

	}
	// END CORE

	// Finishing touches

	// Now give all the stuff the correct permssion and ownership
	err = nextstepPermissionManager(plj)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Now update the state
	err = saveState(plj, cfg, resultversion, state)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil

}
