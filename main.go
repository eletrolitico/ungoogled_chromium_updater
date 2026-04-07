package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

const downloadPath = "downloads"

type asset struct {
	Url  string `json:"browser_download_url"`
	Name string `json:"name"`
}
type apiRes struct {
	Assets []asset `json:"assets"`
}

func main() {
	ucPath, gcPath, err := downloadAssets()
	if err != nil {
		log.Fatalf("erro baixando assets: %v", err)
	}

	log.Println("installing...")
	err = install(ucPath, gcPath)
	if err != nil {
		log.Fatalf("install error: %v", err)
	}

	log.Println("done!")
}

func getLatestAsset(owner, repo, search string) (*asset, error) {
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

	for _, a := range r.Assets {
		if strings.Contains(a.Name, search) {
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

	bar := progressbar.DefaultBytes(
		res.ContentLength,
		fmt.Sprintf("downloading %s...", a.Name),
	)

	_, err = io.Copy(io.MultiWriter(f, bar), res.Body)
	if err != nil {
		return "", err
	}

	return path, nil
}
