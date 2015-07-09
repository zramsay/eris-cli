package commands

import (
	"fmt"
	"strings"

	chns "github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/loaders"
	srv "github.com/eris-ltd/eris-cli/services"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//----------------------------------------------------------------------
// cli definitions

// Primary Chains Sub-Command
var Chains = &cobra.Command{
	Use:   "chains",
	Short: "Start, Stop, and Manage Blockchains.",
	Long: `Start, Stop, and Manage Blockchains.

The chains subcommand is used to work on erisdb smart contract
blockchain networks. The name is not perfect, as eris is able
to operate a wide variety of blockchains out of the box. Most
of those existing blockchains should be ran via the

eris services ...

commands. As they fall under the rubric of "things I just want
to turn on or off". While you can develop against those
blockchains, you generally aren't developing those blockchains
themselves.

Eris chains is built to help you build blockchains. It is our
opinionated gateway to the wonderful world of permissioned
smart contract networks.

Your own baby blockchain/smart contract machine is just an

eris chains new

away!`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// Build the chains subcommand
func buildChainsCommand() {
	Chains.AddCommand(chainsNew)
	Chains.AddCommand(chainsInstall)
	Chains.AddCommand(chainsImport)
	Chains.AddCommand(chainsListKnown)
	Chains.AddCommand(chainsList)
	Chains.AddCommand(chainsEdit)
	Chains.AddCommand(chainsStart)
	Chains.AddCommand(chainsLogs)
	Chains.AddCommand(chainsListRunning)
	Chains.AddCommand(chainsInspect)
	Chains.AddCommand(chainsExec)
	Chains.AddCommand(chainsStop)
	Chains.AddCommand(chainsExport)
	Chains.AddCommand(chainsRename)
	Chains.AddCommand(chainsUpdate)
	Chains.AddCommand(chainsRemove)
	Chains.AddCommand(chainsGraduate)
	Chains.AddCommand(chainsCat)
	addChainsFlags()
}

// Chains Sub-sub-Commands
var chainsNew = &cobra.Command{
	Use:   "new [name]",
	Short: "Hashes a new blockchain.",
	Long: `Hashes a new blockchain.

Will use a default genesis.json unless a --genesis flag is passed.
Still a WIP.`,
	Run: func(cmd *cobra.Command, args []string) {
		NewChain(cmd, args)
	},
}

var chainsInstall = &cobra.Command{
	Use:   "install [chainID]",
	Short: "Install a blockchain.",
	Long: `Install a blockchain.

Install an existing erisdb based blockchain for use locally.

Still a WIP.`,
	Run: func(cmd *cobra.Command, args []string) {
		InstallChain(cmd, args)
	},
}

var chainsListKnown = &cobra.Command{
	Use:   "known",
	Short: "List all the blockchains Eris knows about.",
	Long: `Lists the blockchains which Eris has installed for you.

To hash a new blockchain, use:

eris chains new

To install and fetch a blockchain from a chain definition
file, use:

eris chains install

Services include all other chain types supported by the
Eris platform.

Services are handled using the [eris services] command.`,
	Run: func(cmd *cobra.Command, args []string) {
		ListKnownChains()
	},
}

var chainsImport = &cobra.Command{
	Use:   "import [name] [location]",
	Short: "Import a chain definition file from Github or IPFS.",
	Long: `Import a chain definition for your platform.

By default, Eris will import from ipfs.

To list known chains use: [eris chains known].`,
	Example: "  eris chains import 2gather ipfs:QmNUhPtuD9VtntybNqLgTTevUmgqs13eMvo2fkCwLLx5MX",
	Run: func(cmd *cobra.Command, args []string) {
		ImportChain(cmd, args)
	},
}

var chainsList = &cobra.Command{
	Use:   "ls",
	Short: "Lists all known blockchains in the Eris tree.",
	Long: `Lists all known blockchains in the Eris tree.

To list the known chains: [eris chains known]
To list the running chains: [eris chains ps]
To start a chain use: [eris chains start chainName].
`,
	Run: func(cmd *cobra.Command, args []string) {
		ListChains()
	},
}

var chainsEdit = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit a blockchain.",
	Long: `Edit a blockchain definition file.


Edit will utilize your default editor.
`,
	Run: func(cmd *cobra.Command, args []string) {
		EditChain(cmd, args)
	},
}

var chainsStart = &cobra.Command{
	Use:   "start",
	Short: "Start a blockchain.",
	Long: `Start a blockchain.

[eris chains start name] by default will put the chain into the
background so its logs will not be viewable from the command line.

To stop the chain use:      [eris chains stop chainName].
To view a chain's logs use: [eris chains logs chainName].
`,
	Run: func(cmd *cobra.Command, args []string) {
		StartChain(cmd, args)
	},
}

var chainsLogs = &cobra.Command{
	Use:   "logs",
	Short: "Display the logs of a blockchain.",
	Long:  `Display the logs of a blockchain.`,
	Run: func(cmd *cobra.Command, args []string) {
		LogChain(cmd, args)
	},
}

var chainsExec = &cobra.Command{
	Use:   "exec [serviceName]",
	Short: "Run a command or interactive shell",
	Long: `Run a command or interactive shell in a container
with volumes-from the data container`,
	Run: func(cmd *cobra.Command, args []string) {
		ExecChain(cmd, args)
	},
}

var chainsListRunning = &cobra.Command{
	Use:   "ps",
	Short: "List the running blockchains.",
	Long:  `List the running blockchains.`,
	Run: func(cmd *cobra.Command, args []string) {
		ListRunningChains()
	},
}

var chainsStop = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop a running blockchain.",
	Long:  `Stop a running blockchain.`,
	Run: func(cmd *cobra.Command, args []string) {
		KillChain(cmd, args)
	},
}

