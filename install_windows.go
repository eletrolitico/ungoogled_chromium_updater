package main

import (
	"fmt"
)

const owner = "ungoogled-software"
const repo = "ungoogled-chromium-windows"

var gcAsset = &asset{
	Name: "googlechrome.exe",
	Url:  "https://dl.google.com/chrome/install/standalonesetup64.exe",
}

func downloadAssets() (string, string, error) {
	ucAsset, err := getLatestAsset(owner, repo, "windows_x64.zip")
	if err != nil {
		return "", "", fmt.Errorf("erro obtendo latest: %v", err)
	}

	ucPath, err := downloadFile(ucAsset)
	if err != nil {
		return "", "", err
	}

	gcPath, err := downloadFile(gcAsset)
	if err != nil {
		return "", "", err
	}

	return ucPath, gcPath, nil
}

func install(ucPath, gcPath string) error {

	return nil
}
