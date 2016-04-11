package pkgs

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

var pwd string

func RunPackage(do *definitions.Do) error {
	log.Debug("Welcome! Say the marmots. Running package")
	var err error
	pwd, err = os.Getwd()
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"host path": do.Path,
		"pwd":       pwd,
	}).Debug()
	pkg, err := loaders.LoadPackage(do.Path, do.ChainName)
	if err != nil {
		do.Result = "could not load package"
		return err
	}

	if err := BootServicesAndChain(do, pkg); err != nil {
		do.Result = "could not boot chain or services"
		CleanUp(do, pkg)
		return err
	}

	if err := DefinePkgActionService(do, pkg); err != nil {
		do.Result = "could not define pkg action service"
		CleanUp(do, pkg)
		return err
	}

	if err := PerformAppActionService(do, pkg); err != nil {
		do.Result = "could not perform pkg action service"
		CleanUp(do, pkg)
		return err
	}

	do.Result = "success"
	return CleanUp(do, pkg)
}

func BootServicesAndChain(do *definitions.Do, pkg *definitions.Package) error {
	var err error
	var srvs []*definitions.ServiceDefinition
	do.ServicesSlice = append(do.ServicesSlice, pkg.Dependencies.Services...)

	// assemble the services
	for _, s := range do.ServicesSlice {
		t, err := services.BuildServicesGroup(s, srvs...)
		if err != nil {
			return err
		}
		srvs = append(srvs, t...)
	}

	// boot the services
	if len(srvs) >= 1 {
		if err := services.StartGroup(srvs); err != nil {
			return err
		}
	}

	// boot the chain
	switch do.ChainName { // switch on the flag
	case "":
		switch pkg.ChainName { // switch on the package.json
		case "":
			head, _ := util.GetHead() // checks the checkedout chain
			if head != "" {           // used checked out chain
				log.WithField("=>", head).Info("No chain flag or in package file. Booting chain from checked out chain")
				err = bootChain(head, do)
			} else { // if no chain is checked out and no --chain given, default to a throwaway
				log.Info("No chain was given, booting a throwaway chain")
				err = bootThrowAwayChain(pkg.Name, do)
			}
		case "$chain":
			head, _ := util.GetHead() // checks the checkedout chain
			if head != "" {           // used checked out chain
				log.WithField("=>", head).Info("No chain flag or in package file. Booting chain from checked out chain")
				err = bootChain(head, do)
			} else { // if no chain is checked out and no --chain given, default to a throwaway
				return fmt.Errorf("The package definition file needs a checked out chain to continue. Please check out the appropriate chain or rerun with a chain flag")
			}
		case "t", "tmp", "temp", "temporary", "throwaway", "thr", "throw":
			log.Info("No chain was given, booting a throwaway chain")
			err = bootThrowAwayChain(pkg.Name, do)
		default:
			log.WithField("=>", pkg.ChainName).Info("No chain flag used. Booting chain from package file")
			err = bootChain(pkg.ChainName, do)
		}
	case "t", "tmp", "temp", "temporary", "throwaway", "thr", "throw":
		log.Info("No chain was given, booting a throwaway chain")
		err = bootThrowAwayChain(pkg.Name, do)
	default:
		log.WithField("=>", do.ChainName).Info("Booting chain from chain flag")
		err = bootChain(do.ChainName, do)
	}

	pkg.ChainName = do.Chain.Name
	if err != nil {
		return err
	}

	return nil
}

// build service that will run
func DefinePkgActionService(do *definitions.Do, pkg *definitions.Package) error {
	do.Service.Name = pkg.Name + "_tmp_" + do.Name
	do.Service.Image = path.Join(version.ERIS_REG_DEF, version.ERIS_IMG_PM)
	do.Service.AutoData = true
	do.Service.EntryPoint = "epm --chain chain:46657 --sign keys:4767"
	do.Service.WorkDir = path.Join(common.ErisContainerRoot, "apps", filepath.Base(do.Path))
	do.Service.User = "eris"

	srv := definitions.BlankServiceDefinition()
	srv.Service = do.Service
	srv.Operations = do.Operations
	loaders.ServiceFinalizeLoad(srv)
	do.Service = srv.Service
	do.Operations = srv.Operations
	do.Operations.Follow = true

	prepareEpmAction(do, pkg)
	linkAppToChain(do, pkg)

	log.Debug("App action built")
	return nil
}

