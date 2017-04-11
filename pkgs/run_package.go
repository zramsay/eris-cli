package pkgs

import (
	"fmt"

	"github.com/monax/cli/definitions"
	"github.com/monax/cli/loaders"
	"github.com/monax/cli/log"
	"github.com/monax/cli/util"
)

func RunPackage(do *definitions.Do) error {
	// sets do.ChainIP and do.ChainPort
	if err := setChainIPandPort(do); err != nil {
		return err
	}

	printPathPackage(do)

	do.ChainURL = fmt.Sprintf("tcp://%s:%s", do.ChainIP, do.ChainPort)

	loadedJobs, err := loaders.LoadJobs(do)
	if err != nil {
		return err
	}

	return loadedJobs.RunJobs()
}

func setChainIPandPort(do *definitions.Do) error {

	if !util.IsChain(do.ChainName, true) {
		return fmt.Errorf("chain (%s) is not running", do.ChainName)
	}

	containerName := util.ContainerName(definitions.TypeChain, do.ChainName)

	cont, err := util.DockerClient.InspectContainer(containerName)
	if err != nil {
		return util.DockerError(err)
	}

	do.ChainIP = cont.NetworkSettings.IPAddress
	do.ChainPort = "46657" // [zr] this can be hardcoded even if [--publish] is used

	return nil
}

func printPathPackage(do *definitions.Do) {
	log.WithField("=>", do.Compiler).Info("Using Compiler at")
	log.WithField("=>", do.ChainName).Info("Using Chain at")
	log.WithField("=>", do.ChainID).Debug("With ChainID")
	log.WithField("=>", do.Signer).Info("Using Signer at")
}
