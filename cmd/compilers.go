package commands

import (
	"github.com/spf13/cobra"
)

var Compilers = &cobra.Command{
	Use:   "compilers",
	Short: "manage smart contract compiler versions",
	Long: `manage smart contract compiler versions. 

	The compilers subcommand is a manager of different smart contract development languages versions.
	For now the marmots only provide support for solidity.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func buildCompilersCommand() {
	Compilers.AddCommand(compilersUse)
	Compilers.AddCommand(compilersInstall)
	Compilers.AddCommand(compilersList)
	//addCompilersFlags()
}

var compilersUse = &cobra.Command{
	Use:   "use VERSION",
	Short: "use a specific version of a compiler.",
	Long: `use a specific version of a compiler.
	Assigns the default compiler to a specific version. 
	Must already have the version installed to use.`,
	Run: CheckoutCompiler,
}

var compilersInstall = &cobra.Command{
	Use:   "install VERSION",
	Short: "install a compiler with a specific version",
	Long: `install a compiler with a specific version. 

	Pulls a docker image of a compiler with a specific tag for that version`,
	Run: InstallCompiler,
}

var compilersList = &cobra.Command{
	Use:   "ls",
	Short: "list different versions available",
	Long:  `list different versions available.`,
	Run:   ListCompilers,
}

var compilersRun = &cobra.Command{
	Use:   "compile",
	Short: "compiles a string of files",
	Long: `compiles a string of files. 
	Maps to an exec statement, all of these commands are hand coded.`,
	//Run:   CompileCompiler,
}

func CheckoutCompiler(cmd *cobra.Command, args []string) {

}

func InstallCompiler(cmd *cobra.Command, args []string) {

}

func ListCompilers(cmd *cobra.Command, args []string) {

}
