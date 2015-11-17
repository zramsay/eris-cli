package commands

import (
	"path"
	"strings"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/data"
	srv "github.com/eris-ltd/eris-cli/services"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var Keys = &cobra.Command{
	Use:   "keys",
	Short: "Do specific tasks with keys *for dev only*.",
	Long: `The keys subcommand is an opiniated wrapper around
	eris-keys and requires a keys container to be running. 

	It is for development only. Advanced functionality is available via:
		$ eris services exec keys "eris-keys ..."
		
	see https://github.com/eris-ltd/eris-keys for more info.
	
	Mint tools are available via an erisdb container:
		$ eris chains exec chainName "mint..."

	see https://github.com/eris-ltd/mint-client for more info`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func buildKeysCommand() {
	Keys.AddCommand(keysGen)
	Keys.AddCommand(keysPub)
	Keys.AddCommand(keysExport)
	Keys.AddCommand(keysConvert)
}

var keysGen = &cobra.Command{
	Use:   "gen",
	Short: "Generates a key using the keys container.",
	Long: `Generates a key using the keys container.

	Key is saved in keys data container and can be exported to host with:
		$ eris keys export

	Command is equivalent to:
		$ eris services exec keys "eris-keys gen --no-pass"`,
	Run: GenerateKey,
}

var keysPub = &cobra.Command{
	Use:   "pub ADDR",
	Short: "Returns a pubkey given an address.",
	Long: `Returns a pubkey given an address.
	
	Command is equivalent to:
		$ eris services exec keys "eris-keys pub --addr ADDR"`,
	Run: GetPubKey,
}

var keysExport = &cobra.Command{
	Use:   "export",
	Short: "Export all keys from container to host.",
	Long: `Export all keys from container to host.
	
	Takes the contents of 
		/home/eris/.eris/keys/data/ 

	in the keys container and copies everything to
		/home/user/.eris/keys/data/

	on the host.`,
	Run: ExportKey,
}

var keysConvert = &cobra.Command{
	Use:   "convert ADDR",
	Short: "Convert and eris-keys key to tendermint key",
	Long: `Convert and eris-keys key to tendermint key
	
	Command is equivalent to:
		$ eris chains exec someChain "mintkey mint ADDR"
	without requiring a pre-existing chain, however.
	
	Usually, it's output will be piped into 
		~/.eris/chains/newChain/priv_validator.json`,
	Run: ConvertKey,
}

func GenerateKey(cmd *cobra.Command, args []string) {

	IfExit(ArgCheck(0, "eq", cmd, args)) // no args needed!

	do.Name = "keys"
	do.Operations.ContainerNumber = 1
	IfExit(srv.EnsureRunning(do))
	do.Operations.Interactive = false
	do.Operations.Args = []string{"eris-keys", "gen", "--no-pass"}

	IfExit(srv.ExecService(do))
}

func GetPubKey(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args)) // the addr
	addr := strings.TrimSpace(args[0])

	do.Name = "keys"
	do.Operations.ContainerNumber = 1
	IfExit(srv.EnsureRunning(do))
	do.Operations.Interactive = false
	do.Operations.Args = []string{"eris-keys", "pub", "--addr", addr}

	IfExit(srv.ExecService(do))
}

//from /home/eris/.eris/keys/data/ to /home/user/.eris/keys/data/
func ExportKey(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(0, "eq", cmd, args)) // all keys for now; TODO specify addr

	do.Name = "keys" //for cont as well as path-joined for final dir
	IfExit(srv.EnsureRunning(do))
	do.ErisPath = KeysPath
	//src in container
	do.Path = path.Join(ErisContainerRoot, "keys", "data")

	IfExit(data.ExportData(do))
}

func ConvertKey(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args)) // the addr

	do.Name = "keys"
	IfExit(srv.EnsureRunning(do))

	do.Chain.ChainType = "throwaway"
	do.Name = "default"
	tmp := do.Name
	do.Operations.PublishAllPorts = true
	IfExit(chains.ThrowAwayChain(do))
	do.Chain.Name = do.Name // setting this for tear down purposes
	logger.Debugf("ThrowAwayChain booted =>\t%s\n", do.Name)

	do.Name = tmp
	addr := strings.TrimSpace(args[0])
	do.Operations.Args = []string{"mintkey", "mint", addr}

	err := chains.ExecChain(do)
	if err != nil {
		do.Rm = true
		do.RmD = true
		IfExit(chains.CleanUp(do))
	}

	do.Rm = true
	do.RmD = true
	IfExit(chains.CleanUp(do))

}
