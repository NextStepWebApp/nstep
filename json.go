package main

import (
	"encoding/json"
	"errors"
	"os"
)

type config struct {
	File_path string `json:"file_path"`
}

func Getpackagedir(configpath string) (filepath string, err error) {

	packagejson, err := os.Open(configpath)
	if err != nil {
		return "", errors.New("Cannot open the package.json file")
	}
	defer packagejson.Close()

	configItem := config{}
	decoder := json.NewDecoder(packagejson)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&configItem); err != nil {
		return "", errors.New("Decode error")
	}

	// Get the file path of the package.json
	packagefile := configItem.File_path

	return packagefile, nil
}
