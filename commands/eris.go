package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

const VERSION = "0.10.0"

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
		verbose := cmd.Flags().Lookup("verbose").Changed
		debug := cmd.Flags().Lookup("debug").Changed

		common.InitErisDir()
		util.DockerConnect(verbose)

		var logLevel int
		if verbose {
			logLevel = 1
		} else if debug {
			logLevel = 2
		}
		log.SetLoggers(logLevel, util.GlobalConfig.Writer, util.GlobalConfig.ErrorWriter)
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
	ErisCmd.AddCommand(Config)
	ErisCmd.AddCommand(Version)
	ErisCmd.AddCommand(Init)
}

// Global Flags
var Verbose bool
var Debug bool
var ContainerNumber int

// Flags that are to be used by commands
var Force bool
var Interactive bool
var Pull bool
var SkipPull bool
var Quiet bool
var All bool
var Follow bool
var Tail string
var Rm bool
var RmD bool
var Lines int

// Define the persistent commands (globals)
func AddGlobalFlags() {
	ErisCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	ErisCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "debug level output")
	ErisCmd.PersistentFlags().IntVarP(&ContainerNumber, "num", "n", 1, "container number")
	Init.Flags().BoolVarP(&Pull, "pull", "p", false, "skip the pulling feature; for when git is not installed")
}

func InitializeConfig() {
	var err error
	var out io.Writer
	var erw io.Writer

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
