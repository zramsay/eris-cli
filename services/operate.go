package services

import (
	"fmt"
	"strings"

	"github.com/eris-ltd/eris-cli/perform"
	def "github.com/eris-ltd/eris-cli/definitions"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Start(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(StartServiceRaw(args[0]))
}

func Logs(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(LogsServiceRaw(args[0]))
}

func Exec(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	srv := args[0]
	args = args[1:]
	if len(args) == 1 {
		args = strings.Split(args[0], " ")
	}
	IfExit(ExecServiceRaw(srv, args, cmd.Flags().Lookup("interactive").Changed))
}

func Kill(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(KillServiceRaw(args[0]))
}

func StartServiceRaw(servName string) error {
	service, err := LoadServiceDefinition(servName)
	if err != nil {
		return err
	}

	if IsServiceRunning(service.Service) {
		logger.Infoln("Service already started. Skipping.")
	} else {
		return StartServiceByService(service.Service, service.Operations)
	}

	return nil
}

func LogsServiceRaw(servName string) error {
	service, err := LoadServiceDefinition(servName)
	if err != nil {
		return err
	}
	return LogsServiceByService(service.Service, service.Operations)
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

func KillServiceRaw(servName string) error {
	service, err := LoadServiceDefinition(servName)
	if err != nil {
		return err
	}

	if IsServiceRunning(service.Service) {
		err := KillServiceByService(service.Service, service.Operations)
		if err != nil {
			return err
		}
	} else {
		logger.Infoln("Service not currently running. Skipping.")
	}
	return nil
}

func LogsServiceByService(srv *def.Service, ops *def.ServiceOperation) error {
	return perform.DockerLogs(srv, ops)
}

func StartServiceByService(srvMain *def.Service, ops *def.ServiceOperation) error {
	for _, srv := range srvMain.ServiceDeps {
		go StartServiceRaw(srv)
	}
	return perform.DockerRun(srvMain, ops)
}

func ExecServiceByService(srvMain *def.Service, ops *def.ServiceOperation, cmd []string, attach bool) error {
	return perform.DockerExec(srvMain, ops, cmd, attach)
}

func KillServiceByService(srvMain *def.Service, ops *def.ServiceOperation) error {
	for _, srv := range srvMain.ServiceDeps {
		go KillServiceRaw(srv)
	}
	return perform.DockerStop(srvMain, ops)
}
