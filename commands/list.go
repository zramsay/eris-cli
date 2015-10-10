package commands

import (
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var ListEverything = &cobra.Command{
	Use:   "ls",
	Short: "Lists all the things eris knows about.",
	Long: `Lists all known definition files for services
and chains. Also lists all existing and running services and 
chains and, data containers. 

For more specific output, use:

	$ eris services ls
	$ eris chains ls
	$ eris data ls

with respective flags (--known, --existing, --running).`,

	Run: func(cmd *cobra.Command, args []string) {
		ListAllTheThings()
	},
}

func ListAllTheThings() {
	//do.All for known/existing/running
	do.All = true

	typs := []string{"services", "chains", "actions", "data"}
	for _, typ := range typs {
		if err := util.ListAll(do, typ); err != nil {
			return
		}
	}
}
