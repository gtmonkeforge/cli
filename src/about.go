// info cmd

package main

import "github.com/pterm/pterm"

func about(modId string) {
	var mod *Mod
	var err error

	if modId == "bepinex" {
		mod = getBepInEx()
	} else {
		mod, err = queryMod(modId)
	}

	if err != nil {
		errStyle.Printf("err: %v\n", err)
		return
	}

	pterm.BgLightBlue.Println("About " + mod.GUID)

	tableData := pterm.TableData{
		{"Field", "Value", "Notes"},
		{"GUID", mod.GUID, "Package ID used for installation operations"},
		{"Name", mod.Name, "Display name"},
		{"Version", mod.Version, "Latest version of the mod"},
		{"URL", mod.Url, "Link to mod information page on MonkeForge"},
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
}
