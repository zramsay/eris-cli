package pkgs

import (
	"fmt"
	"runtime"

	"github.com/monax/cli/definitions"
	"github.com/monax/cli/loaders"
	"github.com/monax/cli/log"
	"github.com/monax/cli/pkgs/jobs"
	"github.com/monax/cli/services"
	"github.com/monax/cli/util"
)

func RunPackage(do *definitions.Do) error {

	// clears job_outputs file
	jobs.ClearJobResults()
	// useful for debugging
	printPathPackage(do)

	// sets do.ChainIP and do.ChainPort
	if err := setChainIPandPort(do); err != nil {
		return err
	}

	do.ChainURL = fmt.Sprintf("tcp://%s:%s", do.ChainIP, do.ChainPort)
	if err := util.GetChainID(do); err != nil {
		return err
	}

	var err error
	// Load the package if it doesn't exist
	if do.Package == nil {
		do.Package, err = loaders.LoadPackage(do.YAMLPath)
		if err != nil {
			return err
		}
	}

	if !do.RemoteCompiler {
		if err := bootCompiler(); err != nil {
			return err
		}
		if err = getLocalCompilerData(do); err != nil {
			return err
		}
	}

	return jobs.RunJobs(do)
}

func setChainIPandPort(do *definitions.Do) error {

	if !util.IsChain(do.ChainName, true) {
		return fmt.Errorf("chain (%s) is not running", do.ChainName)
	}
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		do.ChainIP = "127.0.0.1"
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

func bootCompiler() error {

	// add the compilers to the local services if the flag is pushed
	// [csk] note - when we move to default local compilers we'll remove
	// the compilers service completely and this will need to get
	// reworked to utilize DockerRun with a populated service def.
	doComp := definitions.NowDo()
	doComp.Name = "compilers"
	return services.EnsureRunning(doComp)
}

// getLocalCompilerData populates the IP:port combo for the compilers.
func getLocalCompilerData(do *definitions.Do) error {
	// [csk]: note this is brittle we should only expose one port in the
	// docker file by default for the compilers service we can expose more
	// forcibly
	var IPAddress string
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		IPAddress = "127.0.0.1"
	} else {
		containerName := util.ServiceContainerName("compilers")

		cont, err := util.DockerClient.InspectContainer(containerName)
		if err != nil {
			return util.DockerError(err)
		}

		IPAddress = cont.NetworkSettings.IPAddress
	}

	do.Compiler = fmt.Sprintf("http://%s:9099", IPAddress)
	return nil
}

func printPathPackage(do *definitions.Do) {
	log.WithField("=>", do.Compiler).Info("Using Compiler at")
	log.WithField("=>", do.ChainName).Info("Using Chain at")
	log.WithField("=>", do.ChainID).Debug("With ChainID")
	log.WithField("=>", do.Signer).Info("Using Signer at")
}
