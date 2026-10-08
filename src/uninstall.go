package main

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/fatih/color"
	"github.com/pterm/pterm"
)

func uninstall(ids []string) {
	if config.GamePath == "" {
		if _, err := getGamePath(false); err != nil {
			errStyle.Println("unable to get game path; exiting")
			os.Exit(1)
		}
	}

	loadForgeJson()

	var mods []*ModManifest

	for _, modId := range ids {
		idx := slices.IndexFunc(forgeJson.Installed, func(i ModManifest) bool {
			return i.GUID == modId
		})

		if idx == -1 {
			// also check registry
			mod, err := queryMod(modId)
			if err != nil {
				errStyle.Println(err)
			}

			idx = slices.IndexFunc(forgeJson.Installed, func(i ModManifest) bool {
				return i.GUID == mod.GUID
			})

			if idx == -1 {
				infoStyle.Printf("Skipping [%s] because it is not installed\n", mod.GUID)
				continue
			}
		}

		mods = append(mods, &forgeJson.Installed[idx])
	}

	if len(mods) == 0 {
		infoStyle.Println("No items are queued for uninstallation")
		return
	}

	pterm.FgBlue.Println("Uninstallation queue:")
	var listItems []pterm.BulletListItem
	for _, mod := range mods {
		listItems = append(listItems, pterm.BulletListItem{
			Level: 0,
			Text:  fmt.Sprintf(color.RedString("- "+"%s [%s] v%s"), mod.Name, mod.GUID, mod.Version),
		})
	}
	pterm.DefaultBulletList.WithItems(listItems).Render()

	confirmed, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(true).
		Show("Does this look okay?")

	if !confirmed {
		os.Exit(0)
	}

	for _, manifest := range mods {
		removeMod(manifest.GUID)
	}

	saveForgeJson()
}

func removeMod(guid string) error {
	idx := slices.IndexFunc(forgeJson.Installed, func(i ModManifest) bool {
		return i.GUID == guid
	})

	if idx == -1 {
		return errors.New("Mod not installed")
	}

	cleanManifest(forgeJson.Installed[idx])
	forgeJson.Installed = slices.Delete(forgeJson.Installed, idx, idx+1)

	return nil
}

func cleanManifest(manifest ModManifest) error {
	for _, folder := range manifest.Folders {
		err := os.RemoveAll(folder)
		if err != nil {
			return err
		}
	}

	for _, file := range manifest.Files {
		err := os.Remove(file)
		if err != nil {
			return err
		}
	}

	return nil
}
