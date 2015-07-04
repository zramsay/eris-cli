package commands

import (
	"github.com/eris-ltd/eris-cli/files"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// Primary Files Sub-Command
var Files = &cobra.Command{
	Use:   "files",
	Short: "Manage Files containers for your Application.",
	Long: `The files subcommand is used to import, and export
files into containers for use by your application.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// build the files subcommand
func buildFilesCommand() {
	Files.AddCommand(filesImport)
	Files.AddCommand(filesExport)
}

var filesImport = &cobra.Command{
	Use:   "get [hash] [fileName]",
	Short: "Pull a file from IPFS via its hash and save it locally.",
	Long:  `Pull a file from IPFS via its hash and save it locally.`,
	Run: func(cmd *cobra.Command, args []string) {
		Get(cmd, args)
	},
}

var filesExport = &cobra.Command{
	Use:   "put [fileName]",
	Short: "Post a file to IPFS.",
	Long:  `Post a file to IPFS.`,
	Run: func(cmd *cobra.Command, args []string) {
		Put(cmd, args)
	},
}

func Get(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		cmd.Help()
		return
	}
	IfExit(files.GetFilesRaw(args[0], args[1]))
}

func Put(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Help()
		return
	}
	hash, err := files.PutFilesRaw(args[0])
	IfExit(err)
	logger.Println(hash)
}
