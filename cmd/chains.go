package commands

import (
	"fmt"
	"os"
	"strings"

	chns "github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/list"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//----------------------------------------------------------------------
// cli definitions

// Primary Chains Sub-Command
var Chains = &cobra.Command{
	Use:   "chains",
	Short: "Start, stop, and manage blockchains.",
	Long: `Start, stop, and manage blockchains.

The chains subcommand is used to work on erisdb smart contract
blockchain networks. The name is not perfect, as eris is able
to operate a wide variety of blockchains out of the box. Most
of those existing blockchains should be ran via the [eris services ...]
commands. As they fall under the rubric of "things I just want
to turn on or off". While you can develop against those
blockchains, you generally aren't developing those blockchains
themselves. [eris chains ...] commands are built to help you build
blockchains. It is our opinionated gateway to the wonderful world
of permissioned smart contract networks.

Your own blockchain/smart contract machine is just an [eris chains new]
away!`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// Build the chains subcommand
func buildChainsCommand() {
	Chains.AddCommand(chainsMake)
	Chains.AddCommand(chainsNew)
	Chains.AddCommand(chainsImport)
	Chains.AddCommand(chainsListAll)
	Chains.AddCommand(chainsCheckout)
	Chains.AddCommand(chainsHead)
	Chains.AddCommand(chainsPorts)
	Chains.AddCommand(chainsEdit)
	Chains.AddCommand(chainsStart)
	Chains.AddCommand(chainsLogs)
	Chains.AddCommand(chainsInspect)
	Chains.AddCommand(chainsStop)
	Chains.AddCommand(chainsExec)
	Chains.AddCommand(chainsCat)
	Chains.AddCommand(chainsExport)
	Chains.AddCommand(chainsRename)
	Chains.AddCommand(chainsUpdate)
	Chains.AddCommand(chainsRemove)
	Chains.AddCommand(chainsGraduate)
	// Chains.AddCommand(chainsMakeGenesis)
	addChainsFlags()
}

// Chains Sub-sub-Commands
var chainsMake = &cobra.Command{
	Use:   "make NAME",
	Short: "Create keys and a genesis block for your chain.",
	Long: `Create the new required files for your chain.

Make is an opinionated gateway to the basic types of chains which most eris users
will make most of the time. Make is a command line wizard in which you will let
the marmots know how you would like your genesis created.

Make can also be used with a variety of flags for fast chain making.

When using make with the --known flag the marmots will *not* create keys for you
and will instead assume that the keys exist somewhere. When using make with the
wizard (no flags) or when using with the other flags then keys will be made along
with the genesis.jsons and priv_validator.jsons so that everything is ready to go
for you to [eris chains new].

Optionally chains make provides packages of outputed priv_validator and genesis.json
which you can email or send on your slack to your coworkers. These packages can
be tarballs or zip files, and **they will contain the private keys** so please
be aware of that.

The make process will *not* start a chain for you. You will want to use

[eris chains new chainName --dir chainName]

for that which will import all of the files which make creates into containers and
start your shiny new chain.

If you have any questions on eris chains make, please see the eris-cm (chain manager)
documentation here:
https://docs.erisindustries.com/documentation/eris-cm/latest/eris-cm/
`,
	Example: `$ eris chains make myChain -- will use the chain-making wizard and make your chain named myChain (interactive)
$ eris chains make myChain --chain-type=simplechain --  will use the chain type definition files to make your chain named myChain (non-interactive)
$ eris chains make myChain --account-types=Root:1,Developer:0,Validator:0,Participant:1 -- will use the flag to make your chain named myChain (non-interactive)
$ eris chains make myChain --account-types=Root:1,Developer:0,Validator:0,Participant:1 --chain-type=simplechain -- account types trump chain types, this command will use the flags to make the chain (non-interactive)
$ eris chains make myChain --known --validators /path/to/validators.csv --accounts /path/to/accounts.csv -- will use the csv file to make your chain named myChain (non-interactive) (won't make keys)
$ eris chains make myChain --tar -- will create the chain and save each of the "bundles" as tarballs which can be used by colleagues to start their chains
`,
	Run: MakeChain,
}

var chainsNew = &cobra.Command{
	Use:   "new NAME",
	Short: "Create and start a new blockhain.",
	Long: `Create and start a new blockchain.

The creation process will both create a blockchain on the current machine
as well as start running that chain.

If you need to update a chain after creation, you can update any of the
appropriate settings in the chains definition file for the named chain
(which will be located at ~/.eris/chains/NAME.toml) and then
utilize [eris chains update NAME -p] to update the blockchain appropriately
(using the -p flag will force eris not to pull the most recent docker image
for eris:db).

Will use a default genesis.json from ~/.eris/chains/default/genesis.json
unless a --genesis or --dir flag is passed.

Will use a default config.toml from ~/.eris/chains/default/config.toml
unless the --options or --dir flag is passed.

Will use a default eris:db server config from ~/.eris/chains/default/server_conf.toml
unless the --serverconf or --dir flag is passed.

If you would like to create a genesis.json then please utilize [eris chains make]`,
	Run: NewChain,
}

var chainsRegister = &cobra.Command{
	Use:   "register NAME",
	Short: "Register a blockchain on etcb (a blockchain for registering other blockchains).",
	Long: `Register a blockchain on etcb.

etcb is Eris's blockchain which is a public blockchain that can be used to
register *other* blockchains. In other words it is an easy way to "share"
your blockchains with others. [eris chains register] is made to work
seemlessly with [eris chains install] so that other users and/or colleagues
should be able to use your registered blockchain by simply using the install
command.

The [eris chains register] command is not the *only* way to
share your blockchains. You can also export your chain definition file and
genesis.json to IPFS, and share the hash of the chain definition file and
genesis.json with any colleagues or users who need to be able to connect
into the blockchain.`,
	Run: RegisterChain,
}

var chainsInstall = &cobra.Command{
	Use:   "install NAME",
	Short: "Install a blockchain from the etcb registry.",
	Long: `Install a blockchain from the etcb registry.

Install an existing erisdb based blockchain for use locally.

(Currently a work in progress.)`,
	Run: InstallChain,
}

var chainsListAll = &cobra.Command{
	Use:   "ls",
	Short: "Lists everything chain related.",
	Long: `Lists all: chain definition files (--known), current existing
containers for each chain (--existing), current running containers for each
chain (--running).

If no known chains exist yet, create a new blockchain with: [eris chains new NAME]
command.

Services are handled using the [eris services] command.`,
	Run: ListAllChains,
}

var chainsImport = &cobra.Command{
	Use:   "import NAME LOCATION",
	Short: "Import a chain definition file from Github or IPFS.",
	Long: `Import a chain definition for your platform.

By default, Eris will import from IPFS.

To list known chains use: [eris chains ls --known].`,
	Example: "$ eris chains import 2gather QmNUhPtuD9VtntybNqLgTTevUmgqs13eMvo2fkCwLLx5MX",
	Run:     ImportChain,
}

var chainsCheckout = &cobra.Command{
	Use:   "checkout [NAME]",
	Short: "Check out a chain.",
	Long: `Check out a chain.

Checkout is a convenience feature. For any Eris command which accepts a
--chain or $chain variable, the checked out chain can replace manually
passing in a --chain flag. If a --chain is passed to any command accepting
--chain, the --chain which is passed will overwrite any checked out chain.

If command is given without arguments it will clear the head and there will
be no chain checked out.`,
	Run: CheckoutChain,
}

var chainsPorts = &cobra.Command{
	Use:   "ports NAME [PORT]...",
	Short: "Print port mappings.",
	Long: `Print port mappings.

The [eris chains ports] command is mostly a developer
convenience function. It returns a machine readable
port mapping of a port which is exposed inside the
container to what that port is mapped to on the host.

This is useful when stitching together chain networks which
need to know how to connect into a specific chain (perhaps
with or without a container number) container.`,
	Example: `$ eris chains ports myChain 1337 -- will display what port on the host is mapped to the eris:db API port
$ eris chains ports myChain 46656 -- will display what port on the host is mapped to the eris:db peer port
$ eris chains ports myChain 46657 -- will display what port on the host is mapped to the eris:db rpc port
$ eris chains ports myChain -- will display all mappings`,
	Run: PortsChain,
}

var chainsHead = &cobra.Command{
	Use:   "current",
	Short: "The currently checked out chain.",
	Long: `Displays the name of the currently checked out chain.

To checkout a new chain use [eris chains checkout NAME].

To "uncheckout" a chain use [eris chains checkout] without arguments.`,
	Run: CurrentChain,
}

var chainsEdit = &cobra.Command{
	Use:   "edit NAME",
	Short: "Edit a blockchain.",
	Long: `Edit a blockchain definition file.

Edit will utilize the default editor set for your current shell
or if none is set, it will use *vim*. Sorry for the bias Emacs
users, but we had to pick one and more marmots are known vim
users. Emacs users can set their EDITOR variable and eris
will default to that if you wise.`,
	Run: EditChain,
}

var chainsStart = &cobra.Command{
	Use:   "start",
	Short: "Start a blockchain.",
	Long: `Start running a blockchain.

[eris chains start NAME] by default will put the chain into the
background so its logs will not be viewable from the command line.

To stop the chain use:      [eris chains stop NAME].
To view a chain's logs use: [eris chains logs NAME].`,
	Run: StartChain,
}

var chainsLogs = &cobra.Command{
	Use:   "logs NAME",
	Short: "Display the logs of a blockchain.",
	Long:  `Display the logs of a blockchain.`,
	Run:   LogChain,
}

var chainsExec = &cobra.Command{
	Use:   "exec NAME",
	Short: "Run a command or interactive shell",
	Long: `Run a command or interactive shell in a container
with volumes-from the data container`,
	Run: ExecChain,
}

var chainsStop = &cobra.Command{
	Use:   "stop NAME",
	Short: "Stop a running blockchain.",
	Long:  `Stop a running blockchain.`,
	Run:   KillChain,
}

var chainsInspect = &cobra.Command{
	Use:   "inspect NAME [KEY]",
	Short: "Machine readable chain operation details.",
	Long: `Display machine readable details about running containers.

Information available to the inspect command is provided by the
Docker API. For more information about return values,
see: https://github.com/fsouza/go-dockerclient/blob/master/container.go#L235`,
	Example: `$ eris chains inspect 2gather -- will display the entire information about 2gather containers
$ eris chains inspect 2gather name -- will display the name in machine readable format
$ eris chains inspect 2gather host_config.binds -- will display only that value`,
	Run: InspectChain,
}

var chainsExport = &cobra.Command{
	Use:   "export NAME",
	Short: "Export a chain definition file to IPFS.",
	Long: `Export a chain definition file to IPFS.

Command will return a machine readable version of the IPFS hash.`,
	Run: ExportChain,
}

var chainsRename = &cobra.Command{
	Use:   "rename OLD_NAME NEW_NAME",
	Short: "Rename a blockchain.",
	Long:  `Rename a blockchain.`,
	Run:   RenameChain,
}

var chainsRemove = &cobra.Command{
	Use:   "rm NAME",
	Short: "Remove an installed chain.",
	Long: `Remove an installed chain.

Command will remove the chain's container but will not
remove the chain definition file.`,
	Run: RmChain,
}

var chainsUpdate = &cobra.Command{
	Use:   "update NAME",
	Short: "Update an installed chain.",
	Long: `Update an installed chain, or install it if it has not been installed.

Functionally this command will perform the following sequence:

1. Stop the chain (if it is running).
2. Remove the container which ran the chain.
3. Pull the image the container uses from a hub.
4. Rebuild the container from the updated image.
5. Restart the chain (if it was previously running).

NOTE: If the chain uses data containers those will not be affected
by the update command.
`,
	Run: UpdateChain,
}

var chainsGraduate = &cobra.Command{
	Use:   "graduate NAME",
	Short: "Graduate a chain to a service.",
	Long: `Graduate a chain to a service.

Graduate works by translating the chain's definition into a service definition
file with the chain_id set as the service name and everything set for you to
more simply turn the chain on or off.

Graduate should be used whenever you are "finished" working "on" the chain and
you feel the chain is stable. While chains work just fine by turning them "on"
or "off" with [eris chains start] and [eris chains stop], some feel that it is
easier to work with chains as a service rather than as a chain when they are
stable and not longer need to be worked "on" which is why this functionality
exists. Ultimately, graduate is a convenience function as there is little to
no difference in how chains and services "run", however the [eris chains]
functions have more convenience functions for working "on" chains themselves.`,
	Run: GraduateChain,
}

var chainsCat = &cobra.Command{
	Use:     "cat NAME [config|genesis]",
	Short:   "Display chain information.",
	Long:    `Display chain information.`,
	Aliases: []string{"plop"},
	Example: `$ eris chains cat simplechain -- will display the chain definition file
$ eris chains cat simplechain config -- will display the config.toml file from inside the container
$ eris chains cat simplechain genesis -- will display the genesis.json file from the container`,
	Run: CatChain,
}

var chainsMakeGenesis = &cobra.Command{
	Use:   "make-genesis NAME KEY",
	Short: "Generates a genesis file.",
	Long: `Generates a genesis file with chainNAME and a single pubkey.

Command is equivalent to: [eris chains exec someChain "mintgen known NAME KEY"]

but does not require a pre-existing chain to execute.

see https://github.com/eris-ltd/mint-client for more info`,
	Run: MakeGenesisFile,
}

//----------------------------------------------------------------------

func addChainsFlags() {
	chainsMake.PersistentFlags().StringSliceVarP(&do.AccountTypes, "account-types", "", []string{}, "what number of account types should we use? find these in ~/.eris/chains/account_types; incompatible with and overrides chain-type")
	chainsMake.PersistentFlags().StringVarP(&do.ChainType, "chain-type", "", "", "which chain type definition should we use? find these in ~/.eris/chains/chain_types")
	chainsMake.PersistentFlags().BoolVarP(&do.Tarball, "tar", "", false, "instead of making directories in ~/.eris/chains, make tarballs; incompatible with and overrides zip")
	chainsMake.PersistentFlags().BoolVarP(&do.ZipFile, "zip", "", false, "instead of making directories in ~/.eris/chains, make zip files")
	chainsMake.PersistentFlags().BoolVarP(&do.Output, "output", "", true, "should eris-cm provide an output of its job")
	chainsMake.PersistentFlags().BoolVarP(&do.Known, "known", "", false, "use csv for a set of known keys to assemble genesis.json (requires both --accounts and --validators flags")
	chainsMake.PersistentFlags().StringVarP(&do.ChainMakeActs, "accounts", "", "", "comma separated list of the accounts.csv files you would like to utilize (requires --known flag)")
	chainsMake.PersistentFlags().StringVarP(&do.ChainMakeVals, "validators", "", "", "comma separated list of the validators.csv files you would like to utilize (requires --known flag)")
	buildFlag(chainsMake, do, "data", "chain-make")

	// [csk]: the flags below are commented out to reduce the complexity of this command. chains new should just use dirs....
	// buildFlag(chainsNew, do, "config", "chain")
	// buildFlag(chainsNew, do, "csv", "chain")
	// buildFlag(chainsNew, do, "serverconf", "chain")
	// chainsNew.PersistentFlags().StringVarP(&do.GenesisFile, "genesis", "g", "", "genesis.json file")
	// chainsNew.PersistentFlags().StringSliceVarP(&do.ConfigOpts, "options", "", nil, "space separated <key>=<value> pairs to set in config.toml")
	// chainsNew.PersistentFlags().StringVarP(&do.Priv, "priv", "", "", "pass in a priv_validator.json file (dev-only!)")
	// chainsNew.PersistentFlags().UintVarP(&do.N, "N", "", 1, "make a new genesis.json with this many validators and create data containers for each")
	buildFlag(chainsNew, do, "dir", "chain")
	buildFlag(chainsNew, do, "env", "chain")
	buildFlag(chainsNew, do, "publish", "chain")
	buildFlag(chainsNew, do, "links", "chain")
	buildFlag(chainsNew, do, "api", "chain")
	chainsNew.PersistentFlags().BoolVarP(&do.Force, "force", "f", true, "overwrite data in  ~/.eris/data/chainName")
	chainsNew.PersistentFlags().BoolVarP(&do.Logsrotate, "logsrotate", "z", false, "turn on logsrotate as a dependency to handle long output")

	// buildFlag(chainsRegister, do, "links", "chain")
	// buildFlag(chainsRegister, do, "env", "chain")
	// chainsRegister.PersistentFlags().StringVarP(&do.Pubkey, "pub", "p", "", "pubkey to use for registering the chain in etcb")
	// chainsRegister.PersistentFlags().StringVarP(&do.Gateway, "etcb-host", "", "interblock.io:46657", "set the address of the etcb chain")
	// chainsRegister.PersistentFlags().StringVarP(&do.ChainID, "etcb-chain", "", "etcb_testnet", "set the chain id of the etcb chain")

	// buildFlag(chainsInstall, do, "publish", "chain")
	// buildFlag(chainsInstall, do, "links", "chain")
	// buildFlag(chainsInstall, do, "env", "chain")
	// buildFlag(chainsInstall, do, "config", "chain")
	// buildFlag(chainsInstall, do, "serverconf", "chain")
	// buildFlag(chainsInstall, do, "dir", "chain")
	// chainsInstall.PersistentFlags().StringVarP(&do.ChainID, "id", "", "", "id of the chain to fetch")
	// chainsInstall.PersistentFlags().StringVarP(&do.Gateway, "etcb-host", "", "interblock.io:46657", "set the address of the etcb chain")
	// chainsInstall.PersistentFlags().IntVarP(&do.Operations.ContainerNumber, "N", "N", 1, "container number")

	buildFlag(chainsStart, do, "publish", "chain")
	buildFlag(chainsStart, do, "env", "chain")
	buildFlag(chainsStart, do, "links", "chain")
	buildFlag(chainsStart, do, "api", "chain")
	chainsStart.PersistentFlags().BoolVarP(&do.Logsrotate, "logsrotate", "z", false, "turn on logsrotate as a dependency to handle long output")

	buildFlag(chainsLogs, do, "follow", "chain")
	buildFlag(chainsLogs, do, "tail", "chain")

	buildFlag(chainsExec, do, "publish", "chain")
	buildFlag(chainsExec, do, "interactive", "chain")
	buildFlag(chainsExec, do, "links", "chain")
	chainsExec.Flags().StringVarP(&do.Image, "image", "", "", "Docker image")

	buildFlag(chainsRemove, do, "force", "chain")
	buildFlag(chainsRemove, do, "file", "chain")
	buildFlag(chainsRemove, do, "data", "chain")
	buildFlag(chainsRemove, do, "rm-volumes", "chain")

	buildFlag(chainsUpdate, do, "pull", "chain")
	buildFlag(chainsUpdate, do, "timeout", "chain")
	buildFlag(chainsUpdate, do, "env", "chain")
	buildFlag(chainsUpdate, do, "links", "chain")

	buildFlag(chainsStop, do, "rm", "chain")
	buildFlag(chainsStop, do, "data", "chain")
	buildFlag(chainsStop, do, "force", "chain")
	buildFlag(chainsStop, do, "timeout", "chain")
	buildFlag(chainsStop, do, "volumes", "chain")

	buildFlag(chainsListAll, do, "known", "chain")
	buildFlag(chainsListAll, do, "existing", "chain")
	buildFlag(chainsListAll, do, "running", "chain")
	buildFlag(chainsListAll, do, "quiet", "chain")
}

//----------------------------------------------------------------------
// cli command wrappers

func StartChain(cmd *cobra.Command, args []string) {
	// [csk]: if no args should we just start the checkedout chain?
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(chns.StartChain(do))
}

func LogChain(cmd *cobra.Command, args []string) {
	// [csk]: if no args should we just start the checkedout chain?
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(chns.LogsChain(do))
}

func ExecChain(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))

	do.Name = args[0]
	// if interactive, we ignore args. if not, run args as command
	args = args[1:]
	if !do.Operations.Interactive {
		if len(args) == 0 {
			Exit(fmt.Errorf("Non-interactive exec sessions must provide arguments to execute"))
		}
	}
	if len(args) == 1 {
		args = strings.Split(args[0], " ")
	}
	do.Operations.Args = args
	config.GlobalConfig.InteractiveWriter = os.Stdout
	config.GlobalConfig.InteractiveErrorWriter = os.Stderr
	_, err := chns.ExecChain(do)
	IfExit(err)
}

