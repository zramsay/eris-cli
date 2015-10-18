package contracts

import (
	"fmt"
	"os"
	"path"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func RunPackage(do *definitions.Do) error {
	logger.Debugf("Welcome! Say the Marmots. Running DApp package.\n")
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Could not get the present working directory. Are you on Mars?\nError: %v\n", err)
	}

	logger.Debugf("\twith Host Path =>\t%s:%s\n", do.Path, pwd)
	dapp, err := loaders.LoadContractPackage(do.Path, do.ChainName, do.Name, do.Type)
	if err != nil {
		do.Result = "could not load package"
		return err
	}

	if err := BootServicesAndChain(do, dapp); err != nil {
		do.Result = "could not boot chain or services"
		CleanUp(do, dapp)
		return err
	}

	do.Path = pwd
	if err := DefineDappActionService(do, dapp); err != nil {
		do.Result = "could not define dapp action service"
		CleanUp(do, dapp)
		return err
	}

	if err := PerformDappActionService(do, dapp); err != nil {
		do.Result = "could not perform dapp action service"
		CleanUp(do, dapp)
		return err
	}

	if err := CleanUp(do, dapp); err != nil {
		do.Result = "could not cleanup"
		return err
	}

	do.Result = "success"
	return nil
}

func BootServicesAndChain(do *definitions.Do, dapp *definitions.Contracts) error {
	var err error
	var srvs []*definitions.ServiceDefinition

	// launch the services
	for _, s := range do.ServicesSlice {
		t, err := services.BuildServicesGroup(s, do.Operations.ContainerNumber, srvs...)
		if err != nil {
			return err
		}
		srvs = append(srvs, t...)
	}

	if len(srvs) >= 1 {
		if err := services.StartGroup(srvs); err != nil {
			return err
		}
	}

	// boot the chain
	switch do.ChainName {
	case "":
		if dapp.ChainName == "" {
			logger.Infof("No chain was given, booting a throwaway chain.\n")
			err = bootThrowAwayChain(dapp.Name, do)
		} else {
			logger.Infof("Booting chain =>\t\t%s\n", dapp.ChainName)
			err = bootChain(dapp.ChainName, do)
		}
	case "t", "tmp", "temp":
		logger.Infof("No chain was given, booting a throwaway chain.\n")
		err = bootThrowAwayChain(dapp.Name, do)
	default:
		logger.Infof("Booting chain =>\t\t%s\n", do.ChainName)
		err = bootChain(do.ChainName, do)
	}

	dapp.ChainName = do.Chain.Name
	if err != nil {
		return err
	}

	return nil
}

func DefineDappActionService(do *definitions.Do, dapp *definitions.Contracts) error {
	var cmd string

	switch do.Name {
	case "test":
		cmd = dapp.DappType.TestCmd
	case "deploy":
		cmd = dapp.DappType.DeployCmd
	default:
		return fmt.Errorf("I do not know how to perform that task (%s)\nPlease check what you can do with contracts by typing [eris contracts].\n", do.Name)
	}

	// if manual, set task
	if dapp.DappType.Name == "manual" {
		switch do.Name {
		case "test":
			cmd = dapp.TestTask
		case "deploy":
			cmd = dapp.DeployTask
		}
	}

	// task flag override
	if do.Task != "" {
		dapp.DappType = definitions.GulpDapp()
		cmd = do.Task
	}

	if cmd == "nil" {
		return fmt.Errorf("I cannot perform that task against that dapp type.\n")
	}

	// dapp-specific tests
	if dapp.DappType.Name == "pyepm" {
		if do.ConfigFile == "" {
			return fmt.Errorf("The pyepm dapp type requires a --yaml flag for the package definition you would like to deploy.\n")
		} else {
			cmd = do.ConfigFile
		}
	}

	// build service that will run
	do.Service.Name = dapp.Name + "_tmp_" + do.Name
	do.Service.Image = dapp.DappType.BaseImage
	do.Service.AutoData = true
	do.Service.EntryPoint = dapp.DappType.EntryPoint
	do.Service.Command = cmd
	if do.NewName != "" {
		do.Service.WorkDir = do.NewName // do.NewName is actually where the workdir inside the container goes
	}
	do.Service.User = "eris"

	srv := definitions.BlankServiceDefinition()
	srv.Service = do.Service
	srv.Operations = do.Operations
	loaders.ServiceFinalizeLoad(srv)
	do.Service = srv.Service
	do.Operations = srv.Operations
	do.Operations.Remove = true

	linkDappToChain(do, dapp)

	// make data container and import do.Path to do.NewName (if exists)
	doData := definitions.NowDo()
	doData.Name = do.Service.Name
	doData.Operations = do.Operations
	if do.NewName != "" {
		doData.Path = do.NewName
	}

	loca := path.Join(common.DataContainersPath, doData.Name)
	logger.Debugf("Creating Dapp Data Cont =>\t%s:%s\n", do.Path, loca)
	common.Copy(do.Path, loca)
	data.ImportData(doData)
	do.Operations.DataContainerName = util.DataContainersName(doData.Name, doData.Operations.ContainerNumber)

	logger.Debugf("DApp Action Built.\n")

	return nil
}