func PerformAppActionService(do *definitions.Do, app *definitions.Package) error {
	if err := getDataContainerSorted(do, true); err != nil {
		return err
	}

	log.Warn("Performing action. This can sometimes take a wee while")
	log.WithFields(log.Fields{
		"service": do.Service.Name,
		"image":   do.Service.Image,
	}).Info()
	log.WithFields(log.Fields{
		"workdir":    do.Service.WorkDir,
		"entrypoint": do.Service.EntryPoint,
	}).Debug()

	do.Operations.ContainerType = definitions.TypeService
	buf, err := perform.DockerExecService(do.Service, do.Operations)
	if err != nil {
		log.Error(buf)
		do.Result = "could not perform app action"
		return err
	}

	io.Copy(config.GlobalConfig.Writer, buf)

	log.Info("Finished performing action")
	return nil
}

func CleanUp(do *definitions.Do, pkg *definitions.Package) error {
	log.Info("Cleaning up")

	if do.Chain.ChainType == "throwaway" {
		log.WithField("=>", do.Chain.Name).Debug("Destroying throwaway chain")
		doRm := definitions.NowDo()
		doRm.Operations = do.Operations
		doRm.Name = do.Chain.Name
		doRm.Rm = true
		doRm.RmD = true
		chains.KillChain(doRm)

		latentDir := filepath.Join(common.DataContainersPath, do.Chain.Name)
		latentFile := filepath.Join(common.ChainsPath, do.Chain.Name+".toml")
		log.WithFields(log.Fields{
			"dir":  latentDir,
			"file": latentFile,
		}).Debug("Removing latent dir and file")

		os.RemoveAll(latentDir)
		os.Remove(latentFile)
	} else {
		log.Debug("No throwaway chain to destroy")
	}

	if err := getDataContainerSorted(do, false); err != nil {
		return err // errors marmotified in getDataContainerSorted
	}

	if !do.Rm {
		doRemove := definitions.NowDo()
		doRemove.Operations.SrvContainerName = do.Operations.DataContainerName
		log.WithField("=>", doRemove.Operations.SrvContainerName).Debug("Removing data container")
		if err := perform.DockerRemove(nil, doRemove.Operations, false, true, false); err != nil {
			return err
		}
	}

	if !do.RmD {
		log.WithField("dir", filepath.Join(common.DataContainersPath, do.Service.Name)).Debug("Removing data dir on host")
		os.RemoveAll(filepath.Join(common.DataContainersPath, do.Service.Name))
	}

	return nil
}

func bootChain(name string, do *definitions.Do) error {
	do.Chain.ChainType = "service" // setting this for tear down purposes
	startChain := definitions.NowDo()
	startChain.Name = name
	startChain.Operations = do.Operations
	f, err := os.Stat(filepath.Join(common.ChainsPath, startChain.Name))
	switch {
	case util.IsChain(name, true):
		log.WithField("name", startChain.Name).Info("Starting Chain")
		if err := chains.StartChain(startChain); err != nil {
			return err
		}
		do.Chain.ChainType = "chain" // setting this for tear down purposes
	case !os.IsNotExist(err) && f.IsDir():
		log.WithField("name", startChain.Name).Info("Trying New Chain")
		startChain.Path = filepath.Join(common.ChainsPath, startChain.Name)
		if err := chains.NewChain(startChain); err != nil {
			return err
		}
		do.Chain.ChainType = "chain" // setting this for tear down purposes
	case util.IsService(name, false):
		log.WithField("name", name).Info("Chain exists as a service")
		startService := definitions.NowDo()
		startService.Operations = do.Operations
		startService.Operations.Args = []string{name}
		err = services.StartService(startService)
		if err != nil {
			return err
		}
		do.Chain.ChainType = "service" // setting this for tear down purposes
	default:
		return fmt.Errorf("The marmots could not find that chain name. Please review and rerun the command")
	}

	do.Chain.Name = name // setting this for tear down purposes

	// let the chain boot properly
	time.Sleep(5 * time.Second)
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
	log.WithField("=>", do.Name).Debug("Throwaway chain booted")

	// let the chain boot properly
	time.Sleep(5 * time.Second)

	do.Name = tmp
	return nil
}

func linkAppToChain(do *definitions.Do, pkg *definitions.Package) {
	var newLink string

	if do.Chain.ChainType == "service" {
		newLink = util.ServiceContainerName(pkg.ChainName) + ":" + "chain"
	} else {
		newLink = util.ChainContainerName(pkg.ChainName) + ":" + "chain"
	}
	newLink2 := util.ServiceContainerName("keys") + ":" + "keys"
	do.Service.Links = append(do.Service.Links, newLink)
	do.Service.Links = append(do.Service.Links, newLink2)

	for _, s := range do.ServicesSlice {
		l := util.ServiceContainerName(s) + ":" + s
		do.Service.Links = append(do.Service.Links, l)
	}
}

