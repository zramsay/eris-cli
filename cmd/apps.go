package commands

import (
	"github.com/eris-ltd/eris-cli/apps"

	. "github.com/eris-ltd/common/go/common"

	"github.com/spf13/cobra"
)

var Applications = &cobra.Command{
	Use:   "applications",
	Short: "start, stop, and wanage applications",
	Long: `start, stop, and manage applications

Within the Eris platform, applications are a bundle of services,
and actions which are configured to run in a specific manner.
Applications may be defined either by a package.json file in the
root of an application's directory or via a docker-compose.yml
file in the root of an application's directory. Applications are
given a human readable name so that Eris can checkout and
operate the application or application.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func buildApplicationsCommand() {
	Applications.AddCommand(applicationsNew)
	Applications.AddCommand(applicationsInstall)
	Applications.AddCommand(applicationsStart)
	Applications.AddCommand(applicationsEdit)
	Applications.AddCommand(applicationsStop)
	Applications.AddCommand(applicationsRm)
}

var applicationsNew = &cobra.Command{
	Use:   "new [name]",
	Short: "create a new application definition file",
	Long:  `create a new application definition file optionally from a template`,
	Run:   NewApplication,
}

var applicationsInstall = &cobra.Command{
	Use:   "install NAME DEFINITION",
	Short: "install a application's dependencies",
	Long:  `install a application's dependencies if those dependencies are defined services`,
	Run:   InstallApplication,
}

var applicationsStart = &cobra.Command{
	Use:   "start NAME",
	Short: "start a application registered with Eris",
	Long: `start a application registered with Eris. If no NAME is give Eris
will simply start the currently checked out application. To stop a
application use: [eris applications stop name]`,
	Run: StartApplication,
}

var applicationsEdit = &cobra.Command{
	Use:   "edit NAME",
	Short: "edit a application registered with Eris",
	Long:  ``,
	Run:   EditApplication,
}

var applicationsStop = &cobra.Command{
	Use:   "stop NAME",
	Short: "stop a running application",
	Long: `stop a running application. If no NAME is give Eris
will simply stop the currently checked out application`,
	Run: StopApplication,
}

var applicationsRm = &cobra.Command{
	Use:   "rm NAME",
	Short: "remove an application registered with Eris",
	Long: `remove a application registered with Eris. Will not delete the
application's data (chains, etc.) To remove all of the application's
data use: [eris application clean name]`,
	Run: RmApplication,
}

func NewApplication(cmd *cobra.Command, args []string) {
	IfExit(apps.NewApps(do))
}

func InstallApplication(cmd *cobra.Command, args []string) {
	IfExit(apps.InstallApps(do))
}

func StartApplication(cmd *cobra.Command, args []string) {
	IfExit(apps.StartApps(do))
}

func EditApplication(cmd *cobra.Command, args []string) {
	IfExit(apps.EditApps(do))
}

func StopApplication(cmd *cobra.Command, args []string) {
	IfExit(apps.StopApps(do))
}

func RmApplication(cmd *cobra.Command, args []string) {
	IfExit(apps.RmApps(do))
}
