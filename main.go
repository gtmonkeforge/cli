package main

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

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
	Use:   "install",
	Short: "Install packages",

	Args: cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		install(args)
	},
}

func main() {
	errStyle = pterm.NewStyle(pterm.FgRed, pterm.Bold)
	debugStyle = pterm.NewStyle(pterm.FgGray)

	rootCmd.AddCommand(installCmd)

	if err := rootCmd.Execute(); err != nil {
		errStyle.Println(err)
		os.Exit(1)
	}
}
