package commands

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/version"

	"github.com/spf13/cobra"
)

var quiet bool

var VerSion = &cobra.Command{
	Use:   "version",
	Short: "display Eris's platform version",
	Long:  `display the current installed version of Eris`,
	Run:   DisplayVersion,
}

func buildVerSionCommand() {
	addVerSionFlags()
}

func addVerSionFlags() {
	VerSion.Flags().BoolVarP(&quiet, "quiet", "q", false, "machine readable output")
}

func DisplayVersion(cmd *cobra.Command, args []string) {
	var versionMessage string
	if version.COMMIT == "HEAD" {
		versionMessage = version.VERSION
	} else {
		versionMessage = fmt.Sprintf("%s (%s)", version.VERSION, version.COMMIT)
	}

	if !quiet {
		fmt.Println("Eris CLI Version: " + versionMessage)
	} else {
		fmt.Println(versionMessage)
	}
}
