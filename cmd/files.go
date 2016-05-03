package commands

import (
	"github.com/eris-ltd/eris-cli/files"

	log "github.com/Sirupsen/logrus"
	. "github.com/eris-ltd/common/go/common"
	"github.com/spf13/cobra"
)

// Primary Files Sub-Command
// Flags to add: ipfsHost
var Files = &cobra.Command{
	Use:   "files",
	Short: "Manage files needed for your application using IPFS.",
	Long: `The files subcommand is used to import, and export
files to and from IPFS for use on the host machine.

These commands are provided in addition to the various
functionality which is included throughout the tool, such as
services import or services export which operate more
precisely. The eris files command is used as a general wrapper
around an IPFS gateway which would be running as eris services ipfs.

At times, due to the manner in which IPFS boots files commands
will fail. If you get errors when running eris files commands
then please run [eris services start ipfs] give that a second
or two to boot and then retry the eris files command which failed.`,
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
	Use:   "get HASH FILE/DIR",
	Short: "Pull files/objects from IPFS via a hash and save them locally.",
	Long:  `Pull files/objects from IPFS via a hash and save them locally.`,
	Run:   FilesGet,
}

var filesExport = &cobra.Command{
	Use:   "put FILE/DIR",
	Short: "Post files or whole directories to IPFS.",
	Long: `Post files or whole directories to IPFS.
Directories will be added as objects in the MerkleDAG.`,
	Run: FilesPut,
}

var filesCache = &cobra.Command{
	Use:   "cache HASH",
	Short: "Cache files to IPFS.",
	Long: `Cache files to IPFS' local daemon.

It caches files locally via IPFS pin, by hash.

NOTE: "put" will "cache" recursively by default.`,
	Run: FilesPin,
}

var filesCat = &cobra.Command{
	Use:   "cat HASH",
	Short: "Cat the contents of a file from IPFS.",
	Long:  "Cat the contents of a file from IPFS.",
	Run:   FilesCat,
}

var filesList = &cobra.Command{
	Use:   "ls HASH",
	Short: "List links from an IPFS object.",
	//TODO [zr] test listing up and down through DAG
	Long: "List an object named by HASH/FILE and display the link it contains.",
	Run:  FilesList,
}

var filesCached = &cobra.Command{
	Use:   "cached",
	Short: "List files cached locally.",
	Long:  `Display list of files cached locally.`,
	Run:   FilesManageCached,
}

//--------------------------------------------------------------
// cli flags
func addFilesFlags() {

	filesExport.Flags().StringVarP(&do.Gateway, "gateway", "", "", "specify a hosted gateway. default is IPFS' gateway; type \"eris\" for our gateway, or use your own with \"http://yourhost\"")

	filesCached.Flags().BoolVarP(&do.Rm, "rma", "", false, "remove all cached files")
	filesCached.Flags().StringVarP(&do.Hash, "rm", "", "", "remove a cached file by hash")
}

func FilesGet(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "eq", cmd, args))
	do.Hash = args[0]
	do.Path = args[1] // where it is saved
	// TODO make above a flag with `-o` (--output)
	// similar to curl GET -o
	IfExit(files.GetFiles(do))
}

func FilesPut(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args))
	do.Name = args[0]
	IfExit(files.PutFiles(do))
	log.Warn(do.Result)
}

func FilesPin(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args))
	do.Name = args[0]
	IfExit(files.PinFiles(do))
	log.Warn(do.Result)
}

func FilesCat(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args))
	do.Name = args[0]
	IfExit(files.CatFiles(do))
	log.Warn(do.Result)

}

func FilesList(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "eq", cmd, args))
	do.Name = args[0]
	IfExit(files.ListFiles(do))
	log.Warn(do.Result)
}

func FilesManageCached(cmd *cobra.Command, args []string) {
	IfExit(files.ManagePinned(do))
	log.Warn(do.Result)
}
