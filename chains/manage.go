package chains

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/data"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//----------------------------------------------------------------------

// fetch and install a chain
func Install(cmd *cobra.Command, args []string) {
	checkChainGiven(args)
	chainID := args[0]

	chainType := cmd.Flags().Lookup("type").Value.String()
	config := cmd.Flags().Lookup("config").Value.String()
	dir := cmd.Flags().Lookup("dir").Value.String()

	if err := InstallChainRaw(chainType, chainID, config, dir); err != nil {
		fmt.Println(err)
	}

}

// create a new chain
// TODO: interactive option for building genesis?
func New(cmd *cobra.Command, args []string) {
	chainType := cmd.Flags().Lookup("type").Value.String()
	genesis := cmd.Flags().Lookup("genesis").Value.String()
	config := cmd.Flags().Lookup("config").Value.String()
	dir := cmd.Flags().Lookup("dir").Value.String()

	if err := NewChainRaw(chainType, genesis, config, dir); err != nil {
		fmt.Println(err)
	}

}

// import a chain definition file
func Import(cmd *cobra.Command, args []string) {
	IfExit(checkChainGiven(args))
	if len(args) != 3 {
		logger.Println("Please give me: eris chains import [type] [name] [location]")
		return
	}
	IfExit(ImportChainRaw(args[0], args[1], args[2]))
}

// edit a chain definition file
func Edit(cmd *cobra.Command, args []string) {
	IfExit(checkChainGiven(args))
	chainType, chainID := args[0], args[1]
	var configVals []string
	if len(args) > 2 {
		configVals = args[2:]
	}
	IfExit(EditChainRaw(chainType, chainID, configVals))
}

func Inspect(cmd *cobra.Command, args []string) {
	IfExit(checkChainGiven(args))
	chainType, chainID := args[0], args[1]
	var field string
	if len(args) == 2 {
		field = "all"
	} else {
		field = args[2]
	}
	chain, err := LoadChainDefinition(chainType, chainID)
	IfExit(err)
	if IsChainExisting(chain) {
		IfExit(services.InspectServiceByService(chain.Service, chain.Operations, field))
	}
}

func Export(cmd *cobra.Command, args []string) {
	IfExit(checkChainGiven(args))
	IfExit(ExportChainRaw(args[0], args[1]))
}

func Rename(cmd *cobra.Command, args []string) {
	// TODO: on hold until we have refs system
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
	IfExit(checkChainGiven(args))
	if len(args) != 2 {
		fmt.Println("Please give me: eris chains rename [oldName] [newName]")
		return
	}
	IfExit(RenameChainRaw(args[0], args[1]))
}

func Update(cmd *cobra.Command, args []string) {
	IfExit(checkChainGiven(args))
	IfExit(UpdateChainRaw(args[0], args[1]))
}

func Rm(cmd *cobra.Command, args []string) {
	IfExit(checkChainGiven(args))
	IfExit(RmChainRaw(args[0], args[1], cmd.Flags().Lookup("force").Changed))
}

//----------------------------------------------------------------------

func getChainIDFromGenesis(genesis, dir string) (string, error) {
	var hasChainID = struct {
		ChainID string `json:"chain_id"`
	}{}

	if genesis == "" {
		genesis = path.Join(dir, "genesis.json")
		if _, err := os.Stat(genesis); err != nil {
			// (if no genesis, we'll have to
			// copy into a random "scratch" location,
			// start the node so it lays a genesis, read the chainid
			// from that, and then copy to appropriate destination, sigh)
			return "", fmt.Errorf("Please provide a genesis.json explicitly or in a specified directory")
		}
	}

	b, err := ioutil.ReadFile(genesis)
	if err != nil {
		return "", fmt.Errorf("Error reading genesis file: %v", err)
	}

	if err = json.Unmarshal(b, &hasChainID); err != nil {
		return "", fmt.Errorf("Error reading chain id from genesis file: %v", err)
	}

	chainID := hasChainID.ChainID
	if chainID == "" {
		return "", fmt.Errorf("Genesis file must contain chain_id field")
	}
	return chainID, nil
}

func NewChainRaw(chainType, genesis, config, dir string) error {
	if chainType == "" {
		return fmt.Errorf("Please specify a chain type with the --type flag")
	}

	// read chainID from genesis. genesis may be in dir
	chainID, err := getChainIDFromGenesis(genesis, dir)
	if err != nil {
		return err
	}

	// setup data container and service container
	chain, err := LoadChainDefinition(chainType, chainID)
	if err != nil {
		return err
	}

	containerName := chainType + "_" + chainID

	// run containers and exit (creates data container)
	chain.Service.Command = fmt.Sprintf("echo \"Creating new %s containers with chainID %s\"", chainType, chainID)
	if err := perform.DockerCreateDataContainer(containerName); err != nil {
		return fmt.Errorf("Error creating data container %v", err)
	}

	// copy dir, genesis, config into container
	dst := path.Join(DataContainersPath, containerName, "blockchains", chainType, chainID)
	if err := os.MkdirAll(dst, 0700); err != nil {
		return fmt.Errorf("Error making data directory: %v", err)
	}

	if dir != "" {
		if err := Copy(dir, dst); err != nil {
			return err
		}
	}
	if genesis != "" {
		if err := Copy(genesis, path.Join(dst, "genesis.json")); err != nil {
			return err
		}
	}
	if config != "" {
		if err := Copy(config, path.Join(dst, "config."+path.Ext(config))); err != nil {
			return err
		}
	}

	// copy from host to container
	if err := data.ImportDataRaw(containerName); err != nil {
		return err
	}

	// run "new" cmd in chains definition
	// typically this should parse the genesis and write
	// a genesis state to the db. we might also have it
	// post the new chain's id and other info to an etcb, etc. (pun intended)
	chain.Service.Command = chain.Manage.NewCmd
	containerNumber := 1
	chain.Operations.DataContainerName = fmt.Sprintf("eris_data_%s_%d", containerName, containerNumber)
	chain.Operations.Remove = true
	services.StartServiceByService(chain.Service, chain.Operations)
	return nil
}

