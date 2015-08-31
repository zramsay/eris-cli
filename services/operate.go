package services

import (
	"fmt"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"
)

func StartService(do *definitions.Do) (err error) {
	var services []*definitions.ServiceDefinition

	cNum := do.Operations.ContainerNumber
	do.Args = append(do.Args, do.ServicesSlice...)
	logger.Debugf("Building the Services Group =>\t%v\n", do.Args)
	for _, srv := range do.Args {
		// this forces CLI/Agent level overwrites of the Operations.
		// if this needs to get reversed, we should discuss on GH.
		s, e := BuildServicesGroup(srv, cNum)
		if e != nil {
			return e
		}
		services = append(services, s...)
	}
	for _, s := range services {
		//XXX does AutoMagic elim need for this?
		util.OverWriteOperations(s.Operations, do.Operations)
	}

	logger.Debugln("services before build chain")
	for _, s := range services {
		logger.Debugln("\t", s.Name, s.ServiceDeps, s.Service.Links, s.Service.VolumesFrom)
	}
	services, err = BuildChainGroup(do.ChainName, services)
	if err != nil {
		return err
	}
	logger.Debugln("services after build chain")
	for _, s := range services {
		logger.Debugln("\t", s.Name, s.ServiceDeps, s.Service.Links, s.Service.VolumesFrom)
	}

	return StartGroup(services)
}

func KillService(do *definitions.Do) error {
	var services []*definitions.ServiceDefinition

	for _, servName := range do.Args {
		s, e := BuildServicesGroup(servName, do.Operations.ContainerNumber)
		if e != nil {
			return e
		}
		services = append(services, s...)
	}

	var err error
	services, err = BuildChainGroup(do.ChainName, services)
	if err != nil {
		return err
	}

	// if force flag given, this will override any timeout flag
	if do.Force {
		do.Timeout = 0
	}

	for _, service := range services {
		if IsServiceRunning(service.Service, service.Operations) {
			logger.Debugf("Stopping Service =>\t\t%s:%d\n", service.Service.Name, service.Operations.ContainerNumber)
			if err := perform.DockerStop(service.Service, service.Operations, do.Timeout); err != nil {
				return err
			}

		} else {
			logger.Infoln("Service not currently running. Skipping.")
		}

		if do.Rm {
			if err := perform.DockerRemove(service.Service, service.Operations, do.RmD); err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO: test this recursion and service deps generally
func BuildServicesGroup(srvName string, cNum int, services ...*definitions.ServiceDefinition) ([]*definitions.ServiceDefinition, error) {
	logger.Debugf("BuildServicesGroup for =>\t%s:%d\n", srvName, len(services))
	srv, err := loaders.LoadServiceDefinition(srvName, false, cNum)
	if err != nil {
		return nil, err
	}
	if srv.ServiceDeps != nil {
		for _, sName := range srv.ServiceDeps.Dependencies {
			logger.Debugf("Found service dependency =>\t%s\n", sName)
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
// TODO: Add ONE Chain
func StartGroup(group []*definitions.ServiceDefinition) error {
	logger.Debugf("Starting services group =>\t%d Services\n", len(group))
	for _, srv := range group {
		logger.Debugf("Telling Docker to start srv =>\t%s\n", srv.Name)
		if err := perform.DockerRun(srv.Service, srv.Operations); err != nil {
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
		// if there's a $chain and no flag, we err
		return nil, fmt.Errorf("Marmot disapproval face. You tried to start a service which has a $chain variable but didn't give me a chain.")
	} else {
		// otherwise we just use the chainName in the service definition
	}
	s, err := loaders.ChainsAsAService(chainName, false, srv.Operations.ContainerNumber)
	if err != nil {
		return nil, err
	}
	// link the service container linked to the chain
	loaders.ConnectToAChain(srv, chainName, internalName, link, mount)
	// XXX: we may have name collision here if we're not careful.

	util.OverWriteOperations(s.Operations, srv.Operations)
	return s, nil
}

func StartServiceByService(srvMain *definitions.Service, ops *definitions.Operation) error {
	return perform.DockerRun(srvMain, ops)
}

func LogsServiceByService(srv *definitions.Service, ops *definitions.Operation, follow bool, tail string) error {
	return perform.DockerLogs(srv, ops, follow, tail)
}

func KillServiceByService(srvMain *definitions.Service, ops *definitions.Operation, timeout uint) error {
	return perform.DockerStop(srvMain, ops, timeout)
}

// ------------------------------------------------------------------------------------------
// Helpers
