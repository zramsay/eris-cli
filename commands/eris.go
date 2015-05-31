package commands

import (
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
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
}

func Execute() {
	InitializeConfig()
	AddGlobalFlags()
	AddCommands()
	AddCommandFlags()
	ErisCmd.Execute()
	// utils.StopOnErr(ErisCmd.Execute())
}

// Define the commands
func AddCommands() {
	buildProjectsCommand()
	ErisCmd.AddCommand(Projects)
	buildChainsCommand()
	ErisCmd.AddCommand(Chains)
	buildServicesCommand()
	ErisCmd.AddCommand(Services)
	buildActionsCommand()
	ErisCmd.AddCommand(Actions)
	buildRemotesCommand()
	ErisCmd.AddCommand(Remotes)
	buildKeysCommand()
	ErisCmd.AddCommand(Keys)
	buildConfigCommand()
	ErisCmd.AddCommand(Config)
	ErisCmd.AddCommand(Version)
}

// Flags that are to be used by commands
var Verbose bool

// Define the persistent commands (globals)
func AddGlobalFlags() {
	ErisCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

// Add command flags
func AddCommandFlags() {

}

// Properly scope the globalConfig
var globalConfig *viper.Viper

func InitializeConfig() {
	globalConfig = viper.New()
	util.LoadGlobalConfig(globalConfig)
}
