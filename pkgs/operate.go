package pkgs

import (
	"fmt"
	"io"
	"os"
	"os/user"
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

	"github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
)

var pwd string

// entrypoint for eris pkgs do. first loads and populates the pkg struct. Then boots the dependent
// services and chains. Then builds the appropriate pkg service to be ran in docker and properly
// connected to all other containers. Then runs the service and finally operates a cleanup.
//
//  do.Path      - root directory of the pkg
//  do.ChainName - name of the chain to run the pkgs do against
//
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

// ensures that dependent services are started and that the appropriate chain is booted
//
//  do.ServicesSlice - slice of dependent services to boot before the eris-pm runs
//  do.ChainName - name of the chain to ensure is booted (if "" then will check the checkedout chain)
//
//  pkg.Name - name of the package (defaults to os.Base(pwd))
//  pkg.Dependencies.Services  - slice of dependent services to boot before the eris-pm runs (appends to do.ServicesSlice)
//  pkg.ChainName - chain name from the pkg overwrites do.ChainName if do.ChainName blank
//
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

	// overwrite do.ChainName with pkg.ChainName if do.ChainName blank
	if do.ChainName == "" {
		do.ChainName = pkg.ChainName
	}

	// boot the chain
	switch do.ChainName { // switch on the flag
	case "", "$chain":
		head, _ := util.GetHead() // checks the checkedout chain
		if head != "" {           // used checked out chain
			log.WithField("=>", head).Info("No chain flag or in package file. Booting chain from checked out chain")
			err = bootChain(head, do)
		} else { // if no chain is checked out and no --chain given, default to a throwaway
			log.Warn("No chain was given, booting a throwaway chain")
			err = bootThrowAwayChain(pkg.Name, do)
		}
	case "temp", "temporary", "throw", "throwaway":
		log.Info("No chain was given, booting a throwaway chain")
		err = bootThrowAwayChain(pkg.Name, do)
	default:
		log.WithField("=>", pkg.ChainName).Info("No chain flag used. Booting chain from package file")
		err = bootChain(do.ChainName, do)
	}

	// populate pkg.ChainName from do.Chain.Name that was added during the bootChain/bootThrowAwayChain func
	pkg.ChainName = do.Chain.Name
	if err != nil {
		return err
	}

	return nil
}

// Builds a service that will run.
//
//  do.Name      - name. [csk] unused?
//  do.Path      - pkg root path
//  do.ChainPort - port number (as a string) to chain's RPC
//  do.KeysPort  - port number (as a string) to eris-keys signing pipe
//  pkg.Name     - name [csk, why do we have two?]; defaults to dirName(".")
//
func DefinePkgActionService(do *definitions.Do, pkg *definitions.Package) error {
	do.Service.Name = pkg.Name + "_tmp_" + do.Name
	do.Service.Image = path.Join(version.ERIS_REG_DEF, version.ERIS_IMG_PM)
	do.Service.AutoData = true
	do.Service.EntryPoint = fmt.Sprintf("epm --chain chain:%s --sign keys:%s", do.ChainPort, do.KeysPort)
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

// controls the operation of eris-pm. meaning it runs the service container.
//
//  do.Service    - properly populated
//  do.Operations - properly populated
//
func PerformAppActionService(do *definitions.Do, pkg *definitions.Package) error {
	// import into data container
	if err := getDataContainerSorted(do, true); err != nil {
		return err
	}

	// run service, get result from its buffer
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
		do.Result = "could not perform pkg action"
		return err
	}

	// copy output to global writer. [csk note] this is a bit weird cause no output until the whole thing has finished...
	io.Copy(config.GlobalConfig.Writer, buf)

	log.Info("Finished performing action")
	return nil
}

