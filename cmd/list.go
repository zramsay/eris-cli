package commands

import (
	"github.com/eris-ltd/eris-cli/list"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var List = &cobra.Command{
	Use:   "ls",
	Short: "List all the things eris knows about.",
	Long: `List all Eris services, chains, and data containers.

The default output shows containers in three sections. The -a flag adds a few
columns with additional information on container images and open ports.

The -r flag limits the output to running service or chains only.

The --json flag dumps the container information in the JSON format.

The -f flag specifies an alternate format for the list, using the syntax
of Go text templates.

The default output is equivalent to this format:

  {{asterisk .Info.State.Running}} {{.ShortName}}\t
  {{short .Info.ID}}\t{{short (dependent .ShortName)}}

The struct being passed to the template is:

  type Details struct {
    Type      string          // container type
    ShortName string          // chain, service, or data name
    FullName  string          // container name

    Labels map[string]string  // container labels
    Info   *docker.Container  // Docker client library Container info 
  }

The full list of useful Container struct fields can be checked by issuing
the [eris ls --json] command.

Each field in the in the format input can be separated with the '\t' symbol
to columnize the output.

The are a few helper functions available to prefix the fields with:

  quote       quote the value
  toupper     make the value upper case
  ports       prettify exposed container ports (used as {{ports .Info}})
  short       shorten the container ID (or any other value) to 10 symbols
  asterisk    show the '*' symbol if the value is true
  dependent   find a dependent data container for the given service or chain
`,
	Example: `$ eris ls -rf "{{.ShortName}}, {{.Type}}, {{ports .Info}}"
$ eris ls  -f "{{.ShortName}}\t{{.Type}}\t{{.Info.NetworkSettings.IPAddress}}"`,
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

	IfExit(list.Containers("all", do.Format, do.Running))
}
