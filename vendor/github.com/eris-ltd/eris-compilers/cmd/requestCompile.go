package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/eris-ltd/eris-compilers/perform"
	"github.com/eris-ltd/eris-compilers/version"
	"github.com/eris-ltd/eris-cli/log"

	"github.com/spf13/cobra"
)

func BuildCompileCommand() {
	CompilersCmd.AddCommand(compileCmd)
	addCompileFlags()
}

var (
	compilerPort  string
	compilerUrl   string
	compilerDir   string
	libraries     string
	compilerSSL   bool
	compilerLocal bool
	optimizeSolc  bool
)

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "compile your contracts either remotely or locally",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Errorf("Specify a contract to compile \n\n")
			CompilersCmd.Help()
			os.Exit(0)
		}
		url := createUrl()
		output, err := perform.RequestCompile(url, args[0], optimizeSolc, libraries)
		if err != nil {
			log.Error(err)
		}
		perform.PrintResponse(*output, true)
	},
}

func addCompileFlags() {
	compileCmd.Flags().StringVarP(&compilerPort, "port", "p", setDefaultPort(), "call listening port")
	compileCmd.Flags().StringVarP(&compilerUrl, "url", "u", setDefaultURL(), "set the url for where to compile your contracts (no http(s) or port, please)")
	compileCmd.Flags().StringVarP(&compilerDir, "dir", "D", setDefaultDirectoryRoute(), "directory location to search for on the remote server")
	compileCmd.Flags().StringVarP(&libraries, "libs", "L", "", "libraries string (libName:Address[, or whitespace]...)")
	compileCmd.Flags().BoolVarP(&compilerSSL, "ssl", "s", setCompilerSSL(), "call https")
	compileCmd.Flags().BoolVarP(&compilerLocal, "local", "l", setCompilerLocal(), "use local compilers to compile message (good for debugging or if server goes down)")
	compileCmd.Flags().BoolVarP(&optimizeSolc, "optimize", "o", setOptimizeSolc(), "optimize code (solidity only)")
}

func createUrl() string {
	if compilerLocal {
		return ""
	} else {
		if compilerSSL {
			return "https://" + compilerUrl + ":" + compilerPort + compilerDir
		} else {
			return "http://" + compilerUrl + ":" + compilerPort + compilerDir
		}
	}
}

func setOptimizeSolc() bool {
	return false
}

func setCompilerLocal() bool {
	return false
}

func setCompilerSSL() bool {
	return false
}

func setDefaultDirectoryRoute() string {
	return "/"
}

func setDefaultURL() string {
	return "compilers.monax.io"
}

func setDefaultPort() string {
	verSplit := strings.Split(version.VERSION, ".")
	maj, _ := strconv.Atoi(verSplit[0])
	min, _ := strconv.Atoi(verSplit[1])
	pat, _ := strconv.Atoi(verSplit[2])
	return fmt.Sprintf("1%01d%02d%01d", maj, min, pat)
}
