// +build !arm

package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/pkgs"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

	"github.com/eris-ltd/common/go/common"

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

		// Populates chainID from the chain (if its not passed)
		common.IfExit(util.GetChainID(do))
	},
	Run: PackagesDo,
}

func addPackagesFlags() {
	packagesDo.Flags().StringVarP(&do.ChainName, "chain", "c", "", "chain to be used for deployment")
	packagesDo.Flags().StringVarP(&do.Signer, "sign", "s", defaultSigner(), "<ip:port> of signer daemon which EPM should use; default respects $EPM_SIGNER_ADDR")
	//packagesDo.Flags().StringSliceVarP(&do.ServicesSlice, "services", "s", []string{}, "comma separated list of services to start")
	packagesDo.Flags().StringVarP(&do.Path, "dir", "i", "", "root directory of app (will use $pwd by default)") //what's this actually used for?
	//packagesDo.Flags().BoolVarP(&do.Rm, "rm", "r", true, "remove containers after stopping")
	//packagesDo.Flags().BoolVarP(&do.RmD, "rm-data", "x", true, "remove artifacts from host")
	packagesDo.Flags().StringVarP(&do.DefaultOutput, "output", "o", defaultOutput(), "output format which epm should use [csv,json]; default respects $EPM_OUTPUT_FORMAT")
	packagesDo.Flags().StringVarP(&do.YAMLPath, "file", "f", defaultFile(), "path to package file which EPM should use; default respects $EPM_FILE")
	packagesDo.Flags().StringSliceVarP(&do.DefaultSets, "set", "e", defaultSets(), "default sets to use; operates the same way as the [set] jobs, only before the epm file is ran (and after default address; default respects $EPM_SETS")
	packagesDo.Flags().BoolVarP(&do.SummaryTable, "summary", "", true, "output a table summarizing epm jobs")
	packagesDo.Flags().StringVarP(&do.ContractsPath, "contracts-path", "p", defaultContracts(), "path to the contracts EPM should use; default respects $EPM_CONTRACTS_PATH")
	packagesDo.Flags().StringVarP(&do.ABIPath, "abi-path", "", defaultAbi(), "path to the abi directory EPM should use when saving ABIs after the compile process; default respects $EPM_ABI_PATH")
	packagesDo.Flags().StringVarP(&do.DefaultGas, "gas", "g", defaultGas(), "default gas to use; can be overridden for any single job; default respects $EPM_GAS")
	packagesDo.Flags().StringVarP(&do.Compiler, "compiler", "l", formCompilers(), "IP:PORT of compiler which Eris PM should use")
	packagesDo.Flags().StringVarP(&do.DefaultAddr, "address", "a", defaultAddr(), "default address to use; operates the same way as the [account] job, only before the epm file is ran; default respects $EPM_ADDRESS")
	packagesDo.Flags().StringVarP(&do.DefaultFee, "fee", "n", defaultFee(), "default fee to use; default respects $EPM_FEE")
	packagesDo.Flags().StringVarP(&do.DefaultAmount, "amount", "u", defaultAmount(), "default amount to use; default respects $EPM_AMOUNT")
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
	// clears epm.log file
	util.ClearJobResults()
	util.IfExit(pkgs.RunPackage(do))
}

func formCompilers() string {
	verSplit := strings.Split(version.VERSION, ".")
	maj, _ := strconv.Atoi(verSplit[0])
	min, _ := strconv.Atoi(verSplit[1])
	pat, _ := strconv.Atoi(verSplit[2])
	return fmt.Sprintf("https://compilers.monax.io:1%01d%02d%01d", maj, min, pat)
}

// ---------------------------------------------------
// Defaults
func defaultSigner() string {
	return setDefaultString("EPM_SIGNER_ADDR", "localhost:4767")
}

func defaultFile() string {
	return setDefaultString("EPM_FILE", "./epm.yaml")
}

func defaultContracts() string {
	return setDefaultString("EPM_CONTRACTS_PATH", "./contracts")
}

func defaultAbi() string {
	return setDefaultString("EPM_ABI_PATH", "./abi")
}

func defaultAddr() string {
	return setDefaultString("EPM_ADDRESS", "")
}

func defaultOutput() string {
	return setDefaultString("EPM_OUTPUT_FORMAT", "csv")
}

func defaultFee() string {
	return setDefaultString("EPM_FEE", "1234")
}

func defaultAmount() string {
	return setDefaultString("EPM_AMOUNT", "9999")
}

func defaultSets() []string {
	return setDefaultStringSlice("EPM_SETS", []string{})
}

func defaultGas() string {
	return setDefaultString("EPM_GAS", "1111111111")
}

func setDefaultBool(envVar string, def bool) bool {
	env := os.Getenv(envVar)
	if env != "" {
		i, _ := strconv.ParseBool(env)
		return i
	}
	return def
}

func setDefaultString(envVar, def string) string {
	env := os.Getenv(envVar)
	if env != "" {
		return env
	}
	return def
}

func setDefaultStringSlice(envVar string, def []string) []string {
	env := os.Getenv(envVar)
	if env != "" {
		return strings.Split(env, ";")
	}
	return def
}
