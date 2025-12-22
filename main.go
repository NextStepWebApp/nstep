package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {

	// Get the command and error handling
	command, err := getCommand(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// nstep commands
	switch command {
	case "update":
		UpdateNextStep()
	case "rollback":
		fmt.Println("rollbacker")
	case "status":
		fmt.Println("statuses")
	case "help", "--help", "-h":
		printUsage()
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
	fmt.Println("	status       Show installation status and health")
}
