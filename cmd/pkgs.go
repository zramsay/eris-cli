package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/eris-ltd/eris-cli/pkgs"
	"github.com/eris-ltd/eris-cli/version"

	. "github.com/eris-ltd/common/go/common"

	"github.com/spf13/cobra"
)

// Primary Packages Sub-Command
var Packages = &cobra.Command{
	Use:   "pkgs",
	Short: "Deploy, Test, and Manage Your Smart Contract Packages.",
	Long: `The pkgs subcommand is used to test and deploy
smart contract packages for use by your application.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// build the contracts subcommand
func buildPackagesCommand() {
	Packages.AddCommand(packagesDo)
	Packages.AddCommand(packagesImport)
	Packages.AddCommand(packagesExport)
	addPackagesFlags()
}

var packagesImport = &cobra.Command{
	Use:   "import HASH PACKAGE",
	Short: "Pull a package of smart contracts from IPFS.",
	Long: `Pull a package of smart contracts from IPFS
via its hash and save it locally to ~/.eris/apps/PACKAGE.`,
	Run: PackagesImport,
}

var packagesExport = &cobra.Command{
	Use:   "export DIR",
	Short: "Post a package of smart contracts to IPFS.",
	Long:  `Post a package of smart contracts to IPFS.`,
	Run:   PackagesExport,
}

var packagesDo = &cobra.Command{
	Use:   "do",
	Short: "Deploy or test a package of smart contracts to a chain.",
	Long: `Deploy or test a package of smart contracts to a chain.

eris pkgs do will perform the required functionality included
in a package definition file.`,
	Run: PackagesDo,
}

//----------------------------------------------------
// XXX todo deduplicate flags -> [zr] things get wonky with epm
func addPackagesFlags() {
	packagesDo.Flags().StringVarP(&do.ChainName, "chain", "c", "", "chain to be used for deployment")
	packagesDo.Flags().StringSliceVarP(&do.ServicesSlice, "services", "s", []string{}, "comma separated list of services to start")
	packagesDo.Flags().StringVarP(&do.Path, "dir", "i", "", "root directory of app (will use $pwd by default)")
	packagesDo.Flags().BoolVarP(&do.Rm, "rm", "r", true, "remove containers after stopping")
	packagesDo.Flags().BoolVarP(&do.RmD, "rm-data", "x", true, "remove artifacts from host")
	packagesDo.Flags().StringVarP(&do.CSV, "output", "o", "", "results output type")
	packagesDo.Flags().StringVarP(&do.EPMConfigFile, "file", "f", "./epm.yaml", "path to package file which EPM should use")
	packagesDo.Flags().StringSliceVarP(&do.ConfigOpts, "set", "e", []string{}, "default sets to use; operates the same way as the [set] jobs, only before the epm file is ran (and after default address")
	packagesDo.Flags().BoolVarP(&do.OutputTable, "summary", "u", true, "output a table summarizing epm jobs")
	packagesDo.Flags().StringVarP(&do.PackagePath, "contracts-path", "p", "./contracts", "path to the contracts EPM should use")
	packagesDo.Flags().StringVarP(&do.ABIPath, "abi-path", "b", "./abi", "path to the abi directory EPM should use when saving ABIs after the compile process")
	packagesDo.Flags().StringVarP(&do.DefaultGas, "gas", "g", "1111111111", "default gas to use; can be overridden for any single job")
	packagesDo.Flags().StringVarP(&do.Compiler, "compiler", "l", formCompilers(), "<ip:port> of compiler which EPM should use")
	packagesDo.Flags().StringVarP(&do.DefaultAddr, "address", "a", "", "default address to use; operates the same way as the [account] job, only before the epm file is ran")
	packagesDo.Flags().StringVarP(&do.DefaultFee, "fee", "w", "1234", "default fee to use")
	packagesDo.Flags().StringVarP(&do.DefaultAmount, "amount", "y", "9999", "default amount to use")
	packagesDo.Flags().BoolVarP(&do.Overwrite, "overwrite", "t", true, "overwrite jobs of the same name")
}

//----------------------------------------------------

func PackagesImport(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "eq", cmd, args))
	do.Hash = args[0]
	do.Name = args[1]
	IfExit(pkgs.ImportPackage(do))
}

func PackagesExport(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args))
	do.Name = args[0]
	IfExit(pkgs.ExportPackage(do))
	//log.Warn(do.Result) -> handled in above func
}

func PackagesDo(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(0, "eq", cmd, args))
	if do.Path == "" {
		var err error
		do.Path, err = os.Getwd()
		IfExit(err)
	}
	if do.ChainName == "" {
		IfExit(fmt.Errorf("please provide the name of a running chain with --chain"))
	}
	if do.DefaultAddr == "" {
		IfExit(fmt.Errorf("please provide the address to deploy from with --address"))
	}
	IfExit(pkgs.RunPackage(do))
}

func formCompilers() string {
	verSplit := strings.Split(version.VERSION, ".")
	maj, _ := strconv.Atoi(verSplit[0])
	min, _ := strconv.Atoi(verSplit[1])
	pat, _ := strconv.Atoi(verSplit[2])
	return fmt.Sprintf("https://compilers.eris.industries:1%01d%02d%01d", maj, min, pat)
}
