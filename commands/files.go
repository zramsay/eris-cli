package commands

import (
	"github.com/eris-ltd/eris-cli/files"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// Primary Files Sub-Command
// Flags to add: ipfsHost
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
	Files.AddCommand(filesCache)
	Files.AddCommand(filesCat)
	Files.AddCommand(filesList)
	Files.AddCommand(filesCached)
	//	addFilesFlags()
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

var filesCache = &cobra.Command{
	Use:   "cache [fileHash]",
	Short: "Cache a file to IPFS.",
	Long: `Cache a file to IPFS' local daemon.
	
Caches a file locally via IPFS pin, by hash.`,
	Run: func(cmd *cobra.Command, args []string) {
		PinIt(cmd, args)
	},
}

var filesCat = &cobra.Command{
	Use:   "cat [fileHash]",
	Short: "Cat the contents of a file from IPFS.",
	Long:  "Cat the contents of a file from IPFS.",
	Run: func(cmd *cobra.Command, args []string) {
		CatIt(cmd, args)
	},
}

var filesList = &cobra.Command{
	Use:   "ls [objectHash]",
	Short: "List links from an IPFS object.",
	//TODO test listing up and down through DAG / Zach just learn the DAG.
	Long: "Lists object named by [objectHash/Path] and displays the link it contains.",
	Run: func(cmd *cobra.Command, args []string) {
		ListIt(cmd, args)
	},
}

var filesCached = &cobra.Command{
	Use:   "cached",
	Short: "Lists files cached locally.",
	Long:  "Displays list of files cached locally.",
	Run: func(cmd *cobra.Command, args []string) {
		PinnedLs(cmd, args)
	},
}

//--------------------------------------------------------------
// cli flags

func Get(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		cmd.Help()
		return
	}
	do.Name = args[0]
	do.Path = args[1]
	IfExit(files.GetFiles(do))
}

func Put(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Help()
		return
	}
	do.Name = args[0]
	err := files.PutFiles(do)
	IfExit(err)
	logger.Println(do.Result)
}

func PinIt(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Help()
		return
	}
	do.Name = args[0]
	err := files.PinFiles(do)
	IfExit(err)
	logger.Println(do.Result)
}

func CatIt(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Help()
		return
	}
	do.Name = args[0]
	err := files.CatFiles(do)
	IfExit(err)
	logger.Println(do.Result)

}

func ListIt(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Help()
		return
	}
	do.Name = args[0]
	err := files.ListFiles(do)
	IfExit(err)
	logger.Println(do.Result)
}

func PinnedLs(cmd *cobra.Command, args []string) {
	err := files.ListPinned(do)
	IfExit(err)
	logger.Println(do.Result)
}
