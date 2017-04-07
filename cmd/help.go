package commands

import "github.com/spf13/cobra"

const helpTemplate = `Usage: {{.UseLine}}{{if .Runnable}}{{if .HasSubCommands}} COMMAND{{end}}{{if .HasFlags}} [FLAG...]{{end}}{{end}}

{{.Long}}
{{if gt .Aliases 0}}
Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}
Examples:
{{ .Example }}{{end}}{{ if .HasAvailableSubCommands}}
Available commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{ if .HasInheritedFlags}}

Global flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsHelpCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasSubCommands }}

Use "{{.CommandPath}} COMMAND --help" for more information about a command.{{end}}
`

var Help = &cobra.Command{
	Use:   "help COMMAND",
	Short: "Help about a command",
	Long: `Provide help for any command in the application.
Type monax help COMMAND for full details.`,
	PersistentPreRun:  func(cmd *cobra.Command, args []string) {},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {},

	Run: func(c *cobra.Command, args []string) {
		cmd, _, e := c.Root().Find(args)
		if cmd == nil || e != nil {
			c.Printf("Unknown help topic %#q.", args)
			c.Root().Usage()
		} else {

			helpFunc := cmd.HelpFunc()
			helpFunc(cmd, args)
		}
	},
}
