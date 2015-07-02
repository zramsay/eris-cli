package chains

import (
	"fmt"
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
		var ok bool
		chain.Service.Command, ok = chain.Manager["start"]
		if !ok {
			return fmt.Errorf("%s service definition must include '%s' command under Manager", chain.Type, "start")
		}
		if err := services.StartServiceByService(chain.Service, chain.Operations); err != nil {
			return err
		}
	}
	return nil
}

func LogsChainRaw(chainName string, follow bool, lines, containerNumber int) error {
	chain, err := LoadChainDefinition(chainName, containerNumber)
	if err != nil {
		return err
	}
	err = services.LogsServiceByService(chain.Service, chain.Operations, follow)
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

func KillChainRaw(chainName string, rm, data bool, containerNumber int) error {
	chain, err := LoadChainDefinition(chainName, containerNumber)
	if err != nil {
		return err
	}

	if IsChainRunning(chain) {
		err := services.KillServiceByService(true, rm, data, chain.Service, chain.Operations)
		if err != nil {
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

	if data {
		// TODO: data container from chain container
	}

	return nil
}
