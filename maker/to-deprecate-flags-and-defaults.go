package maker

/*
import (
	"fmt"
	"os"

	"github.com/eris-ltd/eris-cm/maker"
	"github.com/eris-ltd/eris-cm/util"
	"github.com/eris-ltd/eris-cm/version"

	. "github.com/eris-ltd/common/go/common"
	keys "github.com/eris-ltd/eris-keys/eris-keys"
	log "github.com/eris-ltd/eris-cli/log"
	"github.com/spf13/cobra"
)

var MakerCmd = &cobra.Command{
	Use:   "make",
	Short: "The Eris Chain Maker is a utility for easily creating the files necessary to build eris chains",
	Long:  `The Eris Chain Maker is a utility for easily creating the files necessary to build eris chains.`,
	Example: `$ eris-cm make myChain -- will use the chain-making wizard and make your chain named myChain using eris-keys defaults (available via localhost) (interactive)
$ eris-cm make myChain --chain-type=simplechain --  will use the chain type definition files to make your chain named myChain using eris-keys defaults (non-interactive)
$ eris-cm make myChain --account-types=Root:1,Developer:0,Validator:0,Participant:1 -- will use the flag to make your chain named myChain using eris-keys defaults (non-interactive)
$ eris-cm make myChain --account-types=Root:1,Developer:0,Validator:0,Participant:1 --chain-type=simplechain -- account types trump chain types, this command will use the flags to make the chain (non-interactive)
$ eris-cm make myChain --csv /path/to/csv -- will use the csv file to make your chain named myChain using eris-keys defaults (non-interactive)`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// loop through chains directories to make sure they exist
		for _, d := range []string{ChainsPath, AccountsTypePath, ChainTypePath} {
			if _, err := os.Stat(d); os.IsNotExist(err) {
				os.MkdirAll(d, 0755)
			}
		}

		// drop default tomls into eris' location
		IfExit(util.CheckDefaultTypes(AccountsTypePath, "account-types"))
		IfExit(util.CheckDefaultTypes(ChainTypePath, "chain-types"))

		keys.DaemonAddr = keysAddr

		// Welcomer....
		log.Info("Hello! I'm the marmot who makes eris chains.")
	},
	Run:     MakeChain,
	PostRun: Archive,
}

// build the data subcommand
func buildMakerCommand() {
	AddMakerFlags()
}

// Flags that are to be used by commands are handled by the Do struct
// Define the persistent commands (globals)
func AddMakerFlags() {
	MakerCmd.PersistentFlags().StringVarP(&keysAddr, "keys-server", "k", defaultKeys(), "keys server which should be used to generate keys; default respects $ERIS_KEYS_PATH")
	MakerCmd.PersistentFlags().StringSliceVarP(&do.AccountTypes, "account-types", "t", defaultActTypes(), "what number of account types should we use? find these in ~/.eris/chains/account_types; incompatible with and overrides chain-type; default respects $ERIS_CHAINMANAGER_ACCOUNTTYPES")
	MakerCmd.PersistentFlags().StringVarP(&do.ChainType, "chain-type", "c", defaultChainType(), "which chain type definition should we use? find these in ~/.eris/chains/chain_types; default respects $ERIS_CHAINMANAGER_CHAINTYPE")
	MakerCmd.PersistentFlags().StringVarP(&do.CSV, "csv-file", "s", defaultCsvFiles(), "csv file in the form `account-type,number,tokens,toBond,perms; default respects $ERIS_CHAINMANAGER_CSVFILE")
	MakerCmd.PersistentFlags().BoolVarP(&do.Tarball, "tar", "r", defaultTarball(), "instead of making directories in ~/.chains, make tarballs; incompatible with and overrides zip; default respects $ERIS_CHAINMANAGER_TARBALLS")
	MakerCmd.PersistentFlags().BoolVarP(&do.Zip, "zip", "z", defaultZip(), "instead of making directories in ~/.chains, make zip files; default respects $ERIS_CHAINMANAGER_ZIPFILES")

	// Service related command flags
	MakerCmd.PersistentFlags().StringVarP(&do.ChainImageName, "image-name", "", defaultChainImageName(), "specify the chain image name; default respects $ERIS_CHAINMANAGER_CHAIN_IMAGE_NAME")
	MakerCmd.PersistentFlags().BoolVarP(&do.UseDataContainer, "use-data-container", "", defaultUseDataContainer(), "set whether to attach the data container to the chain; default respects $ERIS_CHAINMANAGER_USE_DATA_CONTAINER")
	MakerCmd.PersistentFlags().StringSliceVarP(&do.ExportedPorts, "ports", "", defaultExportedPorts(), "list the ports that need to be exported on the container; default respects $ERIS_CHAINMANAGER_EXPORTED_PORTS")
	MakerCmd.PersistentFlags().StringVarP(&do.ContainerEntrypoint, "entrypoint", "", defaultContainterEntrypoint(), "specifiy the entrypoint for the chain service; default respects $ERIS_CHAINMANAGER_CONTAINER_ENTRYPOINT")
}

//----------------------------------------------------
// functions

func MakeChain(cmd *cobra.Command, args []string) {
	argsMin := 1
	if len(args) < argsMin {
		cmd.Help()
		IfExit(fmt.Errorf("\n**Note** you sent our marmots the wrong number of arguments.\nPlease send the marmots at least %d argument(s).", argsMin))
	}
	do.Name = args[0]
	IfExit(maker.MakeChain(do))
}

func Archive(cmd *cobra.Command, args []string) {
	if do.Tarball {
		IfExit(util.Tarball(do))
	} else if do.Zip {
		IfExit(util.Zip(do))
	}
	if do.Output {
		IfExit(util.SaveAccountResults(do))
	}
}

// ---------------------------------------------------
// Defaults

func defaultKeys() string {
	return setDefaultString("ERIS_KEYS_PATH", fmt.Sprintf("http://localhost:4767"))
}

func defaultChainType() string {
	return setDefaultString("ERIS_CHAINMANAGER_CHAINTYPE", "")
}

func defaultActTypes() []string {
	return setDefaultStringSlice("ERIS_CHAINMANAGER_ACCOUNTTYPES", []string{})
}

func defaultCsvFiles() string {
	return setDefaultString("ERIS_CHAINMANAGER_CSVFILE", "")
}

func defaultTarball() bool {
	return setDefaultBool("ERIS_CHAINMANAGER_TARBALLS", false)
}

func defaultZip() bool {
	return setDefaultBool("ERIS_CHAINMANAGER_ZIPFILES", false)
}

// Chain service defaults

func defaultChainImageName() string {
	const imageHost string = "quay.io/eris/db:" + version.VERSION
	return setDefaultString("ERIS_CHAINMANAGER_IMAGE_NAME", imageHost)
}

func defaultUseDataContainer() bool {
	return setDefaultBool("ERIS_CHAINMANAGER_USE_DATA_CONTAINER", true)
}

// TODO: [ben] this is technical debt; these ports need to be detailed controlled
// and match the ports detailed in the configuration file itself.
func defaultExportedPorts() []string {
	return setDefaultStringSlice("ERIS_CHAINMANAGER_EXPORTED_PORTS",
		[]string{"1337", "46656", "46657"})
}

// [csk] this functionality has been removed from the eris:db container
func defaultContainterEntrypoint() string {
	return setDefaultString("ERIS_CHAINMANAGER_CONTAINER_ENTRYPOINT",
		"")
}*/
