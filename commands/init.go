package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var Init = &cobra.Command{
	Use:   "init",
	Short: "Initialize the ~/.eris directory with some default services and actions",
	Long: `Create the ~/.eris directory with actions and services subfolders and clone eris-ltd/eris-actions
		eris-ltd/eris-services into them, respectively`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(common.ErisRoot); err != nil {
			err := common.InitErisDir()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Root eris directory (%s) already exists. Please type `eris` to see the help.\n", common.ErisRoot)
			os.Exit(0)
		}

		c := exec.Command("git", "clone", "https://github.com/eris-ltd/eris-actions", common.ActionsPath)

		if Verbose {
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
		}
		if err := c.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)

		}

		c = exec.Command("git", "clone", "https://github.com/eris-ltd/eris-services", common.ServicesPath)
		if Verbose {
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
		}
		if err := c.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Initialized eris root directory (%s) with default actions and service files\n", common.ErisRoot)
	},
}
