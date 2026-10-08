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

	var mod Mod
	unmarshalErr := json.Unmarshal(jsonBytes, &mod)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &mod, nil
}
