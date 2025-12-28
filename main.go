package main

import (
	"errors"
	"fmt"
	"os"
)

const (
	nstepconfigfile = "/home/william/Documents/programming/PWS/nstep/config.json"
)

func main() {
	// Load the config
	cfg, err := Loadconfig(nstepconfigfile)
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
	case "update":
		UpdateNextStep(cfg)
	case "rollback":
		fmt.Println("rollbacker")
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
	fmt.Println("	update       Update NextStep to latest version")
	fmt.Println("	rollback     Rollback to previous version")
	fmt.Println("	unlock       Clear stuck update lock")
}
