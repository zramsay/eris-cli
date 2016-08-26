package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"

	chns "github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/list"

	. "github.com/eris-ltd/common/go/common"
	"github.com/spf13/cobra"
)

var Chains = &cobra.Command{
	Use:   "chains",
	Short: "start, stop, and manage blockchains",
	Long: `start, stop, and manage blockchains

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

func buildChainsCommand() {
	Chains.AddCommand(chainsMake)
	Chains.AddCommand(chainsNew)
	Chains.AddCommand(chainsList)
	Chains.AddCommand(chainsCheckout)
	Chains.AddCommand(chainsHead)
	Chains.AddCommand(chainsPorts)
	Chains.AddCommand(chainsStart)
	Chains.AddCommand(chainsLogs)
	Chains.AddCommand(chainsInspect)
	Chains.AddCommand(chainsStop)
	Chains.AddCommand(chainsExec)
	Chains.AddCommand(chainsCat)
	Chains.AddCommand(chainsUpdate)
	Chains.AddCommand(chainsRestart)
	Chains.AddCommand(chainsRemove)
	addChainsFlags()
}

var chainsMake = &cobra.Command{
	Use:   "make NAME",
	Short: "create keys and a genesis block for your chain",
	Long: `create the new required files for your chain

Make is an opinionated gateway to the basic types of chains which most eris users
will make most of the time. Make is a command line wizard in which you will let
the marmots know how you would like your genesis created.

Make can also be used with a variety of flags for fast chain making.

When using make with the --known flag the marmots will *not* create keys for you
and will instead assume that the keys exist somewhere. When using make with the
wizard (no flags) or when using with the other flags then keys will be made along
with the genesis.jsons and priv_validator.jsons so that everything is ready to go
for you to [eris chains new].

Optionally chains make provides packages of outputted priv_validator and genesis.json
which you can email or send on your slack to your coworkers. These packages can
be tarballs or zip files, and **they will contain the private keys** so please
be aware of that.

The make process will *not* start a chain for you. You will want to use
the [eris chains new chainName --dir chainName] for that which will import all
of the files which make creates into containers and start your shiny new chain.

If you have any questions on eris chains make, please see the eris-cm (chain manager)
documentation here:
https://docs.erisindustries.com/documentation/eris-cm/latest/eris-cm/`,
	Example: `$ eris chains make myChain -- will use the chain-making wizard and make your chain named myChain (interactive)
$ eris chains make myChain --chain-type=simplechain --  will use the chain type definition files to make your chain named myChain (non-interactive)
$ eris chains make myChain --account-types=Root:1,Developer:0,Validator:0,Participant:1 -- will use the flag to make your chain named myChain (non-interactive)
$ eris chains make myChain --account-types=Root:1,Developer:0,Validator:0,Participant:1 --chain-type=simplechain -- account types trump chain types, this command will use the flags to make the chain (non-interactive)
$ eris chains make myChain --known --validators /path/to/validators.csv --accounts /path/to/accounts.csv -- will use the csv file to make your chain named myChain (non-interactive) (won't make keys)
$ eris chains make myChain --tar -- will create the chain and save each of the "bundles" as tarballs which can be used by colleagues to start their chains`,
	Run: MakeChain,
}