func prepareEpmAction(do *definitions.Do, app *definitions.Package) {
	// todo: rework these so they all just append to the environment variables rather than use flags as it will be more stable
	if do.Verbose {
		do.Service.EntryPoint = do.Service.EntryPoint + " --verbose "

		// do.Service.Environment = append(do.Service.Environment, "EPM_VERBOSE=true")
	}
	if do.Debug {
		do.Service.EntryPoint = do.Service.EntryPoint + " --debug "

		// do.Service.Environment = append(do.Service.Environment, "EPM_DEBUG=true")
	}

	if do.CSV != "" {
		log.WithField("format", do.CSV).Debug("Setting output format to")
		do.Service.EntryPoint = do.Service.EntryPoint + " --output " + do.CSV
	} else {
		do.Service.EntryPoint = do.Service.EntryPoint + " --output json"
	}

	if len(do.ConfigOpts) != 0 {
		var toAdd string
		log.WithField("sets file", do.ConfigOpts).Debug("Setting sets file to")
		for _, s := range do.ConfigOpts {
			toAdd = toAdd + "," + s
		}
		do.Service.EntryPoint = do.Service.EntryPoint + " --set " + toAdd
	}

	if do.OutputTable {
		do.Service.EntryPoint = do.Service.EntryPoint + " --summary "
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

func getDataContainerSorted(do *definitions.Do, inbound bool) error {
	if inbound {
		log.WithField("dir", "inbound").Info("Getting data container situated")
	} else {
		log.WithField("dir", "outbound").Info("Getting data container situated")
	}

	doData := definitions.NowDo()
	doData.Name = do.Service.Name
	doData.Operations = loaders.LoadDataDefinition(doData.Name)
	util.Merge(doData.Operations, do.Operations)

	if inbound && util.Exists(definitions.TypeData, doData.Name) == false {
		doData.Operations.DataContainerName = util.DataContainerName(doData.Name)
		doData.Operations.ContainerType = definitions.TypeData
		if err := perform.DockerCreateData(doData.Operations); err != nil {
			return err
		}
	}

	oldDoPath := do.Path
	oldPkgPath := do.PackagePath
	oldAbiPath := do.ABIPath

	var err error
	do.Path, err = filepath.Abs(do.Path)
	do.PackagePath, err = filepath.Abs(do.PackagePath)
	do.ABIPath, err = filepath.Abs(do.ABIPath)
	do.EPMConfigFile, err = filepath.Abs(do.EPMConfigFile)
	if err != nil {
		return err
	}

	fi, err := os.Stat(do.Path)
	if err == nil && !fi.IsDir() {
		do.Path = filepath.Dir(do.Path)
		log.WithField("=>", do.Path).Debug("Setting do.Path")
	}

	fi, err = os.Stat(do.PackagePath)
	if err == nil && !fi.IsDir() {
		do.PackagePath = filepath.Dir(do.PackagePath)
		log.WithField("=>", do.PackagePath).Debug("Setting do.PackagePath")
	}

	fi, err = os.Stat(do.ABIPath)
	if err == nil && !fi.IsDir() {
		do.ABIPath = filepath.Dir(do.ABIPath)
		log.WithField("=>", do.ABIPath).Debug("Setting do.ABIPath")
	}

	// import contracts path (if exists)
	if _, err := os.Stat(do.Path); !os.IsNotExist(err) {
		if inbound {
			doData.Source = do.Path
			doData.Destination = path.Join(common.ErisContainerRoot, "apps", filepath.Base(do.Path))
		} else {
			doData.Source = path.Join(common.ErisContainerRoot, "apps", filepath.Base(do.Path))
			doData.Destination = filepath.Dir(do.Path) // on exports we always need the parent of the directory
		}

		log.WithFields(log.Fields{
			"source": doData.Source,
			"dest":   doData.Destination,
		}).Debug("Setting app data for container")
		if inbound {
			if err := data.ImportData(doData); err != nil {
				return err
			}
		} else {
			if err := data.ExportData(doData); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("That path does not exist. Please rerun command with a proper path")
	}

	// import contracts path (if exists)
	if _, err := os.Stat(do.PackagePath); !os.IsNotExist(err) && !strings.Contains(do.PackagePath, do.Path) {
		log.WithFields(log.Fields{
			"path":        do.Path,
			"packagePath": do.PackagePath,
		}).Debug("Package path exists and is not in pkg path. Proceeding with Import/Export sequence")
		if inbound {
			doData.Destination = path.Join(common.ErisContainerRoot, "apps", filepath.Base(do.Path), "contracts")
			doData.Source = do.PackagePath

			log.WithFields(log.Fields{
				"source": doData.Source,
				"dest":   doData.Destination,
			}).Debug("Importing contracts path.")
			if err := data.ImportData(doData); err != nil {
				return err
			}
		} else {
			log.WithFields(log.Fields{
				"source": filepath.Join(do.Path, "contracts"),
				"dest":   do.PackagePath,
			}).Debug("Moving contracts into position")
			if err := os.Rename(filepath.Join(do.Path, "contracts"), do.PackagePath); err != nil {
				return err
			}
		}
	} else if !strings.Contains(do.PackagePath, do.Path) {
		log.WithFields(log.Fields{
			"source": filepath.Join(do.Path, "contracts"),
			"dest":   do.PackagePath,
		}).Debug("Moving contracts into position")
		if err := os.Rename(filepath.Join(do.Path, "contracts"), do.PackagePath); err != nil {
			return err
		}
	} else {
		log.Info("Package path does not exist on the host or is inside the pkg path")
	}

	// import abi path (if exists)
	if _, err := os.Stat(do.ABIPath); !os.IsNotExist(err) && !strings.Contains(do.ABIPath, do.Path) {
		log.WithFields(log.Fields{
			"path":    do.Path,
			"abiPath": do.ABIPath,
		}).Debug("ABI path exists and is not in pkg path. Proceeding with Import/Export sequence")
		if inbound {
			doData.Destination = path.Join(common.ErisContainerRoot, "apps", filepath.Base(do.Path), "abi")
			doData.Source = do.ABIPath

			log.WithFields(log.Fields{
				"source": doData.Source,
				"dest":   doData.Destination,
			}).Debug("Importing ABI path")
			if err := data.ImportData(doData); err != nil {
				return err
			}
		} else {
			log.WithFields(log.Fields{
				"source": filepath.Join(do.Path, "abi"),
				"dest":   do.ABIPath,
			}).Debug("Moving ABI into position")
			if err := os.Rename(filepath.Join(do.Path, "abi"), do.ABIPath); err != nil {
				return err
			}
		}
	} else if !strings.Contains(do.ABIPath, do.Path) {
		log.WithFields(log.Fields{
			"source": filepath.Join(do.Path, "abi"),
			"dest":   do.ABIPath,
		}).Debug("Moving ABI into position")
		if err := os.Rename(filepath.Join(do.Path, "abi"), do.ABIPath); err != nil {
			return err
		}
	} else {
		log.Info("ABI path does not exist on the host or is inside the pkg path")
	}

	// import epm.yaml (if it is in a weird place)
	if inbound && !strings.Contains(do.EPMConfigFile, do.Path) { // note <- is the default, if we change the default we'll have to change this.
		doData.Destination = path.Join(common.ErisContainerRoot, "apps", filepath.Base(do.Path))
		doData.Source = do.EPMConfigFile
		log.WithFields(log.Fields{
			"source": doData.Source,
			"dest":   doData.Destination,
		}).Debug("Importing eris-pm config file")
		if err := data.ImportData(doData); err != nil {
			return err
		}
	} else if !strings.Contains(do.EPMConfigFile, do.Path) {
		file, err := filepath.Abs(do.EPMConfigFile)
		if err != nil {
			return err
		}
		dirToMoveTo := filepath.Dir(file)
		epmFiles, err := filepath.Glob(filepath.Join(do.Path, "epm*"))
		if err != nil {
			return err
		}
		log.WithFields(log.Fields{
			"source": epmFiles,
			"dest":   dirToMoveTo,
		}).Debug("Moving eris-pm artifacts")
		for _, epmFile := range epmFiles {
			var err error
			epmFile, err = filepath.Abs(epmFile)
			if err != nil {
				return err
			}
			err = os.Rename(epmFile, filepath.Join(dirToMoveTo, filepath.Base(epmFile)))
			if err != nil {
				return err
			}
		}
	} else {
		log.Info("EPM files do not exist on the host or are inside the pkg path")
	}

	do.Operations.DataContainerName = util.DataContainerName(doData.Name)
	do.Path = oldDoPath
	do.PackagePath = oldPkgPath
	do.ABIPath = oldAbiPath
	return nil
}
