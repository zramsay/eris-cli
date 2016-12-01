package pkgs

import (
	"fmt"
	//"io"
	"os"
	//"os/user"
	"path"
	"path/filepath"
	//"strings"
	"time"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/config"
	//"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/log"
	//"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"
)

var pwd string

// RunPackage runs a package pointed to by the do.Path directory. It first loads
// and populates the pkg struct. Then boots the dependent services and chains.
// Then builds the appropriate pkg service to be ran in docker and properly
// connected to all other containers. Then runs the service and finally operates
// a cleanup.
//
//  do.Path      - root directory of the pkg
//  do.ChainName - name of the chain to run the pkgs do against
//

// do.ChainID (todo, maybe) XXX
func RunPackageSkip(do *definitions.Do) error {
	log.Warn("Performing action. This can sometimes take a wee while")
	var err error
	pwd, err = os.Getwd()
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"host path": do.Path,
		"pwd":       pwd,
	}).Debug()

	pkg, err := loaders.LoadPackageOLD(do.Path, do.ChainName)
	if err != nil {
		return err
	}

	if err := BootServicesAndChain(do, pkg); err != nil {
		CleanUp(do, pkg)
		return fmt.Errorf("Could not boot chain or services: %v", err)
	}

	if err := DefinePkgActionService(do, pkg); err != nil {
		CleanUp(do, pkg)
		return fmt.Errorf("Could not define pkg action service: %v", err)
	}

	//if err := PerformAppActionService(do, pkg); err != nil {
	//	CleanUp(do, pkg)
	//	return fmt.Errorf("Could not perform pkg action service: %v", err)
	//}

	return CleanUp(do, pkg)
}

// BootServicesAndChain ensures that dependent services are started and that
// the appropriate chain is booted.
//
//  do.ServicesSlice - slice of dependent services to boot
//                     before the eris-pm runs
//  do.ChainName     - name of the chain to ensure is booted
//                     (if "" then will check the checkedout chain)
//  do.LocalCompiler - use a local compiler service
//
//  pkg.Name                   - name of the package (defaults to current dir)
//  pkg.Dependencies.Services  - slice of dependent services to boot before
//                               the Eris PM runs (appends to do.ServicesSlice)
//  pkg.ChainName              - chain name from the pkg overwrites do.ChainName
//                               if do.ChainName blank
//
func BootServicesAndChain(do *definitions.Do, pkg *definitions.Package) error {

	var srvs []*definitions.ServiceDefinition
	do.ServicesSlice = append(do.ServicesSlice, pkg.Dependencies.Services...)

	// add the compilers to the local services if the flag is pushed
	// [csk] note - when we move to default local compilers we'll remove
	// the compilers service completely and this will need to get
	// reworked to utilize DockerRun with a populated service def.
	if do.LocalCompiler {
		do.ServicesSlice = append(do.ServicesSlice, "compilers")
	}

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

	return nil
}

// DefinePkgActionService Builds a service that will run.
//
//  do.Name      - name. [csk] unused?
//  do.Path      - pkg root path
//  do.ChainPort - port number (as a string) to chain's RPC
//  do.KeysPort  - port number (as a string) to eris-keys signing pipe
//  pkg.Name     - name [csk, why do we have two?]; defaults to dirName(".")
//
func DefinePkgActionService(do *definitions.Do, pkg *definitions.Package) error {
	do.Service.Name = pkg.Name + "_tmp_" + do.Name
	//do.Service.Image = path.Join(config.Global.DefaultRegistry, config.Global.ImagePM)
	do.Service.AutoData = true
	do.Service.EntryPoint = fmt.Sprintf("eris-pm --chain tcp://chain:%s --sign http://keys:%s", do.ChainPort, do.KeysPort)
	do.Service.WorkDir = path.Join(config.ErisContainerRoot, "apps", filepath.Base(do.Path))
	do.Service.User = "eris"

	srv := definitions.BlankServiceDefinition()
	srv.Service = do.Service
	srv.Operations = do.Operations
	loaders.ServiceFinalizeLoad(srv)
	do.Service = srv.Service
	do.Operations = srv.Operations
	do.Operations.Follow = true

	//prepareEpmAction(do, pkg)
	linkAppToChain(do)

	log.Debug("App action built")
	return nil
}

