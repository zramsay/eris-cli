package commands

import (
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/keys"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/spf13/cobra"
)

var Keys = &cobra.Command{
	Use:   "keys",
	Short: "do specific tasks with keys",
	Long: `the keys subcommand is an opiniated wrapper around
[eris-keys] and requires a keys container to be running

It is for development only. Advanced functionality is available via
the [eris services exec keys "eris-keys CMD"] command.

See https://monax.io/docs/documentation/keys/ for more info.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func buildKeysCommand() {
	Keys.AddCommand(keysGen)
	Keys.AddCommand(keysExport)
	Keys.AddCommand(keysImport)
	Keys.AddCommand(keysList)
	addKeysFlags()
}

var keysGen = &cobra.Command{
	Use:   "gen",
	Short: "generates an unsafe key in the keys container",
	Long: `generates an unsafe key in the keys container

Key is created in keys data container and can be exported to host
by using the [--save] flag or by running [eris keys export ADDR].`,
	Run: GenerateKey,
}

var keysExport = &cobra.Command{
	Use:   "export ADDR",
	Short: "export a key from container to host",
	Long: `export a key from container to host

Takes a key from /home/eris/.eris/keys/data/ADDR/ADDR in the keys container
and copies it to ` + util.Tilde(filepath.Join(config.KeysDataPath, "ADDR", "ADDR")) + ` on the host.`,
	Run: ExportKey,
}

var keysImport = &cobra.Command{
	Use:   "import ADDR",
	Short: "import a key to container from host",
	Long: `import a key to container from host

Takes a key from ` + util.Tilde(filepath.Join(config.KeysDataPath, "ADDR", "ADDR")) + `
on the host and copies it to /home/eris/.eris/keys/data/ADDR/ADDR
in the keys container.`,
	Run: ImportKey,
}

var keysList = &cobra.Command{
	Use:   "ls",
	Short: "list keys on host and in running keys container",
	Long: `list keys on host and in running keys container

Specify location with flags --host or ---container.

Latter flag is equivalent to: [eris services exec keys "ls /home/eris/.eris/keys/data"]`,
	Run: ListKeys,
}

func addKeysFlags() {
	// [zr] eventually we'll want to flip (both?) these bools. definitely the latter, probably the former
	keysGen.Flags().BoolVarP(&do.Save, "save", "", false, "export the key to host following creation")
	//keysGen.Flags().BoolVarP(&do.Password, "passwd", "", false, "require a password prompt to generate the key")

	keysExport.Flags().StringVarP(&do.Address, "addr", "", "", "address of key to export")
	keysExport.Flags().BoolVarP(&do.All, "all", "", false, "export all keys. do not provide any arguments")

	keysImport.Flags().StringVarP(&do.Address, "addr", "", "", "address of key to import")
	keysImport.Flags().BoolVarP(&do.All, "all", "", false, "import all keys. do not provide any arguments")

	keysList.Flags().BoolVarP(&do.Host, "host", "", false, "list keys on host in "+util.Tilde(config.KeysDataPath))
	keysList.Flags().BoolVarP(&do.Container, "container", "", false, "list keys in container in /home/eris/.eris/keys/data")

}

func GenerateKey(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(0, "eq", cmd, args))

	// TODO implement once we move to using keys client exclusively
	// if do.Password {}

	util.IfExit(keys.GenerateKey(do))
}

func ExportKey(cmd *cobra.Command, args []string) {
	if do.All {
		util.IfExit(ArgCheck(0, "eq", cmd, args))
	} else {
		util.IfExit(ArgCheck(1, "eq", cmd, args))
		do.Address = strings.TrimSpace(args[0])
	}
	util.IfExit(keys.ExportKey(do))
}

func ImportKey(cmd *cobra.Command, args []string) {
	if do.All {
		util.IfExit(ArgCheck(0, "eq", cmd, args))
	} else {
		util.IfExit(ArgCheck(1, "eq", cmd, args))
		do.Address = strings.TrimSpace(args[0])
	}
	util.IfExit(keys.ImportKey(do))
}

func ListKeys(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(0, "eq", cmd, args))
	if !do.Host && !do.Container {
		do.Host = true
		do.Container = true
	}
	_, err := keys.ListKeys(do)
	util.IfExit(err)
}
