package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// For online json
type packageOnlineJson struct {
	Version       string `json:"version"`
	Url           string `json:"download_url"`
	Release_notes string `json:"release_notes"`
}

// For local json
type packageLocalJson struct {
	NextStep nextStep `json:"nextstep"`
}
type nextStep struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Remote  string `json:"remote_project"`
}

func UpdateNextStep() {
	versionchecker()
}

func versionchecker() {
	// Get local version
	jsonLocalFile, err := os.Open("/home/william/Documents/programming/PWS/nstep/package.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonLocalFile.Close()

	packageLocalItem := packageLocalJson{}
	decoder := json.NewDecoder(jsonLocalFile)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&packageLocalItem); err != nil {
		log.Fatal("Decode error: ", err)
	}

	localVersion := packageLocalItem.NextStep.Version
	fmt.Println(localVersion)

	// Get the url to see the verion of the project
	remotePackageUrl := packageLocalItem.NextStep.Remote
	fmt.Println(remotePackageUrl)

	response, err := http.Get(remotePackageUrl)

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Fatal(err)
	}

	packageOnlineItem := packageOnlineJson{}
	decoder = json.NewDecoder(response.Body)

	if err := decoder.Decode(&packageOnlineItem); err != nil {
		log.Fatal("Decode error: ", err)
	}

	onlineVersion := packageOnlineItem.Version
	fmt.Println(onlineVersion)
}
