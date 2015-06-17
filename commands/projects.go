package commands

import (
	prj "github.com/eris-ltd/eris-cli/projects"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// Primary Projects Sub-Command
var Projects = &cobra.Command{
	Use:   "projects",
	Short: "Start, Stop, and Manage Projects or Applications.",
	Long: `Start, stop, and manage projects or applications.

Within the Eris platform, projects are a bundle of services,
and actions which are configured to run in a specific manner.
Projects may be defined either by a package.json file in the
root of an application's directory or via a docker-compose.yml
file in the root of an application's directory. Projects are
given a human readable name so that Eris can checkout and
operate the application or project.`,
}

// Build the projects subcommand
func buildProjectsCommand() {
	Projects.AddCommand(projectsGet)
	Projects.AddCommand(projectsNew)
	Projects.AddCommand(projectsAdd)
	Projects.AddCommand(projectsInstall)
	Projects.AddCommand(projectsList)
	Projects.AddCommand(projectsCheckout)
	Projects.AddCommand(projectsConfig)
	Projects.AddCommand(projectsServices)
	Projects.AddCommand(projectsActions)
	Projects.AddCommand(projectsStart)
	Projects.AddCommand(projectsStop)
	Projects.AddCommand(projectsRename)
	Projects.AddCommand(projectsRedefine)
	Projects.AddCommand(projectsRm)
	Projects.AddCommand(projectsClean)
}

// get a project definition file from a remote (currently limited to github.com and ipfs)
// flags to add: --checkout
var projectsGet = &cobra.Command{
	Use:   "get [name] [github.com/USER/REPO] || [name] [ipfs hash]",
	Short: "Get a project from Github or IPFS.",
	Long: `Retrieve a project from the internet (utilizes git clone or ipfs)
and install the project's dependencies.

NOTE: This functionality is currently limited to github.com and IPFS.`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.Get(cmd, args)
	},
}

// new builds a project definition file
// flags to add: --template, --checkout, --format
var projectsNew = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new project definition file.",
	Long:  `Create a new project definition file optionally from a template.`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.New(cmd, args)
	},
}

// add brings a project into the eris projects tree
// flags to add: --checkout
var projectsAdd = &cobra.Command{
	Use:   "add [name] [project-definition-file]",
	Short: "Add a project to Eris.",
	Long: `Projects may be defined either by a package.json file in the root
of an application's directory or via a docker-compose.yml file
in the root of an application's directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.Add(cmd, args)
	},
}

// install dependencies
// flags to add: --checkout
var projectsInstall = &cobra.Command{
	Use:   "install [name] [package-definition-file]",
	Short: "Install a project's dependencies.",
	Long:  `Install a project's dependencies if those dependencies are defined services.`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.Install(cmd, args)
	},
}

// list known projects
var projectsList = &cobra.Command{
	Use:   "ls",
	Short: "List projects registered with Eris.",
	Long: `List all projects registered with Eris. To add a project use:
[eris projects add project-definition-file]`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.ListProjects()
	},
}

// checkout a known project
var projectsCheckout = &cobra.Command{
	Use:   "checkout [project-name]",
	Short: "Checkout a project registered with Eris.",
	Long:  `Checkout a project registered with Eris.`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.Checkout(cmd, args)
	},
}

// configure known projects
var projectsConfig = &cobra.Command{
	Use:   "config [service] [key]:[val]",
	Short: "Configure projects registered with Eris.",
	Long:  `Configure projects registered with Eris.`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.Configure(cmd, args)
	},
}

// list the services associated with the currently checked out project
// flags to add: --verbose
var projectsServices = &cobra.Command{
	Use:   "services [name]",
	Short: "List services for a project.",
	Long: `List services for a project. If no arguments are given, will
display the services for the currently checked out project.`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.ListServices(cmd, args)
	},
}

// list the actions associated with the currently checked out project
// flags to add: --verbose
var projectsActions = &cobra.Command{
	Use:   "actions [name]",
	Short: "List actions for a project.",
	Long: `List actions for a project. If no arguments are given, will
display the actions for the currently checked out project.`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.ListActions(cmd, args)
	},
}

// start a project
// flags to add: --method (execution method)
var projectsStart = &cobra.Command{
	Use:   "start [name]",
	Short: "Start a project registered with Eris.",
	Long: `Start a project registered with Eris. If no [name] is give Eris
will simply start the currently checked out project. To stop a
project use: [eris projects kill name].`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.Start(cmd, args)
	},
}

// stop a running project
var projectsStop = &cobra.Command{
	Use:   "kill [name]",
	Short: "Stop a running project.",
	Long: `Stop a running project. If no [name] is give Eris
will simply stop the currently checked out project.`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.Kill(cmd, args)
	},
}

// rename known projects
var projectsRename = &cobra.Command{
	Use:   "rename [old] [new]",
	Short: "Rename a project registered with Eris.",
	Long: `Rename a project registered with Eris. To add a project use:
eris project add [project-definition-file]`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.Rename(cmd, args)
	},
}

// change the package definition file for a known project
var projectsRedefine = &cobra.Command{
	Use:   "redefine [name] [package-definition-file]",
	Short: "Change a project's definition file.",
	Long:  `Change a project's definition file.`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.Redefine(cmd, args)
	},
}

// remove a known projects
// flags to add: --clean
var projectsRm = &cobra.Command{
	Use:   "rm [name]",
	Short: "Remove a project registered with Eris.",
	Long: `Remove a project registered with Eris. Will not delete the
project's data (chains, etc.). To remove all of the project's
data use: [eris project clean name]`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.Rm(cmd, args)
	},
}

// clean a project's data from the machine
// flags to add: --force (no confirm)
var projectsClean = &cobra.Command{
	Use:   "clean",
	Short: "Clean a project's data from the machine.",
	Long: `Clean a project's data from the machine and unregister the
project with Eris.`,
	Run: func(cmd *cobra.Command, args []string) {
		prj.Clean(cmd, args)
	},
}
