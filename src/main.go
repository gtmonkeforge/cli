package main

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var infoStyle *pterm.Style
var errStyle *pterm.Style
var debugStyle *pterm.Style

var rootCmd = &cobra.Command{
	Use:   "mforge",
	Short: "MonkeForge CLI",
	Long: `MonkeForge CLI - https://monkeforge.org/download
	
	This CLI is for advanced users and developers. If you know what you're doing,
	have fun!
	
	Install mods - mforge install <id> <id2> ...
	Uninstall mods - mforge uninstall <id> <id2>
	Get information about a package - mforge info <id>`,
}

var installCmd = &cobra.Command{
	Use:   "install [guids...]",
	Short: "Install mods",

	Args: cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		no_verify, _ := cmd.Flags().GetBool("no-verify")
		channel, _ := cmd.Flags().GetString("channel")

		install(args, no_verify, channel)
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall [guids...]",
	Short: "Unnstall/remove mods",

	Args: cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		uninstall(args)
	},
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade all installed mods if updates are avaliable",

	Run: func(cmd *cobra.Command, args []string) {
		no_verify, _ := cmd.Flags().GetBool("no-verify")
		channel, _ := cmd.Flags().GetString("channel")

		upgrade(no_verify, channel)
	},
}

var aboutCmd = &cobra.Command{
	Use:   "about [guids]",
	Short: "View information about a GUID",

	Args: cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		about(args[0])
	},
}

var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Re-find your Steam game path or input one manually",

	Run: func(cmd *cobra.Command, args []string) {
		manual, _ := cmd.Flags().GetBool("manual")
		getGamePath(manual)
	},
}

func main() {
	infoStyle = pterm.NewStyle(pterm.FgBlue)
	errStyle = pterm.NewStyle(pterm.FgRed, pterm.Bold)
	debugStyle = pterm.NewStyle(pterm.FgGray)

	loadConfig()

	installFlags := installCmd.Flags()
	installFlags.BoolP("no-verify", "N", false, "Disable verifying mod downloads with SHA256")
	installFlags.StringP("channel", "r", "stable", "Choose release channel to download from")

	upgradeFlags := upgradeCmd.Flags()
	upgradeFlags.BoolP("no-verify", "N", false, "Disable verifying mod downloads with SHA256")
	upgradeFlags.StringP("channel", "r", "stable", "Choose release channel to download from")

	selectCmd.Flags().
		BoolP("manual", "m", false, "Manually select the game path and skip auto-detection")

	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(aboutCmd)
	rootCmd.AddCommand(selectCmd)

	if err := rootCmd.Execute(); err != nil {
		errStyle.Println(err)
		os.Exit(1)
	}

	saveConfig()
}
