package commands

import (
	"errors"
	"fmt"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/files"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/spf13/cobra"
)

var Files = &cobra.Command{
	Use:   "files",
	Short: "manage files needed for your application using IPFS",
	Long: `the files subcommand is used to import, and export
files to and from IPFS for use on the host machine

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
	Use:   "get HASH",
	Short: "pull files/objects from IPFS via a hash and save them locally, requires the [--output] flag",
	Long:  `pull files/objects from IPFS via a hash and save them locally, requires the [--output] flag`,
	Run:   FilesGet,
}

var filesExport = &cobra.Command{
	Use:   "put FILE|DIR",
	Short: "post files or whole directories to IPFS",
	Long: `post files or whole directories to IPFS
Directories will be added as objects in the MerkleDAG.`,
	Run: FilesPut,
}

var filesCache = &cobra.Command{
	Use:   "cache HASH",
	Short: "cache files to IPFS",
	Long: `cache files to IPFS' local daemon

It caches files locally via IPFS pin, by hash.

NOTE: "put" will "cache" recursively by default.`,
	Run: FilesPin,
}

var filesCat = &cobra.Command{
	Use:   "cat HASH",
	Short: "cat the contents of a file from IPFS",
	Long:  "cat the contents of a file from IPFS",
	Run:   FilesCat,
}

var filesList = &cobra.Command{
	Use:   "ls HASH",
	Short: "list links from an IPFS object",
	Long:  "List an object named by HASH/FILE and display the link it contains",
	Run:   FilesList,
}

var filesCached = &cobra.Command{
	Use:   "cached",
	Short: "list files cached locally",
	Long:  `display list of files cached locally`,
	Run:   FilesManageCached,
}

func addFilesFlags() {
	filesImport.Flags().StringVarP(&do.Path, "output", "o", "", "specify a path/name to output the file/directory. this flag is required")

	filesExport.Flags().StringVarP(&do.Gateway, "gateway", "", "", "specify a hosted gateway. default is IPFS' gateway; type \"eris\" for our gateway, or use your own with \"http://yourhost\"")

	filesCached.Flags().BoolVarP(&do.Rm, "rma", "", false, "remove all cached files")
	filesCached.Flags().StringVarP(&do.Hash, "rm", "", "", "remove a cached file by hash")
}

func FilesGet(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "eq", cmd, args))
	do.Hash = args[0]
	if do.Path == "" {
		util.IfExit(errors.New("please specify a path to output your file with the [--output] flag"))
	}
	util.IfExit(files.GetFiles(do))
}

func FilesPut(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "eq", cmd, args))
	do.Name = args[0]
	out, err := files.PutFiles(do)
	util.IfExit(err)
	fmt.Fprintln(config.Global.Writer, out)
}

func FilesPin(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "eq", cmd, args))
	do.Name = args[0]
	out, err := files.PinFiles(do)
	util.IfExit(err)
	fmt.Fprintln(config.Global.Writer, out)
}

func FilesCat(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "eq", cmd, args))
	do.Name = args[0]
	out, err := files.CatFiles(do)
	util.IfExit(err)
	fmt.Fprintln(config.Global.Writer, out)
}

func FilesList(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "eq", cmd, args))
	do.Name = args[0]
	out, err := files.ListFiles(do)
	util.IfExit(err)
	fmt.Fprintln(config.Global.Writer, out)
}

func FilesManageCached(cmd *cobra.Command, args []string) {
	out, err := files.ManagePinned(do)
	util.IfExit(err)
	fmt.Fprintln(config.Global.Writer, out)
}
