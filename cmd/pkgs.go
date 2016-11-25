// +build !arm

package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/pkgs"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

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

[eris pkgs do] will perform the required functionality included
in a package definition file`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// clears epm.log file
		util.ClearJobResults()

		// Welcomer....
		log.Info("Hello! I'm EPM.")

		// Fixes path issues and controls for mint-client / eris-keys assumptions
		// util.BundleHttpPathCorrect(do)
		util.PrintPathPackage(do)

	},
	Run: PackagesDo,
}

func addPackagesFlags() {
	packagesDo.Flags().StringVarP(&do.ChainName, "chain", "c", "", "chain to be used for deployment")
	packagesDo.Flags().StringVarP(&do.Signer, "keys", "s", "http://172.17.0.2:4767", "IP:PORT of keys daemon which EPM should use")
	packagesDo.Flags().StringVarP(&do.Path, "dir", "i", "", "root directory of app (will use $pwd by default)") //what's this actually used for?
	packagesDo.Flags().StringVarP(&do.DefaultOutput, "output", "o", "json", "output format which epm should use [csv,json]")
	packagesDo.Flags().StringVarP(&do.YAMLPath, "file", "f", "./epm.yaml", "path to package file which EPM should use")
	packagesDo.Flags().StringSliceVarP(&do.DefaultSets, "set", "e", []string{}, "default sets to use; operates the same way as the [set] jobs, only before the epm file is ran (and after default address")
	packagesDo.Flags().StringVarP(&do.ContractsPath, "contracts-path", "p", "./contracts", "path to the contracts EPM should use")
	packagesDo.Flags().StringVarP(&do.ABIPath, "abi-path", "", "./abi", "path to the abi directory EPM should use when saving ABIs after the compile process")
	packagesDo.Flags().StringVarP(&do.DefaultGas, "gas", "g", "1111111111", "default gas to use; can be overridden for any single job")
	packagesDo.Flags().StringVarP(&do.Compiler, "compiler", "l", formCompilers(), "IP:PORT of compiler which Eris PM should use")
	packagesDo.Flags().StringVarP(&do.DefaultAddr, "address", "a", "", "default address to use; operates the same way as the [account] job, only before the epm file is ran")
	packagesDo.Flags().StringVarP(&do.DefaultFee, "fee", "n", "1234", "default fee to use")
	packagesDo.Flags().StringVarP(&do.DefaultAmount, "amount", "u", "9999", "default amount to use")
	packagesDo.Flags().BoolVarP(&do.Overwrite, "overwrite", "t", true, "overwrite jobs of the same name")
	packagesDo.Flags().BoolVarP(&do.LocalCompiler, "local-compiler", "z", false, "use a local compiler service; overwrites anything added to compilers flag")
}

func PackagesDo(cmd *cobra.Command, args []string) {
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

	var err error
	do.ChainIP, err = chains.GetChainIP(do)
	if err != nil {
		util.IfExit(err)
	}

	util.IfExit(pkgs.RunPackage(do))
}

func formCompilers() string {
	verSplit := strings.Split(version.VERSION, ".")
	maj, _ := strconv.Atoi(verSplit[0])
	min, _ := strconv.Atoi(verSplit[1])
	pat, _ := strconv.Atoi(verSplit[2])
	return fmt.Sprintf("https://compilers.monax.io:1%01d%02d%01d", maj, min, pat)
}
