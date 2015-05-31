package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/eris-ltd/eris-cli/perform"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var Version = &cobra.Command{
	Use:   "version",
	Short: "Display Eris's Platform Version.",
	Long:  `Display the versions of what your platform.`,
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
	},
}

func printVersion() {
	fmt.Println("Eris CLI Version: " + VERSION)
	epmVer, err := perform.NativeCommandRaw("epm", "--version")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.TrimSpace(epmVer))
}
