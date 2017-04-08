package services

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/monax/cli/config"
	"github.com/monax/cli/definitions"
	//"github.com/monax/cli/initialize"
	"github.com/monax/cli/loaders"
	"github.com/monax/cli/log"
	"github.com/monax/cli/perform"
	"github.com/monax/cli/util"
)

func StartService(do *definitions.Do) (err error) {
	var services []*definitions.ServiceDefinition

	do.Operations.Args = append(do.Operations.Args, do.ServicesSlice...)
	log.WithField("args", do.Operations.Args).Info("Building services group")
	for _, srv := range do.Operations.Args {
		s, e := BuildServicesGroup(srv)
		if e != nil {
			return e
		}
		services = append(services, s...)
	}

	// [csk]: controls for ops reconciliation, overwrite will, e.g., merge the maps and stuff
	for _, s := range services {
		util.Merge(s.Operations, do.Operations)
	}

	log.Debug("Preparing to build chain")
	for _, s := range services {
		log.WithFields(log.Fields{
			"name":         s.Name,
			"dependencies": s.Dependencies,
			"links":        s.Service.Links,
			"volumes from": s.Service.VolumesFrom,
		}).Debug()

		// Spacer.
		log.Debug()
	}
	if do.ChainName != "" {
		services, err = BuildChainGroup(do.ChainName, services)
		if err != nil {
			return err
		}
	}
	log.Debug("Checking services after build chain")
	for _, s := range services {
		log.WithFields(log.Fields{
			"name":         s.Name,
			"dependencies": s.Dependencies,
			"links":        s.Service.Links,
			"volumes from": s.Service.VolumesFrom,
		}).Debug()

		// Spacer.
		log.Debug()
	}

	// NOTE: the top level service should be at the end of the list
	topService := services[len(services)-1]
	topService.Service.Environment = append(topService.Service.Environment, do.Env...)
	topService.Service.Links = append(topService.Service.Links, do.Links...)
	services[len(services)-1] = topService

	return StartGroup(services)
}

func KillService(do *definitions.Do) (err error) {
	var services []*definitions.ServiceDefinition

	log.WithField("args", do.Operations.Args).Info("Building services group")
	for _, servName := range do.Operations.Args {
		s, e := BuildServicesGroup(servName)
		if e != nil {
			return e
		}
		services = append(services, s...)
	}

	// if force flag given, this will override any timeout flag
	if do.Force {
		do.Timeout = 0
	}

	for _, service := range services {
		if util.IsService(service.Service.Name, true) {
			log.WithField("=>", service.Service.Name).Debug("Stopping service")
			if err := perform.DockerStop(service.Service, service.Operations, do.Timeout); err != nil {
				return err
			}

		} else {
			log.WithField("=>", service.Service.Name).Info("Service not currently running. Skipping")
		}

		if do.Rm {
			if err := perform.DockerRemove(service.Service, service.Operations, do.RmD, do.Volumes, do.Force); err != nil {
				return err
			}
		}
	}

	return nil
}

func ExecService(do *definitions.Do) (buf *bytes.Buffer, err error) {
	service, err := loaders.LoadServiceDefinition(do.Name)
	if err != nil {
		return nil, err
	}

	util.Merge(service.Operations, do.Operations)
	if do.Service.User != "" {
		service.Service.User = do.Service.User
	}

	// Get the main service container name, check if it's running.
	main := util.ServiceContainerName(do.Name)
	if util.IsService(do.Name, true) {
		if service.Service.ExecHost == "" {
			log.Info("exec_host not found in service definition file")
			log.WithField("service", do.Name).Info("May not be able to communicate with the service")
		} else {
			// Grab the Container to inspect the service's IP address
			cont, err := util.DockerClient.InspectContainer(main)
			if err != nil {
				return nil, util.DockerError(err)
			}
			service.Service.Environment = append(service.Service.Environment,
				fmt.Sprintf("%s=%s", service.Service.ExecHost,
					cont.NetworkSettings.IPAddress))
		}

		// Use service's short name as a link alias.
		service.Service.Links = append(service.Service.Links, fmt.Sprintf("%s:%s", main, do.Name))
	}

	// Override links on the command line.
	if len(do.Links) > 0 {
		service.Service.Links = do.Links
	}

	return perform.DockerExecService(service.Service, service.Operations)
}

