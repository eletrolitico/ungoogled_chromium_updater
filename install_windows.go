package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	unarr "github.com/gen2brain/go-unarr"
	"github.com/tc-hib/winres"
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
	local := os.Getenv("LOCALAPPDATA")
	installPath := filepath.Join(local, "Chromium", "Application")

	if _, err := os.Stat(installPath); err == nil {
		log.Println("removing old version...")
		if err = os.RemoveAll(installPath); err != nil {
			return fmt.Errorf("error removing old install: %w", err)
		}
	}

	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	internalFolder := strings.TrimSuffix(filepath.Base(ucPath), ".zip")

	if err := unzip(ucPath, installPath, internalFolder); err != nil {
		return fmt.Errorf("error unzipping ungoogled chromium: %w", err)
	}

	os.MkdirAll("tmp", 0755)

	if err := extractWinRes(gcPath, "tmp"); err != nil {
		return fmt.Errorf("error unzipping google chrome: %w", err)
	}

	if err := unzip(filepath.Join("tmp", "UPDATER.PACKED.7Z"), "tmp", ""); err != nil {
		return err
	}

	if err := unzip(filepath.Join("tmp", "updater.7z"), "tmp", ""); err != nil {
		return err
	}

	path, err := findFile("tmp/bin", "installer.exe")
	if err != nil {
		return err
	}

	if err := extractWinRes(path, "tmp"); err != nil {
		return fmt.Errorf("error unzipping google chrome installer: %w", err)
	}

	if err := unzip(filepath.Join("tmp", "CHROME.PACKED.7Z"), "tmp", ""); err != nil {
		return err
	}

	if err := unzip(filepath.Join("tmp", "chrome.7z"), "tmp", ""); err != nil {
		return err
	}

	widevinePath, err := findFile(filepath.Join("tmp", "Chrome-bin"), "WidevineCdm")
	if err != nil {
		return err
	}

	log.Println("moving widevine folder...")
	err = os.Rename(widevinePath, filepath.Join(installPath, "WidevineCdm"))
	if err != nil {
		return err
	}

	os.RemoveAll("tmp")

	return nil
}

func unzip(src string, dest, prefix string) (err error) {
	log.Printf("unzipping %s\n", src)
	r, err := unarr.NewArchive(src)
	if err != nil {
		return err
	}
	defer func() { err = r.Close() }()

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func() error {
		fName := r.Name()
		if prefix != "" {
			if !strings.HasPrefix(fName, prefix) {
				return nil
			}
			fName = strings.TrimPrefix(fName, prefix)
		}

		path := filepath.Join(dest, fName)
		os.MkdirAll(filepath.Dir(path), 0755)
		data, err := r.ReadAll()
		if err != nil {
			return err
		}

		return os.WriteFile(path, data, 0644)
	}

	for {
		if err = r.Entry(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		if err = extractAndWriteFile(); err != nil {
			return err
		}
	}

	return nil
}

func extractWinRes(file string, dest string) error {
	log.Printf("extracting winres %s\n", file)
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	rs, err := winres.LoadFromEXE(f)
	if err != nil {
		return err
	}

	rs.WalkType(winres.Name("B7"), func(resID winres.Identifier, langID uint16, data []byte) bool {
		err = os.WriteFile(filepath.Join(dest, fmt.Sprint(resID)), data, 0644)
		if err != nil {
			panic(err)
		}
		return true
	})

	return nil
}

func findFile(root, search string) (string, error) {
	var path string

	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if strings.Contains(p, search) {
			path = p
			return filepath.SkipAll
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if path == "" {
		return "", fmt.Errorf("file not found")
	}

	return path, nil
}
