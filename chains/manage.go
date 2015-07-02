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
)

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

func setupChain(chainType, chainID, chainName, cmd, dir, genesis, config string, containerNumber int) (err error) {
	containerName := chainType + "_" + chainID
	if chainName != "" {
		containerName = chainName
	}

	// run containers and exit (creates data container)
	logger.Infof("Creating data container for %s\n", containerName)
	if err := perform.DockerCreateDataContainer(containerName, containerNumber); err != nil {
		return fmt.Errorf("Error creating data container %v", err)
	}

	// if something goes wrong, cleanup
	defer func() {
		if err != nil {
			if err2 := services.RmServiceRaw([]string{containerName}, containerNumber, false, true); err2 != nil {
				err = fmt.Errorf("Tragic! We encountered an error during setupChain (%v), and failed to cleanup after ourselves (remove containers) due to another error: %v", err, err2)
			}
		}
	}()

	// copy dir, genesis, config into container
	dst := path.Join(DataContainersPath, containerName, "blockchains", containerName)
	if err = os.MkdirAll(dst, 0700); err != nil {
		return fmt.Errorf("Error making data directory: %v", err)
	}

	if dir != "" {
		if err = Copy(dir, dst); err != nil {
			return err
		}
	}
	if genesis != "" {
		if err = Copy(genesis, path.Join(dst, "genesis.json")); err != nil {
			return err
		}
	}
	if config != "" {
		if err = Copy(config, path.Join(dst, "config."+path.Ext(config))); err != nil {
			return err
		}
	}

	// copy from host to container
	if err = data.ImportDataRaw(containerName, containerNumber); err != nil {
		return err
	}

	chain := &def.Chain{
		Name:    containerName,
		Type:    chainType,
		ChainID: chainID,
		Service: &def.Service{},
		Manager: make(map[string]string),
	}

	// write the chain definition file ...
	fileName := filepath.Join(BlockchainsPath, containerName) + ".toml"
	d := path.Dir(fileName)
	if _, err = os.Stat(d); err != nil {
		if err = os.MkdirAll(d, 0700); err != nil {
			err = fmt.Errorf("Error making directory (%s): %v", d, err)
			return
		}
	}

	if err = WriteChainDefinitionFile(chain, fileName); err != nil {
		err = fmt.Errorf("error writing chain definition to file: %v", err)
		return
	}

	chain, err = LoadChainDefinition(containerName, containerNumber)
	if err != nil {
		return err
	}

	// run "new" cmd in chains definition
	// typically this should parse the genesis and write
	// a genesis state to the db. we might also have it
	// post the new chain's id and other info to an etcb, etc. (pun intended)
	var ok bool
	chain.Service.Command, ok = chain.Manager[cmd]
	if !ok {
		return fmt.Errorf("%s service definition must include '%s' command under Manager", chainType, cmd)
	}
	chain.Operations.DataContainerName = fmt.Sprintf("eris_data_%s_%d", containerName, containerNumber)
	chain.Operations.Remove = true
	err = services.StartServiceByService(chain.Service, chain.Operations)
	return
}

func NewChainRaw(chainType, name, genesis, config, dir string, containerNumber int) error {
	if chainType == "" {
		return fmt.Errorf("Please specify a chain type with the --type flag")
	}

	// read chainID from genesis. genesis may be in dir
	chainID, err := getChainIDFromGenesis(genesis, dir)
	if err != nil {
		return err
	}

	return setupChain(chainType, chainID, name, "new", dir, genesis, config, containerNumber)
}

func InstallChainRaw(chainType, chainID, chainName, config, dir string, containerNumber int) error {
	if chainType == "" {
		return fmt.Errorf("Please specify a chain type with the --type flag")
	}

	return setupChain(chainType, chainID, chainName, "fetch", dir, "", config, containerNumber)
}

func ImportChainRaw(chainName, path string) error {
	fileName := filepath.Join(BlockchainsPath, chainName)
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

// export a chain definition file
func ExportChainRaw(chainName string) error {
	chain, err := LoadChainDefinition(chainName, 1) //TODO:CNUM
	if err != nil {
		return err
	}
	if IsChainExisting(chain) {
		ipfsService, err := services.LoadServiceDefinition("ipfs", 1)
		if err != nil {
			return err
		}

		if services.IsServiceRunning(ipfsService.Service, ipfsService.Operations) {
			logger.Infoln("IPFS is running. Adding now.")

			hash, err := exportFile(chainName)
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

			hash, err := exportFile(chainName)
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

// XXX: What's going on here?
func RenameChainRaw(oldName, newName string) error {
	if oldName == newName {
		return fmt.Errorf("Cannot rename to same name")
	}

	newNameBase := strings.Replace(newName, filepath.Ext(newName), "", 1)
	transformOnly := newNameBase == oldName

	if isKnownChain(oldName) {
		logger.Infoln("Renaming chain", oldName, "to", newNameBase)

		chainDef, err := LoadChainDefinition(oldName, 1) // TODO:CNUM
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
			err = data.RenameDataRaw(oldName, newNameBase, 1) // TODO:CNUM
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

func UpdateChainRaw(chainName string, pull bool, containerNumber int) error {
	// DockerRebuild is built for services, adding false to the final
	//   variable will mean it pulls. But we want the opposite default
	//   behaviour for chains as we do for services in this regard
	//   so we flip the variable.
	skipPull := !pull

	chain, err := LoadChainDefinition(chainName, containerNumber)
	if err != nil {
		return err
	}
	err = perform.DockerRebuild(chain.Service, chain.Operations, skipPull)
	if err != nil {
		return err
	}
	return nil
}

func RmChainRaw(chainName string, rmData bool, force bool, containerNumber int) error {
	chain, err := LoadChainDefinition(chainName, containerNumber)
	if err != nil {
		return err
	}
	err = perform.DockerRemove(chain.Service, chain.Operations)
	if err != nil {
		return err
	}

	if rmData {
		mockServ, mockOp := data.MockService(chainName, containerNumber)
		err = perform.DockerRemove(mockServ, mockOp)
		if err != nil {
			return err
		}
	}

	if force {
		oldFile, err := configFileNameFromChainName(chainName)
		if err != nil {
			return err
		}
		os.Remove(oldFile)
	}
	return nil
}

func exportFile(chainName string) (string, error) {
	fileName, err := configFileNameFromChainName(chainName)
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
