package commands

import (
	"github.com/eris-ltd/eris-cli/config"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var Config = &cobra.Command{
	Use:   "config",
	Short: "Manage Configuration Settings for Eris.",
	Long: `Display and Manage configuration settings for various components of the
Eris platform and for the platform itself.

NOTE: [eris config] is only for configuring the Eris platform
it will not work to configure any of the blockchains, services
or projects which are managed by the Eris platform. To configure
blockchains use [eris chains config]; to configure services
use [eris services config]; to configure projects use
[eris projects config].`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// build the config subcommand
func buildConfigCommand() {
	Config.AddCommand(configPlop)
	Config.AddCommand(configSet)
	Config.AddCommand(configEdit)
}

// set
var configSet = &cobra.Command{
	Use:   "set [key]:[var]",
	Short: "Set a config for the Eris Platform CLI.",
	Long: `Set a config for the Eris Platform CLI.

Note [eris config set] only operates on the settings for the eris
cli. To set the config for a blockchain use [eris chains config]
and to set the config for a service use [eris services config].`,
	Run: func(cmd *cobra.Command, args []string) {
		config.Set(args)
	},
}

// show
var configPlop = &cobra.Command{
	Use:   "show",
	Short: "Display the config for the Eris Platform CLI.",
	Long:  `Display the config for the Eris Platform CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		config.PlopEntireConfig(globalConfig, args)
	},
}

// edit
var configEdit = &cobra.Command{
	Use:   "edit",
	Short: "Edit a config for the Eris Platform CLI in an editor.",
	Long:  `Edit a config for the Eris Platform CLI in your default editor.`,
	Run: func(cmd *cobra.Command, args []string) {
		config.Edit()
	},
}