var chainsInspect = &cobra.Command{
	Use:   "inspect [chainName] [key]",
	Short: "Machine readable chain operation details.",
	Long: `Displays machine readable details about running containers.

Information available to the inspect command is provided by the
Docker API. For more information about return values,
see: https://github.com/fsouza/go-dockerclient/blob/master/container.go#L235`,
	Example: `  eris chains inspect 2gather -> will display the entire information about 2gather containers
  eris chains inspect 2gather name -> will display the name in machine readable format
  eris chains inspect 2gather host_config.binds -> will display only that value`,
	Run: func(cmd *cobra.Command, args []string) {
		InspectChain(cmd, args)
	},
}

var chainsExport = &cobra.Command{
	Use:   "export [chainName]",
	Short: "Export a chain definition file to IPFS.",
	Long: `Export a chain definition file to IPFS.

Command will return a machine readable version of the IPFS hash
`,
	Run: func(cmd *cobra.Command, args []string) {
		ExportChain(cmd, args)
	},
}

var chainsRename = &cobra.Command{
	Use:   "rename [old] [new]",
	Short: "Rename a blockchain.",
	Long:  `Rename a blockchain.`,
	Run: func(cmd *cobra.Command, args []string) {
		RenameChain(cmd, args)
	},
}

var chainsRemove = &cobra.Command{
	Use:   "rm [name]",
	Short: "Removes an installed chain.",
	Long: `Removes an installed chain.

Command will remove the chain's container but will not
remove the chain definition file.

Use the --force flag to also remove the chain definition file.`,
	Run: func(cmd *cobra.Command, args []string) {
		RmChain(cmd, args)
	},
}

var chainsUpdate = &cobra.Command{
	Use:   "update [name]",
	Short: "Updates an installed chain.",
	Long: `Updates an installed chain, or installs it if it has not been installed.

Functionally this command will perform the following sequence:

1. Stop the chain (if it is running)
2. Remove the container which ran the chain
3. Pull the image the container uses from a hub
4. Rebuild the container from the updated image
5. Restart the chain (if it was previously running)

**NOTE**: If the chain uses data containers those will not be affected
by the update command.`,
	Run: func(cmd *cobra.Command, args []string) {
		UpdateChain(cmd, args)
	},
}

var chainsGraduate = &cobra.Command{
	Use:   "graduate",
	Short: "Graduates a chain to a service.",
	Long:  `Graduates a chain to a service by laying a service definition file with the chain_id`,
	Run: func(cmd *cobra.Command, args []string) {
		GraduateChain(cmd, args)
	},
}

var chainsCat = &cobra.Command{
	Use:   "cat [name]",
	Short: "Displays service file.",
	Long: `Displays service file.

Command will cat local service definition file.`,
	Run: func(cmd *cobra.Command, args []string) {
		CatChain(cmd, args)
	},
}

//----------------------------------------------------------------------

