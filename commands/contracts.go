package commands

import (
	"os"

	"github.com/eris-ltd/eris-cli/contracts"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// Primary Contracts Sub-Command
var Contracts = &cobra.Command{
	Use:   "contracts",
	Short: "Deploy, Test, and Manage Your Smart Contracts.",
	Long: `The contracts subcommand is used to test and deploy
smart contracts for use by your application.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// build the contracts subcommand
func buildContractsCommand() {
	// TODO: finish when the PR which is blocking
	//   eris files put --dir is integrated into
	//   ipfs
	// [zr] looks like that's now the case ...
	// XXX see https://github.com/ipfs/go-ipfs/pull/1845
	// Contracts.AddCommand(contractsImport)
	// Contracts.AddCommand(contractsExport)
	Contracts.AddCommand(contractsTest)
	Contracts.AddCommand(contractsDeploy)
	addContractsFlags()
}

var contractsImport = &cobra.Command{
	Use:   "import HASH PACKAGE",
	Short: "Pull a package of smart contracts from IPFS.",
	Long: `Pull a package of smart contracts from IPFS
via its hash and save it locally.`,
	Run: ContractsImport,
}

var contractsExport = &cobra.Command{
	Use:   "export",
	Short: "Post a package of smart contracts to IPFS.",
	Long:  `Post a package of smart contracts to IPFS.`,
	Run:   ContractsExport,
}

var contractsTest = &cobra.Command{
	Use:   "test",
	Short: "Test a package of smart contracts.",
	Long: `Test a package of smart contracts.

Tests can be structured using three different
test types.

1. epm - epm apps can be tested against tendermint style blockchains.
2. embark - embark apps can be tested against ethereum style blockchains.
3. truffle - HELP WANTED!
4. solUnit - pure solidity smart contract packages may be tested via solUnit test framework.
5. manual - a simple gulp task can be given to the test environment.`,
	Run: ContractsTest,
}

var contractsDeploy = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a package of smart contracts to a chain.",
	Long: `Deploy a package of smart contracts to a chain.

Deployments can be structured using three different
deploy types.

1. epm - epm apps can be deployed to tendermint style blockchains simply.
2. embark - embark apps can be deployed to an ethereum style blockchain simply.
3. truffle - HELP WANTED!
4. pyepm - IF THIS IS STILL A THING, HELP WANTED!
5. manual - a simple gulp task can be given to the deployer.`,
	Run: ContractsDeploy,
}

//----------------------------------------------------
// XXX todo deduplicate flags -> [zr] things get wonky with epm
func addContractsFlags() {
	contractsTest.Flags().StringVarP(&do.ChainName, "chain", "c", "", "chain to be used for testing")
	contractsTest.Flags().StringSliceVarP(&do.ServicesSlice, "services", "s", []string{}, "comma separated list of services to start")
	contractsTest.Flags().StringVarP(&do.Type, "type", "t", "mint", "app type paradigm to be used for testing (overrides package.json)")
	contractsTest.Flags().StringVarP(&do.Task, "task", "k", "", "gulp task to be ran (overrides package.json; forces --type manual)")
	contractsTest.Flags().StringVarP(&do.Path, "dir", "i", "", "root directory of app (will use $pwd by default)")
	contractsTest.Flags().BoolVarP(&do.Rm, "rm", "r", true, "remove containers after stopping")
	contractsTest.Flags().BoolVarP(&do.RmD, "rm-data", "x", true, "remove artifacts from host")
	contractsTest.Flags().StringVarP(&do.CSV, "output", "o", "", "results output type (EPM only)")
	contractsTest.Flags().StringVarP(&do.EPMConfigFile, "file", "f", "./epm.yaml", "path to package file which EPM should use (EPM only)")
	contractsTest.Flags().StringSliceVarP(&do.ConfigOpts, "set", "e", []string{}, "default sets to use; operates the same way as the [set] jobs, only before the epm file is ran (and after default address (EPM only)")
	contractsTest.Flags().BoolVarP(&do.OutputTable, "summary", "u", true, "output a table summarizing epm jobs (EPM only)")
	contractsTest.Flags().StringVarP(&do.ContractsPath, "contracts-path", "p", "./contracts", "path to the contracts EPM should use (EPM only)")
	contractsTest.Flags().StringVarP(&do.ABIPath, "abi-path", "b", "./abi", "path to the abi directory EPM should use when saving ABIs after the compile process (EPM only)")
	contractsTest.Flags().StringVarP(&do.DefaultGas, "gas", "g", "1111111111", "default gas to use; can be overridden for any single job (EPM only)")
	contractsTest.Flags().StringVarP(&do.Compiler, "compiler", "l", "https://compilers.eris.industries:9090", "<ip:port> of compiler which EPM should use (EPM only)")
	contractsTest.Flags().StringVarP(&do.DefaultAddr, "address", "a", "", "default address to use; operates the same way as the [account] job, only before the epm file is ran (EPM only)")
	contractsTest.Flags().StringVarP(&do.DefaultFee, "fee", "w", "1234", "default fee to use (EPM only)")
	contractsTest.Flags().StringVarP(&do.DefaultAmount, "amount", "y", "9999", "default amount to use (EPM only)")

	contractsDeploy.Flags().StringVarP(&do.ChainName, "chain", "c", "", "chain to be used for deployment")

	contractsDeploy.Flags().StringSliceVarP(&do.ServicesSlice, "services", "s", []string{}, "comma separated list of services to start")
	contractsDeploy.Flags().StringVarP(&do.Type, "type", "t", "mint", "app type paradigm to be used for deployment (overrides package.)")
	contractsDeploy.Flags().StringVarP(&do.Task, "task", "k", "", "gulp task to be ran (overrides package.json; forces --type manual)")
	contractsDeploy.Flags().StringVarP(&do.Path, "dir", "i", "", "root directory of app (will use $pwd by default)")

	contractsDeploy.Flags().BoolVarP(&do.Rm, "rm", "r", true, "remove containers after stopping")
	contractsDeploy.Flags().BoolVarP(&do.RmD, "rm-data", "x", true, "remove artifacts from host")

	contractsDeploy.Flags().StringVarP(&do.CSV, "output", "o", "", "results output type (EPM only)")
	contractsDeploy.Flags().StringVarP(&do.EPMConfigFile, "file", "f", "./epm.yaml", "path to package file which EPM should use (EPM only)")
	contractsDeploy.Flags().StringSliceVarP(&do.ConfigOpts, "set", "e", []string{}, "default sets to use; operates the same way as the [set] jobs, only before the epm file is ran (and after default address (EPM only)")
	contractsDeploy.Flags().BoolVarP(&do.OutputTable, "summary", "u", true, "output a table summarizing epm jobs (EPM only)")
	contractsDeploy.Flags().StringVarP(&do.ContractsPath, "contracts-path", "p", "./contracts", "path to the contracts EPM should use (EPM only)")
	contractsDeploy.Flags().StringVarP(&do.ABIPath, "abi-path", "b", "./abi", "path to the abi directory EPM should use when saving ABIs after the compile process (EPM only)")
	contractsDeploy.Flags().StringVarP(&do.DefaultGas, "gas", "g", "1111111111", "default gas to use; can be overridden for any single job (EPM only)")
	contractsDeploy.Flags().StringVarP(&do.Compiler, "compiler", "l", "https://compilers.eris.industries:9090", "<ip:port> of compiler which EPM should use (EPM only)")
	contractsDeploy.Flags().StringVarP(&do.DefaultAddr, "address", "a", "", "default address to use; operates the same way as the [account] job, only before the epm file is ran (EPM only)")
	contractsDeploy.Flags().StringVarP(&do.DefaultFee, "fee", "w", "1234", "default fee to use (EPM only)")
	contractsDeploy.Flags().StringVarP(&do.DefaultAmount, "amount", "y", "9999", "default amount to use (EPM only)")
}

//----------------------------------------------------

func ContractsImport(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "eq", cmd, args))
	do.Name = args[0]
	do.Path = args[1]
	IfExit(contracts.GetPackage(do))
}

func ContractsExport(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args))
	do.Name = args[0]
	IfExit(contracts.PutPackage(do))
	log.Println(do.Result)
}

func ContractsTest(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(0, "eq", cmd, args))
	if do.Path == "" {
		do.Path, _ = os.Getwd() // we aren't catching this error, but revisit later if it becomes a problem
	}
	do.Name = "test"
	IfExit(contracts.RunPackage(do))
}

func ContractsDeploy(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(0, "eq", cmd, args))
	if do.Path == "" {
		do.Path, _ = os.Getwd() // we aren't catching this error, but revisit later if it becomes a problem
	}
	do.Name = "deploy"
	IfExit(contracts.RunPackage(do))
}
