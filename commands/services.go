package commands

import (
  srv "github.com/eris-ltd/eris-cli/services"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// Primary Services Sub-Command
var Services = &cobra.Command{
  Use:   "services",
  Short: "Start, Stop, and Manage Services Required for your Application.",
  Long:  `Start, Stop, and Manage Services Required for your Application.

Services are all services known and used by the Eris platform with the
exception of blockchain services and key management services. Blockchain
services are managed and operated via the [eris chain] command while key
management services are managed via the [eris keys] command.`,
  Run:   func(cmd *cobra.Command, args []string) {
           srv.ListInstalled()
         },
}

// build the services subcommand
func buildServicesCommand() {
  Services.AddCommand(servicesListKnown)
  Services.AddCommand(servicesInstall)
  Services.AddCommand(servicesListInstalled)
  Services.AddCommand(servicesConfig)
  Services.AddCommand(servicesStart)
  Services.AddCommand(servicesListRunning)
  Services.AddCommand(servicesLogs)
  Services.AddCommand(servicesStop)
  Services.AddCommand(servicesUpdate)
  Services.AddCommand(servicesRm)
}

// list-known lists the services which eris can automagically install
// flags to add: --list-versions
var servicesListKnown = &cobra.Command{
  Use:   "known",
  Short: "List all the services which Eris can install.",
  Long:  `Lists the services which eris can install for your platform. To install
a service, use: [eris services install].

Services include all executable services supported by the Eris platform which are
NOT blockchains. Blockchains are handled using the [eris chains] command.`,
  Run:   func(cmd *cobra.Command, args []string) {
           srv.ListKnown()
         },
}

// install a service
var servicesInstall = &cobra.Command{
  Use:   "install [name] [version]",
  Short: "Install a Known Service Locally.",
  Long:  `Install a service for your platform. By default, Eris will install the
most recent version of a service unless another version is passed
as an argument. To list known services use:
[eris services known].`,
  Run:   func(cmd *cobra.Command, args []string) {
           srv.Install(cmd, args)
         },
}

// ls lists the services available locally
var servicesListInstalled = &cobra.Command{
  Use:   "ls",
  Short: "List the installed services.",
  Long:  `Lists the installed services which eris knows about. To start a service
use: [eris services start service].`,
  Run:   func(cmd *cobra.Command, args []string) {
           srv.ListInstalled()
         },
}

// configure a service definition
var servicesConfig = &cobra.Command{
  Use:   "config [name] [key]:[val]",
  Short: "Configure a service.",
  Long:  `Configures a service by reading from and writing to a service configuration file
which is kept in ~/.eris/services.

NOTE: Do not use this command for configuring a *specific* service. This
command will only operate on *service configuration file* which tell Eris
how to start and stop a specific service. How that service is used for a
specific project is handled from project definition files. For more
information on project definition files please see: [eris help projects].`,
  Run:   func(cmd *cobra.Command, args []string) {
           srv.Configure(cmd, args)
         },
}

// start a service
// flags to add: --identity (alias for --key), --foreground
var servicesStart = &cobra.Command{
  Use:   "start [name]",
  Short: "Start a service.",
  Long:  `Starts a service according to the service operational definition file which
eris stores in the ~/.eris/services directory. To stop the service use:
[eris services kill service].

[eris services start name] by default will put the service into the
background so its logs will not be viewable from the command line.
To view a service's logs use [eris services logs name].`,
  Run:   func(cmd *cobra.Command, args []string) {
           srv.Start(cmd, args)
         },
}

// ps lists the services which are currently running
// flags to add: --all (list installed)
var servicesListRunning = &cobra.Command{
  Use:   "ps",
  Short: "Lists the running services.",
  Long:  `Lists the services which are currently running.`,
  Run:   func(cmd *cobra.Command, args []string) {
           srv.ListRunning()
         },
}

// logs [name] displays the logs of a running service
// flags to add: --tail
var servicesLogs = &cobra.Command{
  Use:   "logs [name]",
  Short: "Displays the logs of a running service.",
  Long:  `Displays the logs of a running service.`,
  Run:   func(cmd *cobra.Command, args []string) {
           srv.Logs(cmd, args)
         },
}

// kill stops a running service
var servicesStop = &cobra.Command{
  Use:   "kill [name]",
  Short: "Stops a running service.",
  Long:  `Stops a services which is currently running.`,
  Run:   func(cmd *cobra.Command, args []string) {
           srv.Kill(cmd, args)
         },
}

// updates an installed service
var servicesUpdate = &cobra.Command{
  Use:   "update [name]",
  Short: "Updates an installed service.",
  Long:  `Updates an installed service, or installs it if it has not been installed.`,
  Run:   func(cmd *cobra.Command, args []string) {
           srv.Update(cmd, args)
         },
}

// rm [name] removes a service
// flags to add: --all (?), --force (no confirm)
var servicesRm = &cobra.Command{
  Use:   "rm [name]",
  Short: "Removes an installed service.",
  Long:  `Removes an installed service.`,
  Run:   func(cmd *cobra.Command, args []string) {
           srv.Rm(cmd, args)
         },
}