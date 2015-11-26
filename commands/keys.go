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

It is for development only. Advanced functionality is available via:
$ eris services exec keys "eris-keys CMD"
	
see https://docs.erisindustries.com/documentation/eris-keys/ for more info`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func buildKeysCommand() {
	Keys.AddCommand(keysGen)
	Keys.AddCommand(keysPub)
	Keys.AddCommand(keysExport)
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

//TODO optional ADDR
var keysExport = &cobra.Command{
	Use:   "export",
	Short: "Export keys from container to host.",
	Long: `Export keys from container to host.
	
Takes the contents (or a single key via addr flag) of
/home/eris/.eris/keys/data/

in the keys container and copies everything (by default) to
$HOME/user/.eris/keys/data/

on the host. Use the addr flag to export single keys.`,
	Run: ExportKey,
}

//TODO optional ADDR & cmd description
var keysImport = &cobra.Command{
	Use:   "import",
	Short: "Import keys to container from host.",
	Long: `Import keys to container from host.

Takes the contents (or a single key via addr flag) of
$HOME/user/.eris/keys/data/

on the host and copies everything (by default) to
/home/eris/.eris/keys/data/

in the keys container. Use the addr flag to import single keys.`,
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

func addKeysFlags() {
	keysExport.Flags().StringVarP(&do.Destination, "dest", "", "", "destination for export on host")
	keysExport.Flags().StringVarP(&do.Address, "addr", "", "", "address of key to export")
	keysImport.Flags().StringVarP(&do.Source, "src", "", "", "source on host to import from")
	keysImport.Flags().StringVarP(&do.Address, "addr", "", "", "address of key to import")

}

func GenerateKey(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(0, "eq", cmd, args)) // no args needed!

	IfExit(keys.GenerateKey(do))
}

func GetPubKey(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args)) // the addr
	do.Address = strings.TrimSpace(args[0])
	IfExit(keys.GetPubKey(do))
}

//from /home/eris/.eris/keys/data/ to /home/user/.eris/keys/data/
func ExportKey(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(0, "eq", cmd, args)) // all keys for now; TODO specify addr
	IfExit(keys.ExportKey(do))
}

func ImportKey(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(0, "eq", cmd, args)) // all keys for now; TODO specify addr
	IfExit(keys.ImportKey(do))
}

func ConvertKey(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args)) // the addr
	do.Address = strings.TrimSpace(args[0])
	IfExit(keys.ConvertKey(do))
}