func KillChain(cmd *cobra.Command, args []string) {
	// [csk]: if no args should we just start the checkedout chain?
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(chns.KillChain(do))
}

// fetch and install a chain
//
// the idea here is you will either specify a chainName as the arg and that will
// double as the chainID, or you want a local reference name for the chain, so you specify
// the chainID with a flag and give your local reference name as the arg
func InstallChain(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(chns.InstallChain(do))
}

// make the genesis files for a chain
//
// MakeChain will assemble the necessary files for a new blockchain. it does not
// actually make the chain or start it (that happens in new), but it does create
// the predicate files for more complex chains than is typically used in default
func MakeChain(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	if do.Known && (do.ChainMakeActs == "" || do.ChainMakeVals == "") {
		cmd.Help()
		IfExit(fmt.Errorf("\nIf you are using the --known flag the --validators *and* the --accounts flags are both required."))
	}
	if len(do.AccountTypes) > 0 && do.ChainType != "" {
		cmd.Help()
		IfExit(fmt.Errorf("\nThe --account-types flag is incompatible with the --chain-type flag. Please use one or the other."))
	}
	if (len(do.AccountTypes) > 0 || do.ChainType != "") && do.Known {
		cmd.Help()
		IfExit(fmt.Errorf("\nThe --account-types and --chain-type flags are incompatible with the --known flag. Please use only one of these."))
	}
	IfExit(chns.MakeChain(do))
}

