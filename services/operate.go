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
	logger.Debugf("Building the Services Group =>\t%v\n", do.Args)

	for _, srv := range do.Args {
		s, e := BuildGroup(srv)
		if e != nil {
			return e
		}
		services = append(services, s...)
	}

	// Gives us a chance to overwrite operational functionality
	// which has been passed via command line flags or otherwise
	for _, srv := range services {
		util.OverwriteOps(srv.Operations, do.Operations)
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

func LogsServiceRaw(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}
	return LogsServiceByService(service.Service, service.Operations, do.Follow, do.Tail)
}

func ExecServiceRaw(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}

	if IsServiceExisting(service.Service, service.Operations) {
		return ExecServiceByService(service.Service, service.Operations, do.Args, do.Interactive)
	} else {
		return fmt.Errorf("Services does not exist. Please start the service container with eris services start %s.\n", do.Name)
	}

	return nil
}

func KillServiceRaw(do *definitions.Do) error {
	var services []*definitions.ServiceDefinition

	for _, servName := range do.Args {
		s, e := BuildGroup(servName)
		if e != nil {
			return e
		}
		services = append(services, s...)
	}

	for _, service := range services {
		if IsServiceRunning(service.Service, service.Operations) {
			if err := perform.DockerStop(service.Service, service.Operations); err != nil {
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
func BuildGroup(srvName string, services ...*definitions.ServiceDefinition) ([]*definitions.ServiceDefinition, error) {
	srv, err := loaders.LoadServiceDefinition(srvName) // TODO: populate cNum in load process
	if err != nil {
		return nil, err
	}
	for _, sName := range srv.ServiceDeps {
		s, e := BuildGroup(sName) // TODO: populate cNum in load process
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
	for _, srv := range group {
		wg.Add(1)

		go func(s *definitions.ServiceDefinition) {
			logger.Debugf("Telling Docker to start srv =>\t%s\n", s.Name)

			if err := perform.DockerRun(srv.Service, srv.Operations); err != nil {
				logger.Debugln("Error starting service (%s): %v\n", s.Name, err)
				ch <- err
			}

			wg.Done()
		}(srv)
	}
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

func KillServiceByService(srvMain *definitions.Service, ops *definitions.Operation) error {
	return perform.DockerStop(srvMain, ops)
}

// ------------------------------------------------------------------------------------------
// Helpers

