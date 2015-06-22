package chains

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/services"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Start(cmd *cobra.Command, args []string) {
	checkChainGiven(args)
	StartChainRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func Logs(cmd *cobra.Command, args []string) {
	checkChainGiven(args)
	LogsChainRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func Kill(cmd *cobra.Command, args []string) {
	checkChainGiven(args)
	KillChainRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func StartChainRaw(chainName string, verbose bool) {
	chain := LoadChainDefinition(chainName)

	if IsChainRunning(chain) {
		if verbose {
			fmt.Println("Chain already started. Skipping.")
		}
	} else {
		services.StartServiceByService(chain.Service, chain.Operations, verbose)
	}
}

func LogsChainRaw(chainName string, verbose bool) {
	chain := LoadChainDefinition(chainName)
	services.LogsServiceByService(chain.Service, chain.Operations, verbose)
}

func KillChainRaw(chainName string, verbose bool) {
	chain := LoadChainDefinition(chainName)

	if IsChainRunning(chain) {
		services.KillServiceByService(chain.Service, chain.Operations, verbose)
	} else {
		if verbose {
			fmt.Println("Chain not currently running. Skipping.")
		}
	}
}
