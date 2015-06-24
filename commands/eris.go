package commands

import (
	"fmt"
	"io"
	"os"

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
		common.InitErisDir()
		util.DockerConnect(cmd)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		err := util.SaveGlobalConfig(GlobalConfig.Config)
		if err != nil {
			fmt.Fprintln(GlobalConfig.ErrorWriter, err)
		}
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
	buildKeysCommand()
	ErisCmd.AddCommand(Keys)
	buildConfigCommand()
	ErisCmd.AddCommand(Config)
	ErisCmd.AddCommand(Version)
	ErisCmd.AddCommand(Init)
}

// Global Flags
var Verbose bool

// Flags that are to be used by commands
var Force bool

// Define the persistent commands (globals)
func AddGlobalFlags() {
	ErisCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

// Properly scope the globalConfig
var GlobalConfig *util.ErisCli

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

	GlobalConfig, err = util.SetGlobalObject(out, erw)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