var chainsNew = &cobra.Command{
	Use:   "new NAME",
	Short: "create and start a new blockhain",
	Long: `create and start a new blockchain

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

If you would like to create a genesis.json then please utilize [eris chains make]

You can redefine the chain ports accessible over the network with the --ports flag.
`,
	Run: NewChain,
	Example: `$ eris chains new simplechain --ports 4000 -- map the first port from the definition file to the host port 40000
$ eris chains new simplechain --ports 40000,50000- -- redefine the first and the second port mapping and autoincrement the rest
$ eris chains new simplechain --ports 46656:50000 -- redefine the specific port mapping (published host port:exposed container port)`,
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
output for the [eris ls] command.`,

	Run: ListChains,
	Example: `$ eris chains ls -f '{{.ShortName}}\t{{.Info.Config.Image}}\t{{ports .Info}}'
$ eris chains ls -f '{{.ShortName}}\t{{.Info.State}}'`,
}

var chainsCheckout = &cobra.Command{
	Use:   "checkout NAME",
	Short: "check out a chain",
	Long: `check out a chain

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
	Short: "print port mappings",
	Long: `print port mappings

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
	Short: "the currently checked out chain",
	Long: `displays the name of the currently checked out chain

To checkout a new chain use [eris chains checkout NAME].

To "uncheckout" a chain use [eris chains checkout] without arguments.`,
	Run: CurrentChain,
}

var chainsStart = &cobra.Command{
	Use:   "start",
	Short: "start a blockchain",
	Long: `start running a blockchain

[eris chains start NAME] by default will put the chain into the
background so its logs will not be viewable from the command line.

To stop the chain use:      [eris chains stop NAME]
To view a chain's logs use: [eris chains logs NAME]

You can redefine the chain ports accessible over the network with the --ports flag.
See the [eris chains new] command for examples.
`,
	Run: StartChain,
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
	Run:   KillChain,
}

var chainsInspect = &cobra.Command{
	Use:   "inspect NAME [KEY]",
	Short: "machine readable chain operation details",
	Long: `display machine readable details about running containers

Information available to the inspect command is provided by the
Docker API. For more information about return values,
see: https://github.com/fsouza/go-dockerclient/blob/master/container.go#L235`,
	Example: `$ eris chains inspect 2gather -- will display the entire information about 2gather containers
$ eris chains inspect 2gather name -- will display the name in machine readable format
$ eris chains inspect 2gather host_config.binds -- will display only that value`,
	Run: InspectChain,
}

var chainsRemove = &cobra.Command{
	Use:   "rm NAME",
	Short: "remove an installed chain",
	Long: `remove an installed chain

Command will remove the chain's container but will not
remove the chain definition file.`,
	Run: RmChain,
}

