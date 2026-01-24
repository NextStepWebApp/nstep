package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func statusNextStep(plj *packageLocalJson, cfg config, state *state) error {
	var err error

	// Check to see if the web app is installed
	_, err = os.ReadDir(plj.getLocalWebpath())
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s is not installed. Run 'sudo nstep install'", plj.getname())
		} else {
			return fmt.Errorf("cannot check installation status: %w", err)
		}
	}

	fmt.Println(":: Status")
	fmt.Println(strings.Repeat("-", 20))

	fmt.Printf("%-18s %s\n", plj.getname()+" version:", state.getInstalledWebAppVersion())
	fmt.Printf("%-18s %d\n", getPackageName(cfg)+" version:", state.getInstalledPackageVersion())

	fmt.Println("\n:: Timeline")
	fmt.Println(strings.Repeat("-", 20))

	installDate := state.installationDate()
	fmt.Printf(
		"%-18s %s (%s)\n",
		"Installed on:",
		installDate.Format("Jan 02, 2006"),
		humanDuration(installDate),
	)

	fmt.Printf(
		"%-18s %s (%s)\n",
		"Last updated:",
		state.LastUpdate.Format("Jan 02, 2006"),
		humanDuration(state.LastUpdate),
	)

	return nil
}

func humanDuration(from time.Time) string {
	d := time.Since(from)

	if d < 0 {
		return "in the future"
	}

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%d months ago", int(d.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%d years ago", int(d.Hours()/(24*365)))
	}
}