func InstallChainRaw(chainType, chainID, config, dir string) error {
	// check known chain type, or if empty default to mint

	// setup data container and service container

	// create dir, possibly copy in config and dir if present

	// run "fetch" cmd in chains definition
	return nil
}

func ImportChainRaw(chainType, chainID, path string) error {
	fileName := filepath.Join(BlockchainsPath, chainType, chainID)
	if filepath.Ext(fileName) == "" {
		fileName = fileName + ".toml"
	}

	s := strings.Split(path, ":")
	if s[0] == "ipfs" {

		var err error
		if logger.Level > 0 {
			err = util.GetFromIPFS(s[1], fileName, logger.Writer)
		} else {
			err = util.GetFromIPFS(s[1], fileName, bytes.NewBuffer([]byte{}))
		}

		if err != nil {
			return err
		}
		return nil
	}

	if strings.Contains(s[0], "github") {
		logger.Println("https://twitter.com/ryaneshea/status/595957712040628224")
		return nil
	}

	return fmt.Errorf("I do not know how to get that file. Sorry.")
}

func ExportChainRaw(chainType, chainID string) error {
	chain, err := LoadChainDefinition(chainType, chainID)
	if err != nil {
		return err
	}
	if IsChainExisting(chain) {
		ipfsService, err := services.LoadServiceDefinition("ipfs")
		if err != nil {
			return err
		}

		if services.IsServiceRunning(ipfsService.Service) {
			logger.Infoln("IPFS is running. Adding now.")

			hash, err := exportFile(chainType, chainID)
			if err != nil {
				return err
			}
			logger.Println(hash)
		} else {
			logger.Infoln("IPFS is not running. Starting now.")
			err := services.StartServiceByService(ipfsService.Service, ipfsService.Operations)
			if err != nil {
				return err
			}

			hash, err := exportFile(chainType, chainID)
			if err != nil {
				return err
			}
			logger.Println(hash)
		}

	} else {
		return fmt.Errorf(`I don't known of that chain.
Please retry with a known chain.
To find known chains use: eris chains known`)
	}
	return nil
}

func EditChainRaw(chainType, chainID string, configVals []string) error {
	chainConf, err := readChainDefinition(chainType, chainID)
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

/* // suspended until we have a ref system in place for chains
// right now a chain's name is chainType/chainID and that's that
func RenameChainRaw(oldName, newName string) error {
	if oldName == newName {
		return fmt.Errorf("Cannot rename to same name")
	}

	newNameBase := strings.Replace(newName, filepath.Ext(newName), "", 1)
	transformOnly := newNameBase == oldName

	if isKnownChain(oldName) {
		logger.Infoln("Renaming chain", oldName, "to", newNameBase)

		chainDef, err := LoadChainDefinition(oldName)
		if err != nil {
			return err
		}

		if !transformOnly {
			err = perform.DockerRename(chainDef.Service, chainDef.Operations, oldName, newNameBase)
			if err != nil {
				return err
			}
		}

		oldFile, err := configFileNameFromChainName(oldName)
		if err != nil {
			return err
		}

		if filepath.Base(oldFile) == newName {
			logger.Infoln("Those are the same file. Not renaming")
			return nil
		}

		var newFile string
		if filepath.Ext(newName) == "" {
			newFile = strings.Replace(oldFile, oldName, newName, 1)
		} else {
			newFile = filepath.Join(BlockchainsPath, newName)
		}

		chainDef.Name = newNameBase
		chainDef.Service.Name = ""
		chainDef.Service.Image = ""
		err = WriteChainDefinitionFile(chainDef, newFile)
		if err != nil {
			return err
		}

		if !transformOnly {
			err = data.RenameDataRaw(oldName, newNameBase)
			if err != nil {
				return err
			}
		}

		os.Remove(oldFile)
	} else {
		return fmt.Errorf("I cannot find that chain. Please check the chain name you sent me.")
	}
	return nil
}
*/

func UpdateChainRaw(chainType, chainID string) error {
	chain, err := LoadChainDefinition(chainType, chainID)
	if err != nil {
		return err
	}
	err = perform.DockerRebuild(chain.Service, chain.Operations, false)
	if err != nil {
		return err
	}
	return nil
}

func RmChainRaw(chainType, chainID string, force bool) error {
	chain, err := LoadChainDefinition(chainType, chainID)
	if err != nil {
		return err
	}
	err = perform.DockerRemove(chain.Service, chain.Operations)
	if err != nil {
		return err
	}

	if force {
		oldFile, err := configFileNameFromChainName(chainType, chainID)
		if err != nil {
			return err
		}
		os.Remove(oldFile)
	}
	return nil
}

func exportFile(chainType, chainID string) (string, error) {
	fileName, err := configFileNameFromChainName(chainType, chainID)
	if err != nil {
		return "", err
	}

	var hash string
	if logger.Level > 0 {
		hash, err = util.SendToIPFS(fileName, logger.Writer)
	} else {
		hash, err = util.SendToIPFS(fileName, bytes.NewBuffer([]byte{}))
	}

	if err != nil {
		return "", err
	}

	return hash, nil
}
