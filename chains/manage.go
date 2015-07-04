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

const (
	ErisChainType    = "erisdb"
	ErisChainStart   = "run"
	ErisChainInstall = "install"
	ErisChainNew     = "new"
)

//------------------------------------------------------------------------

// genesis file either given directly, in dir, or not found (empty)
func resolveGenesisFile(genesis, dir string) string {
	if genesis == "" {
		genesis = path.Join(dir, "genesis.json")
		if _, err := os.Stat(genesis); err != nil {
			return ""
		}
	}
	return genesis
}

// "chain_id" should be in the genesis.json
// or else is set to name
func getChainIDFromGenesis(genesis, name string) (string, error) {
	var hasChainID = struct {
		ChainID string `json:"chain_id"`
	}{}

	b, err := ioutil.ReadFile(genesis)
	if err != nil {
		return "", fmt.Errorf("Error reading genesis file: %v", err)
	}

	if err = json.Unmarshal(b, &hasChainID); err != nil {
		return "", fmt.Errorf("Error reading chain id from genesis file: %v", err)
	}

	chainID := hasChainID.ChainID
	if chainID == "" {
		chainID = name
	}
	return chainID, nil
}

// the main function for setting up a chain container
// handles both "new" and "fetch" - most of the differentiating logic is in the container
func setupChain(chainID, chainName, cmd, dir, genesis, config string, containerNumber int) (err error) {
	// chainName is mandatory
	if chainName == "" {
		return fmt.Errorf("setupChain requires a chainName")
	}
	containerName := util.NameAndNumber(chainName, containerNumber)
	if chainID == "" {
		chainID = chainName
	}

	// TODO: check if data container already exists
	// run containers and exit (creates data container)
	logger.Infof("Creating data container for %s\n", chainName)
	if err := perform.DockerCreateDataContainer(chainName, containerNumber); err != nil {
		return fmt.Errorf("Error creating data container %v", err)
	}

	// if something goes wrong, cleanup
	defer func() {
		if err != nil {
			logger.Infof("\nError on setupChain: %v\n", err)
			logger.Infoln("Cleaning up...")
			if err2 := RmChainRaw(chainName, true, false, containerNumber); err2 != nil {
				err = fmt.Errorf("Tragic! We encountered an error during setupChain for %s, and failed to cleanup after ourselves (remove containers) due to another error.\n\nFirst error:  %v\nCleanup error: %v", containerName, err, err2)
			}
		}
	}()

	// copy dir, genesis, config into container
	containerDst := path.Join("blockchains", chainName)           // path in container
	dst := path.Join(DataContainersPath, chainName, containerDst) // path on host
	// TODO: deal with containerNumbers ....!
	// we probably need to update Import

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
	} else {
		// TODO: run mintgen and open the genesis in editor
	}

	if config != "" {
		if err = Copy(config, path.Join(dst, "config."+path.Ext(config))); err != nil {
			return err
		}
	}

	// copy from host to container
	if err = data.ImportDataRaw(chainName, containerNumber); err != nil {
		return err
	}

	chain := &def.Chain{
		Name:    chainName,
		ChainID: chainID,
		Service: &def.Service{},
	}

	chain.Service.AutoData = true

	// write the chain definition file ...
	fileName := filepath.Join(BlockchainsPath, chainName) + ".toml"
	if err = WriteChainDefinitionFile(chain, fileName); err != nil {
		err = fmt.Errorf("error writing chain definition to file: %v", err)
		return
	}

	chain, err = LoadChainDefinition(chainName, containerNumber)
	if err != nil {
		return err
	}

	// cmd should "new" or "install"
	chain.Service.Command = cmd

	// set chainid and other vars
	chain.Service.Environment = append(chain.Service.Environment, "CHAIN_ID="+chainID)
	chain.Service.Environment = append(chain.Service.Environment, "CONTAINER_NAME="+containerName)
	chain.Service.Environment = append(chain.Service.Environment, "RUN=false") // TODO new should take a run flag
	// TODO mint vs. erisdb (in terms of rpc)
	// chain.Service.Environment = append(chain.Service.Environment, "CHAIN_TYPE=mint")
	if genesis == "" {
		chain.Service.Environment = append(chain.Service.Environment, "GENERATE_GENESIS=true")
	}

	chain.Operations.DataContainerName = fmt.Sprintf("eris_data_%s", util.NameAndNumber(chainName, containerNumber))
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		chain.Operations.Remove = true
	}
	err = services.StartServiceByService(chain.Service, chain.Operations)

	return
}

func NewChainRaw(name, genesis, config, dir string, containerNumber int) (err error) {
	// read chainID from genesis. genesis may be in dir
	// if no genesis or no genesis.chain_id, chainID = name
	var chainID string
	if genesis = resolveGenesisFile(genesis, dir); genesis == "" {
		chainID = name
	} else {
		chainID, err = getChainIDFromGenesis(genesis, name)
		if err != nil {
			return err
		}
	}

	return setupChain(chainID, name, ErisChainNew, dir, genesis, config, containerNumber)
}

func InstallChainRaw(chainID, chainName, config, dir string, containerNumber int) error {
	return setupChain(chainID, chainName, ErisChainInstall, dir, "", config, containerNumber)
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

func RmChainRaw(chainName string, rmData bool, file bool, containerNumber int) error {
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

	if file {
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
