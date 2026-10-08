package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func getBepInExUrl() (string, string, error) {
	apiURL := "https://api.github.com/repos/BepInEx/BepInEx/releases/latest"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "monkeforge-cli")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("github api returned status: %s", resp.Status)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("failed to decode json response: %w", err)
	}

	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, "win_x64") && strings.HasSuffix(asset.Name, ".zip") {
			return release.TagName[1:], asset.BrowserDownloadURL, nil
		}
	}

	return "", "", fmt.Errorf("could not find a windows x64 zip asset in the latest release (%s)", release.TagName)
}

func getBepInEx() *Mod {
	if config.GamePath == "" {
		if _, err := getGamePath(false); err != nil {
			errStyle.Println("unable to get game path; exiting")
			os.Exit(1)
		}
	}

	loadForgeJson()

	version, url, err := getBepInExUrl()

	if err != nil {
		errStyle.Println(err)
		os.Exit(1)
	}

	return &Mod{
		Name:    "BepInEx",
		GUID:    "bepinex",
		Url:     "https://bepinex.dev",
		Version: version,
		Releases: []Release{
			{
				Version:     version,
				ChannelName: "stable",
				Notes:       "Auto-generated from https://github.com/BepInEx/BepInEx",
				Downloads: []Download{
					{
						Name:   "BepInEx.zip",
						Type:   "ZIP",
						Url:    url,
						Size:   -1,
						Sha256: "",
					},
				},
			},
		},
		Channels: []Channel{
			{
				Name:    "stable",
				Version: version,
			},
		},
	}
}
