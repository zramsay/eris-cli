package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monax/cli/chains"
	"github.com/monax/cli/config"
	"github.com/monax/cli/definitions"
	"github.com/monax/cli/list"
	"github.com/monax/cli/util"

	"github.com/spf13/cobra"
)

var Chains = &cobra.Command{
	Use:   "chains",
	Short: "start, stop, and manage blockchains",
	Long: `start, stop, and manage blockchains

The chains subcommand is used to work on monaxdb smart contract
blockchain networks. The name is not perfect, as monax is able
to operate a wide variety of blockchains out of the box. Most
of those existing blockchains should be ran via the [monax services ...]
commands. As they fall under the rubric of "things I just want
to turn on or off". While you can develop against those
blockchains, you generally aren't developing those blockchains
themselves. [monax chains ...] commands are built to help you build
blockchains. It is our opinionated gateway to the wonderful world
of permissioned smart contract networks.

Your own blockchain/smart contract machine is just an [monax chains start]
away!`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func buildChainsCommand() {
	Chains.AddCommand(chainsMake)
	Chains.AddCommand(chainsList)
	Chains.AddCommand(chainsCheckout)
	Chains.AddCommand(chainsCurrent)
	Chains.AddCommand(chainsPorts)
	Chains.AddCommand(chainsStart)
	Chains.AddCommand(chainsLogs)
	Chains.AddCommand(chainsInspect)
	Chains.AddCommand(chainsIP)
	Chains.AddCommand(chainsStop)
	Chains.AddCommand(chainsExec)
	Chains.AddCommand(chainsCat)
	Chains.AddCommand(chainsRestart)
	Chains.AddCommand(chainsRemove)
	addChainsFlags()
}

var chainsMake = &cobra.Command{
	Use:   "make NAME",
	Short: "create necessary files for your chain",
	Long: `create necessary files for your chain

Make is an opinionated gateway to the basic types of chains which most Monax users
will make most of the time. Make is also a command line wizard in which
you will let the marmots know how you would like your genesis created.

Make can also be used with a variety of flags for fast chain making.

When using make with the --known flag the marmots will *not* create keys for you
and will instead assume that the keys exist somewhere. When using make with the
wizard (no flags) or when using with the other flags then keys will be made along
with the genesis.jsons and priv_validator.jsons so that everything is ready to go
for you to [monax chains start].

Optionally chains make provides packages of outputted priv_validator and genesis.json
which you can email or send on your slack to your coworkers. These packages can
be tarballs or zip files, and **they will contain the private keys** so please
be aware of that.

The make process will *not* start a chain for you. You will want to use
the [monax chains start NAME --init-dir ` + util.Tilde(filepath.Join(config.ChainsPath, "NAME", "ACCOUNT")) + `] for that
which will import all of the files which make creates into containers and
start your shiny new chain.

If you have any questions on [monax chains make], see the documentation here:
https://monax.io/docs`,
	Example: `$ monax chains make myChain --wizard -- will use the interactive chain-making wizard and make your chain named myChain
$ monax chains make myChain -- will use the simplechain definition file to make your chain named myChain (non-interactive); use the [--chain-type] flag to specify chain types
$ monax chains make myChain --account-types=Root:1,Developer:0,Validator:1,Participant:1 -- will use the flag to make your chain named myChain (non-interactive)
$ monax chains make myChain --known --validators /path/to/validators.csv --accounts /path/to/accounts.csv -- will use the csv file to make your chain named myChain (non-interactive) (won't make keys)
$ monax chains make myChain --tar -- will create the chain and save each of the "bundles" as tarballs which can be used by colleagues to start their chains`,
	Run: MakeChain,
}

var chainsList = &cobra.Command{
	Use:   "ls",
	Short: "lists everything chain related",
	Long: `list chains or known chain definition files

The -r flag limits the output to running chains only.

The --json flag dumps the container or known files information
in the JSON format.

The -q flag is equivalent to the '{{.ShortName}}' format.

The -f flag specifies an alternate format for the list, using the syntax
of Go text templates. See the more detailed description in the help
output for the [monax ls] command.`,

	Run: ListChains,
	Example: `$ monax chains ls -f '{{.ShortName}} {{.Info.Config.Image}} {{ports .Info}}'
$ monax chains ls -f '{{.ShortName}}\t{{.Info.State}}'`,
}

var chainsCheckout = &cobra.Command{
	Use:   "checkout NAME",
	Short: "check out a chain",
	Long: `check out a chain

Checkout is a convenience feature. For any Monax command which accepts a
--chain or $chain variable, the checked out chain can replace manually
passing in a --chain flag. If a --chain is passed to any command accepting
--chain, the --chain which is passed will overwrite any checked out chain.

If command is given without arguments it will clear the head and there will
be no chain checked out.`,
	Run: CheckoutChain,
}

var chainsPorts = &cobra.Command{
	Use:   "ports NAME [PORT]...",
	Short: "print port mappings",
	Long: `print port mappings

The [monax chains ports] command is mostly a developer
convenience function. It returns a machine readable
port mapping of a port which is exposed inside the
container to what that port is mapped to on the host.

This is useful when stitching together chain networks which
need to know how to connect into a specific chain (perhaps
with or without a container number) container.`,
	Example: `$ monax chains ports myChain 1337 -- will display what port on the host is mapped to the monax:db API port
$ monax chains ports myChain 46656 -- will display what port on the host is mapped to the monax:db peer port
$ monax chains ports myChain 46657 -- will display what port on the host is mapped to the monax:db rpc port
$ monax chains ports myChain -- will display all mappings`,
	Run: PortsChain,
}

var chainsCurrent = &cobra.Command{
	Use:   "current",
	Short: "the currently checked out chain",
	Long: `displays the name of the currently checked out chain

To checkout a new chain use [monax chains checkout NAME].

To "uncheckout" a chain use [monax chains checkout] without arguments.`,
	Run: CurrentChain,
}

var chainsStart = &cobra.Command{
	Use:   "start NAME",
	Short: "start an existing chain or initialize a new one",
	Long: `start an existing chain or initialize a new one

[monax chains start NAME] by default will put an existing chain into
the background. Its logs will not be viewable from the command line.

To initialize (create) a new chain, the [monax chains make NAME] command
must first be run. This will (by default) create a simple chain with
relevant files in ` + util.Tilde(filepath.Join(config.ChainsPath, "NAME", "ACCOUNT")) + `. The path to this directory is then passed into the [--init-dir] flag like so:

  [monax chains start NAME --init-dir ` + util.Tilde(filepath.Join(config.ChainsPath, "NAME", "ACCOUNT")) + `]

To stop the chain use: [monax chains stop NAME]. To view a chain's logs use:
[monax chains logs NAME].

You can redefine the chain ports accessible over the network with the --ports flag.`,
	Run: StartChain,
	Example: `$ monax chains start simplechain --ports 4000 -- map the first port from the config file to the host port 40000
$ monax chains start simplechain --ports 40000,50000- -- redefine the first and the second port mapping and autoincrement the rest
$ monax chains start simplechain --ports 46656:50000 -- redefine the specific port mapping (published host port:exposed container port)`,
}

var chainsLogs = &cobra.Command{
	Use:   "logs NAME",
	Short: "display the logs of a blockchain",
	Long:  `display the logs of a blockchain`,
	Run:   LogChain,
}

var chainsExec = &cobra.Command{
	Use:   "exec NAME",
	Short: "run a command or interactive shell",
	Long: `run a command or interactive shell in a container
with volumes-from the data container`,
	Run: ExecChain,
}

var chainsStop = &cobra.Command{
	Use:   "stop NAME",
	Short: "stop a running blockchain",
	Long:  `stop a running blockchain`,
	Run:   StopChain,
}

var chainsInspect = &cobra.Command{
	Use:   "inspect NAME [KEY]",
	Short: "machine readable chain operation details",
	Long: `display machine readable details about running containers

Information available to the inspect command is provided by the
Docker API. For more information about return values,
see: https://github.com/fsouza/go-dockerclient/blob/master/container.go#L235`,
	Example: `$ monax chains inspect simplechain -- will display the entire information about simplechain containers
$ monax chains inspect 2gather Name -- will display the name in machine readable format
$ monax chains inspect 2gather HostConfig.Binds -- will display only that value`,
	Run: InspectChain,
}
var chainsIP = &cobra.Command{
	Use:   "ip NAME",
	Short: "display chain IP",
	Long:  `display chain IP`,

	Run: IPChain,
}

var chainsRemove = &cobra.Command{
	Use:   "rm NAME",
	Short: "remove an installed chain",
	Long: `remove an installed chain

Command will remove the chain's container but not its
local directory or data container unless specified.`,
	Run: RmChain,
}

var chainsRestart = &cobra.Command{
	Use:   "restart NAME",
	Short: "restart a chain",
	Long: `restart a chain

Command will gracefully stop then start a chain.`,
	Run: RestartChain,
}

var chainsCat = &cobra.Command{
	Use: "cat NAME [config|genesis]",
	//Use:     "cat NAME [config|genesis|status|validators]",
	Short:   "display chain information",
	Long:    `display chain information`,
	Aliases: []string{"plop"},
	Example: `$ monax chains cat simplechain config -- display the config.toml file from inside the container
$ monax chains cat simplechain genesis -- display the genesis.json file from the container`,
	// [zr] these don't work (mintinfo not found in container)
	// TODO re-implement when monax-client is merged into edb
	// $ monax chains cat simplechain status -- display chain status
	// $ monax chains cat simplechain validators -- display chain validators`,
	Run: CatChain,
}

func addChainsFlags() {
	chainsMake.PersistentFlags().StringSliceVarP(&do.AccountTypes, "account-types", "", []string{}, "specify the kind and number of account types. find these in "+util.Tilde(filepath.Join(config.ChainsPath, "account-types"))+"; incompatible with chain-type")
	chainsMake.PersistentFlags().StringVarP(&do.ChainType, "chain-type", "", "", "specify the type of chain to use. find these in "+util.Tilde(filepath.Join(config.ChainsPath, "chain-types"))+"; incompatible with account-types")
	chainsMake.PersistentFlags().BoolVarP(&do.Tarball, "tar", "", false, "instead of making directories in "+util.Tilde(config.ChainsPath)+", make tarballs; incompatible with and overrides zip")
	chainsMake.PersistentFlags().BoolVarP(&do.ZipFile, "zip", "", false, "instead of making directories in "+util.Tilde(config.ChainsPath)+", make zip files")
	chainsMake.PersistentFlags().BoolVarP(&do.Output, "output", "", true, "should monax-cm provide an output of its job")
	chainsMake.PersistentFlags().BoolVarP(&do.Known, "known", "", false, "use csv for a set of known keys to assemble genesis.json (requires both --accounts and --validators flags)")
	chainsMake.PersistentFlags().StringVarP(&do.ChainMakeActs, "accounts", "", "", "comma separated list of the accounts.csv files you would like to utilize (requires --known flag)")
	chainsMake.PersistentFlags().StringVarP(&do.ChainMakeVals, "validators", "", "", "comma separated list of the validators.csv files you would like to utilize (requires --known flag)")
	chainsMake.PersistentFlags().BoolVarP(&do.Wizard, "wizard", "w", false, "summon the interactive chain making wizard")
	chainsMake.PersistentFlags().StringSliceVarP(&do.SeedsIP, "seeds-ip", "", []string{}, "set a list of seeds (e.g. IP:PORT,IP:PORT) for peers to join the chain")
	// NOTE: [ben] the unsafe flag is introduced to start pushing out bad
	// practices from the tooling with regards to extracting private keys
	// from monax-keys.  Extracting the private keys can be convenient for
	// the development and poc phase, but must be deprecated even in that
	// case.
	chainsMake.PersistentFlags().BoolVarP(&do.Unsafe, "unsafe", "", false, "require explicit confirmation to write private keys from monax-keys to host during make in accounts.json")

	buildFlag(chainsStart, do, "init-dir", "chain")
	buildFlag(chainsStart, do, "publish", "chain")
	buildFlag(chainsStart, do, "ports", "chain")
	buildFlag(chainsStart, do, "env", "chain")
	buildFlag(chainsStart, do, "links", "chain")
	chainsStart.PersistentFlags().BoolVarP(&do.Force, "force", "f", false, "force reinitialize the chain")
	chainsStart.PersistentFlags().BoolVarP(&do.Logrotate, "logrotate", "z", false, "turn on logrotate as a dependency to handle long output")

	buildFlag(chainsLogs, do, "follow", "chain")
	buildFlag(chainsLogs, do, "tail", "chain")

	buildFlag(chainsExec, do, "publish", "chain")
	buildFlag(chainsExec, do, "ports", "chain")
	buildFlag(chainsExec, do, "interactive", "chain")
	buildFlag(chainsExec, do, "links", "chain")
	chainsExec.Flags().StringVarP(&do.Image, "image", "", "", "docker image")

	buildFlag(chainsRemove, do, "force", "chain")
	buildFlag(chainsRemove, do, "data", "chain")
	buildFlag(chainsRemove, do, "rm-volumes", "chain")
	chainsRemove.Flags().BoolVarP(&do.RmHF, "dir", "r", false, "remove the chain directory in "+util.Tilde(config.ChainsPath))

	buildFlag(chainsStop, do, "force", "chain")
	buildFlag(chainsStop, do, "timeout", "chain")

	chainsList.Flags().BoolVarP(&do.JSON, "json", "", false, "machine readable output")
	chainsList.Flags().BoolVarP(&do.All, "all", "a", false, "show extended output")
	chainsList.Flags().BoolVarP(&do.Quiet, "quiet", "q", false, "show a list of chain names")
	chainsList.Flags().StringVarP(&do.Format, "format", "f", "", "alternate format for columnized output")
	chainsList.Flags().BoolVarP(&do.Running, "running", "r", false, "show running containers")
}

func StartChain(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	util.IfExit(chains.StartChain(do))
}

func LogChain(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	util.IfExit(chains.LogsChain(do))
}

func ExecChain(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))

	do.Name = args[0]
	// if interactive, we ignore args. if not, run args as command
	args = args[1:]
	if !do.Operations.Interactive {
		if len(args) == 0 {
			util.Exit(fmt.Errorf("Non-interactive exec sessions must provide arguments to execute"))
		}
	}
	if len(args) == 1 {
		args = strings.Split(args[0], " ")
	}
	do.Operations.Terminal = true
	do.Operations.Args = args
	config.Global.InteractiveWriter = os.Stdout
	config.Global.InteractiveErrorWriter = os.Stderr
	_, err := chains.ExecChain(do)
	util.IfExit(err)
}

