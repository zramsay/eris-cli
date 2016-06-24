package commands

import (
	"github.com/spf13/cobra"
)

var Config = &cobra.Command{
	Use:   "config",
	Short: "manage configuration settings",
	Long: `display and manage configuration settings for various components of the
Eris platform and for the platform itself

The [eris config] command is only for configuring the Eris platform:
it will not work to configure any of the blockchains, services
or projects which are managed by the Eris platform. To configure
blockchains use [eris chains config]; to configure services use [eris services config]; 
to configure projects use [eris projects config] command.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var Cum []int

func buildConfigCommand() {
	Config.AddCommand(configPlop)
	Config.AddCommand(configSet)
	Config.AddCommand(configEdit)
}

var configSet = &cobra.Command{
	Use:   "set KEY:VALUE",
	Short: "set a config value",
	Long: `set a config value
NOTE: the [eris config set] command only operates on the settings 
for the eris CLI. To set the config for a blockchain use [eris chains config]
command, and to set the config for a service use [eris services config] 
command.`,
	Run: func(cmd *cobra.Command, args []string) {},
}

var configPlop = &cobra.Command{
	Use:   "show",
	Short: "display the config",
	Long:  `display the config`,
	Run:   func(cmd *cobra.Command, args []string) {},
}

var configEdit = &cobra.Command{
	Use:   "edit",
	Short: "ddit a config for in an editor",
	Long:  `edit a config for in your default editor`,
	Run:   func(cmd *cobra.Command, args []string) {},
}