// controls the eris pkgs tear down function after an eris pkgs do. destroys throwaway chains.
// runs export process to pull everything out of data containers.
//
//  do.Operations      - must be populated
//  do.Chain.ChainType - if == "throwaway" then will remove; otherwise no effect
//  do.Rm              - remove the service container (defaults to true; false useful for debug and testing purposes only)
//  do.RmD             - remove the data container (defaults to true; false useful for debug and testing purposes only)
//
func CleanUp(do *definitions.Do, pkg *definitions.Package) error {
	log.Info("Cleaning up")

	// destroy throwaway chain
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

	// export process
	if err := getDataContainerSorted(do, false); err != nil {
		return err // errors marmotified in getDataContainerSorted
	}

	// removal of service container
	if !do.Rm {
		doRemove := definitions.NowDo()
		doRemove.Operations.SrvContainerName = do.Operations.DataContainerName
		log.WithField("=>", doRemove.Operations.SrvContainerName).Debug("Removing data container")
		if err := perform.DockerRemove(nil, doRemove.Operations, false, true, false); err != nil {
			return err
		}
	}

	// removal of data container
	if !do.RmD {
		log.WithField("dir", filepath.Join(common.DataContainersPath, do.Service.Name)).Debug("Removing data dir on host")
		os.RemoveAll(filepath.Join(common.DataContainersPath, do.Service.Name))
	}

	return nil
}