func StopChain(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	util.IfExit(chains.StopChain(do))
}

func MakeChain(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "eq", cmd, args))

	do.Name = args[0]

	// TODO clean up this logic
	if do.Known && (do.ChainMakeActs == "" || do.ChainMakeVals == "") {
		cmd.Help()
		util.IfExit(fmt.Errorf("If you are using the --known flag the --validators *and* the --accounts flags are both required"))
	}
	if !do.Known && (do.ChainMakeActs != "" || do.ChainMakeVals != "") {
		cmd.Help()
		util.IfExit(fmt.Errorf("If you are using the --validators and the --accounts flags, --known is also required"))
	}
	if len(do.AccountTypes) > 0 && do.ChainType != "" {
		cmd.Help()
		util.IfExit(fmt.Errorf("The --account-types flag is incompatible with the --chain-type flag. Please use one or the other"))
	}
	if (len(do.AccountTypes) > 0 || do.ChainType != "") && do.Known {
		cmd.Help()
		util.IfExit(fmt.Errorf("The --account-types and --chain-type flags are incompatible with the --known flag. Please use only one of these"))
	}
	if do.Known && do.Wizard {
		cmd.Help()
		util.IfExit(fmt.Errorf("The --known and --wizard flags are incompatible with each other. Please use one one of these"))
	}

	if do.Wizard {
		// TODO ... something ... ?
	} else if len(do.AccountTypes) == 0 && do.ChainType == "" && do.ChainMakeActs == "" && do.ChainMakeVals == "" {
		// no flags given assume simplechain
		do.ChainType = "simplechain"
	}

	util.IfExit(chains.MakeChain(do))
}

