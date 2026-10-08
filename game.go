// game utils

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
)

const (
	gorillaTagAppID   = "1533390"
	gorillaTagDirName = "Gorilla Tag"
)

func getSteamRoot() (string, error) {
	switch runtime.GOOS {
	case "windows":
		programFiles := os.Getenv("ProgramFiles(x86)")

		if programFiles == "" {
			programFiles = os.Getenv("ProgramFiles")
		}

		if programFiles == "" {
			return "", fmt.Errorf("could not find Program Files environment variable")
		}

		return filepath.Join(programFiles, "Steam"), nil

	case "linux":
		home, err := os.UserHomeDir()

		if err != nil {
			return "", err
		}

		standardPath := filepath.Join(home, ".local/share/Steam")
		if _, err := os.Stat(standardPath); err == nil {
			return standardPath, nil
		}

		altPath := filepath.Join(home, ".steam/steam")
		if _, err := os.Stat(altPath); err == nil {
			return altPath, nil
		}

		flatpakPath := filepath.Join(home, ".var/app/com.valvesoftware.Steam/.local/share/Steam")
		if _, err := os.Stat(flatpakPath); err == nil {
			return flatpakPath, nil
		}
		return standardPath, nil

	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func parseLibraryFolders(steamRoot string) ([]string, error) {
	vdfPath := filepath.Join(steamRoot, "steamapps", "libraryfolders.vdf")

	file, err := os.Open(vdfPath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var libraries []string
	scanner := bufio.NewScanner(file)

	currentPath := ""
	inAppsBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "\"path\"") {
			parts := strings.Split(line, "\"")
			if len(parts) >= 4 {
				currentPath = filepath.Clean(parts[3])
			}
		}

		if strings.HasPrefix(line, "\"apps\"") {
			inAppsBlock = true
			continue
		}

		if inAppsBlock {
			if strings.HasPrefix(line, "}") {
				inAppsBlock = false
				currentPath = "" // Reset
				continue
			}

			if strings.Contains(line, "\""+gorillaTagAppID+"\"") {
				if currentPath != "" {
					libraries = append(libraries, currentPath)
				}
			}
		}
	}

	return libraries, scanner.Err()
}

func findGorillaTag() (string, error) {
	steamRoot, err := getSteamRoot()
	if err != nil {
		return "", fmt.Errorf("failed to locate Steam root: %w", err)
	}

	matchingLibraries, err := parseLibraryFolders(steamRoot)
	if err != nil {
		matchingLibraries = []string{steamRoot}
	}

	for _, lib := range matchingLibraries {
		gamePath := filepath.Join(lib, "steamapps", "common", gorillaTagDirName)
		if info, err := os.Stat(gamePath); err == nil && info.IsDir() {
			return gamePath, nil
		}
	}

	return "", fmt.Errorf("could not find gorilla tag")
}

func getGamePath(skip_auto bool) (string, error) {
	if !skip_auto {
		path, autoErr := findGorillaTag()
		if autoErr == nil {
			config.GamePath = path
			return path, nil
		}

		pterm.FgWhite.Println("The path to Gorilla Tag could not be automatically found.")
	}

	pterm.FgWhite.Println("You need to manually input your game path. You can Google how to find")
	pterm.FgWhite.Println("your game folder if needed.")

	result, err := pterm.DefaultInteractiveTextInput.
		WithDefaultText("Paste the path:").
		Show()

	if err != nil {
		return "", err
	}

	config.GamePath = result
	return result, nil
}
