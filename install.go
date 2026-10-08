package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/pterm/pterm"
)

func install(modIds []string, no_verify bool, channel string) {
	if config.GamePath == "" {
		if _, err := getGamePath(false); err != nil {
			errStyle.Println("unable to get game path; exiting")
			os.Exit(1)
		}
	}

	loadForgeJson()

	var mods []*Mod
	modsRemove := []ModManifest{}

	if (!guidIsInstalled("bepinex")) && !slices.Contains(modIds, "bepinex") {
		mods = append(mods, getBepInEx())
	}

	for _, modId := range modIds {
		var mod *Mod
		var err error

		if modId == "bepinex" {
			mod = getBepInEx()

			if guidIsInstalled(mod.GUID) {
				modsRemove = append(modsRemove, *getManifest(mod.GUID))
			}

			mods = append(mods, mod)
			continue
		} else {
			mod, err = queryMod(modId)
		}

		if err != nil {
			errStyle.Printf("err: %v\n", err)
			return
		}

		idx := slices.IndexFunc(mod.Channels, func(c Channel) bool {
			return c.Name == channel
		})

		if idx == -1 {
			errStyle.Printf("no release candidate on mod [%s] with channel %s\n", mod.GUID, channel)
			return
		} else {
			mod.Version = mod.Channels[idx].Version
		}

		if versionIsInstalled(mod.GUID, mod.Version) {
			infoStyle.Printf("Skipping [%s] v%s because it is already installed\n", mod.GUID, mod.Version)
			continue
		}

		if guidIsInstalled(mod.GUID) {
			modsRemove = append(modsRemove, *getManifest(mod.GUID))
		}

		mods = append(mods, mod)
	}

	if len(mods) == 0 {
		infoStyle.Println("No items are queued for installation")
		return
	}

	pterm.FgBlue.Println("Installation queue:")
	var listItems []pterm.BulletListItem
	for _, mod := range modsRemove {
		listItems = append(listItems, pterm.BulletListItem{
			Level: 0,
			Text:  fmt.Sprintf(color.RedString("- "+"%s [%s] v%s"), mod.Name, mod.GUID, mod.Version),
		})
	}

	for _, mod := range mods {
		listItems = append(listItems, pterm.BulletListItem{
			Level: 0,
			Text:  fmt.Sprintf("%s [%s] v%s", mod.Name, mod.GUID, mod.Version),
		})
	}

	pterm.DefaultBulletList.WithItems(listItems).Render()

	confirmed, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(true).
		Show("Does this look okay?")

	if !confirmed {
		os.Exit(0)
	}

	for _, manifest := range modsRemove {
		removeMod(manifest.GUID)
	}

	for _, mod := range mods {
		msgFormat := fmt.Sprintf("%s [v%s]", mod.GUID, mod.Version)

		var release Release
		var download Download
		if mod.GUID != "bepinex" {
			releases, err := queryReleases(mod.GUID, channel)
			if err != nil {
				errStyle.Println(err)
				return
			}

			release = (*releases)[0]
		} else {
			release = mod.Releases[0]
		}

		download = release.Downloads[0]

		p, _ := pterm.DefaultProgressbar.
			WithTotal(100).
			WithTitle(msgFormat).
			Start()

		stream, streamErr := DownloadStreamWithProgress(download.Url, func(downloaded, total int64) {
			completion := (float64(downloaded) / float64(total)) * 100

			delta := int(completion) - p.Current
			if delta > 0 {
				p.Add(delta)
			}
		})

		if streamErr != nil {
			errStyle.Println(streamErr)
			return
		}

		defer stream.Close()

		bytes, err := io.ReadAll(stream)
		if err != nil {
			errStyle.Println(err)
			return
		}

		sha := sha256.Sum256(bytes)
		if !(download.Sha256 == "" || no_verify) && download.Sha256 != hex.EncodeToString(sha[:]) {
			errStyle.Printf("error: sha256 verification failed for [%s] v%s download.. skipping\n", mod.GUID, mod.Version)
		}

		if strings.ToLower(download.Type) == "zip" {
			hasBep, err := HasBepInExFolder(bytes)

			if err != nil {
				errStyle.Println(err)
				return
			}

			path := config.GamePath

			if !hasBep {
				path = filepath.Join(config.GamePath, "BepInEx", "plugins", mod.GUID)

				mkErr := os.MkdirAll(path, os.ModePerm)
				if mkErr != nil {
					errStyle.Println(mkErr)
					return
				}
			}

			UnzipBytes(bytes, path)
			manifest, err := GenerateManifest(bytes, path)

			if err != nil {
				errStyle.Println(err)
				return
			}

			manifest.Version = mod.Version
			manifest.GUID = mod.GUID
			manifest.Name = mod.Name

			forgeJson.Installed = append(forgeJson.Installed, *manifest)
		} else if strings.ToLower(download.Type) == "dll" {
			path := filepath.Join(config.GamePath, "BepInEx", "plugins", mod.GUID)

			mkErr := os.MkdirAll(path, os.ModePerm)
			if mkErr != nil {
				errStyle.Println(mkErr)
				return
			}

			err := os.WriteFile(filepath.Join(path, mod.Name+".dll"), bytes, 0644)
			if err != nil {
				errStyle.Println(err)
			}

			manifest := ModManifest{
				GUID:    mod.GUID,
				Name:    mod.Name,
				Version: mod.Version,
				Files:   []string{},
				Folders: []string{filepath.Join(config.GamePath, "BepInEx", "plugins", mod.GUID)},
			}

			forgeJson.Installed = append(forgeJson.Installed, manifest)
		}

		saveForgeJson()
	}
}
