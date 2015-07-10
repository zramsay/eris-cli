package services

import (
	"fmt"
	"sync"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"
)

func StartServiceRaw(do *definitions.Do) error {
	var services []*definitions.ServiceDefinition
	if do.Operations.ContainerNumber == 0 { // TODO: automagic, see #67
		do.Operations.ContainerNumber = 1
	}

	do.Args = append(do.Args, do.ServicesSlice...)
	logger.Debugf("Building the Services Group =>\t%v\n", do.Args)
	for _, srv := range do.Args {
		// this forces CLI/Agent level overwrites of the Operations.
		// if this needs to get reversed, we should discuss on GH.
		s, e := BuildServicesGroup(srv)
		if e != nil {
			return e
		}
		services = append(services, s...)
	}

	for _, s := range services {
		util.OverWriteOperations(s.Operations, do.Operations)
	}

	var err error
	services, err = BuildChainGroup(do.ChainName, services)
	if err != nil {
		return err
	}

	wg, ch := new(sync.WaitGroup), make(chan error, 1)
	StartGroup(ch, wg, services) // TODO, add the chain
	go func() {
		wg.Wait()
		ch <- nil
	}()
	if err := <-ch; err != nil {
		return err
	}

	return nil
}

func KillServiceRaw(do *definitions.Do) error {
	var services []*definitions.ServiceDefinition

	for _, servName := range do.Args {
		s, e := BuildServicesGroup(servName)
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

	if do.Force {
		if do.Timeout == 10 { // default set by flags
			do.Timeout = 0
		}
	}

	for _, service := range services {
		if IsServiceRunning(service.Service, service.Operations) {
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
func BuildServicesGroup(srvName string, services ...*definitions.ServiceDefinition) ([]*definitions.ServiceDefinition, error) {
	logger.Debugf("BuildServicesGroup for =>\t%s:%d\n", srvName, len(services))
	srv, err := loaders.LoadServiceDefinition(srvName)
	if err != nil {
		return nil, err
	}
	for _, sName := range srv.ServiceDeps {
		logger.Debugf("Found service dependency =>\t%s\n", sName)
		s, e := BuildServicesGroup(sName)
		if e != nil {
			return nil, e
		}
		services = append(services, s...)
	}
	services = append(services, srv)
	return services, nil
}

// start a group of chains or services. catch errors on a channel so we can stop as soon as something goes wrong
// TODO: Add ONE Chain
func StartGroup(ch chan error, wg *sync.WaitGroup, group []*definitions.ServiceDefinition) {
	logger.Debugf("Starting services group =>\t%d Services\n", len(group))
	for _, srv := range group {
		wg.Add(1)

		go func(s *definitions.ServiceDefinition) {

			logger.Debugf("Telling Docker to start srv =>\t%s\n", s.Name)
			if err := perform.DockerRun(s.Service, s.Operations); err != nil {
				logger.Debugf("Error starting service (%s): %v\n", s.Name, err)
				ch <- err
			}

			wg.Done()
		}(srv)

	}
}

// Note chainName in this command refers mostly to a chain which has been passed as a flag
// the command will add to the group a single chain passed into the group as well as
// individualized chains that each service may individually rely upon.
func BuildChainGroup(chainName string, services []*definitions.ServiceDefinition) ([]*definitions.ServiceDefinition, error) {
	var chains []*definitions.ServiceDefinition

	for _, srv := range services {
		if srv.Chain == "$chain" && chainName == "" {
			return nil, fmt.Errorf("Marmot disapproval face. You tried to start a service which has a $chain variable but didn't give me a chain.")
		}
		if chainName != "" {
			s, err := ChainConnectedToAService(chainName, srv)
			if err != nil {
				return nil, err
			}
			chains = append(chains, s)
		}
		if srv.Chain == "$chain" {
			continue
		}
		if srv.Chain != "" {
			s, err := ChainConnectedToAService(srv.Chain, srv)
			if err != nil {
				return nil, err
			}
			chains = append(chains, s)
		}
	}

	return append(services, chains...), nil
}

func ChainConnectedToAService(chainName string, srv *definitions.ServiceDefinition) (*definitions.ServiceDefinition, error) {
	s, err := loaders.ChainsAsAService(chainName, srv.Operations.ContainerNumber)
	if err != nil {
		return nil, err
	}

	loaders.ConnectToAService(srv, chainName) // first make the service container linked to the chain
	loaders.ConnectToAService(s, srv.Name)   // now make the chain container linked to the service container
	// XXX: we may have name collision here if we're not careful.

	util.OverWriteOperations(s.Operations, srv.Operations)
	return s, nil
}

func StartServiceByService(srvMain *definitions.Service, ops *definitions.Operation) error {
	return perform.DockerRun(srvMain, ops)
}

func ExecServiceByService(srvMain *definitions.Service, ops *definitions.Operation, cmd []string, attach bool) error {
	return perform.DockerExec(srvMain, ops, cmd, attach)
}

func LogsServiceByService(srv *definitions.Service, ops *definitions.Operation, follow bool, tail string) error {
	return perform.DockerLogs(srv, ops, follow, tail)
}

func KillServiceByService(srvMain *definitions.Service, ops *definitions.Operation, timeout uint) error {
	return perform.DockerStop(srvMain, ops, timeout)
}

// ------------------------------------------------------------------------------------------
// Helpers