func addChainsFlags() {
	chainsNew.PersistentFlags().StringVarP(&do.GenesisFile, "genesis", "g", "", "genesis.json file")
	chainsNew.PersistentFlags().StringVarP(&do.ConfigFile, "config", "c", "", "main config file for the chain")
	chainsNew.PersistentFlags().StringVarP(&do.DirToCopy, "dir", "", "", "a directory whose contents should be copied into the chain's main dir")
	chainsNew.PersistentFlags().BoolVarP(&do.Run, "run", "r", false, "run the chain after creating")

	chainsStart.PersistentFlags().BoolVarP(&do.PublishAllPorts, "publish", "p", false, "publish all ports")

	chainsInstall.PersistentFlags().StringVarP(&do.ConfigFile, "config", "c", "", "main config file for the chain")
	chainsInstall.PersistentFlags().StringVarP(&do.DirToCopy, "dir", "", "", "a directory whose contents should be copied into the chain's main dir")
	chainsInstall.PersistentFlags().StringVarP(&do.ChainID, "id", "", "", "id of the chain to fetch")
	chainsInstall.PersistentFlags().BoolVarP(&do.Operations.PublishAllPorts, "publish", "p", false, "publish all ports")

	chainsLogs.Flags().BoolVarP(&do.Follow, "follow", "f", false, "follow logs, like tail -f")
	chainsLogs.Flags().StringVarP(&do.Tail, "tail", "t", "all", "number of lines to show from end of logs")

	chainsRemove.Flags().BoolVarP(&do.File, "file", "f", false, "remove chain definition file as well as chain container")
	chainsRemove.Flags().BoolVarP(&do.RmD, "data", "x", false, "remove data containers also")

	chainsExec.Flags().BoolVarP(&do.Interactive, "interactive", "i", false, "interactive shell")

	chainsUpdate.Flags().BoolVarP(&do.SkipPull, "pull", "p", true, "pull an updated version of the chain's base service image from docker hub")

	chainsStop.Flags().BoolVarP(&do.Rm, "rm", "r", false, "remove containers after stopping")
	chainsStop.Flags().BoolVarP(&do.RmD, "data", "x", false, "remove data containers after stopping")
	chainsStop.Flags().BoolVarP(&do.Force, "force", "f", false, "kill the container instantly without waiting to exit")
	chainsStop.Flags().UintVarP(&do.Timeout, "timeout", "t", 10, "manually set the timeout; overridden by --force")

	chainsList.Flags().BoolVarP(&do.Quiet, "quiet", "q", false, "machine parsable output")
	chainsListRunning.Flags().BoolVarP(&do.Quiet, "quiet", "q", false, "machine parsable output")
}

//----------------------------------------------------------------------
// cli command wrappers

func StartChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Name = args[0]
	IfExit(chns.StartChainRaw(do))
}

func LogChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Name = args[0]
	IfExit(chns.LogsChainRaw(do))
}

func ExecChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}

	do.Name = args[0]
	// if interactive, we ignore args. if not, run args as command
	if !do.Interactive {
		if len(args) < 2 {
			Exit(fmt.Errorf("Non-interactive exec sessions must provide arguments to execute"))
		}
		args = args[1:]
	}
	if len(args) == 1 {
		args = strings.Split(args[0], " ")
	}
	do.Args = args
	IfExit(chns.ExecChainRaw(do))
}

func KillChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Name = args[0]
	IfExit(chns.KillChainRaw(do))
}

// fetch and install a chain
func InstallChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	// the idea here is you will either specify a chainName as the arg and that will
	// double as the chainID, or you want a local reference name for the chain, so you specify
	// the chainID with a flag and give your local reference name as the arg
	do.Name = args[0]
	IfExit(chns.InstallChainRaw(do))
}

// create a new chain
// genesis is either given or a simple single-validator genesis will be laid for you
func NewChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Name = args[0]
	IfExit(chns.NewChainRaw(do))
}

// import a chain definition file
func ImportChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	if len(args) != 2 {
		cmd.Help()
		return
	}
	do.Name = args[0]
	do.Path = args[1]
	IfExit(chns.ImportChainRaw(do))
}

// edit a chain definition file
func EditChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	var configVals []string
	if len(args) > 1 {
		configVals = args[1:]
	}
	do.Name = args[0]
	do.Args = configVals
	IfExit(chns.EditChainRaw(do))
}

func InspectChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}

	do.Name = args[0]
	if len(args) == 1 {
		do.Args = []string{"all"}
	} else {
		do.Args = []string{args[1]}
	}

	chain, err := loaders.LoadChainDefinition(do.Name, do.Operations.ContainerNumber)
	IfExit(err)

	if chns.IsChainExisting(chain) {
		IfExit(srv.InspectServiceByService(chain.Service, chain.Operations, do.Args[0]))
	}
}

func ExportChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Name = args[0]
	IfExit(chns.ExportChainRaw(do))
}

func ListKnownChains() {
	if err := chns.ListKnownRaw(do); err != nil {
		return
	}

	fmt.Println(do.Result)
}

func ListChains() {
	if err := chns.ListExistingRaw(do); err != nil {
		return
	}
}

func ListRunningChains() {
	if err := chns.ListRunningRaw(do); err != nil {
		return
	}
}

func RenameChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	if len(args) != 2 {
		cmd.Help()
		return
	}
	do.Name = args[0]
	do.NewName = args[1]
	IfExit(chns.RenameChainRaw(do))
}

func UpdateChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Name = args[0]
	IfExit(chns.UpdateChainRaw(do))
}

func RmChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Name = args[0]
	IfExit(chns.RmChainRaw(do))
}

func GraduateChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Name = args[0]
	IfExit(chns.GraduateChainRaw(do))
}

func CatChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Name = args[0]
	IfExit(chns.CatChainRaw(do))
}

func checkChainGiven(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Please provide a chain")
	}
	return nil
}
