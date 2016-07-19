package commands

import (
	"strings"

	"github.com/eris-ltd/eris-cli/keys"

	. "github.com/eris-ltd/common/go/common"
	"github.com/spf13/cobra"
)

var Keys = &cobra.Command{
	Use:   "keys",
	Short: "do specific tasks with keys",
	Long: `the keys subcommand is an opiniated wrapper around
eris-keys and requires a keys container to be running

It is for development only. Advanced functionality is available via
the [eris services exec keys "eris-keys CMD"] command.

See https://docs.erisindustries.com/documentation/eris-keys/ for more info.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func buildKeysCommand() {
	Keys.AddCommand(keysGen)
	Keys.AddCommand(keysPub)
	Keys.AddCommand(keysExport)
	Keys.AddCommand(keysImport)
	Keys.AddCommand(keysConvert)
	Keys.AddCommand(keysList)
	addKeysFlags()
}

var keysGen = &cobra.Command{
	Use:   "gen",
	Short: "generates an unsafe key using the keys container",
	Long: `generates a key using the keys container

Key is saved in keys data container and can be exported to host
with the [eris keys export] command.

Command is equivalent to [eris services exec keys "eris-keys gen --no-pass"]`,
	Run: GenerateKey,
}

var keysPub = &cobra.Command{
	Use:   "pub ADDR",
	Short: "returns a machine readable pubkey given an address",
	Long: `returns a machine readable pubkey given an address

Command is equivalent to [eris services exec keys "eris-keys pub --addr ADDR"]`,
	Run: GetPubKey,
}

var keysExport = &cobra.Command{
	Use:   "export ADDR",
	Short: "export a key from container to host",
	Long: `export a key from container to host

Takes a key from /home/eris/.eris/keys/data/ADDR/ADDR in the keys container
and copies it to $HOME/user/.eris/keys/data/ADDR/ADDR on the host.

Optionally specify host destination with --dest.`,
	Run: ExportKey,
}

var keysImport = &cobra.Command{
	Use:   "import ADDR",
	Short: "import a key to container from host",
	Long: `import a key to container from host

Takes a key from $HOME/user/.eris/keys/data/ADDR/ADDR
on the host and copies it to /home/eris/.eris/keys/data/ADDR/ADDR
in the keys container.`,
	Run: ImportKey,
}

var keysConvert = &cobra.Command{
	Use:   "convert ADDR",
	Short: "convert and eris-keys key to Tendermint key",
	Long: `convert and eris-keys key to Tendermint key

Command is equivalent to [eris services exec keys "mintkey mint ADDR"]

Usually will be piped into $HOME/.eris/chains/newChain/priv_validator.json`,
	Run: ConvertKey,
}

var keysList = &cobra.Command{
	Use:   "ls",
	Short: "list keys on host and in running keys container",
	Long: `list keys on host and in running keys container

Specify location with flags --host or ---container.

Latter flag is equivalent to:
the [eris actions do keys list] command, which itself wraps
the [eris services exec keys "ls /home/eris/.eris/keys/data"] command.`,
	Run: ListKeys,
}

func addKeysFlags() {
	//keysExport.Flags().StringVarP(&do.Destination, "dest", "", DefKeysPathHost, "destination for export on host")
	keysExport.Flags().StringVarP(&do.Address, "addr", "", "", "address of key to export")
	keysExport.Flags().BoolVarP(&do.All, "all", "", false, "export all keys. do not provide any arguments")

	//keysImport.Flags().StringVarP(&do.Source, "src", "", DefKeysPathHost, "source on host to import from")
	keysImport.Flags().StringVarP(&do.Address, "addr", "", "", "address of key to import")
	keysImport.Flags().BoolVarP(&do.All, "all", "", false, "import all keys. do not provide any arguments")

	keysList.Flags().BoolVarP(&do.Host, "host", "", false, "list keys on host: looks in $HOME/.eris/keys/data")
	keysList.Flags().BoolVarP(&do.Container, "container", "", false, "list keys in container: looks in /home/eris/.eris/keys/data")

}

func GenerateKey(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(0, "eq", cmd, args))
	IfExit(keys.GenerateKey(do))
}

func GetPubKey(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args))
	do.Address = strings.TrimSpace(args[0])
	IfExit(keys.GetPubKey(do))
}

func ExportKey(cmd *cobra.Command, args []string) {
	if do.All {
		IfExit(ArgCheck(0, "eq", cmd, args))
	} else {
		IfExit(ArgCheck(1, "eq", cmd, args))
		do.Address = strings.TrimSpace(args[0])
	}
	//do.Source = KeysContainerPath // placeholders that ought to go
	//do.Destination = KeysDataPath
	IfExit(keys.ExportKey(do))
}

func ImportKey(cmd *cobra.Command, args []string) {
	if do.All {
		IfExit(ArgCheck(0, "eq", cmd, args))
	} else {
		IfExit(ArgCheck(1, "eq", cmd, args))
		do.Address = strings.TrimSpace(args[0])
	}
	//do.Source = KeysDataPath // placeholders that ought to go
	//do.Destination = KeysContainerPath
	IfExit(keys.ImportKey(do))
}

func ConvertKey(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args))
	do.Address = strings.TrimSpace(args[0])
	IfExit(keys.ConvertKey(do))
}

func ListKeys(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(0, "eq", cmd, args))
	if !do.Host && !do.Container {
		do.Host = true
		do.Container = true
	}
	IfExit(keys.ListKeys(do))
}
