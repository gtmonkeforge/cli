// Manages forge.json file for tracking mods

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
)

type ModManifest struct {
	Folders []string `json:"folders"`
	Files   []string `json:"files"`
	GUID    string   `json:"guid"`
	Name    string   `json:"name"`
	Version string   `json:"version"`
}

type ForgeJson struct {
	Installed []ModManifest `json:"installed"`
}

var forgeJson ForgeJson

func saveForgeJson() {
	forgePath := filepath.Join(config.GamePath, "forge.json")

	forgeBytes, err := json.MarshalIndent(forgeJson, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(forgePath, forgeBytes, 0644)
}

func loadForgeJson() {
	forge_path := filepath.Join(config.GamePath, "forge.json")

	if _, err := os.Stat(forge_path); os.IsNotExist(err) {
		forgeJson = ForgeJson{
			Installed: []ModManifest{},
		}
		return
	}

	forgeBytes, err := os.ReadFile(forge_path)
	if err != nil {
		forgeJson = ForgeJson{
			Installed: []ModManifest{},
		}
		return
	}

	if err := json.Unmarshal(forgeBytes, &forgeJson); err != nil {
		forgeJson = ForgeJson{
			Installed: []ModManifest{},
		}
		return
	}
}

func getManifest(guid string) *ModManifest {
	item := slices.IndexFunc(forgeJson.Installed, func(i ModManifest) bool {
		return i.GUID == guid
	})

	return &forgeJson.Installed[item]
}

func guidIsInstalled(guid string) bool {
	item := slices.IndexFunc(forgeJson.Installed, func(i ModManifest) bool {
		return i.GUID == guid
	})

	return item != -1
}

func versionIsInstalled(guid string, version string) bool {
	item := slices.IndexFunc(forgeJson.Installed, func(i ModManifest) bool {
		return i.GUID == guid && i.Version == version
	})

	return item != -1
}
