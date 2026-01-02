package main

import (
	"errors"
	"fmt"
	"os"
)

const (
	nstepconfigfile = "/etc/nstep/config.json"
)

func main() {
	// Load the config json
	cfg, err := Loadconfig(nstepconfigfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Load the local package json
	plj, err := loadlocalpackage(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Check to see if the directories exist (config.go)
	err = cfg.Diravailable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating filepath", err)
		os.Exit(1)
	}

	// Get the command and error handling
	command, err := getCommand(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// nstep commands
	switch command {

	case "install":
		err := InstallNextStep(plj)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}

	case "update":
		err := UpdateNextStep(cfg, plj)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
	case "rollback":
		fmt.Println("rollbacker")
	case "unlock":
		err := UnlockNstep(cfg)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
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
	fmt.Println("	install      install the nextstep webapp")
	fmt.Println("	update       Update NextStep to latest version")
	fmt.Println("	rollback     Rollback to previous version")
	fmt.Println("	unlock       Clear stuck update lock")
}
