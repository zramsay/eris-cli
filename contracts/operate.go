package contracts

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

var pwd string

func RunPackage(do *definitions.Do) error {
	logger.Debugf("Welcome! Say the Marmots. Running App package.\n")
	var err error
	pwd, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("Could not get the present working directory. Are you on Mars?\nError: %v\n", err)
	}

	logger.Debugf("\twith Host Path =>\t%s:%s\n", do.Path, pwd)
	app, err := loaders.LoadContractPackage(do.Path, do.ChainName, do.Name, do.Type)
	if err != nil {
		do.Result = "could not load package"
		return err
	}

	if err := BootServicesAndChain(do, app); err != nil {
		do.Result = "could not boot chain or services"
		CleanUp(do, app)
		return err
	}

	do.Path = pwd
	if err := DefineAppActionService(do, app); err != nil {
		do.Result = "could not define app action service"
		CleanUp(do, app)
		return err
	}

	if err := PerformAppActionService(do, app); err != nil {
		do.Result = "could not perform app action service"
		CleanUp(do, app)
		return err
	}

	do.Result = "success"
	return CleanUp(do, app)
}

func BootServicesAndChain(do *definitions.Do, app *definitions.Contracts) error {
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
		if app.ChainName == "" {
			// TODO [csk]: first check if there is a chain checked out. if not, then use throwAway
			logger.Infof("No chain was given, booting a throwaway chain.\n")
			err = bootThrowAwayChain(app.Name, do)
		} else {
			logger.Infof("Booting chain =>\t\t%s\n", app.ChainName)
			err = bootChain(app.ChainName, do)
		}
	case "t", "tmp", "temp":
		logger.Infof("No chain was given, booting a throwaway chain.\n")
		err = bootThrowAwayChain(app.Name, do)
	default:
		logger.Infof("Booting chain =>\t\t%s\n", do.ChainName)
		err = bootChain(do.ChainName, do)
	}

	app.ChainName = do.Chain.Name
	if err != nil {
		return err
	}

	return nil
}

func DefineAppActionService(do *definitions.Do, app *definitions.Contracts) error {
	var cmd string

	switch do.Name {
	case "test":
		cmd = app.AppType.TestCmd
	case "deploy":
		cmd = app.AppType.DeployCmd
	default:
		return fmt.Errorf("I do not know how to perform that task (%s)\nPlease check what you can do with contracts by typing [eris contracts].\n", do.Name)
	}

	// if manual, set task
	if app.AppType.Name == "manual" {
		switch do.Name {
		case "test":
			cmd = app.TestTask
		case "deploy":
			cmd = app.DeployTask
		}
	}

	// task flag override
	if do.Task != "" {
		app.AppType = definitions.GulpApp()
		cmd = do.Task
	}

	if cmd == "nil" {
		return fmt.Errorf("I cannot perform that task against that app type.\n")
	}

	// build service that will run
	do.Service.Name = app.Name + "_tmp_" + do.Name
	do.Service.Image = app.AppType.BaseImage
	do.Service.AutoData = true
	do.Service.EntryPoint = app.AppType.EntryPoint
	do.Service.Command = cmd
	if do.Path != pwd {
		do.Service.WorkDir = do.Path // do.Path is actually where the workdir inside the container goes
	} else {
		do.Service.WorkDir = path.Join(common.ErisContainerRoot, "apps", app.Name)
	}
	do.Service.User = "eris"

	srv := definitions.BlankServiceDefinition()
	srv.Service = do.Service
	srv.Operations = do.Operations
	loaders.ServiceFinalizeLoad(srv)
	do.Service = srv.Service
	do.Operations = srv.Operations
	do.Operations.Follow = true

	linkAppToChain(do, app)

	if app.AppType.Name == "epm" {
		prepareEpmAction(do, app)
	}

	// make data container and import do.Path to do.Path (if exists)
	doData := definitions.NowDo()
	doData.Name = do.Service.Name
	doData.Operations = do.Operations
	if do.Path != pwd {
		doData.Path = do.Path
	} else {
		doData.Path = path.Join(common.ErisContainerRoot)
	}
	var loca string
	if do.Path != pwd {
		loca = path.Join(common.DataContainersPath, doData.Name, do.Path)
	} else {
		loca = path.Join(common.DataContainersPath, doData.Name, "apps", app.Name)
	}
	logger.Debugf("Creating App Data Cont =>\t%s:%s\n", do.Path, loca)
	common.Copy(do.Path, loca)
	if err := data.ImportData(doData); err != nil {
		return err
	}
	do.Operations.DataContainerName = util.DataContainersName(doData.Name, doData.Operations.ContainerNumber)

	logger.Debugf("App Action Built.\n")
	return nil
}

func PerformAppActionService(do *definitions.Do, app *definitions.Contracts) error {
	logger.Println("Performing Action. This can sometimes take a wee while.")
	logger.Infof("\t=>\t\t\t%s:%s\n", do.Service.Name, do.Service.Image)
	logger.Debugf("\t=>\t\t\t%s:%s\n", do.Service.WorkDir, do.Service.EntryPoint)

	do.Operations.ContainerType = definitions.TypeService
	if err := perform.DockerExecService(do.Service, do.Operations); err != nil {
		do.Result = "could not perform app action"
		return err
	}

	logger.Infof("Finished performing App Action.\n")
	return nil
}

