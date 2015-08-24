package commands

import (
	"github.com/eris-ltd/eris-cli/files"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
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
	addFilesFlags()
}

var filesImport = &cobra.Command{
	Use:   "get [hash] [fileName]",
	Short: "Pull files from IPFS via a hash and save them locally.",
	Long: `Pull files from IPFS via a hash and save them locally.
	
Optionally pass in a csv with: get --csv=[fileName]`,
	Run: func(cmd *cobra.Command, args []string) {
		Get(cmd, args)
	},
}

var filesExport = &cobra.Command{
	Use:   "put [fileName]",
	Short: "Post files to IPFS.",
	Long: `Post files to IPFS. 
	
Optionally post all contents of a directory with: put [dirName] --dir`,
	Run: func(cmd *cobra.Command, args []string) {
		Put(cmd, args)
	},
}

var filesCache = &cobra.Command{
	Use:   "cache [fileHash]",
	Short: "Cache files to IPFS.",
	Long: `Cache files to IPFS' local daemon.
	
Caches a files locally via IPFS pin, by hash.
Optionally pass in a csv with: cache --csv=[fileName]
Note: "put" will "cache" recursively by default`,
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
	Long:  `Displays list of files cached locally.`,
	Run: func(cmd *cobra.Command, args []string) {
		ManageCached(cmd, args)
	},
}

//--------------------------------------------------------------
// cli flags
func addFilesFlags() {

	filesImport.Flags().StringVarP(&do.CSV, "csv", "", "", "specify a .csv with entries of format: hash,fileName")
	filesImport.Flags().StringVarP(&do.NewName, "dirname", "", "", "name of new directory to dump IPFS files from --csv")
	//maybe add flag to specify the gateway one wants to use?
	filesExport.Flags().BoolVarP(&do.Gateway, "gateway", "", false, "put files to Eris' hosted gateway")
	filesExport.Flags().BoolVarP(&do.AddDir, "dir", "", false, "add all files from a directory (note: this will not create an ipfs object). returns a log file (ipfs_hashes.csv) to pass into `eris files get`")

	//command will ignore fileName but that's ok
	filesCache.Flags().StringVarP(&do.CSV, "csv", "", "", "specify a .csv with entries of format: hash,fileName")

	filesCached.Flags().BoolVarP(&do.Rm, "rma", "", false, "remove all cached files")
	filesCached.Flags().StringVarP(&do.Hash, "rm", "", "", "remove a cached file by hash")
}

func Get(cmd *cobra.Command, args []string) {
	if do.CSV == "" {
		IfExit(ArgCheck(2, "eq", cmd, args))
		do.Name = args[0]
		do.Path = args[1]
	} else {
		do.Name = ""
		do.Path = ""
	}
	IfExit(files.GetFiles(do))
}

func Put(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args))

	do.Name = args[0]
	err := files.PutFiles(do)
	IfExit(err)
	logger.Println(do.Result)
}

func PinIt(cmd *cobra.Command, args []string) {
	if do.CSV == "" {
		if len(args) != 1 {
			cmd.Help()
			return
		}
		do.Name = args[0]
	} else {
		do.Name = ""
	}
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

func ManageCached(cmd *cobra.Command, args []string) {
	err := files.ManagePinned(do)
	IfExit(err)
	logger.Println(do.Result)
}
