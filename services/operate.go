package services

import (
	"fmt"
	"strconv"

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

	// if interactive, we ignore args. if not, run args as command
	interactive := cmd.Flags().Lookup("interactive").Changed
	if !interactive {
		if len(args) < 2 {
			Exit(fmt.Errorf("Non-interactive exec sessions must provide arguments to execute"))
		}
		args = args[1:]
	}

	IfExit(ExecServiceRaw(srv, interactive, args))
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

func ExecServiceRaw(name string, interactive bool, args []string) error {
	if parseKnown(name) {
		logger.Infoln("Running exec on container with volumes from data container for " + name)
		containerNumber := 1
		name = "eris_service_" + name + "_" + strconv.Itoa(containerNumber)
		if err := perform.DockerRunVolumesFromContainer(name, interactive, args); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
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
