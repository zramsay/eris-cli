package services

import (
	"fmt"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	"sync"
)

func StartServiceRaw(servName string, containerNumber int, ops *def.Operation) error {
	srv, err := LoadServiceDefinition(servName, containerNumber)
	if err != nil {
		return err
	}

	if IsServiceRunning(srv.Service, srv.Operations) {
		logger.Infoln("Service already started. Skipping", servName)
	} else {
		util.OverwriteOps(srv.Operations, ops)

		for _, dep := range srv.ServiceDeps {
			newLink := util.ServiceContainersName(dep, ops.ContainerNumber) + ":" + dep
			srv.Service.Links = append(srv.Service.Links, newLink)
		}

		return StartServiceByService(srv.Service, srv.Operations, srv.ServiceDeps)
	}

	return nil
}

func LogsServiceRaw(servName string, follow bool, tail string, containerNumber int) error {
	service, err := LoadServiceDefinition(servName, containerNumber)
	if err != nil {
		return err
	}
	return LogsServiceByService(service.Service, service.Operations, follow, tail)
}

func ExecServiceRaw(name string, args []string, attach bool, containerNumber int) error {
	service, err := LoadServiceDefinition(name, containerNumber)
	if err != nil {
		return err
	}

	if IsServiceExisting(service.Service, service.Operations) {
		logger.Infoln("Service exists.")
		return ExecServiceByService(service.Service, service.Operations, args, attach)
	} else {
		return fmt.Errorf("Services does not exist. Please start the service container with eris services start %s.\n", name)
	}

	return nil
}

func KillServiceRaw(all, rm, rmData bool, containerNumber int, servNames ...string) error {
	for _, servName := range servNames {
		service, err := LoadServiceDefinition(servName, containerNumber)
		if err != nil {
			return err
		}

		if IsServiceRunning(service.Service, service.Operations) {
			if err := KillServiceByService(service.Service, service.Operations, service.ServiceDeps, all, rm, rmData); err != nil {
				return err
			}

		} else {
			logger.Infoln("Service not currently running. Skipping.")
		}
	}

	if rm {
		if err := RmServiceRaw(servNames, containerNumber, false, rmData); err != nil {
			return err
		}
	}

	return nil
}

func LogsServiceByService(srv *def.Service, ops *def.Operation, follow bool, tail string) error {
	return perform.DockerLogs(srv, ops, follow, tail)
}

// start a group of chains or services. catch errors on a channel so we can stop as soon as something goes wrong
func StartGroup(ch chan error, wg *sync.WaitGroup, group, running []string, name string, ops *def.Operation, start func(string, int, *def.Operation) error) {
	num := ops.ContainerNumber
	var skip bool
	for _, srv := range group {

		if srv == "" {
			continue
		}

		// XXX: is this redundant with what happens in StartServiceRaw ?
		skip = false
		for _, run := range running {
			if srv == run {
				logger.Infof("%s already started, skipping: %s\n", name, srv)
				skip = true
			}
		}
		if skip {
			continue
		}

		wg.Add(1)

		go func(s string) {
			logger.Debugln("starting service", s)

			if err := start(s, num, ops); err != nil {
				logger.Debugln("error starting service", s, err)
				ch <- err
			}

			wg.Done()

		}(srv)
	}
}

func StartServiceByService(srvMain *def.Service, ops *def.Operation, servDeps []string) error {
	wg, ch := new(sync.WaitGroup), make(chan error, 1)
	StartGroup(ch, wg, servDeps, nil, "service", ops, StartServiceRaw)
	go func() {
		wg.Wait()
		ch <- nil
	}()
	if err := <-ch; err != nil {
		return err
	}
	return perform.DockerRun(srvMain, ops)
}

func ExecServiceByService(srvMain *def.Service, ops *def.Operation, cmd []string, attach bool) error {
	return perform.DockerExec(srvMain, ops, cmd, attach)
}

func KillServiceByService(srvMain *def.Service, ops *def.Operation, servDeps []string, all, rm, rmData bool) error {
	if all {
		for _, srv := range servDeps {
			go KillServiceRaw(all, rm, rmData, ops.ContainerNumber, srv)
		}
	}
	return perform.DockerStop(srvMain, ops)
}
