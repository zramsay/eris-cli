package commands

import (
	"strings"

	"github.com/eris-ltd/eris-cli/keys"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var Keys = &cobra.Command{
	Use:   "keys",
	Short: "Do specific tasks with keys *for dev only*.",
	Long: `The keys subcommand is an opiniated wrapper around
eris-keys and requires a keys container to be running. 

It is for development only. 
Advanced functionality is available via: [eris services exec keys "eris-keys CMD"]
	
see https://docs.erisindustries.com/documentation/eris-keys/ for more info`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func buildKeysCommand() {
	Keys.AddCommand(keysGen)
	Keys.AddCommand(keysPub)
	Keys.AddCommand(keysExport)
	Keys.AddCommand(keysImport)
	Keys.AddCommand(keysConvert)
	addKeysFlags()
}

var keysGen = &cobra.Command{
	Use:   "gen",
	Short: "Generates an unsafe key using the keys container.",
	Long: `Generates a key using the keys container.
WARNING: this command is not safe for production.
For development only.

Key is saved in keys data container and can be 
exported to host with: [eris keys export]

Command is equivalent to: [eris services exec keys "eris-keys gen --no-pass"]`,
	Run: GenerateKey,
}

var keysPub = &cobra.Command{
	Use:   "pub ADDR",
	Short: "Returns a machine readable pubkey given an address.",
	Long: `Returns a machine readable pubkey given an address.
	
Command is equivalent to: [eris services exec keys "eris-keys pub --addr ADDR"]`,
	Run: GetPubKey,
}

var keysExport = &cobra.Command{
	Use:   "export ADDR",
	Short: "Export a key from container to host.",
	Long: `Export a key from container to host.
	

Takes a key from:
/home/eris/.eris/keys/data/ADDR/ADDR

in the keys container and copies it to
$HOME/user/.eris/keys/data/ADDR/ADDR

on the host. Optionally specify host destination with --dest.`,
	Run: ExportKey,
}

var keysImport = &cobra.Command{
	Use:   "import ADDR",
	Short: "Import a key to container from host.",
	Long: `Import a key to container from host.

Takes a key from:
$HOME/user/.eris/keys/data/ADDR/ADDR

on the host and copies it to
/home/eris/.eris/keys/data/ADDR/ADDR

in the keys container.`,
	Run: ImportKey,
}

var keysConvert = &cobra.Command{
	Use:   "convert ADDR",
	Short: "Convert and eris-keys key to tendermint key",
	Long: `Convert and eris-keys key to tendermint key

Command is equivalent to: [eris services exec keys "mintkey mint ADDR"]

Usually, it's output will be piped into
$HOME/.eris/chains/newChain/priv_validator.json`,
	Run: ConvertKey,
}

//the container path is always hardcoded to /home/eris/.eris/keys/data
func addKeysFlags() {
	keysExport.Flags().StringVarP(&do.Destination, "dest", "", "", "destination for export on host")
	keysExport.Flags().StringVarP(&do.Address, "addr", "", "", "address of key to export")
	keysImport.Flags().StringVarP(&do.Source, "src", "", "", "source on host to import from. give full filepath to key")
	keysImport.Flags().StringVarP(&do.Address, "addr", "", "", "address of key to import")

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
	IfExit(ArgCheck(1, "eq", cmd, args))
	do.Address = strings.TrimSpace(args[0])
	IfExit(keys.ExportKey(do))
}

func ImportKey(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args))
	do.Address = strings.TrimSpace(args[0])
	IfExit(keys.ImportKey(do))
}

func ConvertKey(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args))
	do.Address = strings.TrimSpace(args[0])
	IfExit(keys.ConvertKey(do))
}
