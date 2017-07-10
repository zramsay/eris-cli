package pkgs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/monax/monax/definitions"
	"github.com/monax/monax/loaders"
	"github.com/monax/monax/log"
	"github.com/monax/monax/pkgs/jobs"
	"github.com/monax/monax/services"
	"github.com/monax/monax/util"
)

func RunPackage(do *definitions.Do) error {
	var gotwd string
	if do.Path == "" {
		var err error
		gotwd, err = os.Getwd()
		if err != nil {
			return err
		}
		do.Path = gotwd
	}

	if do.ChainURL == "" {
		// sets do.ChainIP and do.ChainPort if do.ChainURL is not populated
		if err := setChainIPandPort(do); err != nil {
			return err
		}
		do.ChainURL = fmt.Sprintf("tcp://%s:%s", do.ChainIP, do.ChainPort)
	}

	if err := util.GetChainID(do); err != nil {
		return err
	}

	// note: [zr] this could be problematic with a combo of
	// other flags, however, at least the --dir flag isn't
	// completely broken now
	if do.Path != gotwd {
		originalYAMLPath := do.YAMLPath

		// if --dir given, assume *.yaml is in there
		do.YAMLPath = filepath.Join(do.Path, originalYAMLPath)

		// if do.YAMLPath does not exist, try $pwd
		if _, err := os.Stat(do.YAMLPath); os.IsNotExist(err) {
			do.YAMLPath = filepath.Join(gotwd, originalYAMLPath)
		}

		// if it still cannot be found, abort
		if _, err := os.Stat(do.YAMLPath); os.IsNotExist(err) {
			return fmt.Errorf("could not find jobs file (%s), ensure correct used of the --file flag")
		}

		if do.BinPath == "./bin" {
			do.BinPath = filepath.Join(do.Path, "bin")
		}
		if do.ABIPath == "./abi" {
			do.ABIPath = filepath.Join(do.Path, "abi")
		}
		// TODO enable this feature
		// if do.ContractsPath == "./contracts" {
		//do.ContractsPath = filepath.Join(do.Path, "contracts")
		//}
	}

	// useful for debugging
	printPathPackage(do)

	var err error
	// Load the package if it doesn't exist
	if do.Package == nil {
		do.Package, err = loaders.LoadPackage(do.YAMLPath)
		if err != nil {
			return err
		}
	}

	if do.Path != gotwd {
		for _, job := range do.Package.Jobs {
			if job.Deploy != nil {
				job.Deploy.Contract = filepath.Join(do.Path, job.Deploy.Contract)
			}
		}
	}

	// should be EnsureRunning
	if err := bootCompiler(); err != nil {
		return err
	}
	if err = getLocalCompilerData(do); err != nil {
		return err
	}

	return jobs.RunJobs(do)
}

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
	var err error
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		IPAddress, err = util.DockerWindowsAndMacIP(do)
		if err != nil {
			return err
		}
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
	log.WithField("=>", do.ChainName).Info("Using ChainName")
	log.WithField("=>", do.ChainID).Info("With ChainID")
	log.WithField("=>", do.ChainURL).Info("With ChainURL")
	log.WithField("=>", do.Signer).Info("Using Signer at")
}