func PerformDappActionService(do *definitions.Do, dapp *definitions.Contracts) error {
	logger.Infof("Performing DAPP Action =>\t%s:%s:%s\n", do.Service.Name, do.Service.Image, do.Service.Command)

	if err := perform.DockerRun(do.Service, do.Operations); err != nil {
		do.Result = "could not perform dapp action"
		return err
	}

	logger.Infof("Finished performing DAPP Action.\n")
	return nil
}

func CleanUp(do *definitions.Do, dapp *definitions.Contracts) error {
	logger.Infof("Commensing CleanUp.\n")

	if do.Chain.ChainType == "throwaway" {
		logger.Debugf("Destroying Throwaway Chain =>\t%s\n", do.Chain.Name)
		doRm := definitions.NowDo()
		doRm.Operations = do.Operations
		doRm.Name = do.Chain.Name
		doRm.Rm = true
		doRm.RmD = true
		chains.KillChain(doRm)

		logger.Debugf("Removing latent files/dirs =>\t%s:%s\n", path.Join(common.DataContainersPath, do.Chain.Name), path.Join(common.BlockchainsPath, do.Chain.Name+".toml"))
		os.RemoveAll(path.Join(common.DataContainersPath, do.Chain.Name))
		os.Remove(path.Join(common.BlockchainsPath, do.Chain.Name+".toml"))
	} else {
		logger.Debugf("No Throwaway Chain to destroy.\n")
	}

	logger.Debugf("Removing data dir on host =>\t%s\n", path.Join(common.DataContainersPath, do.Service.Name))
	os.RemoveAll(path.Join(common.DataContainersPath, do.Service.Name))

	logger.Debugf("Removing tmp srv contnr =>\t%s\n", do.Operations.SrvContainerName)
	perform.DockerRemove(do.Service, do.Operations, true, true)
	return nil
}

func bootChain(name string, do *definitions.Do) error {
	do.Chain.ChainType = "service" // setting this for tear down purposes
	startChain := definitions.NowDo()
	startChain.Name = name
	startChain.Operations = do.Operations
	err := chains.StartChain(startChain)
	// errors *could* be because the chain was actually a service.
	if err != nil {
		if util.IsServiceContainer(name, do.Operations.ContainerNumber, true) {
			startService := definitions.NowDo()
			startService.Operations = do.Operations
			startService.Args = []string{name}
			err = services.StartService(startService)
			if err != nil {
				return err
			}
		}
	} else {
		do.Chain.ChainType = "chain" // setting this for tear down purposes
	}
	do.Chain.Name = name // setting this for tear down purposes
	return nil
}

func bootThrowAwayChain(name string, do *definitions.Do) error {
	do.Chain.ChainType = "throwaway"

	tmp := do.Name
	do.Name = name
	err := chains.ThrowAwayChain(do)
	if err != nil {
		do.Name = tmp
		return err
	}

	do.Chain.Name = do.Name // setting this for tear down purposes
	logger.Debugf("ThrowAwayChain booted =>\t%s\n", do.Name)

	do.Name = tmp
	return nil
}

func linkDappToChain(do *definitions.Do, dapp *definitions.Contracts) {
	var newLink string

	if do.Chain.ChainType == "service" {
		newLink = util.ServiceContainersName(dapp.ChainName, do.Operations.ContainerNumber) + ":" + "chain"
	} else {
		newLink = util.ChainContainersName(dapp.ChainName, do.Operations.ContainerNumber) + ":" + "chain"
	}
	do.Service.Links = append(do.Service.Links, newLink)
}
