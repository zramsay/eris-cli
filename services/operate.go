package services

import (
	"fmt"
	"io"
	"os"

	"github.com/eris-ltd/eris-cli/perform"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Start(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(StartServiceRaw(args[0], cmd.Flags().Lookup("verbose").Changed, os.Stdout))
}

func Logs(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(LogsServiceRaw(args[0], cmd.Flags().Lookup("verbose").Changed, os.Stdout))
}

func Kill(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(KillServiceRaw(args[0], cmd.Flags().Lookup("verbose").Changed, os.Stdout))
}

func StartServiceRaw(servName string, verbose bool, w io.Writer) error {
	service, err := LoadServiceDefinition(servName)
	if err != nil {
		return err
	}

	if IsServiceRunning(service.Service) {
		if verbose {
			w.Write([]byte("Service already started. Skipping."))
		}
	} else {
		err := StartServiceByService(service.Service, service.Operations, verbose, w)
		if err != nil {
			return err
		}
	}

	return nil
}

func LogsServiceRaw(servName string, verbose bool, w io.Writer) error {
	service, err := LoadServiceDefinition(servName)
	if err != nil {
		return err
	}
	err = LogsServiceByService(service.Service, service.Operations, verbose, w)
	if err != nil {
		return err
	}
	return nil
}

func KillServiceRaw(servName string, verbose bool, w io.Writer) error {
	service, err := LoadServiceDefinition(servName)
	if err != nil {
		return err
	}

	if IsServiceRunning(service.Service) {
		err := KillServiceByService(service.Service, service.Operations, verbose, w)
		if err != nil {
			return err
		}
	} else {
		if verbose {
			fmt.Println("Service not currently running. Skipping.")
		}
	}
	return nil
}

func LogsServiceByService(srv *def.Service, ops *def.ServiceOperation, verbose bool, w io.Writer) error {
	err := perform.DockerLogs(srv, ops, verbose, w)
	if err != nil {
		return err
	}
	return nil
}

func StartServiceByService(srvMain *def.Service, ops *def.ServiceOperation, verbose bool, w io.Writer) error {
	for _, srv := range srvMain.ServiceDeps {
		go StartServiceRaw(srv, verbose, w)
	}
	err := perform.DockerRun(srvMain, ops, verbose, w)
	if err != nil {
		return err
	}
	return nil
}

func KillServiceByService(srvMain *def.Service, ops *def.ServiceOperation, verbose bool, w io.Writer) error {
	for _, srv := range srvMain.ServiceDeps {
		go KillServiceRaw(srv, verbose, w)
	}
	err := perform.DockerStop(srvMain, ops, verbose, w)
	if err != nil {
		return err
	}
	return nil
}
