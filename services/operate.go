package services

import (
  "fmt"

  "github.com/eris-ltd/eris-cli/perform"

  def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"
  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Start(cmd *cobra.Command, args []string) {
  checkServiceGiven(args)
  StartServiceRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func Logs(cmd *cobra.Command, args []string) {
  checkServiceGiven(args)
  LogsServiceRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func Kill(cmd *cobra.Command, args []string) {
  checkServiceGiven(args)
  KillServiceRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func StartServiceRaw(servName string, verbose bool) {
  service := LoadServiceDefinition(servName)

  if IsServiceRunning(service.Service) {
    if verbose {
      fmt.Println("Service already started. Skipping.")
    }
  } else {
    StartServiceByService(service.Service, service.Operations, verbose)
  }
}

func LogsServiceRaw(servName string, verbose bool) {
  service := LoadServiceDefinition(servName)
  LogsServiceByService(service.Service, service.Operations, verbose)
}

func KillServiceRaw(servName string, verbose bool) {
  service := LoadServiceDefinition(servName)

  if IsServiceRunning(service.Service) {
    KillServiceByService(service.Service, service.Operations, verbose)
  } else {
    if verbose {
      fmt.Println("Service not currently running. Skipping.")
    }
  }
}

func LogsServiceByService(srv *def.Service, ops *def.ServiceOperation, verbose bool) {
  perform.DockerLogs(srv, ops, verbose)
}

func StartServiceByService(srvMain *def.Service, ops *def.ServiceOperation, verbose bool) {
  for _, srv := range srvMain.ServiceDeps {
    go StartServiceRaw(srv, verbose)
  }
  perform.DockerRun(srvMain, ops, verbose)
}

func KillServiceByService(srvMain *def.Service, ops *def.ServiceOperation, verbose bool) {
  for _, srv := range srvMain.ServiceDeps {
    go KillServiceRaw(srv, verbose)
  }
  perform.DockerStop(srvMain, ops, verbose)
}
