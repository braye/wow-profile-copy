package main

import (
	"errors"
	"github.com/pterm/pterm"
	"io"
	"os"
	"path/filepath"
)

// deduplicates slices by throwing them into a map
// not mine, credit to @kylewbanks
func deduplicateStringSlice(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)
	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}
	return u
}

// small wrapper around os and io to copy files from source to destination
func copyFile(src string, dest string) (bytes int64, err error) {
	if src == dest {
		return -1, errors.New("source and destination are the same")
	}
	srcFileHandle, err := os.Open(src)
	if err != nil {
		return -1, err
	}
	defer srcFileHandle.Close()

	dstFileHandle, err := os.Create(dest)
	if err != nil {
		return -1, err
	}
	defer dstFileHandle.Close()

	bytes, err = io.Copy(dstFileHandle, srcFileHandle)
	return bytes, err
}

// determines if a given path appears to contain a WoW install
func isWowInstallDirectory(dir string) bool {
	var isInstallDir = false

	files, err := os.ReadDir(dir)
	if err != nil {
		// directory probably doesn't exist
		return false
	}

	for _, file := range files {
		// if this directory contains a wow instance folder name, it's probably where WoW is installed
		_, matchesInstanceName := _wowInstanceFolderNames[file.Name()]
		if file.IsDir() && matchesInstanceName {
			isInstallDir = true
			break
		}
	}

	return isInstallDir
}

func promptForWowDirectory(dir string) (wowDir string, err error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var fileChoices = []string{".. (go back)"}

	for _, file := range files {
		fileChoices = append(fileChoices, file.Name())
	}

	selectedFile, _ := pterm.DefaultInteractiveSelect.
		WithOptions(fileChoices).
		WithDefaultText("Select a WoW Install directory").
		WithMaxHeight(15).
		Show()
	var fullSelectedPath string
	if selectedFile == ".. (go back)" {
		fullSelectedPath = filepath.Clean(filepath.Join(dir, ".."))
	} else {
		fullSelectedPath = filepath.Join(dir, selectedFile)
	}
	isWowDir := isWowInstallDirectory(fullSelectedPath)
	if !isWowDir {
		return promptForWowDirectory(fullSelectedPath)
	} else {
		return fullSelectedPath, nil
	}
}
