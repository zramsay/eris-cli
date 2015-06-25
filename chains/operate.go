package chains

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/perform"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//----------------------------------------------------------------------

func Start(cmd *cobra.Command, args []string) {
	IfExit(checkChainGiven(args))
	IfExit(StartChainRaw(args[0]))
}

func Logs(cmd *cobra.Command, args []string) {
	IfExit(checkChainGiven(args))
	IfExit(LogsChainRaw(args[0]))
}

func Exec(cmd *cobra.Command, args []string) {
	IfExit(checkChainGiven(args))
	srv := args[0]

	// if interactive, we ignore args. if not, run args as command
	interactive := cmd.Flags().Lookup("interactive").Changed
	if !interactive {
		if len(args) < 2 {
			Exit(fmt.Errorf("Non-interactive exec sessions must provide arguments to execute"))
		}
		args = args[1:]
	}

	IfExit(ExecChainRaw(srv, interactive, args))
}

func Kill(cmd *cobra.Command, args []string) {
	IfExit(checkChainGiven(args))
	IfExit(KillChainRaw(args[0]))
}

//----------------------------------------------------------------------

func StartChainRaw(chainName string) error {
	chain, err := LoadChainDefinition(chainName)
	if err != nil {
		return err
	}
	if IsChainRunning(chain) {
		logger.Infoln("Chain already started. Skipping.")
	} else {
		err := services.StartServiceByService(chain.Service, chain.Operations)
		if err != nil {
			return err
		}
	}
	return nil
}

func LogsChainRaw(chainName string) error {
	chain, err := LoadChainDefinition(chainName)
	if err != nil {
		return err
	}
	err = services.LogsServiceByService(chain.Service, chain.Operations)
	if err != nil {
		return err
	}
	return nil
}

func ExecChainRaw(name string, interactive bool, args []string) error {
	if isKnownChain(name) {
		logger.Infoln("Running exec on container with volumes from data container for " + name)
		if err := perform.DockerRunVolumesFromContainer(name, interactive, args); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}
	return nil
}

func KillChainRaw(chainName string) error {
	chain, err := LoadChainDefinition(chainName)
	if err != nil {
		return err
	}

	if IsChainRunning(chain) {
		err := services.KillServiceByService(chain.Service, chain.Operations)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Chain not currently running. Skipping.")
	}
	return nil
}
