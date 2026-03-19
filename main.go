package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/itchio/damage"
	"github.com/itchio/damage/hdiutil"
	"github.com/itchio/headway/state"
)

const owner = "ungoogled-software"
const repo = "ungoogled-chromium-macos"

const downloadPath = "downloads"
const appPath = "/Applications/Chromium.app"
const wideVineSrcPath = "gc/Google Chrome.app/Contents/Frameworks/Google Chrome Framework.framework/Libraries/WidevineCdm"
const wideVineDestPath = "/Applications/Chromium.app/Contents/Frameworks/Chromium Framework.framework/Libraries/WidevineCdm"

type asset struct {
	Url  string `json:"browser_download_url"`
	Name string `json:"name"`
}
type apiRes struct {
	Assets []asset `json:"assets"`
}

var consumer *state.Consumer = &state.Consumer{
	OnMessage: func(level, msg string) { log.Printf("[%s] %s\n", level, msg) },
}

var gcAsset = &asset{
	Name: "googlechrome.dmg",
	Url:  "https://dl.google.com/chrome/mac/universal/stable/GGRO/googlechrome.dmg",
}

func main() {
	ucAsset, err := getLatestAsset()
	if err != nil {
		log.Fatalf("erro obtendo latest: %v", err)
	}

	log.Println("downloading ungoogled chromium...")
	ucPath, err := downloadFile(ucAsset)
	if err != nil {
		log.Fatalf("download error: %v", err)
	}

	log.Println("downloading google chrome...")
	gcPath, err := downloadFile(gcAsset)
	if err != nil {
		log.Fatalf("download error: %v", err)
	}

	log.Println("installing...")
	err = install(ucPath, gcPath)
	if err != nil {
		log.Fatalf("install error: %v", err)
	}

	log.Println("done!")
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

func getLatestAsset() (*asset, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	res, err := http.Get(url)
	if err != nil {
		log.Fatalf("erro buscando versão: %v", err)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if res.StatusCode >= 400 {
		log.Fatalf("status %d - %s", res.StatusCode, string(data))
	}

	var r apiRes
	if err = json.Unmarshal(data, &r); err != nil {
		log.Fatalln(err)
	}

	plat := runtime.GOARCH
	if plat == "amd64" {
		plat = "x86_64"
	}

	for _, a := range r.Assets {
		if strings.Contains(a.Name, plat) {
			return &a, nil
		}
	}

	return nil, errors.New("asset not found")
}

func downloadFile(a *asset) (string, error) {
	os.MkdirAll(downloadPath, 0755)
	path := filepath.Join(downloadPath, a.Name)

	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	res, err := http.Get(a.Url)
	if err != nil {
		return "", err
	}

	if res.StatusCode >= 400 {
		return "", fmt.Errorf("status %d", res.StatusCode)
	}

	if _, err = io.Copy(f, res.Body); err != nil {
		return "", err
	}

	return path, nil
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