// create a new chain
//
// genesis is either given or a simple single-validator genesis will be laid for you
func NewChain(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(chns.NewChain(do))
}

// register a chain in the etcb chain registry
func RegisterChain(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "ge", cmd, args))
	do.Name = args[0]
	do.Operations.Args = args[1:]
	IfExit(chns.RegisterChain(do))
}

// import a chain definition file
func ImportChain(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "eq", cmd, args))
	do.Name = args[0]
	do.Path = args[1]
	IfExit(chns.ImportChain(do))
}

// checkout a chain
func CheckoutChain(cmd *cobra.Command, args []string) {
	if len(args) >= 1 {
		do.Name = args[0]
	} else {
		do.Name = ""
	}
	IfExit(chns.CheckoutChain(do))
}

func CurrentChain(cmd *cobra.Command, args []string) {
	IfExit(chns.CurrentChain(do))
}

func CatChain(cmd *cobra.Command, args []string) {
	// [csk]: if no args should we just start the checkedout chain?
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	do.Type = "toml"
	if len(args) > 1 {
		do.Type = args[1]
	}
	IfExit(chns.CatChain(do))
}

func PortsChain(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	do.Operations.Args = args[1:]
	IfExit(chns.PortsChain(do))
}

// edit a chain definition file
func EditChain(cmd *cobra.Command, args []string) {
	// [csk]: if no args should we just start the checkedout chain?
	IfExit(ArgCheck(1, "ge", cmd, args))
	var configVals []string
	if len(args) > 1 {
		configVals = args[1:]
	}
	do.Name = args[0]
	do.Operations.Args = configVals
	IfExit(chns.EditChain(do))
}

