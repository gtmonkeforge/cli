package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	GamePath string `json:"path"`
}

var config Config

func saveConfig() {
	config_dir, _ := os.UserConfigDir()
	config_path := filepath.Join(config_dir, "mforge.json")

	bytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(config_path, bytes, 0644)
}

func loadConfig() {
	config_dir, _ := os.UserConfigDir()
	config_path := filepath.Join(config_dir, "mforge.json")

	if _, err := os.Stat(config_path); os.IsNotExist(err) {
		config = Config{
			GamePath: "",
		}
		return
	}

	configBytes, err := os.ReadFile(config_path)
	if err != nil {
		config = Config{
			GamePath: "",
		}
		return
	}

	if err := json.Unmarshal(configBytes, &config); err != nil {
		config = Config{
			GamePath: "",
		}
		return
	}
}