var chainsUpdate = &cobra.Command{
	Use:   "update NAME",
	Short: "update an installed chain",
	Long: `update an installed chain, or install it if it has not been installed

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

var chainsRestart = &cobra.Command{
	Use:   "restart NAME",
	Short: "restart chain",
	Long:  `restart chain`,
	Run:   RestartChain,
}

var chainsCat = &cobra.Command{
	Use:     "cat NAME [config|genesis|status|validators]",
	Short:   "display chain information",
	Long:    `display chain information`,
	Aliases: []string{"plop"},
	Example: `$ eris chains cat simplechain -- display the chain definition file
$ eris chains cat simplechain config -- display the config.toml file from inside the container
$ eris chains cat simplechain genesis -- display the genesis.json file from the container
$ eris chains cat simplechain status -- display chain status
$ eris chains cat simplechain validators -- display chain validators`,
	Run: CatChain,
}

func addChainsFlags() {
	chainsMake.PersistentFlags().StringSliceVarP(&do.AccountTypes, "account-types", "", []string{}, "what number of account types should we use? find these in ~/.eris/chains/account-types; incompatible with and overrides chain-type")
	chainsMake.PersistentFlags().StringVarP(&do.ChainType, "chain-type", "", "", "which chain type definition should we use? find these in ~/.eris/chains/chain-types")
	chainsMake.PersistentFlags().BoolVarP(&do.Tarball, "tar", "", false, "instead of making directories in ~/.eris/chains, make tarballs; incompatible with and overrides zip")
	chainsMake.PersistentFlags().BoolVarP(&do.ZipFile, "zip", "", false, "instead of making directories in ~/.eris/chains, make zip files")
	chainsMake.PersistentFlags().BoolVarP(&do.Output, "output", "", true, "should eris-cm provide an output of its job")
	chainsMake.PersistentFlags().BoolVarP(&do.Known, "known", "", false, "use csv for a set of known keys to assemble genesis.json (requires both --accounts and --validators flags")
	chainsMake.PersistentFlags().StringVarP(&do.ChainMakeActs, "accounts", "", "", "comma separated list of the accounts.csv files you would like to utilize (requires --known flag)")
	chainsMake.PersistentFlags().StringVarP(&do.ChainMakeVals, "validators", "", "", "comma separated list of the validators.csv files you would like to utilize (requires --known flag)")
	chainsMake.PersistentFlags().BoolVarP(&do.RmD, "data", "x", true, "remove data containers after stopping")

	buildFlag(chainsNew, do, "dir", "chain")
	buildFlag(chainsNew, do, "env", "chain")
	buildFlag(chainsNew, do, "publish", "chain")
	buildFlag(chainsNew, do, "ports", "chain")
	buildFlag(chainsNew, do, "links", "chain")
	chainsNew.PersistentFlags().BoolVarP(&do.Logrotate, "logrotate", "z", false, "turn on logrotate as a dependency to handle long output")

	buildFlag(chainsStart, do, "publish", "chain")
	buildFlag(chainsStart, do, "ports", "chain")
	buildFlag(chainsStart, do, "env", "chain")
	buildFlag(chainsStart, do, "links", "chain")
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
	chainsRemove.Flags().BoolVarP(&do.RmHF, "dir", "", false, "remove the chain directory in ~/.eris/chains")

	buildFlag(chainsUpdate, do, "pull", "chain")
	buildFlag(chainsUpdate, do, "timeout", "chain")
	buildFlag(chainsUpdate, do, "env", "chain")
	buildFlag(chainsUpdate, do, "links", "chain")

	buildFlag(chainsStop, do, "rm", "chain")
	buildFlag(chainsStop, do, "data", "chain")
	buildFlag(chainsStop, do, "force", "chain")
	buildFlag(chainsStop, do, "timeout", "chain")
	buildFlag(chainsStop, do, "volumes", "chain")

	chainsList.Flags().BoolVarP(&do.JSON, "json", "", false, "machine readable output")
	chainsList.Flags().BoolVarP(&do.All, "all", "a", false, "show extended output")
	chainsList.Flags().BoolVarP(&do.Quiet, "quiet", "q", false, "show a list of chain names")
	chainsList.Flags().StringVarP(&do.Format, "format", "f", "", "alternate format for columnized output")
	chainsList.Flags().BoolVarP(&do.Running, "running", "r", false, "show running containers")
}

func StartChain(cmd *cobra.Command, args []string) {
	// [csk]: if no args should we just start the checkedout chain?
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	do.Run = true
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
	do.Operations.Terminal = true
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
	if !do.Known {
		config.GlobalConfig.InteractiveWriter = os.Stdout
		config.GlobalConfig.InteractiveErrorWriter = os.Stderr
		do.Operations.Terminal = true
	}

	IfExit(chns.MakeChain(do))
}

func NewChain(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	do.Run = true
	if do.Name != "default" && do.Path == "" { //not default & no --dir given
		IfExit(errors.New("cannot omit the --dir flag unless chainName == default"))
	}
	IfExit(chns.NewChain(do))
}

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
	IfExit(list.Containers(def.TypeChain, do.Format, do.Running))
}

func UpdateChain(cmd *cobra.Command, args []string) {
	// [csk]: if no args should we just start the checkedout chain?
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(chns.UpdateChain(do))
}

func RestartChain(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	do.Pull = false
	IfExit(chns.UpdateChain(do))
}

func RmChain(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(chns.RemoveChain(do))
}

func MakeGenesisFile(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "ge", cmd, args))
	do.Chain.Name = strings.TrimSpace(args[0])
	do.Pubkey = strings.TrimSpace(args[1])
	IfExit(chns.MakeGenesisFile(do))

}
