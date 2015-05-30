package commands

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var config = &cobra.Command{
	Use:   "config [key]:[var]",
	Short: "Manage configuration settings for Eris",
	Long: `Display Manage configuration settings for various components of the
Eris platform and for the platform itself.

NOTE: [eris config] is only for configuring the Eris platform
it will not work to configure any of the blockchains, services
or projects which are managed by the Eris platform. To configure
blockchains use [eris chains name config]; to configure services
use [eris services name config]; to configure projects use
[eris projects name config].`,
	Run: func(cmd *cobra.Command, args []string) {
		plopConfigVals(args)
	},
}

func plopConfigVals(args []string) {
	for _, key := range args {
		fmt.Println(globalConfig.GetString(key))
	}
}