// CleanUp controls the eris pkgs tear down function after an eris pkgs do.
// It runs export process to pull everything out of data containers.
//
//  do.Operations      - must be populated
//  do.Rm              - remove the service container (defaults to true;
//                       false is useful for debug and testing purposes only)
//  do.RmD             - remove the data container (defaults to true;
//                       false useful for debug and testing purposes only)
//
func CleanUp(do *definitions.Do, pkg *definitions.Package) error {
	log.Info("Cleaning up")

	// removal of local compiler; [csk] note we may not want to remove the container for performance reasons
	if do.LocalCompiler {
		log.Debug("Turning off and removing local compiler container")
		doStop := definitions.NowDo()
		doStop.Operations.Args = []string{"compilers"}
		doStop.Rm, doStop.Force, doStop.RmD, doStop.Volumes = true, true, true, true
		if err := services.KillService(doStop); err != nil {
			return err
		}
	}

	return nil
}

// bootChain boots chain as an Eris chain or an eris service depending on the name.
// Assumes a do.Operations struct has been properly populated.
func bootChain(name string, do *definitions.Do) error {
	// Setting this for tear-down purposes.
	do.ChainDefinition.ChainType = "service"

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
		// Setting this for tear-down purposes.
		// [zr]: should no longer be needed.
		do.ChainDefinition.ChainType = "chain"

	// known chain directory; new the chain with the right directory (note this will use only chain root so is only good for single node chains) [zr] this should go too
	case util.DoesDirExist(filepath.Join(config.ChainsPath, startChain.Name)):
		log.WithField("name", startChain.Name).Info("Trying new chain")
		startChain.Path = filepath.Join(config.ChainsPath, startChain.Name)
		if err := chains.StartChain(startChain); err != nil {
			return err
		}
		do.ChainDefinition.ChainType = "chain" // setting this for tear down purposes
	// known service; make sure service is running
	case util.IsService(name, false):
		log.WithField("name", name).Info("Chain exists as a service")
		startService := definitions.NowDo()
		startService.Operations = do.Operations
		startService.Operations.Args = []string{name}
		if err := services.StartService(startService); err != nil {
			return err
		}
		do.ChainDefinition.ChainType = "service" // setting this for tear down purposes
	default:
		return fmt.Errorf("The marmots could not find that chain name. Please review and rerun the command")
	}

	// Setting this for tear-down purposes.
	do.ChainDefinition.Name = name

	// let the chain boot properly
	time.Sleep(5 * time.Second)
	return nil
}

func linkKeys(do *definitions.Do) {
	newLink2 := util.ServiceContainerName("keys") + ":" + "keys"
	do.Signer = newLink2
	// ah this isn't used by run_package
	do.Service.Links = append(do.Service.Links, newLink2)
}

// linkAppToChain ensures chain properly connected to Eris PM services
// container. It assumes a do and pkg struct properly populated.
func linkAppToChain(do *definitions.Do) {
	var newLink string

	if do.ChainDefinition.ChainType == "service" {
		newLink = util.ServiceContainerName(do.ChainName) + ":" + "chain"
	} else {
		newLink = util.ChainContainerName(do.ChainName) + ":" + "chain"
	}
	newLink2 := util.ServiceContainerName("keys") + ":" + "keys"
	do.Signer = newLink2
	do.Service.Links = append(do.Service.Links, newLink)
	do.Service.Links = append(do.Service.Links, newLink2)

	for _, s := range do.ServicesSlice {
		l := util.ServiceContainerName(s) + ":" + s
		do.Service.Links = append(do.Service.Links, l)
	}
}

// getLocalCompilerData populates the IP:port combo for the compilers.
func getLocalCompilerData(do *definitions.Do) {
	// [csk]: note this is brittle we should only expose one port in the
	// docker file by default for the compilers service we can expose more
	// forcibly

	do.Compiler = "http://compilers:9099"
}
