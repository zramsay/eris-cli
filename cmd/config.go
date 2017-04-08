package commands

import (
	"github.com/spf13/cobra"
)

var Config = &cobra.Command{
	Use:   "config",
	Short: "manage configuration settings",
	Long: `display and manage configuration settings for various components of Monax and for the platform itself

The [monax config] command is only for configuring Monax:
it will not work to configure any of the blockchains, services
or projects which are managed by Monax. To configure blockchains 
use [monax chains config]; to configure services use [monax services config]; 
to configure projects use [monax projects config] command.`,
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
NOTE: the [monax config set] command only operates on the settings 
for the monax CLI. To set the config for a blockchain use [monax chains config]
command, and to set the config for a service use [monax services config] 
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
