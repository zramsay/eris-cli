package commands

import (
	"github.com/eris-ltd/eris-cli/list"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var ListEverything = &cobra.Command{
	Use:   "ls",
	Short: "List all the things eris knows about.",
	Long: `List all known definition files for services
and chains. Also lists all existing and running services and
chains and, data containers.

For more detailed output, use [eris services ls], [eris chains ls], 
and [eris data ls] commands with respective flags (--known, --existing, 
--running).`,

	Run: func(cmd *cobra.Command, args []string) {
		ListAllTheThings()
	},
}

func ListAllTheThings() {
	//do.All for known/existing/running
	do.All = true

	typs := []string{"chains", "services"}
	for _, typ := range typs {
		if err := list.ListAll(do, typ); err != nil {
			return
		}
	}
	if err := list.ListDatas(do); err != nil {
		return
	}
	if err := list.ListActions(do); err != nil {
		return
	}
}
