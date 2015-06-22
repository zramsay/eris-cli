package commands

import (
	"github.com/eris-ltd/eris-cli/keys"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// Primary Keys Sub-Command
var Keys = &cobra.Command{
	Use:   "keys",
	Short: "Manage Keys for your Application.",
	Long: `The keys subcommand is used to generate, import, export, and use
the cryptographic keys for your application.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// build the keys subcommand
func buildKeysCommand() {
	Keys.AddCommand(keysNew)
	Keys.AddCommand(keysList)
	Keys.AddCommand(keysRename)
	Keys.AddCommand(keysUse)
	Keys.AddCommand(keysImport)
	Keys.AddCommand(keysRm)
}

// new
// flags to add: --type
var keysNew = &cobra.Command{
	Use:   "new",
	Short: "Generate new keys.",
	Long:  `Generate new keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		keys.Generate(cmd, args)
	},
}

// ls
var keysList = &cobra.Command{
	Use:   "ls",
	Short: "List available keys.",
	Long:  `List available keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		keys.List()
	},
}

// rename
var keysRename = &cobra.Command{
	Use:   "new [old] [new]",
	Short: "Rename a key.",
	Long:  `Rename a key.`,
	Run: func(cmd *cobra.Command, args []string) {
		keys.Rename(cmd, args)
	},
}

// use
var keysUse = &cobra.Command{
	Use:   "use [name]",
	Short: "Use a key.",
	Long:  `Use a key.`,
	Run: func(cmd *cobra.Command, args []string) {
		keys.Use(cmd, args)
	},
}

// import
var keysImport = &cobra.Command{
	Use:   "import [key-file]",
	Short: "Import a key.",
	Long:  `Import a key.`,
	Run: func(cmd *cobra.Command, args []string) {
		keys.Import(cmd, args)
	},
}

// rm
// flags to add: --force
var keysRm = &cobra.Command{
	Use:   "rm [name]",
	Short: "Remove a key.",
	Long: `Remove a key.

**CAUTION**: This command will remove a key entire.`,
	Run: func(cmd *cobra.Command, args []string) {
		keys.Remove(cmd, args)
	},
}
