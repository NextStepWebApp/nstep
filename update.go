package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Online
type nextstepJsonApi struct {
	Tag_name    string `json:"tag_name"`
	Tarball_url string `json:"tarball_url"`
}

// Local
type nextstepJsonLocal struct {
	NextStep nextStep `json:"nextstep"`
}

type nextStep struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	API     api    `json:"api"`
}

type api struct {
	Releases string `json:"releases"`
	Repo     string `json:"repo"`
}

func UpdateNextStep() {
	versionchecker()
}

func versionchecker() {
	// Get version on github
	url := "https://api.github.com/repos/NextStepWebApp/NextStep/releases/latest"
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Fatal(err)
	}

	nextstepJsonApiItem := nextstepJsonApi{}
	decoder := json.NewDecoder(response.Body)

	if err := decoder.Decode(&nextstepJsonApiItem); err != nil {
		log.Fatal("Decode error: ", err)
	}

	onlineVersion := nextstepJsonApiItem.Tag_name
	fmt.Println(onlineVersion)

	// Get local version

	jsonLocalFile, err := os.Open("/home/william/Documents/programming/PWS/nstep/package.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonLocalFile.Close()

	nextstepJsonLocalItem := nextstepJsonLocal{}

	decoder = json.NewDecoder(jsonLocalFile)

	if err := decoder.Decode(&nextstepJsonLocalItem); err != nil {
		log.Fatal("Decode error: ", err)
	}

	localVersion := nextstepJsonLocalItem.NextStep.Version
	fmt.Println(localVersion)

}
