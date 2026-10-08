// handles API communications

package main

import (
	"encoding/json"
	"fmt"
	"io"
)

func queryMod(guid string) (*Mod, error) {
	resp, downErr := DownloadStream(fmt.Sprintf("https://monkeforge.org/api/v1/mods/%s", guid))

	if downErr != nil {
		return nil, downErr
	}

	defer resp.Close()

	jsonBytes, getErr := io.ReadAll(resp)
	if getErr != nil {
		return nil, getErr
	}

	var baseJson map[string]any

	if parseErr := json.Unmarshal(jsonBytes, &baseJson); parseErr != nil {
		fmt.Println("Error parsing JSON:", parseErr)
		return nil, parseErr
	}

	if _, exists := baseJson["error"]; exists {
		if baseJson["error"] == "This mod could not be found." {
			return nil, fmt.Errorf("package %s not found", guid)
		}

		return nil, fmt.Errorf("monkeforge.org returned error \"%s\"", baseJson["error"])
	}

	var mod Mod

	if unmarshalErr := json.Unmarshal(jsonBytes, &mod); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &mod, nil
}

func queryReleases(guid string, channel string) (*[]Release, error) {
	resp, downErr := DownloadStream(fmt.Sprintf("https://monkeforge.org/api/v1/mods/%s/releases?channel=%s", guid, channel))

	if downErr != nil {
		return nil, downErr
	}

	defer resp.Close()

	jsonBytes, getErr := io.ReadAll(resp)
	if getErr != nil {
		return nil, getErr
	}

	var baseJson map[string]any

	if parseErr := json.Unmarshal(jsonBytes, &baseJson); parseErr != nil {
		fmt.Println("Error parsing JSON:", parseErr)
		return nil, parseErr
	}

	if _, exists := baseJson["error"]; exists {
		if baseJson["error"] == "This mod could not be found." {
			return nil, fmt.Errorf("package %s not found", guid)
		}

		return nil, fmt.Errorf("monkeforge.org returned error \"%s\"", baseJson["error"])
	}

	var releases Releases
	if unmarshalErr := json.Unmarshal(jsonBytes, &releases); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &releases.Items, nil
}
