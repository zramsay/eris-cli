package chains

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"
)

func StartChainRaw(do *definitions.Do) error {
	logger.Infoln("Ensuring Key Server is Started.")
	keysService, err := loaders.LoadServiceDefinition("keys", 1)
	if err != nil {
		return err
	}

	err = perform.DockerRun(keysService.Service, keysService.Operations)
	if err != nil {
		return err
	}

	chain, err := loaders.LoadChainDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		logger.Infoln("Cannot start a chain I cannot find. Failing silently.")
		do.Result = "no file"
		return nil
	}

	if chain.Name == "" {
		logger.Infoln("Cannot start a chain without a name.")
		do.Result = "no name"
		return nil
	}

	chain.Service.Command = ErisChainStart
	util.OverwriteOps(chain.Operations, do.Operations)
	chain.Service.Environment = append(chain.Service.Environment, "CHAIN_ID="+chain.ChainID)

	logger.Infof("StartChainRaw to DockerRun =>\t%s\n", chain.Service.Name)
	logger.Debugf("\twith ChainID =>\t\t%v\n", chain.ChainID)
	logger.Debugf("\twith Environment =>\t%v\n", chain.Service.Environment)
	if err := perform.DockerRun(chain.Service, chain.Operations); err != nil {
		do.Result = "error"
		return err
	}

	return nil
}

func LogsChainRaw(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}
	err = perform.DockerLogs(chain.Service, chain.Operations, do.Follow, do.Tail)
	if err != nil {
		return err
	}
	return nil
}

func ExecChainRaw(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}

	if IsChainExisting(chain) {
		logger.Infoln("Chain exists.")
		return perform.DockerExec(chain.Service, chain.Operations, do.Args, do.Interactive)
	} else {
		return fmt.Errorf("Chain does not exist. Please start the chain container with eris chains start %s.\n", do.Name)
	}

	return nil
}

func KillChainRaw(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}

	if IsChainRunning(chain) {
		if err := perform.DockerStop(chain.Service, chain.Operations); err != nil {
			return err
		}
	} else {
		logger.Infoln("Chain not currently running. Skipping.")
	}

	if do.Rm {
		if err := perform.DockerRemove(chain.Service, chain.Operations, do.RmD); err != nil {
			return err
		}
	}

	return nil
}
