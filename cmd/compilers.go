package commands

import (
	"github.com/spf13/cobra"
)

var Compilers = &cobra.Command{
	Use:   "compilers",
	Short: "manage smart contract languages and versions",
	Long: `the compilers subcommand is a manager of different 
	smart contract development languages and their different versions.
	For now we only provide support for solidity but it's very easy to 
	contribute your own smart contracting language into here. Send the marmots a PR!`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func buildCompilersCommand() {
	Compilers.AddCommand(compilersUse)
	Compilers.AddCommand(compilersInstall)
	Compilers.AddCommand(compilersList)
	//addCompilersFlags()
}

var compilersUse = &cobra.Command{
	Use:   "use LANG VER",
	Short: "use a specific version of a compiler",
	Long: `assigns the default compiler to a specific language and version. 
Must already have the language and version to use`,
	Run: CheckoutCompiler,
}

var compilersInstall = &cobra.Command{
	Use:   "install LANG VER",
	Short: "install a compiler with a specific version",
	Long:  `pulls a docker image of a compiler with a specific tag for that version`,
	Run:   InstallCompiler,
}

var compilersList = &cobra.Command{
	Use:   "ls [LANG]",
	Short: "list different languages and versions available",
	Long:  `can also specify specific language`,
	Run:   ListCompilers,
}

var compilersRun = &cobra.Command{
	Use:   "compile",
	Short: "compiles a string of files",
	Long:  "Maps to an exec statement, all of these commands are hand coded.",
	//Run:   CompileCompiler,
}

func CheckoutCompiler(cmd *cobra.Command, args []string) {

}

func InstallCompiler(cmd *cobra.Command, args []string) {

}

func ListCompilers(cmd *cobra.Command, args []string) {

}
