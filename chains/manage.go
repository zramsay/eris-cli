package chains

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//----------------------------------------------------------------------

func Install(cmd *cobra.Command, args []string) {

}

func New(cmd *cobra.Command, args []string) {

}

func Import(cmd *cobra.Command, args []string) {
	checkChainGiven(args)
	if len(args) != 2 {
		fmt.Println("Please give me: eris chains import [name] [location]")
		return
	}
	err := ImportChainRaw(args[0], args[1], cmd.Flags().Lookup("verbose").Changed)
	if err != nil {
		fmt.Println(err)
	}
}

func Edit(cmd *cobra.Command, args []string) {
	checkChainGiven(args)
	name := args[0]
	var configVals []string
	if len(args) > 0 {
		configVals = args[1:]
	}
	IfExit(EditChainRaw(name, configVals))
}

func Inspect(cmd *cobra.Command, args []string) {
	checkChainGiven(args)
	if len(args) == 1 {
		args = append(args, "all")
	}
	chain, err := LoadChainDefinition(args[0])
	IfExit(err)
	if IsChainExisting(chain) {
		services.InspectServiceByService(chain.Service, chain.Operations, args[1], cmd.Flags().Lookup("verbose").Changed)
	}
}

func Export(cmd *cobra.Command, args []string) {
	checkChainGiven(args)
	err := ExportChainRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
	if err != nil {
		fmt.Println(err)
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
	RmChainRaw(args[0], cmd.Flags().Lookup("force").Changed, cmd.Flags().Lookup("verbose").Changed)
}

//----------------------------------------------------------------------

func ImportChainRaw(chainName, path string, verbose bool) error {
	fileName := filepath.Join(BlockchainsPath, chainName)
	if filepath.Ext(fileName) == "" {
		fileName = fileName + ".toml"
	}

	s := strings.Split(path, ":")
	if s[0] == "ipfs" {

		var err error
		if verbose {
			err = util.GetFromIPFS(s[1], fileName, os.Stdout)
		} else {
			err = util.GetFromIPFS(s[1], fileName, bytes.NewBuffer([]byte{}))
		}

		if err != nil {
			return err
		}
		return nil
	}

	if strings.Contains(s[0], "github") {
		fmt.Println("https://twitter.com/ryaneshea/status/595957712040628224")
		return nil
	}

	fmt.Println("I do not know how to get that file. Sorry.")
	return nil
}

func ExportChainRaw(chainName string, verbose bool) error {
	chain, err := LoadChainDefinition(chainName)
	if err != nil {
		return err
	}
	if IsChainExisting(chain) {
		ipfsService := services.LoadServiceDefinition("ipfs")

		if services.IsServiceRunning(ipfsService.Service) {
			if verbose {
				fmt.Println("IPFS is running. Adding now.")
			}

			hash, err := exportFile(chainName, verbose)
			if err != nil {
				return err
			}
			fmt.Println(hash)
		} else {
			if verbose {
				fmt.Println("IPFS is not running. Starting now.")
			}
			services.StartServiceByService(ipfsService.Service, ipfsService.Operations, verbose)

			hash, err := exportFile(chainName, verbose)
			if err != nil {
				return err
			}
			fmt.Println(hash)
		}

	} else {
		return fmt.Errorf(`I don't known of that chain.
Please retry with a known chain.
To find known chains use: eris chains known`)
	}
	return nil
}

func EditChainRaw(chainName string, configVals []string) error {
	chainConf, err := readChainDefinition(chainName)
	if err != nil {
		return err
	}
	if err := util.EditRaw(chainConf, configVals); err != nil {
		return err
	}
	var chain def.Chain
	marshalChainDefinition(chainConf, &chain)
	return WriteChainDefinitionFile(&chain, chainConf.ConfigFileUsed())
}

func ListKnownRaw() []string {
	chns := []string{}
	fileTypes := []string{}
	for _, t := range []string{"*.json", "*.yaml", "*.toml"} {
		fileTypes = append(fileTypes, filepath.Join(BlockchainsPath, t))
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

func ListRunningRaw() []string {
	return listChains(false)
}

func ListExistingRaw() []string {
	return listChains(true)
}

func RenameChainRaw(oldName, newName string, verbose bool) error {
	if isKnownChain(oldName) {
		if verbose {
			fmt.Println("Renaming chain", oldName, "to", newName)
		}

		chainDef, err := LoadChainDefinition(oldName)
		if err != nil {
			return err
		}

		perform.DockerRename(chainDef.Service, chainDef.Operations, oldName, newName, verbose)
		oldFile, err := configFileNameFromChainName(oldName)
		if err != nil {
			return err
		}
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
	return nil
}

func UpdateChainRaw(chainName string, verbose bool) error {
	chain, err := LoadChainDefinition(chainName)
	if err != nil {
		return err
	}
	perform.DockerRebuild(chain.Service, chain.Operations, false, verbose)
	return nil
}

func RmChainRaw(chainName string, force, verbose bool) error {
	chain, err := LoadChainDefinition(chainName)
	if err != nil {
		return err
	}
	perform.DockerRemove(chain.Service, chain.Operations, verbose)

	if force {
		oldFile, err := configFileNameFromChainName(chainName)
		if err != nil {
			return err
		}
		os.Remove(oldFile)
	}
	return nil
}

func exportFile(chainName string, verbose bool) (string, error) {
	fileName, err := configFileNameFromChainName(chainName)
	if err != nil {
		return "", err
	}

	var hash string
	if verbose {
		hash, err = util.SendToIPFS(fileName, os.Stdout)
	} else {
		hash, err = util.SendToIPFS(fileName, bytes.NewBuffer([]byte{}))
	}

	if err != nil {
		return "", err
	}

	return hash, nil
}
