package commands

// TODO salvage
/*
import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/eris-ltd/eris-pm/definitions"
	"github.com/eris-ltd/eris-pm/packages"
	"github.com/eris-ltd/eris-pm/util"
	"github.com/eris-ltd/eris-pm/version"

	common "github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"

	"github.com/spf13/cobra"
)

const VERSION = version.VERSION

// Global Do struct
var do *definitions.Do

// Defining the root command
var EPMCmd = &cobra.Command{
	Use:   "eris-pm",
	Short: "The Eris Package Manager Deploys and Tests Smart Contract Systems",
	Long: `The Eris Package Manager Deploys and Tests Smart Contract Systems

Made with <3 by Monax Industries.

Complete documentation is available at https://monax.io/docs/documentation
` + "\nVersion:\n  " + VERSION,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// TODO: make this better.... need proper epm config
		// need to be able to have variable writers (eventually)
		log.SetLevel(log.WarnLevel)
		if do.Verbose {
			log.SetLevel(log.InfoLevel)
		} else if do.Debug {
			log.SetLevel(log.DebugLevel)
		}

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

	Run: RunPackage,

	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Ensure that errors get written to screen and generally flush the log
		// log.Flush()
	},
}

func Execute() {
	InitEPM()
	AddGlobalFlags()
	EPMCmd.Execute()
}

func InitEPM() {
	do = definitions.NowDo()
}

// Flags that are to be used by commands are handled by the Do struct
// Define the persistent commands (globals)
func AddGlobalFlags() {
	EPMCmd.PersistentFlags().StringVarP(&do.YAMLPath, "file", "f", defaultFile(), "path to package file which EPM should use; default respects $EPM_FILE")
	EPMCmd.PersistentFlags().StringVarP(&do.ContractsPath, "contracts-path", "p", defaultContracts(), "path to the contracts EPM should use; default respects $EPM_CONTRACTS_PATH")
	EPMCmd.PersistentFlags().StringVarP(&do.ABIPath, "abi-path", "a", defaultAbi(), "path to the abi directory EPM should use when saving ABIs after the compile process; default respects $EPM_ABI_PATH")
	EPMCmd.PersistentFlags().StringVarP(&do.Chain, "chain", "c", defaultChain(), "<ip:port> of chain which EPM should use; default respects $EPM_CHAIN_ADDR")
	EPMCmd.PersistentFlags().StringVarP(&do.Signer, "sign", "s", defaultSigner(), "<ip:port> of signer daemon which EPM should use; default respects $EPM_SIGNER_ADDR")
	EPMCmd.PersistentFlags().StringVarP(&do.DefaultGas, "gas", "g", defaultGas(), "default gas to use; can be overridden for any single job; default respects $EPM_GAS")
	EPMCmd.PersistentFlags().StringVarP(&do.Compiler, "compiler", "m", defaultCompiler(), "<ip:port> of compiler which EPM should use; default respects $EPM_COMPILER_ADDR")
	EPMCmd.PersistentFlags().StringVarP(&do.DefaultAddr, "address", "r", defaultAddr(), "default address to use; operates the same way as the [account] job, only before the epm file is ran; default respects $EPM_ADDRESS")
	EPMCmd.PersistentFlags().StringSliceVarP(&do.DefaultSets, "set", "e", defaultSets(), "default sets to use; operates the same way as the [set] jobs, only before the epm file is ran (and after default address; default respects $EPM_SETS")
	EPMCmd.PersistentFlags().StringVarP(&do.DefaultFee, "fee", "n", defaultFee(), "default fee to use; default respects $EPM_FEE")
	EPMCmd.PersistentFlags().StringVarP(&do.DefaultAmount, "amount", "u", defaultAmount(), "default amount to use; default respects $EPM_AMOUNT")
	EPMCmd.PersistentFlags().StringVarP(&do.DefaultOutput, "output", "o", defaultOutput(), "output format which epm should use [csv,json]; default respects $EPM_OUTPUT_FORMAT")
	EPMCmd.PersistentFlags().BoolVarP(&do.Overwrite, "overwrite", "w", defaultOverwrite(), "overwrite jobs of similar names; defaults respects $EPM_OVERWRITE_APPROVE")
	EPMCmd.PersistentFlags().BoolVarP(&do.SummaryTable, "summary", "t", defaultSummaryTable(), "output a table summarizing epm jobs; default respects $EPM_SUMMARY_TABLE")
	EPMCmd.PersistentFlags().BoolVarP(&do.Verbose, "verbose", "v", defaultVerbose(), "verbose output; more output than no output flags; less output than debug level; default respects $EPM_VERBOSE")
	EPMCmd.PersistentFlags().BoolVarP(&do.Debug, "debug", "d", defaultDebug(), "debug level output; the most output available for epm; if it is too chatty use verbose flag; default respects $EPM_DEBUG")
}

//----------------------------------------------------
func RunPackage(cmd *cobra.Command, args []string) {
	common.IfExit(packages.RunPackage(do))
}

// ---------------------------------------------------
// Defaults

func defaultFile() string {
	return setDefaultString("EPM_FILE", "./epm.yaml")
}

func defaultContracts() string {
	return setDefaultString("EPM_CONTRACTS_PATH", "./contracts")
}

func defaultAbi() string {
	return setDefaultString("EPM_ABI_PATH", "./abi")
}

func defaultChain() string {
	return setDefaultString("EPM_CHAIN_ADDR", "localhost:46657")
}

func defaultSigner() string {
	return setDefaultString("EPM_SIGNER_ADDR", "localhost:4767")
}

func defaultCompiler() string {
	verSplit := strings.Split(version.VERSION, ".")
	maj, _ := strconv.Atoi(verSplit[0])
	min, _ := strconv.Atoi(verSplit[1])
	pat, _ := strconv.Atoi(verSplit[2])
	return setDefaultString("EPM_COMPILER_ADDR", fmt.Sprintf("https://compilers.monax.io:1%01d%02d%01d", maj, min, pat))
}

func defaultAddr() string {
	return setDefaultString("EPM_ADDRESS", "")
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

func defaultOutput() string {
	return setDefaultString("EPM_OUTPUT_FORMAT", "csv")
}

func defaultSummaryTable() bool {
	return setDefaultBool("EPM_SUMMARY_TABLE", true)
}

func defaultVerbose() bool {
	return setDefaultBool("EPM_VERBOSE", false)
}

func defaultOverwrite() bool {
	return setDefaultBool("EPM_OVERWRITE_APPROVE", false)
}

func defaultDebug() bool {
	return setDefaultBool("EPM_DEBUG", false)
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
}*/