func InspectChain(cmd *cobra.Command, args []string) {
	// [csk]: if no args should we just start the checkedout chain?
	IfExit(ArgCheck(1, "ge", cmd, args))

	do.Name = args[0]
	if len(args) == 1 {
		do.Operations.Args = []string{"all"}
	} else {
		do.Operations.Args = []string{args[1]}
	}

	IfExit(chns.InspectChain(do))
}

func ExportChain(cmd *cobra.Command, args []string) {
	// [csk]: if no args should we just start the checkedout chain?
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(chns.ExportChain(do))
}

func ListAllChains(cmd *cobra.Command, args []string) {
	//if no flags are set, list all the things
	//otherwise, allow only a single flag
	if !(do.Known || do.Running || do.Existing) {
		do.All = true
	} else {
		flargs := []bool{do.Known, do.Running, do.Existing}
		flags := []string{}

		for _, f := range flargs {
			if f == true {
				flags = append(flags, "true")
			}
		}
		IfExit(FlagCheck(1, "eq", cmd, flags))
	}

	if err := list.ListAll(do, "chains"); err != nil {
		return
	}
	if !do.All {
		fmt.Println(do.Result)
	}
}

func RenameChain(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "eq", cmd, args))
	do.Name = args[0]
	do.NewName = args[1]
	IfExit(chns.RenameChain(do))
}

func UpdateChain(cmd *cobra.Command, args []string) {
	// [csk]: if no args should we just start the checkedout chain?
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(chns.UpdateChain(do))
}

func RmChain(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(chns.RmChain(do))
}

func GraduateChain(cmd *cobra.Command, args []string) {
	// [csk]: if no args should we just start the checkedout chain?
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(chns.GraduateChain(do))
}

func MakeGenesisFile(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "ge", cmd, args))       //eq doesn't fly...
	do.Chain.Name = strings.TrimSpace(args[0]) //trim for bash
	do.Pubkey = strings.TrimSpace(args[1])
	IfExit(chns.MakeGenesisFile(do))

}