// ExecHandler implemements ExecService for use within
// the cli for under the hood functionality
// (wrapping) calls to respective containers
func ExecHandler(srvName string, args []string) (buf *bytes.Buffer, err error) {
	do := definitions.NowDo()
	do.Name = srvName
	do.Operations.Interactive = false
	do.Operations.Args = args
	do.Operations.PublishAllPorts = true
	return ExecService(do)
}

// TODO: test this recursion and service deps generally
func BuildServicesGroup(srvName string, services ...*definitions.ServiceDefinition) ([]*definitions.ServiceDefinition, error) {
	log.WithFields(log.Fields{
		"=>":        srvName,
		"services#": len(services),
	}).Debug("Building services group for")
	srv, err := loaders.LoadServiceDefinition(srvName)
	if err != nil {
		return nil, err
	}
	if srv.Dependencies != nil {
		for _, sName := range srv.Dependencies.Services {
			log.WithField("=>", sName).Debug("Found service dependency")
			s, e := BuildServicesGroup(sName)
			if e != nil {
				return nil, e
			}
			services = append(services, s...)
		}
	}
	services = append(services, srv)
	return services, nil
}

// start a group of chains or services. catch errors on a channel so we can stop as soon as something goes wrong
func StartGroup(group []*definitions.ServiceDefinition) error {
	log.WithField("services#", len(group)).Debug("Starting services group")
	for _, srv := range group {
		log.WithField("=>", srv.Name).Debug("Performing container start")
		if err := perform.DockerRunService(srv.Service, srv.Operations); err != nil {
			return fmt.Errorf("Error starting service %s: %v", srv.Name, err)
		}
	}
	return nil
}

// BuildChainGroup adds the chain specified in each service definition to the service group.
// If chainName is not empty, it will overwrite chains specified in the defs.
// Service defs which don't specify a chain or $chain won't connect to a chain.
// NOTE: chains have to be started before services that depend on them.
func BuildChainGroup(chainName string, services []*definitions.ServiceDefinition) (servicesAndChains []*definitions.ServiceDefinition, err error) {
	if !util.IsChain(chainName, true) {
		return nil, fmt.Errorf("Dependent chain %v is not running", chainName)
	}
	var chains = make(map[string]*definitions.ServiceDefinition)
	for _, srv := range services {
		if srv.Chain != "" {
			s, err := ConnectChainToService(chainName, srv.Chain, srv)
			if err != nil {
				return nil, err
			}
			if _, ok := chains[s.Name]; !ok {
				chains[s.Name] = s
			}
		}
	}
	for _, sd := range chains {
		servicesAndChains = append(servicesAndChains, sd)
	}
	return append(servicesAndChains, services...), nil
}

func ConnectChainToService(chainFlag, chainNameAndOpts string, srv *definitions.ServiceDefinition) (*definitions.ServiceDefinition, error) {
	chainName, internalName, link, mount := util.ParseDependency(chainNameAndOpts)
	if chainFlag != "" {
		// flag overwrites whatever is in the service definition
		chainName = chainFlag
	} else if strings.HasPrefix(srv.Chain, "$chain") {
		// if there's a $chain and no flag or checked out chain, we err
		var err error
		chainName, err = util.GetHead()
		if chainName == "" || err != nil {
			return nil, fmt.Errorf("Oops. You tried to start a service which has a `$chain` variable but didn't give us a chain.\nPlease rerun the command either after [monax chains checkout CHAINNAME] *or* with a --chain flag.\n")
		}
	}
	s, err := loaders.ChainsAsAService(chainName)
	if err != nil {
		return nil, err
	}
	// link the service container linked to the chain
	// XXX: we may have name collision here if we're not careful.
	loaders.ConnectToAChain(srv.Service, srv.Operations, chainName, internalName, link, mount)

	return s, nil
}

