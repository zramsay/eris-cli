package commands

import (
	"fmt"
	"strings"

	chns "github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/data"
	srv "github.com/eris-ltd/eris-cli/services"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var ListKnown = &cobra.Command{
	Use:   "known",
	Short: "List everything Eris knows about.",
	Long: `Lists all the services, chains, & data containers which Eris
has installed locally as definition files.

To install a new service, use: [eris services import].

To install a new chain, use: [eris chains import].

Services include all executable services supported by the Eris platform which are
NOT blockchains or key managers.

Blockchains are handled using the [eris chains] command.
`,
	Run: func(cmd *cobra.Command, args []string) {
		ListAllKnown()
	},
}

var ListExisting = &cobra.Command{
	Use:   "ls",
	Short: "List everything that is installed and built.",
	Long: `Lists the built services and chains known to Eris.

To list the known services: [eris services known]. This will show
all services which exist in the form of locally held service
definition files.

To list the running services: [eris services ps]. This is the service
which are currently running.

To start a service use: [eris services start serviceName].`,
	Run: func(cmd *cobra.Command, args []string) {
		ListAllExisting()
	},
}

func ListAllKnown() {
	if err := srv.ListKnown(do); err != nil {
		return
	}
	fmt.Printf("Services:\n%s \n\n", do.Result)

	if err := chns.ListKnown(do); err != nil {
		return
	}
	fmt.Printf("Chains:\n%s \n", do.Result)

}

//TODO make output pretty
func ListAllExisting() {
	if err := srv.ListExisting(do); err != nil {
		return
	}

	if err := chns.ListExisting(do); err != nil {
		return
	}
	//`data ls` calls this func, hence the confusing semantics
	//should it be renamed?
	if err := data.ListKnown(do); err != nil {
		return
	}

	// https://www.reddit.com/r/television/comments/2755ow/hbos_silicon_valley_tells_the_most_elaborate/
	datasToManipulate := do.Result
	for _, s := range strings.Split(datasToManipulate, "||") {
		fmt.Printf("Data Containers:\n%s \n", s)
	}

}
