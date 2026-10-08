package main

import (
	"os"
)

func upgrade(no_verify bool, channel string) {
	if config.GamePath == "" {
		if _, err := getGamePath(false); err != nil {
			errStyle.Println("unable to get game path; exiting")
			os.Exit(1)
		}
	}

	loadForgeJson()

	ids := []string{}

	for _, manifest := range forgeJson.Installed {
		ids = append(ids, manifest.GUID)
	}

	install(ids, no_verify, channel)
}
