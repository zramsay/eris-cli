package commands

import (
	"fmt"
	"runtime"

	"github.com/monax/monax/pkgs"
	"github.com/monax/monax/util"

	"github.com/spf13/cobra"
)

var Packages = &cobra.Command{
	Use:   "pkgs",
	Short: "deploy, test, and manage your smart contract packages",
	Long: `the pkgs subcommand is used to test and deploy
smart contract packages for use by your application`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func buildPackagesCommand() {
	Packages.AddCommand(packagesDo)
	addPackagesFlags()
}

var packagesDo = &cobra.Command{
	Use:   "do",
	Short: "deploy or test a package of smart contracts to a chain",
	Long: `deploy or test a package of smart contracts to a chain

[monax pkgs do] will perform the required functionality included
in a package definition file`,
	Run: PackagesDo,
}

func addPackagesFlags() {
	packagesDo.Flags().StringVarP(&do.ChainName, "chain", "c", "", "chain name to be used for deployment")
	packagesDo.Flags().StringVarP(&do.ChainURL, "chain-url", "", "", "chain-url to be used in tcp://IP:PORT format (only necessary for cluster and remote operations)")
	packagesDo.Flags().StringVarP(&do.Signer, "keys", "s", defaultSigner(), "IP:PORT of keys daemon which jobs should use")
	packagesDo.Flags().StringVarP(&do.Path, "dir", "i", "", "root directory of app (will use $pwd by default)")
	packagesDo.Flags().StringVarP(&do.DefaultOutput, "output", "o", "epm.output.json", "filename for jobs output file. by default, this name will reflect the name passed in on the optional [--file]")
	packagesDo.Flags().StringVarP(&do.YAMLPath, "file", "f", "epm.yaml", "path to package file which jobs should use. if also using the --dir flag, give the relative path to jobs file, which should be in the same directory")
	packagesDo.Flags().StringSliceVarP(&do.DefaultSets, "set", "e", []string{}, "default sets to use; operates the same way as the [set] jobs, only before the jobs file is ran (and after default address")
	// the package manager does not use this flag!
	// packagesDo.Flags().StringVarP(&do.ContractsPath, "contracts-path", "p", "./contracts", "path to the contracts jobs should use")
	packagesDo.Flags().StringVarP(&do.BinPath, "bin-path", "", "./bin", "path to the bin directory jobs should use when saving binaries after the compile process")
	packagesDo.Flags().StringVarP(&do.ABIPath, "abi-path", "", "./abi", "path to the abi directory jobs should use when saving ABIs after the compile process")
	packagesDo.Flags().StringVarP(&do.DefaultGas, "gas", "g", "1111111111", "default gas to use; can be overridden for any single job")
	packagesDo.Flags().StringVarP(&do.DefaultAddr, "address", "a", "", "default address to use; operates the same way as the [account] job, only before the epm file is ran")
	packagesDo.Flags().StringVarP(&do.DefaultFee, "fee", "n", "9999", "default fee to use")
	packagesDo.Flags().StringVarP(&do.DefaultAmount, "amount", "u", "9999", "default amount to use")
	packagesDo.Flags().BoolVarP(&do.Overwrite, "overwrite", "t", true, "overwrite jobs of the same name")
}

func PackagesDo(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(0, "eq", cmd, args))
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