func CleanUp(do *definitions.Do, app *definitions.Contracts) error {
	logger.Infof("Commensing CleanUp.\n")

	if do.Chain.ChainType == "throwaway" {
		logger.Debugf("Destroying Throwaway Chain =>\t%s\n", do.Chain.Name)
		doRm := definitions.NowDo()
		doRm.Operations = do.Operations
		doRm.Name = do.Chain.Name
		doRm.Rm = true
		doRm.RmD = true
		chains.KillChain(doRm)

		logger.Debugf("Removing latent files/dirs =>\t%s:%s\n", path.Join(common.DataContainersPath, do.Chain.Name), path.Join(common.ChainsPath, do.Chain.Name+".toml"))
		os.RemoveAll(path.Join(common.DataContainersPath, do.Chain.Name))
		os.Remove(path.Join(common.ChainsPath, do.Chain.Name+".toml"))
	} else {
		logger.Debugf("No Throwaway Chain to destroy.\n")
	}

	doData := definitions.NowDo()
	doData.Name = do.Service.Name
	doData.Operations = do.Operations
	if do.Path != pwd {
		doData.Path = do.Path
	} else {
		doData.Path = path.Join(common.ErisContainerRoot, "apps", app.Name)
	}
	var loca string
	if do.Path != pwd {
		loca = path.Join(common.DataContainersPath, doData.Name, do.Path)
	} else {
		loca = path.Join(common.DataContainersPath, doData.Name, "apps", app.Name)
	}

	logger.Debugf("Exporting Results =>\t\t%s:%s\n", doData.Path, loca)
	data.ExportData(doData)

	if app.AppType.Name == "epm" {
		files, _ := filepath.Glob(filepath.Join(loca, "epm*"))
		for _, f := range files {
			dest := filepath.Join(do.Path, filepath.Base(f))
			logger.Debugf("Moving file =>\t\t\t%s:%s\n", f, dest)
			common.Copy(f, dest)
		}
	}

	if !do.RmD {
		logger.Debugf("Removing data dir on host =>\t%s\n", path.Join(common.DataContainersPath, do.Service.Name))
		os.RemoveAll(path.Join(common.DataContainersPath, do.Service.Name))
	}

	if !do.Rm {
		doRemove := definitions.NowDo()
		doRemove.Operations.SrvContainerName = do.Operations.DataContainerName
		logger.Debugf("Removing data contnr =>\t\t%s\n", doRemove.Operations.SrvContainerName)
		if err := perform.DockerRemove(nil, doRemove.Operations, false, true); err != nil {
			return err
		}
	}

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
			startService.Operations.Args = []string{name}
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

func linkAppToChain(do *definitions.Do, app *definitions.Contracts) {
	var newLink string

	if do.Chain.ChainType == "service" {
		newLink = util.ServiceContainersName(app.ChainName, do.Operations.ContainerNumber) + ":" + "chain"
	} else {
		newLink = util.ChainContainersName(app.ChainName, do.Operations.ContainerNumber) + ":" + "chain"
	}
	newLink2 := util.ServiceContainersName("keys", do.Operations.ContainerNumber) + ":" + "keys"
	do.Service.Links = append(do.Service.Links, newLink)
	do.Service.Links = append(do.Service.Links, newLink2)
}

func prepareEpmAction(do *definitions.Do, app *definitions.Contracts) {
	if do.Verbose {
		do.Service.Environment = append(do.Service.Environment, "EPM_VERBOSE=true")
	}
	if do.Debug {
		do.Service.Environment = append(do.Service.Environment, "EPM_DEBUG=true")
	}

	if do.CSV != "" {
		logger.Debugf("Setting output format to =>\t%s\n", do.CSV)
		do.Service.EntryPoint = do.Service.EntryPoint + " --output " + do.CSV
	} else {
		do.Service.EntryPoint = do.Service.EntryPoint + " --output json"
	}

	if do.EPMConfigFile != "" {
		logger.Debugf("Setting config file to =>\t%s\n", do.EPMConfigFile)
		do.Service.EntryPoint = do.Service.EntryPoint + " --file " + path.Join(do.Service.WorkDir, do.EPMConfigFile)
	}

	if len(do.ConfigOpts) != 0 {
		var toAdd string
		logger.Debugf("Setting sets file to =>\t%v\n", do.ConfigOpts)
		for _, s := range do.ConfigOpts {
			toAdd = toAdd + "," + s
		}
		do.Service.EntryPoint = do.Service.EntryPoint + " --set " + toAdd
	}

	if do.OutputTable {
		do.Service.EntryPoint = do.Service.EntryPoint + " --summary "
	}

	if do.ContractsPath != "" {
		do.Service.EntryPoint = do.Service.EntryPoint + " --contracts-path " + path.Join(do.Service.WorkDir, do.ContractsPath)
	}

	if do.ABIPath != "" {
		do.Service.EntryPoint = do.Service.EntryPoint + " --abi-path " + path.Join(do.Service.WorkDir, do.ABIPath)
	}

	if do.DefaultGas != "" {
		do.Service.EntryPoint = do.Service.EntryPoint + " --gas " + do.DefaultGas
	}

	if do.Compiler != "" {
		do.Service.EntryPoint = do.Service.EntryPoint + " --compiler " + do.Compiler
	}

	if do.DefaultAddr != "" {
		do.Service.EntryPoint = do.Service.EntryPoint + " --address " + do.DefaultAddr
	}

	if do.DefaultFee != "" {
		do.Service.EntryPoint = do.Service.EntryPoint + " --fee " + do.DefaultFee
	}

	if do.DefaultAmount != "" {
		do.Service.EntryPoint = do.Service.EntryPoint + " --amount " + do.DefaultAmount
	}
}
