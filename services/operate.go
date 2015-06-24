package services

import (
	"github.com/eris-ltd/eris-cli/perform"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"
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
		err := StartServiceByService(service.Service, service.Operations)
		if err != nil {
			return err
		}
	}

	return nil
}

func LogsServiceRaw(servName string) error {
	service, err := LoadServiceDefinition(servName)
	if err != nil {
		return err
	}
	err = LogsServiceByService(service.Service, service.Operations)
	if err != nil {
		return err
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
	err := perform.DockerLogs(srv, ops)
	if err != nil {
		return err
	}
	return nil
}

func StartServiceByService(srvMain *def.Service, ops *def.ServiceOperation) error {
	for _, srv := range srvMain.ServiceDeps {
		go StartServiceRaw(srv)
	}
	err := perform.DockerRun(srvMain, ops)
	if err != nil {
		return err
	}
	return nil
}

func KillServiceByService(srvMain *def.Service, ops *def.ServiceOperation) error {
	for _, srv := range srvMain.ServiceDeps {
		go KillServiceRaw(srv)
	}
	err := perform.DockerStop(srvMain, ops)
	if err != nil {
		return err
	}
	return nil
}
