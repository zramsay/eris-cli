package services

import (
	"fmt"
	"strings"
	"sync"

	"github.com/eris-ltd/eris-cli/perform"
	def "github.com/eris-ltd/eris-cli/definitions"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Start(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	IfExit(StartServiceRaw(args[0]))
}

func Logs(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	IfExit(LogsServiceRaw(args[0], cmd.Flags().Lookup("tail").Changed))
}

func Exec(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	srv := args[0]
	args = args[1:]
	if len(args) == 1 {
		args = strings.Split(args[0], " ")
	}
	IfExit(ExecServiceRaw(srv, args, cmd.Flags().Lookup("interactive").Changed))
}

func Kill(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	IfExit(KillServiceRaw(cmd.Flags().Lookup("all").Changed, cmd.Flags().Lookup("rm").Changed, args...))
}

func StartServiceRaw(servName string) error {
	service, err := LoadServiceDefinition(servName)
	if err != nil {
		return err
	}

	if IsServiceRunning(service.Service) {
		logger.Infoln("Service already started. Skipping", servName)
	} else {
		return StartServiceByService(service.Service, service.Operations)
	}

	return nil
}

func LogsServiceRaw(servName string, follow bool) error {
	service, err := LoadServiceDefinition(servName)
	if err != nil {
		return err
	}
	return LogsServiceByService(service.Service, service.Operations, follow)
}

func ExecServiceRaw(name string, args []string, attach bool) error {
	service, err := LoadServiceDefinition(name)
	if err != nil {
		return err
	}

	if IsServiceExisting(service.Service) {
		logger.Infoln("Service exists.")
		return ExecServiceByService(service.Service, service.Operations, args, attach)
	} else {
		return fmt.Errorf("Services does not exist. Please start the service container with eris services start %s.\n", name)
	}

	return nil
}

func KillServiceRaw(all, rm bool, servNames ...string) error {
	for _, servName := range servNames {
		service, err := LoadServiceDefinition(servName)
		if err != nil {
			return err
		}

		if IsServiceRunning(service.Service) {
			if err := KillServiceByService(all, rm, service.Service, service.Operations); err != nil {
				return err
			}

		} else {
			logger.Infoln("Service not currently running. Skipping.")
		}
	}

	if rm {
		if err := RmServiceRaw(servNames, false); err != nil {
			return err
		}
	}

	return nil
}

func LogsServiceByService(srv *def.Service, ops *def.ServiceOperation, follow bool) error {
	return perform.DockerLogs(srv, ops, follow)
}

// start a group of chains or services. catch errors on a channel so we can stop as soon as something goes wrong
func StartGroup(ch chan error, wg *sync.WaitGroup, group, running []string, name string, start func(string) error) {
	var skip bool
	for _, srv := range group {

		if srv == "" {
			continue
		}

		skip = false
		// XXX: is this redundant with what happens in StartServiceRaw ?
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
			if err := start(s); err != nil {
				logger.Debugln("error starting service", s, err)
				ch <- err
			}
			wg.Done()
		}(srv)
	}
}

func StartServiceByService(srvMain *def.Service, ops *def.ServiceOperation) error {
	wg, ch := new(sync.WaitGroup), make(chan error, 1)
	StartGroup(ch, wg, srvMain.ServiceDeps, nil, "service", StartServiceRaw)
	go func() {
		wg.Wait()
		ch <- nil
	}()
	if err := <-ch; err != nil {
		return err
	}
	return perform.DockerRun(srvMain, ops)
}

func ExecServiceByService(srvMain *def.Service, ops *def.ServiceOperation, cmd []string, attach bool) error {
	return perform.DockerExec(srvMain, ops, cmd, attach)
}

func KillServiceByService(all, rm bool, srvMain *def.Service, ops *def.ServiceOperation) error {
	if all {
		for _, srv := range srvMain.ServiceDeps {
			go KillServiceRaw(all, rm, srv)
		}
	}
	return perform.DockerStop(srvMain, ops)
}
