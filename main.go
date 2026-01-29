package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
)

const (
	nstepconfigfile = "/etc/nstep/config.json"
)

var (
	yellow = color.New(color.FgYellow, color.Bold).SprintFunc()
	red    = color.New(color.FgRed, color.Bold).SprintFunc()
	green  = color.New(color.FgGreen, color.Bold).SprintFunc()
	blue   = color.New(color.FgBlue, color.Bold).SprintFunc()
)

func main() {
	var err error

	status := &status{install: false, update: false}

	// Load the config json
	cfg, err := loadconfig(nstepconfigfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Load the settings toml
	settings, err := loadSettings(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Load the state package json
	state, err := loadState(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Get the command and error handling
	command, err := getCommand(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if command == "init" {
		err = sudoPowerChecker()
		powerHandler(err)

		// Create directories if not exist
		err = cfg.diravailable(settings)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		// Setup the package
		err = initLocalPackage(cfg, settings)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// If not exit it will move to the switch block
		os.Exit(0)
	}

	// Load the local package json
	plj, err := loadlocalpackage(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// nstep commands
	switch command {

	case "install":
		// Check if running as root
		err = sudoPowerChecker()
		powerHandler(err)

		status.isInstall()

		err = installNextStep(plj, cfg, settings, state, status)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	case "update":
		// Check if running as root
		err = sudoPowerChecker()
		powerHandler(err)

		status.isUpdate()

		err = updateNextStep(cfg, plj, settings, state, status)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "rollback":
		// Check if running as root
		err = sudoPowerChecker()
		powerHandler(err)

		status.isRollback()

		err = rollbackNextStep(cfg, plj, settings, state, status)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	case "status":
		err = statusNextStep(plj, cfg, state)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "remove":
		// check if running as root
		err = sudoPowerChecker()
		powerHandler(err)

		fmt.Println("remover")

	case "unlock":

		// Check if running as root
		err = sudoPowerChecker()
		powerHandler(err)

		err = unlockNstep(cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	case "help", "--help", "-h":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s (use -h for help)\n", command)
		os.Exit(1)
	}
}

// function called in main to get the command line arguments
func getCommand(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("missing operation (use -h for help)")
	}
	if len(args) > 1 {
		return "", errors.New("only one operation is allowed (use -h for help)")
	}
	return args[0], nil
}

func printUsage() {
	fmt.Println("nstep - NextStep Package Manager")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("	nstep <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("	install      Install the nextstep webapp")
	fmt.Println("	update       Update nextstep to latest version")
	fmt.Println("	rollback     Rollback to previous version")
	fmt.Println("	status       See the current version")
	fmt.Println("	unlock       Clear stuck nstep lock")
	fmt.Println("	remove       Remove the nextstep webapp")
}

// Store the status of the command
type status struct {
	install  bool
	update   bool
	rollback bool
}

func (s *status) isUpdate() {
	s.update = true
}

func (s *status) isInstall() {
	s.install = true
}

func (s *status) isRollback() {
	s.rollback = true
}

func (s *status) getStatus() (string, error) {
	// Count how many statuses are true
	trueCount := 0
	if s.install {
		trueCount++
	}
	if s.update {
		trueCount++
	}
	if s.rollback {
		trueCount++
	}

	// Error if multiple statuses are set
	if trueCount > 1 {
		return "", fmt.Errorf("multiple statuses are set: install=%v, update=%v, rollback=%v",
			s.install, s.update, s.rollback)
	}

	// Error if no status is set
	if trueCount == 0 {
		return "", fmt.Errorf("no status is set")
	}

	// Return the single true status
	if s.install {
		return "install", nil
	}
	if s.update {
		return "update", nil
	}
	if s.rollback {
		return "rollback", nil
	}

	return "", fmt.Errorf("unexpected error")
}
