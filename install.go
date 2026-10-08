package main

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
)

func install(modIds []string) {
	var mods []*Mod

	for _, modId := range modIds {
		mod, err := queryMod(modId)

		if err != nil {
			errStyle.Printf("err: %v\n", err)
		}

		mods = append(mods, mod)
	}

	pterm.FgBlue.Println("Installation queue:")
	var listItems []pterm.BulletListItem
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

	for _, mod := range mods {
		msgFormat := fmt.Sprintf("%s [v%s]", mod.GUID, mod.Version)
		p, _ := pterm.DefaultProgressbar.WithTotal(100).WithTitle(msgFormat).Start()

		for i := 0; i < 100; i++ {
			p.Increment()
		}
	}
}
