package commands

import (
	"fmt"
	"strings"

	chns "github.com/eris-ltd/eris-cli/chains"
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

The chains subcommand is used to start, stop, and configure blockchains.
Within the Eris platform, blockchains are the primary method of storing
structured data which is used by the Eris platform in combination with
IPFS (a globally-accessible content-addressable peer to peer file
storage solution).`,
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

Still a WIP.`,
	Run: func(cmd *cobra.Command, args []string) {
		InstallChain(cmd, args)
	},
}

var chainsListKnown = &cobra.Command{
	Use:   "known",
	Short: "List all the blockchains Eris knows about.",
	Long: `Lists the blockchains which Eris has installed for you.

To has a new blockchain from a chain definition file, use: [eris chains new].
To install a new blockchain from a chain definition file, use: [eris chains install].
To install a new chain definition file, use: [eris chains import].

Services include all executable chains supported by the Eris platform which are
NOT blockchains or key managers.

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
	Long:  "Run a command or interactive shell in a container with volumes-from the data container",
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

//----------------------------------------------------------------------
// cli flags

var (
	ChainName   string
	GenesisFile string
	ConfigFile  string
	DirToCopy   string
	ChainID     string
)

func addChainsFlags() {
	chainsNew.PersistentFlags().StringVarP(&GenesisFile, "genesis", "g", "", "genesis.json file")
	chainsNew.PersistentFlags().StringVarP(&ConfigFile, "config", "c", "", "main config file for the chain")
	chainsNew.PersistentFlags().StringVarP(&DirToCopy, "dir", "", "", "a directory whose contents should be copied into the chain's main dir")

	chainsInstall.PersistentFlags().StringVarP(&ConfigFile, "config", "c", "", "main config file for the chain")
	chainsInstall.PersistentFlags().StringVarP(&DirToCopy, "dir", "", "", "a directory whose contents should be copied into the chain's main dir")
	chainsInstall.PersistentFlags().StringVarP(&ChainID, "id", "", "", "id of the chain to fetch")

	chainsLogs.Flags().BoolVarP(&Follow, "tail", "t", false, "follow logs, like tail -f")

	chainsRemove.Flags().BoolVarP(&Force, "file", "f", false, "remove chain definition file as well as chain container")
	chainsRemove.Flags().BoolVarP(&RmD, "data", "x", false, "remove data containers also")

	chainsExec.Flags().BoolVarP(&Interactive, "interactive", "i", false, "interactive shell")

	chainsUpdate.Flags().BoolVarP(&Pull, "pull", "p", false, "pull an updated version of the chain's base service image from docker hub")

	chainsStop.Flags().BoolVarP(&Rm, "rm", "r", false, "remove containers after stopping")
	chainsStop.Flags().BoolVarP(&RmD, "data", "x", false, "remove data containers after stopping")
}

//----------------------------------------------------------------------
// cli command wrappers

func StartChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	IfExit(chns.StartChainRaw(args[0], ContainerNumber))
}

func LogChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	IfExit(chns.LogsChainRaw(args[0], Follow, ContainerNumber))
}

func ExecChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}

	srv := args[0]
	// if interactive, we ignore args. if not, run args as command
	interactive := cmd.Flags().Lookup("interactive").Changed
	if !interactive {
		if len(args) < 2 {
			Exit(fmt.Errorf("Non-interactive exec sessions must provide arguments to execute"))
		}
		args = args[1:]
	}
	if len(args) == 1 {
		args = strings.Split(args[0], " ")
	}

	IfExit(chns.ExecChainRaw(srv, args, interactive, ContainerNumber))
}

func KillChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	IfExit(chns.KillChainRaw(args[0], Rm, RmD, ContainerNumber))
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
	chainName := args[0]
	IfExit(chns.InstallChainRaw(ChainID, chainName, ConfigFile, DirToCopy, ContainerNumber))
}

// create a new chain
// genesis is either given or a simple single-validator genesis will be laid for you
func NewChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	chainName := args[0]
	IfExit(chns.NewChainRaw(chainName, GenesisFile, ConfigFile, DirToCopy, ContainerNumber))
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
	IfExit(chns.ImportChainRaw(args[0], args[1]))
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
	IfExit(chns.EditChainRaw(args[0], configVals))
}

func InspectChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	chainName := args[0]
	var field string
	if len(args) == 1 {
		field = "all"
	} else {
		field = args[1]
	}
	chain, err := chns.LoadChainDefinition(chainName, ContainerNumber)
	IfExit(err)
	if chns.IsChainExisting(chain) {
		IfExit(srv.InspectServiceByService(chain.Service, chain.Operations, field))
	}
}

func ExportChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	IfExit(chns.ExportChainRaw(args[0]))
}

func ListKnownChains() {
	chains := chns.ListKnownRaw()
	for _, s := range chains {
		fmt.Println(s)
	}
}

func ListInstalledChains() {
	chns.ListExistingRaw()
}

func ListChains() {
	chains := chns.ListExistingRaw()
	for _, s := range chains {
		fmt.Println(s)
	}
}

func ListRunningChains() {
	chains := chns.ListRunningRaw()
	for _, s := range chains {
		fmt.Println(s)
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
	IfExit(chns.RenameChainRaw(args[0], args[1]))
}

func UpdateChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	IfExit(chns.UpdateChainRaw(args[0], Pull, ContainerNumber))
}

func RmChain(cmd *cobra.Command, args []string) {
	if err := checkChainGiven(args); err != nil {
		cmd.Help()
		return
	}
	IfExit(chns.RmChainRaw(args[0], RmD, Force, ContainerNumber))
}

func checkChainGiven(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Please provide a chain")
	}
	return nil
}