// boots chain as an eris chain or an eris service depending on the name. assumes a do.Operations struct has
// been properly populated.
func bootChain(name string, do *definitions.Do) error {
	do.Chain.ChainType = "service" // setting this for tear down purposes
	startChain := definitions.NowDo()
	startChain.Name = name
	startChain.Operations = do.Operations

	switch {
	// known chain; make sure chain is running
	case util.IsChain(name, true):
		log.WithField("name", startChain.Name).Info("Starting chain")
		if err := chains.StartChain(startChain); err != nil {
			return err
		}
		do.Chain.ChainType = "chain" // setting this for tear down purposes
	// known chain directory; new the chain with the right directory (note this will use only chain root so is only good for single node chains)
	case util.DoesDirExist(filepath.Join(common.ChainsPath, startChain.Name)):
		log.WithField("name", startChain.Name).Info("Trying new chain")
		startChain.Path = filepath.Join(common.ChainsPath, startChain.Name)
		if err := chains.NewChain(startChain); err != nil {
			return err
		}
		do.Chain.ChainType = "chain" // setting this for tear down purposes
	// known service; make sure service is running
	case util.IsService(name, false):
		log.WithField("name", name).Info("Chain exists as a service")
		startService := definitions.NowDo()
		startService.Operations = do.Operations
		startService.Operations.Args = []string{name}
		if err := services.StartService(startService); err != nil {
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

// if a throwaway chain is noted; booth that chain
//
//  do.Name - temp name for throwaway chain reference
//
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

// ensures chain properly connected to eris-pm services container. assumes a do and pkg struct properly populated
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

// creates the command to be sent to eris-pm by the Service struct. we're using entrypoint and adding
// flags to the entrypoint to properly populate how eris-pm runs
//
//  do.Verbose       - run eris-pm in verbose mode.
//  do.Debug         - run eris-pm in debug mode.
//  do.Overwrite     - approve overwrite of variables.
//  do.CSV           - output an epm.csv instead of an epm.json.
//  do.ConfigOpts    - add set variables in a comma separated list to eris-pm
//  do.EPMConfigFile - path to epm.yaml
//  do.OutputTable   - output table from eris-pm
//  do.DefaultGas    - gas to give eris-pm
//  do.Compiler      - url of compiler service to use (assumes a receiving eris-compilers service at the ip:port combination)
//  do.DefaultAddr   - address of key to use by default (can also be set in epm.yaml); must match a key currently available in eris-keys
//  do.DefaultFee    - default fee to be paid to validators
//  do.DefaultAmount - default amount of tokens to send
//
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
	if do.Overwrite {
		do.Service.EntryPoint = do.Service.EntryPoint + " --overwrite "

		// do.Service.Environment = append(do.Service.Environment, "EPM_OVERWRITE_APPROVE=true")
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

	if do.EPMConfigFile != "" {
		do.Service.EntryPoint = do.Service.EntryPoint + " --file " + path.Join(".", filepath.Base(do.EPMConfigFile))
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

// deals with imports to and exports from eris-pm's data container. this function prob needs optimization [csk]
// function should be given a do struct which is read for operation (namely, has passed pkg loaders and has
// both do.Service && do.Operations properly populated).
//
//  do.Path          - path on host to where the epm.yaml is and where the epm.json will be written to. eris-pm will run from here.
//  do.PackagePath   - path on host to where the root of the package is. eris-pm assumes that contracts are available here or in here/contracts.
//  do.ABIPath       - path on host to where the ABI folder is and will be saved to.
//  do.EPMConfigFile - path on host to where the epm.yaml is located.
//
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

	// on importing create a data container to work with
	if inbound && util.Exists(definitions.TypeData, doData.Name) == false {
		doData.Operations.DataContainerName = util.DataContainerName(doData.Name)
		doData.Operations.ContainerType = definitions.TypeData
		if err := perform.DockerCreateData(doData.Operations); err != nil {
			return err
		}
	}

	// save these for replacing at end of function so that do struct is not changed outside of this func
	oldDoPath := do.Path
	oldPkgPath := do.PackagePath
	oldAbiPath := do.ABIPath

	// the do struct must be populated with absolute paths to reduce uncertainty in import/export phase
	var err error
	do.Path, err = filepath.Abs(do.Path)
	do.PackagePath, err = filepath.Abs(do.PackagePath)
	do.ABIPath, err = filepath.Abs(do.ABIPath)
	do.EPMConfigFile, err = filepath.Abs(do.EPMConfigFile)
	if err != nil {
		return err
	}

	// ensure that settings which expect a directory are actually directories. if not move up a level in filesystem.
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

	// If the ABI path specified is a home directory,
	// append the "abi" subdirectory to it.
	if user, err := user.Current(); err == nil && user.HomeDir == do.ABIPath {
		do.ABIPath = filepath.Join(do.ABIPath, "abi")
	}

	// import/export path
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

	// import/export package path
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
				"source": filepath.Join(do.Path, "contracts"), // [csk] this is an export, on windows this may be a problem... may need to be path.Join...?
				"dest":   do.PackagePath,
			}).Debug("Moving contracts into position")
			if err := util.MoveTree(filepath.Join(do.Path, "contracts"), do.PackagePath); err != nil {
				return err
			}
		}
	} else if !strings.Contains(do.PackagePath, do.Path) { // [csk] why is this needed? (obvi I built this func, but am now unsure why this is here)
		log.WithFields(log.Fields{
			"source": filepath.Join(do.Path, "contracts"),
			"dest":   do.PackagePath,
		}).Debug("Moving contracts into position")
		if err := util.MoveTree(filepath.Join(do.Path, "contracts"), do.PackagePath); err != nil {
			return err
		}
	} else {
		log.Info("Package path does not exist on the host or is inside the pkg path")
	}

	// import/export ABI path
	if inbound {
		if _, err := os.Stat(do.ABIPath); !os.IsNotExist(err) && !strings.Contains(do.ABIPath, do.Path) {
			log.WithFields(log.Fields{
				"path":     do.Path,
				"abi path": do.ABIPath,
			}).Debug("ABI path exists and is not in pkg path. Proceeding with Import/Export sequence")
			doData.Destination = path.Join(common.ErisContainerRoot, "apps", filepath.Base(do.Path), "abi")
			doData.Source = do.ABIPath

			log.WithFields(log.Fields{
				"source": doData.Source,
				"dest":   doData.Destination,
			}).Debug("Importing ABI path")

			if err := data.ImportData(doData); err != nil {
				return err
			}
		}
	} else {
		// Export ABI path.
		if do.ABIPath != filepath.Join(do.Path, "abi") {
			log.WithFields(log.Fields{
				"source": filepath.Join(do.Path, "abi"),
				"dest":   do.ABIPath,
			}).Debug("Moving ABI into position")

			if err := os.RemoveAll(do.ABIPath); err != nil {
				return err
			}

			if err := os.MkdirAll(do.ABIPath, 0755); err != nil {
				return err
			}

			if err := util.MoveTree(filepath.Join(do.Path, "abi"), do.ABIPath); err != nil {
				return err
			}
		}
	}

	// Import epm.yaml (if it is in a weird place).
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

	// put things back the way they were in the do struct
	do.Operations.DataContainerName = util.DataContainerName(doData.Name)
	do.Path = oldDoPath
	do.PackagePath = oldPkgPath
	do.ABIPath = oldAbiPath
	return nil
}
