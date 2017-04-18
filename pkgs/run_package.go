package pkgs

import (
	"fmt"
	"os"
	"runtime"

	"github.com/monax/cli/definitions"
	"github.com/monax/cli/loaders"
	"github.com/monax/cli/log"
	"github.com/monax/cli/util"
)

// Run a package of smart contracts based on inputs from the CLI
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

	// go to the directory where the yaml file is, it makes it far easier to
	// resolve pathways to contracts and whatnot
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if do.YAMLPath != "./epm.yaml" {
		if err = os.Chdir(do.YAMLPath); err != nil {
			return err
		}
		defer os.Chdir(pwd)
	}

	if _, err := os.Stat(do.ABIPath); os.IsNotExist(err) {
		if err := os.Mkdir(do.ABIPath, 0775); err != nil {
			return err
		}
	}
	if _, err := os.Stat(do.BinPath); os.IsNotExist(err) {
		if err := os.Mkdir(do.BinPath, 0775); err != nil {
			return err
		}
	}

	return loadedJobs.RunJobs()
}

// Sets the chain networking information for later interactions via the jobs
func setChainIPandPort(do *definitions.Do) error {

	if !util.IsChain(do.ChainName, true) {
		return fmt.Errorf("chain (%s) is not running", do.ChainName)
	}
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		var err error
		do.ChainIP, err = util.DockerWindowsAndMacIP(do)
		if err != nil {
			return err
		}
	} else {
		containerName := util.ContainerName(definitions.TypeChain, do.ChainName)

		cont, err := util.DockerClient.InspectContainer(containerName)
		if err != nil {
			return util.DockerError(err)
		}

		do.ChainIP = cont.NetworkSettings.IPAddress
	}
	do.ChainPort = "46657" // [zr] this can be hardcoded even if [--publish] is used

	return nil
}

// Utility function for printing out nice information about what we're getting into
func printPathPackage(do *definitions.Do) {
	log.WithField("=>", do.ChainName).Info("Using Chain at")
	log.WithField("=>", do.ChainURL).Debug("Through Chain URL")
	log.WithField("=>", do.Signer).Info("Using Signer at")
}
