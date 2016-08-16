package commands

import (
	//"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/kaihei"

	"github.com/spf13/cobra"

	. "github.com/eris-ltd/common/go/common"
)

var Start = &cobra.Command{}

var Stop = &cobra.Command{}

func buildStartCommand() {
	addStartFlags()
}

func addStartFlags() {
}

func buildStopCommand() {
	addStopFlags()
}

func addStopFlags() {
}

func StartEris(cmd *cobra.Command, args []string) {
	IfExit(kaihei.StartUpEris(do))
}

func StopEris(cmd *cobra.Command, args []string) {
	IfExit(kaihei.ShutUpEris(do))
}
