package pkgs

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/pkgs/jobs"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"
)

func RunPackage(do *definitions.Do) error {
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

	if do.LocalCompiler {
		if err := bootCompiler(); err != nil {
			return err
		}
		getLocalCompilerData(do)
	}

	return jobs.RunJobs(do)
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

func bootCompiler() error {

	// add the compilers to the local services if the flag is pushed
	// [csk] note - when we move to default local compilers we'll remove
	// the compilers service completely and this will need to get
	// reworked to utilize DockerRun with a populated service def.
	doComp := definitions.NowDo()
	doComp.Name = "compilers"
	return services.StartService(doComp)
}

// getLocalCompilerData populates the IP:port combo for the compilers.
func getLocalCompilerData(do *definitions.Do) {
	// [csk]: note this is brittle we should only expose one port in the
	// docker file by default for the compilers service we can expose more
	// forcibly

	do.Compiler = "http://compilers:9099"
}
