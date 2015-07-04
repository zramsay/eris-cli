package chains

import (
	"fmt"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
)

func StartChainRaw(chainName string, containerNumber int) error {
	chain, err := LoadChainDefinition(chainName, containerNumber)
	if err != nil {
		return err
	}
	if IsChainRunning(chain) {
		logger.Infoln("Chain already started. Skipping.")
	} else {
		chain.Service.Command = ErisChainStart
		chain.Service.Environment = append(chain.Service.Environment, "CHAIN_ID="+chain.ChainID)
		if err := services.StartServiceByService(chain.Service, chain.Operations); err != nil {
			return err
		}
	}
	return nil
}

func LogsChainRaw(chainName string, follow bool, tail string, containerNumber int) error {
	chain, err := LoadChainDefinition(chainName, containerNumber)
	if err != nil {
		return err
	}
	err = services.LogsServiceByService(chain.Service, chain.Operations, follow, tail)
	if err != nil {
		return err
	}
	return nil
}

func ExecChainRaw(name string, args []string, attach bool, containerNumber int) error {
	chain, err := LoadChainDefinition(name, containerNumber)
	if err != nil {
		return err
	}

	if IsChainExisting(chain) {
		logger.Infoln("Chain exists.")
		return services.ExecServiceByService(chain.Service, chain.Operations, args, attach)
	} else {
		return fmt.Errorf("Chain does not exist. Please start the chain container with eris chains start %s.\n", name)
	}

	return nil
}

func KillChainRaw(chainName string, rm, rmData bool, containerNumber int) error {
	chain, err := LoadChainDefinition(chainName, containerNumber)
	if err != nil {
		return err
	}

	if IsChainRunning(chain) {
		// this won't kank the data for the chain (only it's dependent services)
		// TODO: refactor service/chain loading so this problem goes away
		if err := services.KillServiceByService(true, rm, rmData, chain.Service, chain.Operations); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Chain not currently running. Skipping.")
	}

	if rm {
		if err := perform.DockerRemove(chain.Service, chain.Operations); err != nil {
			return err
		}
	}

	if rmData {
		if err := data.RmDataRaw(chainName, containerNumber); err != nil {
			return err
		}
	}
	return nil
}
