package commands

import (
	chns "github.com/eris-ltd/eris-cli/chains"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// Primary Chains Sub-Command
var Chains = &cobra.Command{
	Use:   "chains",
	Short: "Start, Stop, and Manage Blockchains.",
	Long: `Start, Stop, and Manage Blockchains.

The chains subcommand is used to start, stop, and configure blockchains.
Within the Eris platform, blockchains are the primary method of storing
structured data which is used by the Eris platform in combination with
IPFS (a globally-accessible content-addressable peer to peer file
storage solution).`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.ListChains()
	},
}

// Build the chains subcommand
func buildChainsCommand() {
	Chains.AddCommand(chainsListKnown)
	Chains.AddCommand(chainsInstall)
	Chains.AddCommand(chainsTypes)
	Chains.AddCommand(chainsNew)
	Chains.AddCommand(chainsAdd)
	Chains.AddCommand(chainsList)
	Chains.AddCommand(chainsCheckout)
	Chains.AddCommand(chainsConfig)
	Chains.AddCommand(chainsStart)
	Chains.AddCommand(chainsLogs)
	Chains.AddCommand(chainsListRunning)
	Chains.AddCommand(chainsKill)
	Chains.AddCommand(chainsRename)
	Chains.AddCommand(chainsRemove)
	Chains.AddCommand(chainsClean)
	Chains.AddCommand(chainsUpdate)
}

// known lists the known chain types which eris can install
// flags to add: --list-versions
var chainsListKnown = &cobra.Command{
	Use:   "known",
	Short: "List all the blockchain types Eris can install.",
	Long: `Lists the blockchain types which Eris can install for your platform. To install
a service, use: eris chains install.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.ListKnown()
	},
}

// install a blockchain library for your machine
var chainsInstall = &cobra.Command{
	Use:   "install [type] [version]",
	Short: "Installs a blockchain library for your platform.",
	Long: `Installs a blockchain library for your platform.  By default, Eris will install
the most recent version of a service unless another version is
passed as an argument. To list known services use:
[eris chains known].`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Install(cmd, args)
	},
}

// types lists the currently installed chain types
var chainsTypes = &cobra.Command{
	Use:   "types",
	Short: "Lists the currently installed blockchain types.",
	Long:  `Lists the currently installed blockchain types.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.ListInstalled()
	},
}

// new
// flags to add: --type, --genesis, --config, --checkout, --force-name
var chainsNew = &cobra.Command{
	Use:   "new [name]",
	Short: "Hashes a new blockchain.",
	Long: `Hashes a new blockchain.

Will use a default genesis.json unless a --genesis flag is passed.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.New(cmd, args)
	},
}

// add
// flags to add: --checkout
var chainsAdd = &cobra.Command{
	Use:   "add [name] [ref]",
	Short: "Adds an existing blockchain to Eris' blockchain tree.",
	Long:  `Adds an existing blockchain to Eris' blockchain tree.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Add(cmd, args)
	},
}

// list
// flags to add: --current, --short, --all
var chainsList = &cobra.Command{
	Use:   "ls",
	Short: "Lists all known blockchains in the Eris tree.",
	Long:  `Lists all known blockchains in the Eris tree.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.ListChains()
	},
}

// checkout
var chainsCheckout = &cobra.Command{
	Use:   "checkout",
	Short: "Checks out a blockchain.",
	Long: `Checks out a blockchain.

Unless Eris is running a particular project, it will only
operate on a single blockchain at a time; namely, it will
operate on the currently *checked* out blockchain. This
command is used to change the currently checked out chain.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Checkout(cmd, args)
	},
}

// config
// flags to add: --chain, --edit
var chainsConfig = &cobra.Command{
	Use:   "config [key]:[val]",
	Short: "Configure a blockchain.",
	Long: `Configure a blockchain.

Multiple config options may be given at the same time.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Config(cmd, args)
	},
}

// start
// flags to add: --commit, --multi, --foreground, --config, --chain
var chainsStart = &cobra.Command{
	Use:   "start",
	Short: "Start a blockchain.",
	Long:  `Start a blockchain.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Start(cmd, args)
	},
}

// logs
// flags to add: --tail
var chainsLogs = &cobra.Command{
	Use:   "logs",
	Short: "Display the logs of a blockchain.",
	Long:  `Display the logs of a running blockchain.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Logs(cmd, args)
	},
}

// ps
var chainsListRunning = &cobra.Command{
	Use:   "ps",
	Short: "List the running blockchains.",
	Long:  `List the running blockchains.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.ListRunning()
	},
}

// kill
var chainsKill = &cobra.Command{
	Use:   "kill [name]",
	Short: "Stop a running blockchains.",
	Long:  `Stop a running blockchains.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Kill(cmd, args)
	},
}

// rename
var chainsRename = &cobra.Command{
	Use:   "rename [old] [new]",
	Short: "Rename a blockchain.",
	Long:  `Rename a blockchain.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Rename(cmd, args)
	},
}

// rm
// flags to add: --force (no confirm), --clean
var chainsRemove = &cobra.Command{
	Use:   "rm [name]",
	Short: "Remove a blockchain's reference from Eris tree.",
	Long: `Remove a blockchain's reference from Eris tree.

[eris chains rm] does not remove any data, just removes
the reference from eris' tree of blockchains. To remove
the blockchain data from the node use: [eris chains clean].`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Remove(cmd, args)
	},
}

// clean
// flags to add: --force (no confirm)
var chainsClean = &cobra.Command{
	Use:   "clean [name]",
	Short: "Clean a blockchains' data from the node.",
	Long: `Clean a blockchains' data from the node.

Clean will remove the blockchain reference as well as
its data.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Clean(cmd, args)
	},
}

// update
var chainsUpdate = &cobra.Command{
	Use:   "update [type]",
	Short: "Update a blockchain library.",
	Long:  `Update a blockchain library.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Update(cmd, args)
	},
}