// Checks that a service is running and starts it if it isn't.
func EnsureRunning(do *definitions.Do) error {
	if os.Getenv("MONAX_SKIP_ENSURE") != "" {
		return nil
	}

	if _, err := loaders.LoadServiceDefinition(do.Name); err != nil {
		return err
	}

	if !util.IsService(do.Name, true) {
		log.WithField("=>", do.Name).Info("Starting service")
		do.Operations.Args = []string{do.Name}
		StartService(do)
	} else {
		log.WithField("=>", do.Name).Info("Service is running")
	}
	return nil
}

func IsServiceKnown(service *definitions.Service, ops *definitions.Operation) bool {
	return parseKnown(service.Name)
}

func FindServiceDefinitionFile(name string) string {
	return util.GetFileByNameAndType("services", name)
}

/*
func MakeService(do *definitions.Do) error {
	srv := definitions.BlankServiceDefinition()
	srv.Name = do.Name
	srv.Service.Name = do.Name
	srv.Service.Image = do.Operations.Args[0]
	srv.Service.AutoData = true

	var err error
	//get maintainer info
	srv.Maintainer.Name, srv.Maintainer.Email, err = config.GitConfigUser()
	if err != nil {
		// don't return -> field not required
		log.Debug(err.Error())
	}

	log.WithFields(log.Fields{
		"service": srv.Service.Name,
		"image":   srv.Service.Image,
	}).Debug("Creating a new service definition file")

	// TODO
	//if err := initialize.WriteServiceDefinitionFile(do.Name, srv); err != nil {
	//	return err
	//}
	return nil
}*/

func EditService(do *definitions.Do) error {
	servDefFile := FindServiceDefinitionFile(do.Name)
	log.WithField("=>", servDefFile).Info("Editing service")
	return config.Editor(servDefFile)
}

func InspectService(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name)
	if err != nil {
		return err
	}
	err = InspectServiceByService(service.Service, service.Operations, do.Operations.Args[0])
	if err != nil {
		return err
	}
	return nil
}

func PortsService(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name)
	if err != nil {
		return err
	}

	if util.IsService(service.Service.Name, false) {
		log.Debug("Service exists, getting port mapping")
		return util.PrintPortMappings(service.Operations.SrvContainerName, do.Operations.Args)
	}

	return nil
}

func LogsService(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name)
	if err != nil {
		return err
	}
	return perform.DockerLogs(service.Service, service.Operations, do.Follow, do.Tail)
}

func UpdateService(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name)
	if err != nil {
		return err
	}
	service.Service.Environment = append(service.Service.Environment, do.Env...)
	service.Service.Links = append(service.Service.Links, do.Links...)
	err = perform.DockerRebuild(service.Service, service.Operations, do.Pull, do.Timeout)
	if err != nil {
		return err
	}
	return nil
}

func RmService(do *definitions.Do) error {
	for _, servName := range do.Operations.Args {
		service, err := loaders.LoadServiceDefinition(servName)
		if err != nil {
			return err
		}
		if util.IsService(service.Service.Name, false) {
			if err := perform.DockerRemove(service.Service, service.Operations, do.RmD, do.Volumes, do.Force); err != nil {
				return err
			}
		}

		if do.RmImage {
			if err := perform.DockerRemoveImage(service.Service.Image, true); err != nil {
				return err
			}
		}

		if do.File {
			oldFile := util.GetFileByNameAndType("services", servName)
			if err != nil {
				return err
			}
			log.WithField("file", oldFile).Warn("Removing file")
			if err := os.Remove(oldFile); err != nil {
				return err
			}
		}
	}
	return nil
}

func CatService(do *definitions.Do) (string, error) {
	configs := util.GetGlobalLevelConfigFilesByType("services", true)
	for _, c := range configs {
		cName := strings.Split(filepath.Base(c), ".")[0]
		if cName == do.Name {
			cat, err := ioutil.ReadFile(c)
			if err != nil {
				return "", err
			}
			return string(cat), nil
		}
	}
	return "", fmt.Errorf("Unknown service %s or invalid file extension", do.Name)
}

func InspectServiceByService(srv *definitions.Service, ops *definitions.Operation, field string) error {
	return perform.DockerInspect(srv, ops, field)
}

func parseKnown(name string) bool {
	known := util.GetGlobalLevelConfigFilesByType("services", false)
	if len(known) != 0 {
		for _, srv := range known {
			if srv == name {
				return true
			}
		}
	}
	return false
}
