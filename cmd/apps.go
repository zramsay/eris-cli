package commands

import (
	"github.com/eris-ltd/eris-cli/apps"

	. "github.com/eris-ltd/common/go/common"

	"github.com/spf13/cobra"
)

// Primary Applications Sub-Command
var Applications = &cobra.Command{
	Use:   "applications",
	Short: "Start, Stop, and Manage Applications.",
	Long: `Start, stop, and manage applications.

Within the Eris platform, applications are a bundle of services,
and actions which are configured to run in a specific manner.
Applications may be defined either by a package.json file in the
root of an application's directory or via a docker-compose.yml
file in the root of an application's directory. Applications are
given a human readable name so that Eris can checkout and
operate the application or application.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// Build the applications subcommand
func buildApplicationsCommand() {
	Applications.AddCommand(applicationsNew)
	Applications.AddCommand(applicationsInstall)
	Applications.AddCommand(applicationsStart)
	Applications.AddCommand(applicationsEdit)
	Applications.AddCommand(applicationsStop)
	Applications.AddCommand(applicationsRm)
	addApplicationsFlags()
}

// new builds a application definition file
var applicationsNew = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new application definition file.",
	Long:  `Create a new application definition file optionally from a template.`,
	Run:   NewApplication,
}

// install dependencies
var applicationsInstall = &cobra.Command{
	Use:   "install [name] [package-definition-file]",
	Short: "Install a application's dependencies.",
	Long:  `Install a application's dependencies if those dependencies are defined services.`,
	Run:   InstallApplication,
}

// start a application
var applicationsStart = &cobra.Command{
	Use:   "start [name]",
	Short: "Start a application registered with Eris.",
	Long: `Start a application registered with Eris. If no [name] is give Eris
will simply start the currently checked out application. To stop a
application use: [eris applications stop name].`,
	Run: StartApplication,
}

// remove a known applications
var applicationsEdit = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit a application registered with Eris.",
	Long:  ``,
	Run:   EditApplication,
}

// stop a running application
var applicationsStop = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop a running application.",
	Long: `Stop a running application. If no [name] is give Eris
will simply stop the currently checked out application.`,
	Run: StopApplication,
}

// remove a known applications
var applicationsRm = &cobra.Command{
	Use:   "rm [name]",
	Short: "Remove a application registered with Eris.",
	Long: `Remove a application registered with Eris. Will not delete the
application's data (chains, etc.). To remove all of the application's
data use: [eris application clean name]`,
	Run: RmApplication,
}

//----------------------------------------------------------------------
// cli flags
func addApplicationsFlags() {
	// buildFlag(actionsDo, do, "quiet", "action")
	// buildFlag(actionsDo, do, "chain", "action")
	// buildFlag(actionsDo, do, "services", "action")

	// buildFlag(actionsRemove, do, "file", "action")

	// actionsList.Flags().BoolVarP(&do.Quiet, "quiet", "", false, "machine readable output; also used in tests")
}

//----------------------------------------------------------------------
// cli command wrappers

func NewApplication(cmd *cobra.Command, args []string) {
	// IfExit(ArgCheck(2, "eq", cmd, args))
	// do.Name = args[0]
	// do.Path = args[1]
	IfExit(apps.NewApps(do))
}

func InstallApplication(cmd *cobra.Command, args []string) {
	// IfExit(ArgCheck(2, "eq", cmd, args))
	// do.Name = args[0]
	// do.Path = args[1]
	IfExit(apps.InstallApps(do))
}

func StartApplication(cmd *cobra.Command, args []string) {
	// IfExit(ArgCheck(2, "eq", cmd, args))
	// do.Name = args[0]
	// do.Path = args[1]
	IfExit(apps.StartApps(do))
}

func EditApplication(cmd *cobra.Command, args []string) {
	// IfExit(ArgCheck(2, "eq", cmd, args))
	// do.Name = args[0]
	// do.Path = args[1]
	IfExit(apps.EditApps(do))
}

func StopApplication(cmd *cobra.Command, args []string) {
	// IfExit(ArgCheck(2, "eq", cmd, args))
	// do.Name = args[0]
	// do.Path = args[1]
	IfExit(apps.StopApps(do))
}

func RmApplication(cmd *cobra.Command, args []string) {
	// IfExit(ArgCheck(2, "eq", cmd, args))
	// do.Name = args[0]
	// do.Path = args[1]
	IfExit(apps.RmApps(do))
}
