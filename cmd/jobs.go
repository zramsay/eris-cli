package commands

import (
	"fmt"
	"os"
	"runtime"

	"github.com/monax/cli/pkgs"
	"github.com/monax/cli/util"

	"github.com/spf13/cobra"
)

func buildRunCommand() {
	addRunFlags()
}

var Run = &cobra.Command{
	Use:   "run",
	Short: "deploy or test a package of smart contracts to a chain",
	Long: `deploy or test a package of smart contracts onto a chain

[monax run] will perform the required functionality included
in a jobs definition file`,
	Run: JobsRun,
}

func addRunFlags() {
	Run.Flags().StringVarP(&do.ChainName, "chain", "c", "", "chain name to be used for deployment")
	// TODO links keys
	Run.Flags().StringVarP(&do.Signer, "keys", "s", defaultSigner(), "IP:PORT of keys daemon which jobs should use")
	Run.Flags().StringVarP(&do.Path, "dir", "i", "", "root directory of app (will use $pwd by default)")              //what's this actually used for?
	Run.Flags().StringVarP(&do.DefaultOutput, "output", "o", "json", "output format which should be used [csv,json]") // [zr] this is not well tested!
	Run.Flags().StringVarP(&do.YAMLPath, "file", "f", "./epm.yaml", "path to package file which jobs should use")
	Run.Flags().StringSliceVarP(&do.DefaultSets, "set", "e", []string{}, "default sets to use; operates the same way as the [set] jobs, only before the jobs file is ran (and after default address")
	Run.Flags().StringVarP(&do.ContractsPath, "contracts-path", "p", "./contracts", "path to the contracts jobs should use")
	Run.Flags().StringVarP(&do.BinPath, "bin-path", "", "./bin", "path to the bin directory jobs should use when saving binaries after the compile process")
	Run.Flags().StringVarP(&do.ABIPath, "abi-path", "", "./abi", "path to the abi directory jobs should use when saving ABIs after the compile process")
	Run.Flags().StringVarP(&do.DefaultGas, "gas", "g", "1111111111", "default gas to use; can be overridden for any single job")
	Run.Flags().StringVarP(&do.DefaultAddr, "address", "a", "", "default address to use; operates the same way as the [account] job, only before the epm file is ran")
	Run.Flags().StringVarP(&do.DefaultFee, "fee", "n", "9999", "default fee to use")
	Run.Flags().StringVarP(&do.DefaultAmount, "amount", "u", "9999", "default amount to use")
	Run.Flags().BoolVarP(&do.Overwrite, "overwrite", "t", true, "overwrite jobs of the same name")
}

func JobsRun(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(0, "eq", cmd, args))
	if do.Path == "" {
		var err error
		do.Path, err = os.Getwd()
		util.IfExit(err)
	}
	if do.ChainName == "" {
		util.IfExit(fmt.Errorf("please provide the name of a running chain with --chain"))
	}
	if do.DefaultAddr == "" { // note that this is not strictly necessary since the addr can be set in the epm.yaml.
		util.IfExit(fmt.Errorf("please provide the address to deploy from with --address"))
	}

	util.IfExit(pkgs.RunPackage(do))
}

func defaultSigner() string {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		ip, err := util.DockerWindowsAndMacIP(do)
		util.IfExit(err)
		return fmt.Sprintf("http://%v:4767", ip)
	} else {
		util.DockerConnect(false, "monax")
		keysName := util.ServiceContainerName("keys")
		cont, err := util.DockerClient.InspectContainer(keysName)
		if err != nil {
			return fmt.Sprintf("error will be caught by cli failing: %v", util.DockerError(err))
		}
		return fmt.Sprintf("http://%s:4767", cont.NetworkSettings.IPAddress)
	}
}
