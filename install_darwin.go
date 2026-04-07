package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/itchio/damage"
	"github.com/itchio/damage/hdiutil"
	"github.com/itchio/headway/state"
)

const owner = "ungoogled-software"
const repo = "ungoogled-chromium-macos"

const appPath = "/Applications/Chromium.app"
const wideVineSrcPath = "gc/Google Chrome.app/Contents/Frameworks/Google Chrome Framework.framework/Libraries/WidevineCdm"
const wideVineDestPath = "/Applications/Chromium.app/Contents/Frameworks/Chromium Framework.framework/Libraries/WidevineCdm"

var consumer *state.Consumer = &state.Consumer{
	OnMessage: func(level, msg string) { log.Printf("[%s] %s\n", level, msg) },
}

var gcAsset = &asset{
	Name: "googlechrome.dmg",
	Url:  "https://dl.google.com/chrome/mac/universal/stable/GGRO/googlechrome.dmg",
}

func downloadAssets() (string, string, error) {
	search := runtime.GOARCH
	if search == "amd64" {
		search = "x86_64"
	}

	ucAsset, err := getLatestAsset(owner, repo, search)
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
	host := hdiutil.NewHost(consumer)

	_, err := damage.Mount(host, ucPath, "uc")
	if err != nil {
		return err
	}
	defer damage.Unmount(host, "uc")

	_, err = damage.Mount(host, gcPath, "gc")
	if err != nil {
		return err
	}
	defer damage.Unmount(host, "gc")

	log.Println("removing old app...")
	if err = os.RemoveAll(appPath); err != nil {
		return err
	}

	log.Println("copying new app...")
	if err = cp("uc/Chromium.app", appPath); err != nil {
		return err
	}

	log.Println("copying widevine...")
	if err = cp(wideVineSrcPath, wideVineDestPath); err != nil {
		return err
	}

	log.Println("fixing signing...")
	if err = fixSigning(); err != nil {
		return err
	}

	return nil
}

func cp(from, to string) error {
	return command("cp", "-R", from, to)
}

func fixSigning() error {
	if err := command("xattr", "-cr", appPath); err != nil {
		return err
	}

	return command("codesign", "--force", "--deep", "-s", "-", appPath)
}

func command(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
