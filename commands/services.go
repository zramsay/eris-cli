package commands

import (
  "github.com/eris-ltd/eris-cli/services"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// Primary Services Sub-Command
var cmdServices = &cobra.Command{
  Use:   "service",
  Short: "Start, Stop, and Manage Services Required for your Application",
  Long:  `The services subcommand is used to install, start, stop, and configure
the services needed to operate your application.`,
  Run:   func(cmd *cobra.Command, args []string) {
           services.ListInstalled()
         },
}

func buildServicesCommand() {
  cmdServices.AddCommand(cmdServicesListKnown)
  cmdServices.AddCommand(cmdServicesInstall)
  cmdServices.AddCommand(cmdServicesListInstalled)
  cmdServices.AddCommand(cmdServicesConfig)
  cmdServices.AddCommand(cmdServicesStart)
  cmdServices.AddCommand(cmdServicesListRunning)
  cmdServices.AddCommand(cmdServicesStop)
  cmdServices.AddCommand(cmdServicesUpdate)
}

// list-known lists the services which eris can automagically install
var cmdServicesListKnown = &cobra.Command{
  Use:   "known",
  Short: "List all the services which eris can install for your platform.",
  Long:  `Lists the services which eris can install for your platform. To install
a service, use: eris service install.`,
  Run:   func(cmd *cobra.Command, args []string) {
           services.ListKnown()
         },
}

// install a service
var cmdServicesInstall = &cobra.Command{
  Use:   "install",
  Short: "Install a Known Service Locally.",
  Long:  `Install a service for your platform. To list known services use:
eris service list-known.`,
  Run:   func(cmd *cobra.Command, args []string) {
           services.Install(cmd, args)
         },
}

// ls lists the services available locally
var cmdServicesListInstalled = &cobra.Command{
  Use:   "ls",
  Short: "List the installed services.",
  Long:  `Lists the installed services which eris knows about. To start a service
use: eris service start [service].`,
  Run:   func(cmd *cobra.Command, args []string) {
           services.ListInstalled()
         },
}

// configure a service definition
var cmdServicesConfig = &cobra.Command{
  Use:   "config",
  Short: "Configure a service definition file.",
  Long:  `Configures a service by reading from and writing to a service definition file
which is kept in ~/.eris/services.

NOTE: Do not use this command for configuring a *specific* blockchain. This
command will only operate on service definition files which tell Eris how to
start and stop a specific service. How that service is used for a specific
project is handled from project definition files. For more information on
project definition files please see: eris help project.`,
  Run:   func(cmd *cobra.Command, args []string) {
           services.Configure(cmd, args)
         },
}

// start a service
var cmdServicesStart = &cobra.Command{
  Use:   "start",
  Short: "Start a service.",
  Long:  `Starts a service according to the service operational definition file which
eris stores in the ~/.eris/services directory. To stop the service use:
eris service kill [service].`,
  Run:   func(cmd *cobra.Command, args []string) {
           services.Start(cmd, args)
         },
}

// ps lists the services which are currently running
var cmdServicesListRunning = &cobra.Command{
  Use:   "ps",
  Short: "Lists the running services.",
  Long:  `Lists the services which are currently running.`,
  Run:   func(cmd *cobra.Command, args []string) {
           services.ListRunning()
         },
}

// kill stops a running service
var cmdServicesStop = &cobra.Command{
  Use:   "kill",
  Short: "Stops a running service.",
  Long:  `Stops a services which is currently running.`,
  Run:   func(cmd *cobra.Command, args []string) {
           services.Kill(cmd, args)
         },
}

// updates an installed service
var cmdServicesUpdate = &cobra.Command{
  Use:   "update",
  Short: "Updates an installed service.",
  Long:  `Updates an installed service, or installs it if it has not been installed.`,
  Run:   func(cmd *cobra.Command, args []string) {
           services.Update(cmd, args)
         },
}