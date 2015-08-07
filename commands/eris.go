package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/log"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

const VERSION = version.VERSION

// Defining the root command
var ErisCmd = &cobra.Command{
	Use:   "eris [command] [flags]",
	Short: "The Blockchain Application Platform",
	Long: `Eris is a platform for building, testing, maintaining, and operating
distributed applications with a blockchain backend. Eris makes it easy
and simple to wrangle the dragons of smart contract blockchains.

Made with <3 by Eris Industries.

Complete documentation is available at https://docs.erisindustries.com
` + "\nVersion:\n  " + VERSION,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var logLevel int
		if do.Verbose {
			logLevel = 1
		} else if do.Debug {
			logLevel = 2
		}
		log.SetLoggers(logLevel, util.GlobalConfig.Writer, util.GlobalConfig.ErrorWriter)

		common.InitErisDir()
		util.DockerConnect(do.Verbose)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		err := util.SaveGlobalConfig(util.GlobalConfig.Config)
		if err != nil {
			logger.Errorln(err)
		}
		log.Flush()
	},
}

func Execute() {
	InitializeConfig()
	AddGlobalFlags()
	AddCommands()
	ErisCmd.Execute()
}

// Define the commands
func AddCommands() {
	// buildProjectsCommand()
	// ErisCmd.AddCommand(Projects)
	buildServicesCommand()
	ErisCmd.AddCommand(Services)
	buildChainsCommand()
	ErisCmd.AddCommand(Chains)
	buildActionsCommand()
	ErisCmd.AddCommand(Actions)
	buildDataCommand()
	ErisCmd.AddCommand(Data)
	buildFilesCommand()
	ErisCmd.AddCommand(Files)
	// buildRemotesCommand()
	// ErisCmd.AddCommand(Remotes)
	buildConfigCommand()
	ErisCmd.AddCommand(ListKnown)
	ErisCmd.AddCommand(ListExisting)
	//	ErisCmd.AddCommand(ListRunning)

	ErisCmd.AddCommand(Config)
	ErisCmd.AddCommand(Version)
	ErisCmd.AddCommand(Init)
}

var do *definitions.Do

// Flags that are to be used by commands are handled by the Do struct
// Define the persistent commands (globals)
func AddGlobalFlags() {
	ErisCmd.PersistentFlags().BoolVarP(&do.Verbose, "verbose", "v", false, "verbose output")
	ErisCmd.PersistentFlags().BoolVarP(&do.Debug, "debug", "d", false, "debug level output")
	ErisCmd.PersistentFlags().IntVarP(&do.Operations.ContainerNumber, "num", "n", 1, "container number")
	Init.Flags().BoolVarP(&do.Pull, "pull", "p", false, "skip the pulling feature; for when git is not installed")
	Init.Flags().BoolVarP(&do.Dev, "dev", "", false, "pull development images")
	Init.Flags().BoolVarP(&do.SkipImages, "no-pull", "", false, "skip pulling default images")
}

func InitializeConfig() {
	var err error
	var out io.Writer
	var erw io.Writer

	do = definitions.NowDo()

	if os.Getenv("ERIS_CLI_WRITER") != "" {
		out, err = os.Open(os.Getenv("ERIS_CLI_WRITER"))
		if err != nil {
			fmt.Printf("Could not open: %s\n", err)
			return
		}
	} else {
		out = os.Stdout
	}

	if os.Getenv("ERIS_CLI_ERROR_WRITER") != "" {
		erw, err = os.Open(os.Getenv("ERIS_CLI_ERROR_WRITER"))
		if err != nil {
			fmt.Printf("Could not open: %s\n", err)
			return
		}
	} else {
		erw = os.Stderr
	}

	util.GlobalConfig, err = util.SetGlobalObject(out, erw)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
