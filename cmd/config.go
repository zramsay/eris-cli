package commands

import (
	"github.com/spf13/cobra"
)

var Config = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings.",
	Long: `Display and manage configuration settings for various components of the
Eris platform and for the platform itself.

The [eris config] command is only for configuring the Eris platform:
it will not work to configure any of the blockchains, services
or projects which are managed by the Eris platform. To configure
blockchains use [eris chains config]; to configure services use [eris services config]; 
to configure projects use [eris projects config] command.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var Cum []int

// build the config subcommand
func buildConfigCommand() {
	Config.AddCommand(configPlop)
	Config.AddCommand(configSet)
	Config.AddCommand(configEdit)
	// configPlop.Flags().IntSliceVar(&Cum, "cum", "c", []int{}, "suppress action output")
}

// set
var configSet = &cobra.Command{
	Use:   "set KEY:VALUE",
	Short: "Set a config value.",
	Long: `Set a config value.
NOTE: the [eris config set] command only operates on the settings 
for the eris CLI. To set the config for a blockchain use [eris chains config]
command, and to set the config for a service use [eris services config] 
command.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO
	},
}

// show
var configPlop = &cobra.Command{
	Use:   "show",
	Short: "Display the config.",
	Long:  `Display the config.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO
	},
}

// edit
var configEdit = &cobra.Command{
	Use:   "edit",
	Short: "Edit a config for in an editor.",
	Long:  `Edit a config for in your default editor.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO
	},
}
