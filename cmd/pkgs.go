// +build !arm

package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
	Run: PackagesDo,
}

func addPackagesFlags() {
	packagesDo.Flags().StringVarP(&do.ChainName, "chain", "c", "", "chain to be used for deployment")
	packagesDo.Flags().StringSliceVarP(&do.ServicesSlice, "services", "s", []string{}, "comma separated list of services to start")
	packagesDo.Flags().StringVarP(&do.Path, "dir", "i", "", "root directory of app (will use $pwd by default)")
	packagesDo.Flags().BoolVarP(&do.Rm, "rm", "r", true, "remove containers after stopping")
	packagesDo.Flags().BoolVarP(&do.RmD, "rm-data", "x", true, "remove artifacts from host")
	packagesDo.Flags().StringVarP(&do.CSV, "output", "o", "", "results output type")
	packagesDo.Flags().StringVarP(&do.EPMConfigFile, "file", "f", "./epm.yaml", "path to package file which Eris PM should use")
	packagesDo.Flags().StringSliceVarP(&do.ConfigOpts, "set", "e", []string{}, "default sets to use; operates the same way as the [set] jobs, only before the epm file is ran (and after default address")
	packagesDo.Flags().BoolVarP(&do.OutputTable, "summary", "u", true, "output a table summarizing epm jobs")
	packagesDo.Flags().StringVarP(&do.PackagePath, "contracts-path", "p", "./contracts", "path to the contracts Eris PM should use")
	packagesDo.Flags().StringVarP(&do.ABIPath, "abi-path", "b", "./abi", "path to the abi directory Eris PM should use when saving ABIs after the compile process")
	packagesDo.Flags().StringVarP(&do.DefaultGas, "gas", "g", "1111111111", "default gas to use; can be overridden for any single job")
	packagesDo.Flags().StringVarP(&do.Compiler, "compiler", "l", formCompilers(), "IP:PORT of compiler which Eris PM should use")
	packagesDo.Flags().StringVarP(&do.DefaultAddr, "address", "a", "", "default address to use; operates the same way as the [account] job, only before the epm file is ran")
	packagesDo.Flags().StringVarP(&do.DefaultFee, "fee", "w", "1234", "default fee to use")
	packagesDo.Flags().StringVarP(&do.DefaultAmount, "amount", "y", "9999", "default amount to use")
	packagesDo.Flags().StringVarP(&do.ChainPort, "chain-port", "", "46657", "chain rpc port")
	packagesDo.Flags().StringVarP(&do.KeysPort, "keys-port", "", "4767", "port for keys server")
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
	util.IfExit(pkgs.RunPackage(do))
}

func formCompilers() string {
	verSplit := strings.Split(version.VERSION, ".")
	maj, _ := strconv.Atoi(verSplit[0])
	min, _ := strconv.Atoi(verSplit[1])
	pat, _ := strconv.Atoi(verSplit[2])
	return fmt.Sprintf("https://compilers.monax.io:1%01d%02d%01d", maj, min, pat)
}
