package commands

import (
	"os"

	"github.com/eris-ltd/eris-cli/pkgs"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
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
	// TODO: finish when the PR which is blocking
	//   eris files put --dir is integrated into
	//   ipfs
	// [zr] looks like that's now the case ...
	// XXX see https://github.com/ipfs/go-ipfs/pull/1845
	// Packages.AddCommand(packagesImport)
	// Packages.AddCommand(packagesExport)
	Packages.AddCommand(packagesDo)
	addPackagesFlags()
}

var packagesImport = &cobra.Command{
	Use:   "import HASH PACKAGE",
	Short: "Pull a package of smart contracts from IPFS.",
	Long: `Pull a package of smart contracts from IPFS
via its hash and save it locally.`,
	Run: PackagesImport,
}

var packagesExport = &cobra.Command{
	Use:   "export",
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
	packagesDo.Flags().StringVarP(&do.Compiler, "compiler", "l", "https://compilers.eris.industries:9090", "<ip:port> of compiler which EPM should use")
	packagesDo.Flags().StringVarP(&do.DefaultAddr, "address", "a", "", "default address to use; operates the same way as the [account] job, only before the epm file is ran")
	packagesDo.Flags().StringVarP(&do.DefaultFee, "fee", "w", "1234", "default fee to use")
	packagesDo.Flags().StringVarP(&do.DefaultAmount, "amount", "y", "9999", "default amount to use")
}

//----------------------------------------------------

func PackagesImport(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "eq", cmd, args))
	do.Name = args[0]
	do.Path = args[1]
	IfExit(pkgs.GetPackage(do))
}

func PackagesExport(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args))
	do.Name = args[0]
	IfExit(pkgs.PutPackage(do))
	log.Warn(do.Result)
}

func PackagesDo(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(0, "eq", cmd, args))
	if do.Path == "" {
		var err error
		do.Path, err = os.Getwd()
		IfExit(err)
	}
	IfExit(pkgs.RunPackage(do))
}
