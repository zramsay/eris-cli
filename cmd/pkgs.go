// +build !arm

package commands

import (
	"github.com/spf13/cobra"
)

// Below is deprecated and on schedule for elimination
var Packages = &cobra.Command{
	Use:   "pkgs",
	Short: "deploy, test, and manage your smart contract packages",
	Long: `the pkgs subcommand is used to test and deploy
smart contract packages for use by your application

Note: this subcommand is deprecated as of this version and will be moving into
the main command thread.`,
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
in a package definition file

Note: this subcommand is deprecated as of this version and will be moving into
the main command thread. Please use "monax run" instead!`,
	Run: JobsRun,
}

func addPackagesFlags() {
	packagesDo.Flags().StringVarP(&do.ChainName, "chain", "c", "", "chain name to be used for deployment")
	// TODO links keys
	packagesDo.Flags().StringVarP(&do.Signer, "keys", "s", defaultSigner(), "IP:PORT of keys daemon which jobs should use")
	packagesDo.Flags().StringVarP(&do.Path, "dir", "i", "", "root directory of app (will use $pwd by default)")              //what's this actually used for?
	packagesDo.Flags().StringVarP(&do.DefaultOutput, "output", "o", "json", "output format which should be used [csv,json]") // [zr] this is not well tested!
	packagesDo.Flags().StringVarP(&do.YAMLPath, "file", "f", "./epm.yaml", "path to package file which jobs should use")
	packagesDo.Flags().StringSliceVarP(&do.DefaultSets, "set", "e", []string{}, "default sets to use; operates the same way as the [set] jobs, only before the jobs file is ran (and after default address")
	packagesDo.Flags().StringVarP(&do.ContractsPath, "contracts-path", "p", "./contracts", "path to the contracts jobs should use")
	packagesDo.Flags().StringVarP(&do.BinPath, "bin-path", "", "./bin", "path to the bin directory jobs should use when saving binaries after the compile process")
	packagesDo.Flags().StringVarP(&do.ABIPath, "abi-path", "", "./abi", "path to the abi directory jobs should use when saving ABIs after the compile process")
	packagesDo.Flags().StringVarP(&do.DefaultGas, "gas", "g", "1111111111", "default gas to use; can be overridden for any single job")
	packagesDo.Flags().StringVarP(&do.DefaultAddr, "address", "a", "", "default address to use; operates the same way as the [account] job, only before the epm file is ran")
	packagesDo.Flags().StringVarP(&do.DefaultFee, "fee", "n", "9999", "default fee to use")
	packagesDo.Flags().StringVarP(&do.DefaultAmount, "amount", "u", "9999", "default amount to use")
	packagesDo.Flags().BoolVarP(&do.Overwrite, "overwrite", "t", true, "overwrite jobs of the same name")
}
