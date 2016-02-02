package services

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

func StartService(do *definitions.Do) (err error) {
	var services []*definitions.ServiceDefinition

	do.Operations.Args = append(do.Operations.Args, do.ServicesSlice...)
	log.WithField("args", do.Operations.Args).Info("Building services group")
	for _, srv := range do.Operations.Args {
		s, e := BuildServicesGroup(srv, do.Operations.ContainerNumber)
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
	services, err = BuildChainGroup(do.ChainName, services)
	if err != nil {
		return err
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
		s, e := BuildServicesGroup(servName, do.Operations.ContainerNumber)
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
		if IsServiceRunning(service.Service, service.Operations) {
			log.WithField("=>", fmt.Sprintf("%s:%d", service.Service.Name, service.Operations.ContainerNumber)).Debug("Stopping service")
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
	service, err := loaders.LoadServiceDefinition(do.Name, false, do.Operations.ContainerNumber)
	if err != nil {
		return nil, err
	}

	util.Merge(service.Operations, do.Operations)

	// Get the main service container name, check if it's running.
	main := util.FindServiceContainer(do.Name, do.Operations.ContainerNumber, false)
	if main != nil {
		if service.Service.ExecHost == "" {
			log.Info("exec_host not found in service definition file")
			log.WithField("service", do.Name).Info("May not be able to communicate with the service")
		} else {
			service.Service.Environment = append(service.Service.Environment,
				fmt.Sprintf("%s=%s", service.Service.ExecHost, do.Name))
		}

		// Use service's short name as a link alias.
		service.Service.Links = append(service.Service.Links, fmt.Sprintf("%s:%s", main.FullName, do.Name))
	}

	// Override links on the command line.
	if len(do.Links) > 0 {
		service.Service.Links = do.Links
	}

	return perform.DockerExecService(service.Service, service.Operations)
}

// TODO: test this recursion and service deps generally
func BuildServicesGroup(srvName string, cNum int, services ...*definitions.ServiceDefinition) ([]*definitions.ServiceDefinition, error) {
	log.WithFields(log.Fields{
		"=>":        srvName,
		"services#": len(services),
	}).Debug("Building services group for")
	srv, err := loaders.LoadServiceDefinition(srvName, false, cNum)
	if err != nil {
		return nil, err
	}
	if srv.Dependencies != nil {
		for _, sName := range srv.Dependencies.Services {
			log.WithField("=>", sName).Debug("Found service dependency")
			s, e := BuildServicesGroup(sName, cNum)
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
			return fmt.Errorf("StartGroup. Err starting srv =>\t%s:%v\n", srv.Name, err)
		}
	}
	return nil
}

// BuildChainGroup adds the chain specified in each service definition to the service group.
// If chainName is not empty, it will overwrite chains specified in the defs.
// Service defs which don't specify a chain or $chain won't connect to a chain.
// NOTE: chains have to be started before services that depend on them.
func BuildChainGroup(chainName string, services []*definitions.ServiceDefinition) (servicesAndChains []*definitions.ServiceDefinition, err error) {
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
			return nil, fmt.Errorf("Marmot disapproval face.\nYou tried to start a service which has a `$chain` variable but didn't give us a chain.\nPlease rerun the command either after [eris chains checkout CHAINNAME] *or* with a --chain flag.\n")
		}
	}
	s, err := loaders.ChainsAsAService(chainName, false, srv.Operations.ContainerNumber)
	if err != nil {
		return nil, err
	}
	// link the service container linked to the chain
	// XXX: we may have name collision here if we're not careful.
	loaders.ConnectToAChain(srv.Service, srv.Operations, chainName, internalName, link, mount)

	return s, nil
}
