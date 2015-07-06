package chains

import (
	"fmt"

	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"
)

func StartChainRaw(chainName string, containerNumber int, ops *def.Operation) error {
	chain, err := LoadChainDefinition(chainName, containerNumber)
	if err != nil {
		return err
	}
	if IsChainRunning(chain) {
		logger.Infoln("Chain already started. Skipping.")
	} else {
		chain.Service.Command = ErisChainStart
		util.OverwriteOps(chain.Operations, ops)
		chain.Service.Environment = append(chain.Service.Environment, "CHAIN_ID="+chain.ChainID)
		if err := services.StartServiceByService(chain.Service, chain.Operations, []string{}); err != nil {
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
		if err := services.KillServiceByService(chain.Service, chain.Operations, []string{}, true, rm, rmData); err != nil {
			return err
		}
	} else {
		logger.Infoln("Chain not currently running. Skipping.")
	}

	if rm {
		if err := perform.DockerRemove(chain.Service, chain.Operations, rmData); err != nil {
			return err
		}
	}
	return nil
}
