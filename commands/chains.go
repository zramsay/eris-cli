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
	Chains.AddCommand(chainsExport)
	Chains.AddCommand(chainsStop)
	Chains.AddCommand(chainsRename)
	Chains.AddCommand(chainsUpdate)
	Chains.AddCommand(chainsRemove)
	addChainsFlags()
}

func addChainsFlags() {
	chainsNew.PersistentFlags().StringVarP(&ChainType, "type", "t", "", "type of chain (service definition file to read from)")
	chainsNew.PersistentFlags().StringVarP(&ChainName, "name", "n", "", "name the chain for future reference")
	chainsNew.PersistentFlags().StringVarP(&GenesisFile, "genesis", "g", "", "genesis.json file")
	chainsNew.PersistentFlags().StringVarP(&ConfigFile, "config", "c", "", "main config file for the chain")
	chainsNew.PersistentFlags().StringVarP(&DirToCopy, "dir", "", "", "a directory whose contents should be copied into the chain's main dir")

	chainsInstall.PersistentFlags().StringVarP(&ChainType, "type", "t", "", "type of chain (service definition file to read from)")
	chainsInstall.PersistentFlags().StringVarP(&ChainName, "name", "n", "", "name the chain for future reference")
	chainsInstall.PersistentFlags().StringVarP(&ConfigFile, "config", "c", "", "main config file for the chain")
	chainsInstall.PersistentFlags().StringVarP(&DirToCopy, "dir", "", "", "a directory whose contents should be copied into the chain's main dir")

	chainsRemove.Flags().BoolVarP(&Force, "force", "f", false, "force action")

	chainsExec.Flags().BoolVarP(&Interactive, "interactive", "i", false, "interactive shell")

	chainsUpdate.Flags().BoolVarP(&Pull, "pull", "p", false, "pull an updated version of the chain's image from docker hub")
}

// known lists the known chain types which eris can install
// flags to add: --list-versions
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
		chns.ListKnown()
	},
}

var chainsInstall = &cobra.Command{
	Use:   "install [chainID]",
	Short: "Install a blockchain.",
	Long: `Install a blockchain.

Still a WIP.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Install(cmd, args)
	},
}

// flags to add: --type, --genesis, --config, --checkout, --force-name
var (
	ChainType   string
	ChainName   string
	GenesisFile string
	ConfigFile  string
	DirToCopy   string
)
var chainsNew = &cobra.Command{
	Use:   "new [name]",
	Short: "Hashes a new blockchain.",
	Long: `Hashes a new blockchain.

Will use a default genesis.json unless a --genesis flag is passed.
Still a WIP.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.New(cmd, args)
	},
}

var chainsImport = &cobra.Command{
	Use:   "import [name] [location]",
	Short: "Import a chain definition file from Github or IPFS.",
	Long: `Import a chain definition for your platform.

By default, Eris will import from ipfs.

To list known chains use: [eris chains known].`,
	Example: "  eris chains 2gather ipfs:QmNUhPtuD9VtntybNqLgTTevUmgqs13eMvo2fkCwLLx5MX",
	Run: func(cmd *cobra.Command, args []string) {
		chns.Import(cmd, args)
	},
}

// flags to add: --current, --short, --all
var chainsList = &cobra.Command{
	Use:   "ls",
	Short: "Lists all known blockchains in the Eris tree.",
	Long: `Lists all known blockchains in the Eris tree.

To list the known chains: [eris chains known]
To list the running chains: [eris chains ps]
To start a chain use: [eris chains start chainName].
`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.ListChains()
	},
}

var chainsEdit = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit a blockchain.",
	Long: `Edit a blockchain definition file.


Edit will utilize your default editor.
`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Edit(cmd, args)
	},
}

// flags to add: --commit, --multi, --foreground, --config, --chain
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
		chns.Start(cmd, args)
	},
}

// flags to add: --tail
var chainsLogs = &cobra.Command{
	Use:   "logs",
	Short: "Display the logs of a blockchain.",
	Long:  `Display the logs of a blockchain.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Logs(cmd, args)
	},
}

var chainsExec = &cobra.Command{
	Use:   "exec [serviceName]",
	Short: "Run a command or interactive shell",
	Long:  "Run a command or interactive shell in a container with volumes-from the data container",
	Run: func(cmd *cobra.Command, args []string) {
		chns.Exec(cmd, args)
	},
}

var chainsListRunning = &cobra.Command{
	Use:   "ps",
	Short: "List the running blockchains.",
	Long:  `List the running blockchains.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.ListRunning()
	},
}

// stop
var chainsStop = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop a running blockchain.",
	Long:  `Stop a running blockchain.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Kill(cmd, args)
	},
}

// inspect running containers
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
		chns.Inspect(cmd, args)
	},
}

// export running containers
var chainsExport = &cobra.Command{
	Use:   "export [chainName]",
	Short: "Export a chain definition file to IPFS.",
	Long: `Export a chain definition file to IPFS.

Command will return a machine readable version of the IPFS hash
`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Export(cmd, args)
	},
}

var chainsRename = &cobra.Command{
	Use:   "rename [old] [new]",
	Short: "Rename a blockchain.",
	Long:  `Rename a blockchain.`,
	Run: func(cmd *cobra.Command, args []string) {
		chns.Rename(cmd, args)
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
		chns.Rm(cmd, args)
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
		chns.Update(cmd, args)
	},
}
