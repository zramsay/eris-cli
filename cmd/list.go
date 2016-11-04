package commands

import (
	"github.com/eris-ltd/eris-cli/list"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/spf13/cobra"
)

var List = &cobra.Command{
	Use:   "ls",
	Short: "list all the things eris knows about",
	Long: `list all Eris service, chain, and data containers

The default output shows containers in three sections. The -a flag
adds a few additional informational columns for each container.

The -r flag limits the output to running services or chains only.

The --json flag dumps the container information in the JSON format.

The -f flag specifies an alternative format for the list, using the syntax
of Go text templates. If the fields to be displayed are separated by the
'\t' tab character, the output will be columnized.

The struct being passed to the template is:

  type Details struct {
    Type      string          // container type
    ShortName string          // chain, service, or data name
    FullName  string          // container name

    Labels map[string]string  // container labels
    Info   *docker.Container  // Docker client library Container info
  }

The full list of available fields can be observed by issuing
the [eris ls --json] command.

The default [eris ls] output is equivalent to this custom format:

  {{.ShortName}}\t{{asterisk .Info.State.Running}}\t
  {{short .Info.ID}}\t{{short (dependent .ShortName)}}

The are a few helper functions available to prefix the fields with:

  quote       quote the value
  toupper     make the value upper case
  ports       prettify exposed container ports (used as {{ports .Info}})
  short       shorten the container ID (or any other value) to 10 symbols
  asterisk    show the '*' symbol if the value is true, '-' otherwise
  dependent   find a dependent data container for the given service or chain
`,
	Example: `$ eris ls -rf '{{.ShortName}}, {{.Type}}, {{ports .Info}}'
$ eris ls  -f '{{.ShortName}}\t{{.Type}}\t{{.Info.NetworkSettings.IPAddress}}'
$ eris ls  -f '{{.ShortName}}\t{{.Info.Config.Env}}'`,
	Run: func(cmd *cobra.Command, args []string) {
		ListAll()
	},
}

func buildListCommand() {
	List.Flags().BoolVarP(&do.All, "all", "a", false, "show extended output")
	List.Flags().BoolVarP(&do.Running, "running", "r", false, "show only running containers")
	List.Flags().BoolVarP(&do.JSON, "json", "", false, "machine readable output")
	List.Flags().StringVarP(&do.Format, "format", "f", "", "alternate format for columnized output")
}

func ListAll() {
	if do.All {
		do.Format = "extended"
	}
	if do.JSON {
		do.Format = "json"
	}

	util.IfExit(list.Containers("all", do.Format, do.Running))
}