func CheckoutChain(cmd *cobra.Command, args []string) {
	if len(args) >= 1 {
		do.Name = args[0]
	} else {
		do.Name = ""
	}
	util.IfExit(chains.CheckoutChain(do))
}

func CurrentChain(cmd *cobra.Command, args []string) {
	out, err := chains.CurrentChain(do)
	util.IfExit(err)
	fmt.Fprintln(config.Global.Writer, out)
}

func CatChain(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(2, "ge", cmd, args))
	do.Name = args[0]
	do.Type = args[1]
	util.IfExit(chains.CatChain(do))
}

func PortsChain(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	do.Operations.Args = args[1:]
	util.IfExit(chains.PortsChain(do))
}

func InspectChain(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))

	do.Name = args[0]
	if len(args) == 1 {
		do.Operations.Args = []string{"all"}
	} else {
		do.Operations.Args = []string{args[1]}
	}

	util.IfExit(chains.InspectChain(do))
}

func IPChain(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))

	do.Name = args[0]
	do.Operations.Args = []string{"NetworkSettings.IPAddress"}
	util.IfExit(chains.InspectChain(do))
}

func ListChains(cmd *cobra.Command, args []string) {
	if do.All {
		do.Format = "extended"
	}
	if do.Quiet {
		do.Format = "{{.ShortName}}"
	}
	if do.JSON {
		do.Format = "json"
	}
	util.IfExit(list.Containers(definitions.TypeChain, do.Format, do.Running))
}

func RestartChain(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	util.IfExit(chains.StopChain(do))
	util.IfExit(chains.StartChain(do))
}

func RmChain(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	util.IfExit(chains.RemoveChain(do))
}
