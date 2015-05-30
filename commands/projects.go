package commands

import (
  prj "github.com/eris-ltd/eris-cli/projects"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// Primary Projects Sub-Command
var projects = &cobra.Command{
  Use:   "project",
  Short: "Start, Stop, and Manage Projects or Applications.",
  Long:  `The projects subcommand is used to start, stop, and configure projects
or applications. Within the Eris platform, projects are a bundle
of services, configured to run in a specific manner. Projects may
be defined either by a package.json file in the root of an
application's directory or via a docker-compose.yml file in the
root of an application's directory. Projects are given a human
readable name so that they can Eris can checkout and operate
the application or project.`,
  Run:   func(cmd *cobra.Command, args []string) {
           prj.ListProjects()
         },
}

// Build the projects subcommand
func buildProjectsCommand() {
  projects.AddCommand(projectsGet)
  projects.AddCommand(projectsInstall)
  projects.AddCommand(projectsAdd)
  projects.AddCommand(projectsList)
  projects.AddCommand(projectsCheckout)
  projects.AddCommand(projectsConfig)
  projects.AddCommand(projectsServices)
  projects.AddCommand(projectsStart)
  projects.AddCommand(projectsStop)
  projects.AddCommand(projectsRename)
  projects.AddCommand(projectsRedefine)
  projects.AddCommand(projectsRm)
  projects.AddCommand(projectsClean)
}

// get a project from a remote (currently limited to github.com)
// flags to add: name, checkout
var projectsGet = &cobra.Command{
  Use:   "get [github.com/USER/REPO]",
  Short: "Get a project from Github.",
  Long:  `Retrieve a project from the internet (utilizes git clone) and install
the project's dependencies.

NOTE: This functionality is currently limited to github.com.`,
  Run:   func(cmd *cobra.Command, args []string) {
           prj.Get(cmd, args)
         },
}

// install dependencies
// flags to add: name, checkout
var projectsInstall = &cobra.Command{
  Use:   "install [package-definition-file]",
  Short: "Install a project's dependencies.",
  Long:  `Install a project's dependencies if those dependencies are defined services.`,
  Run:   func(cmd *cobra.Command, args []string) {
           prj.Install(cmd, args)
         },
}

// add brings a project into the eris projects tree
// flags to add: checkout
var projectsAdd = &cobra.Command{
  Use:   "add [name] [project-definition-file]",
  Short: "Adds a project to Eris.",
  Long:  `Projects may be defined either by a package.json file in the root
of an application's directory or via a docker-compose.yml file
in the root of an application's directory.`,
  Run:   func(cmd *cobra.Command, args []string) {
           prj.Add(cmd, args)
         },
}

// list known projects
var projectsList = &cobra.Command{
  Use:   "ls",
  Short: "List projects registered with Eris.",
  Long:  `List all projects registered with Eris. To add a project use:
eris project add [project-definition-file]`,
  Run:   func(cmd *cobra.Command, args []string) {
           prj.ListProjects()
         },
}

// checkout a known project
var projectsCheckout = &cobra.Command{
  Use:   "checkout [project-name]",
  Short: "Checkout a project registered with Eris.",
  Long:  `Checkout a project registered with Eris.`,
  Run:   func(cmd *cobra.Command, args []string) {
           prj.Checkout(cmd, args)
         },
}

// configure known projects
var projectsConfig = &cobra.Command{
  Use:   "config [service] [key]:[val]",
  Short: "Configure projects registered with Eris.",
  Long:  `Configure projects registered with Eris.`,
  Run:   func(cmd *cobra.Command, args []string) {
           prj.Configure(cmd, args)
         },
}

// list the services associated with the currently checked out project
// flags to add: verbose
var projectsServices = &cobra.Command{
  Use:   "services [name]",
  Short: "List services for a project.",
  Long:  `List services for a project. If no arguments are given, will
display the services for the currently checked out project.`,
  Run:   func(cmd *cobra.Command, args []string) {
           prj.ListServices(cmd, args)
         },
}

// start a project
// flags to add: execMethod
var projectsStart = &cobra.Command{
  Use:   "start [name]",
  Short: "Start a project registered with Eris.",
  Long:  `Start a project registered with Eris. If no [name] is give Eris
will simply start the currently checked out project. To stop a
project use: eris project kill [name].`,
  Run:   func(cmd *cobra.Command, args []string) {
           prj.Start(cmd, args)
         },
}

// stop a running project
var projectsStop = &cobra.Command{
  Use:   "kill [name]",
  Short: "Stop a running project.",
  Long:  `Stop a running project. If no [name] is give Eris
will simply stop the currently checked out project.`,
  Run:   func(cmd *cobra.Command, args []string) {
           prj.Kill(cmd, args)
         },
}

// rename known projects
var projectsRename = &cobra.Command{
  Use:   "rename [old] [new]",
  Short: "Rename a project registered with Eris.",
  Long:  `Rename a project registered with Eris. To add a project use:
eris project add [project-definition-file]`,
  Run:   func(cmd *cobra.Command, args []string) {
           prj.Rename(cmd, args)
         },
}

// change the package definition file for a known project
var projectsRedefine = &cobra.Command{
  Use:   "redefine [name] [package-definition-file]",
  Short: "Change a project's definition file.",
  Long:  `Change a project's definition file.`,
  Run:   func(cmd *cobra.Command, args []string) {
           prj.Redefine(cmd, args)
         },
}

// remove a known projects
// flags to add: clean
var projectsRm = &cobra.Command{
  Use:   "rm [name]",
  Short: "Remove a project registered with Eris.",
  Long:  `Remove a project registered with Eris. Will not delete the
project's data (chains, etc.). To remove all of the project's
data use: eris project clean [name]`,
  Run:   func(cmd *cobra.Command, args []string) {
           prj.Rm(cmd, args)
         },
}

// clean a project's data from the machine
// flags to add: force (no confirm)
var projectsClean = &cobra.Command{
  Use:   "clean",
  Short: "Clean a project's data from the machine.",
  Long:  `Clean a project's data from the machine and unregister the
project with Eris.`,
  Run:   func(cmd *cobra.Command, args []string) {
           prj.Clean(cmd, args)
         },
}
