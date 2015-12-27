package commands

import (
	"strings"

	act "github.com/eris-ltd/eris-cli/actions"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//----------------------------------------------------------------------
// cli definitions

// Primary Actions Sub-Command
// flags to add: --global --project
var Actions = &cobra.Command{
	Use:   "actions",
	Short: "Manage and perform structured actions.",
	Long: `Display and manage actions for various components of the
Eris platform and for the platform itself.

Actions are bundles of commands which rely upon a project
which is currently in scope or on a global set of actions.
Actions are held in yaml, toml, or json action-definition
files within the action folder in the eris tree (globally
scoped actions) or in a directory pointed to by the
actions field of the currently checked out project
(project scoped actions). Actions are a sequence of
commands which operate in a similar fashion to how a
circle.yml file or a .travis.yml script field may operate.

Actions execute in a series of individual sub-shells ran
on the host. Note actions do not run from inside containers
but can interact with containers either via the installed
eris commands or via the docker cli itself or, indeed, any
other programs installed *on the host*.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// Build the actions subcommand
func buildActionsCommand() {
	Actions.AddCommand(actionsNew)
	Actions.AddCommand(actionsImport)
	Actions.AddCommand(actionsList)
	Actions.AddCommand(actionsEdit)
	Actions.AddCommand(actionsDo)
	Actions.AddCommand(actionsExport)
	Actions.AddCommand(actionsRename)
	Actions.AddCommand(actionsRemove)
	addActionsFlags()
}

// Actions Sub-sub-Commands
var actionsImport = &cobra.Command{
	Use:   "import NAME LOCATION",
	Short: "Import an action definition file from Github or IPFS.",
	Long: `Import an action definition for your platform.

By default, Eris will import from ipfs.`,
	Example: "$ eris actions import \"do not use\" QmNUhPtuD9VtntybNqLgTTevUmgqs13eMvo2fkCwLLx5MX",
	Run:     ImportAction,
}

// flags to add: template
var actionsNew = &cobra.Command{
	Use:   "new NAME",
	Short: "Create a new action definition file.",
	Long:  `Create a new action definition file optionally from a template.`,
	Run:   NewAction,
}

//TODO [zr] (eventually) list all + flags, see issue #231
var actionsList = &cobra.Command{
	Use:   "ls",
	Short: "List all registered action definition files.",
	Long:  `List all registered action definition files.`,
	Run:   ListActions,
}

var actionsDo = &cobra.Command{
	Use:   "do NAME",
	Short: "Perform an action.",
	Long: `Perform an action according to the action definition file.

Actions are used to perform functions which are a
semi-scriptable series of steps. These are general
helper functions.

Actions are a series of commands passed to a series of
*individual* subshells. These actions can take a series
of arguments.

Arguments passed into the shells via the command line
(extra arguments which do not match the name) will be
available to the command steps as $1, $2, $3, etc.

In addition, variables will be populated within the
subshell according to the key:val syntax within the
command line.

The shells will be passed the host's environment as
well as any additional env vars added to the action
definition file.`,
	Example: `$ eris actions do dns register -- will run the ~/.eris/actions/dns_register action def file
$ eris actions do dns register name:cutemarm ip:111.111.111.111 -- will populate $name and $ip
$ eris actions do dns register cutemarm 111.111.111.111 -- will populate $1 and $2`,
	Run: DoAction,
}

var actionsEdit = &cobra.Command{
	Use:   "edit NAME",
	Short: "Edit an action definition file.",
	Long:  `Edit an action definition file in the default editor.`,
	Run:   EditAction,
}

var actionsExport = &cobra.Command{
	Use:   "export NAME",
	Short: "Export an action definition file to IPFS.",
	Long: `Export an action definition file to IPFS.

Command will return a machine readable version of the IPFS hash.`,
	Run: ExportAction,
}

var actionsRename = &cobra.Command{
	Use:     "rename OLD_NAME NEW_NAME",
	Short:   "Rename an action.",
	Long:    `Rename an action.`,
	Example: "$ eris actions rename OLD_NAME NEW_NAME",
	Run:     RenameAction,
}

var actionsRemove = &cobra.Command{
	Use:   "remove NAME",
	Short: "Remove an action definition file.",
	Long:  `Remove an action definition file.`,
	Run:   RmAction,
}

//----------------------------------------------------------------------
// cli flags
func addActionsFlags() {
	buildFlag(actionsDo, do, "quiet", "action")
	buildFlag(actionsDo, do, "chain", "action")
	buildFlag(actionsDo, do, "services", "action")

	buildFlag(actionsRemove, do, "file", "action")

}

//----------------------------------------------------------------------
// cli command wrappers

func ImportAction(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "eq", cmd, args))
	do.Name = args[0]
	do.Path = args[1]
	IfExit(act.ImportAction(do))
}

func NewAction(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	//do.Path = args[1] else index out of range...
	do.Operations.Args = args
	IfExit(act.NewAction(do))
}

func ListActions(cmd *cobra.Command, args []string) {
	// TODO: add scoping for when projects done.
	do.Known = true
	do.Running = false
	do.Existing = false
	if err := util.ListAll(do, "actions"); err != nil {
		return
	}
	for _, s := range strings.Split(do.Result, "\n") {
		log.Warn(strings.Replace(s, "_", " ", -1))
	}
}

func EditAction(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = strings.Join(args, "_")
	IfExit(act.EditAction(do))
}

func DoAction(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Operations.Args = args
	IfExit(act.Do(do))
}

func ExportAction(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = strings.Join(args, "_")
	IfExit(act.ExportAction(do))
}

func RenameAction(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "eq", cmd, args))
	do.Name = args[0]
	do.NewName = args[1]
	IfExit(act.RenameAction(do))
}

func RmAction(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Operations.Args = args
	IfExit(act.RmAction(do))
}
