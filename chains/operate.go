package chains

import (
	"fmt"
	"io"
	"os"

	"github.com/eris-ltd/eris-cli/services"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//----------------------------------------------------------------------

func Start(cmd *cobra.Command, args []string) {
	IfExit(checkChainGiven(args))
	IfExit(StartChainRaw(args[0], cmd.Flags().Lookup("verbose").Changed, os.Stdout))
}

func Logs(cmd *cobra.Command, args []string) {
	IfExit(checkChainGiven(args))
	IfExit(LogsChainRaw(args[0], cmd.Flags().Lookup("verbose").Changed, os.Stdout))
}

func Kill(cmd *cobra.Command, args []string) {
	IfExit(checkChainGiven(args))
	IfExit(KillChainRaw(args[0], cmd.Flags().Lookup("verbose").Changed, os.Stdout))
}

//----------------------------------------------------------------------

func StartChainRaw(chainName string, verbose bool, w io.Writer) error {
	chain, err := LoadChainDefinition(chainName)
	if err != nil {
		return err
	}
	if IsChainRunning(chain) {
		if verbose {
			w.Write([]byte("Chain already started. Skipping."))
		}
	} else {
		err := services.StartServiceByService(chain.Service, chain.Operations, verbose, w)
		if err != nil {
			return err
		}
	}
	return nil
}

func LogsChainRaw(chainName string, verbose bool, w io.Writer) error {
	chain, err := LoadChainDefinition(chainName)
	if err != nil {
		return err
	}
	err = services.LogsServiceByService(chain.Service, chain.Operations, verbose, w)
	if err != nil {
		return err
	}
	return nil
}

func KillChainRaw(chainName string, verbose bool, w io.Writer) error {
	chain, err := LoadChainDefinition(chainName)
	if err != nil {
		return err
	}

	if IsChainRunning(chain) {
		err := services.KillServiceByService(chain.Service, chain.Operations, verbose, w)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Chain not currently running. Skipping.")
	}
	return nil
}
