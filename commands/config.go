package commands

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var config = &cobra.Command{
	Use:   "config [variable]",
	Short: "Manage configuration settings for Eris",
	Long: `Display Manage configuration settings for various components of the
Eris platform and for the platform itself.`,
	Run: func(cmd *cobra.Command, args []string) {
		plopConfigVals(args)
	},
}

func plopConfigVals(args []string) {
	for _, key := range args {
		fmt.Println(globalConfig.GetString(key))
	}
}
