package chains

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"

	dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Install(cmd *cobra.Command, args []string) {

}

func New(cmd *cobra.Command, args []string) {

}

func Edit(cmd *cobra.Command, args []string) {
	checkChainGiven(args)
	EditChainRaw(args[0])
}

func Inspect(cmd *cobra.Command, args []string) {
	checkChainGiven(args)
	if len(args) == 1 {
		args = append(args, "all")
	}
	chain := LoadChainDefinition(args[0])
	if IsChainExisting(chain) {
		services.InspectServiceByService(chain.Service, chain.Operations, args[1], cmd.Flags().Lookup("verbose").Changed)
	}
}

func ListKnown() {
	chains := ListKnownRaw()
	for _, s := range chains {
		fmt.Println(s)
	}
}

func ListInstalled() {
	listChains(true)
}

func ListChains() {
	chains := ListExistingRaw()
	for _, s := range chains {
		fmt.Println(s)
	}
}

func ListRunning() {
	chains := ListRunningRaw()
	for _, s := range chains {
		fmt.Println(s)
	}
}

func ListKnownRaw() []string {
	chns := []string{}
	fileTypes := []string{}
	for _, t := range []string{"*.json", "*.yaml", "*.toml"} {
		fileTypes = append(fileTypes, filepath.Join(dir.BlockchainsPath, t))
	}
	for _, t := range fileTypes {
		s, _ := filepath.Glob(t)
		for _, s1 := range s {
			s1 = strings.Split(filepath.Base(s1), ".")[0]
			chns = append(chns, s1)
		}
	}
	return chns
}

func Rename(cmd *cobra.Command, args []string) {
	checkChainGiven(args)
	if len(args) != 2 {
		fmt.Println("Please give me: eris services rename [oldName] [newName]")
		return
	}
	RenameChainRaw(args[0], args[1], cmd.Flags().Lookup("verbose").Changed)
}

func Update(cmd *cobra.Command, args []string) {
	checkChainGiven(args)
	UpdateChainRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func Rm(cmd *cobra.Command, args []string) {
	checkChainGiven(args)
	RmChainRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func EditChainRaw(chainName string) {
	chainConf := loadChainDefinition(chainName)
	filePath := chainConf.ConfigFileUsed()
	dir.Editor(filePath)
}

func ListRunningRaw() []string {
	return listChains(false)
}

func ListExistingRaw() []string {
	return listChains(true)
}

func RenameChainRaw(oldName, newName string, verbose bool) {
	if parseKnown(oldName) {
		if verbose {
			fmt.Println("Renaming chain", oldName, "to", newName)
		}

		chainDef := LoadChainDefinition(oldName)

		perform.DockerRename(chainDef.Service, chainDef.Operations, oldName, newName, verbose)
		oldFile := chainDefFileByChainName(oldName)
		newFile := strings.Replace(oldFile, oldName, newName, 1)

		chainDef.Name = newName
		chainDef.Service.Name = ""
		chainDef.Service.Image = ""
		_ = WriteChainDefinitionFile(chainDef, newFile)

		data.RenameDataRaw(oldName, newName, verbose)

		os.Remove(oldFile)
	} else {
		if verbose {
			fmt.Println("I cannot find that chain. Please check the chain name you sent me.")
		}
	}
}

func UpdateChainRaw(chainName string, verbose bool) {
	chain := LoadChainDefinition(chainName)
	perform.DockerRebuild(chain.Service, chain.Operations, false, verbose)
}

func RmChainRaw(chainName string, verbose bool) {
	chain := LoadChainDefinition(chainName)
	perform.DockerRemove(chain.Service, chain.Operations, verbose)
	oldFile := chainDefFileByChainName(chainName)
	os.Remove(oldFile)
}
