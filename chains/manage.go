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
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
)

const (
	ErisChainType    = "erisdb"
	ErisChainStart   = "erisdb-wrapper run"
	ErisChainInstall = "erisdb-wrapper install"
	ErisChainNew     = "erisdb-wrapper new"
)

func NewChainRaw(do *definitions.Do) error {
	// read chainID from genesis. genesis may be in dir
	// if no genesis or no genesis.chain_id, chainID = name
	var err error
	if do.GenesisFile = resolveGenesisFile(do.GenesisFile, do.DirToCopy); do.GenesisFile == "" {
		do.ChainID = do.Name
	} else {
		do.ChainID, err = getChainIDFromGenesis(do.GenesisFile, do.Name)
		if err != nil {
			return err
		}
	}

	return setupChain(do, ErisChainNew)
}

func InstallChainRaw(do *definitions.Do) error {
	return setupChain(do, ErisChainInstall)
}

func ImportChainRaw(do *definitions.Do) error {
	fileName := filepath.Join(BlockchainsPath, do.Name)
	if filepath.Ext(fileName) == "" {
		fileName = fileName + ".toml"
	}

	s := strings.Split(do.Path, ":")
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
func ExportChainRaw(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name, 1) //TODO:CNUM
	if err != nil {
		return err
	}
	if IsChainExisting(chain) {
		ipfsService, err := loaders.LoadServiceDefinition("ipfs", 1)
		if err != nil {
			return err
		}

		logger.Infoln("IPFS is not running. Starting now.")
		err = perform.DockerRun(ipfsService.Service, ipfsService.Operations) // docker run fails quickly if the service is already running so this is safe to do now
		if err != nil {
			return err
		}

		hash, err := exportFile(do.Name)
		if err != nil {
			return err
		}
		logger.Println(hash)

	} else {
		return fmt.Errorf(`I don't known of that chain.
Please retry with a known chain.
To find known chains use: eris chains known`)
	}
	return nil
}

func EditChainRaw(do *definitions.Do) error {
	chainConf, err := util.LoadViperConfig(path.Join(BlockchainsPath), do.Name, "chain")
	if err != nil {
		return err
	}
	if err := util.EditRaw(chainConf, do.Args); err != nil {
		return err
	}
	var chain definitions.Chain
	loaders.MarshalChainDefinition(chainConf, &chain)
	return WriteChainDefinitionFile(&chain, chainConf.ConfigFileUsed())
}

func ListKnownRaw(do *definitions.Do) error {
	chns := util.GetGlobalLevelConfigFilesByType("chains", false)
	do.Result = strings.Join(chns, "\n")
	return nil
}

func ListRunningRaw(do *definitions.Do) error {
	if do.Quiet {
		do.Result = strings.Join(util.ChainContainerNames(false), "\n")
	} else {
		perform.PrintTableReport("chain", false)
	}

	return nil
}

func ListExistingRaw(do *definitions.Do) error {
	if do.Quiet {
		do.Result = strings.Join(util.ChainContainerNames(true), "\n")
	} else {
		perform.PrintTableReport("chain", true)
	}

	return nil
}

// XXX: What's going on here? => [csk]: magic
func RenameChainRaw(do *definitions.Do) error {
	if do.Name == do.NewName {
		return fmt.Errorf("Cannot rename to same name")
	}

	newNameBase := strings.Replace(do.NewName, filepath.Ext(do.NewName), "", 1)
	transformOnly := newNameBase == do.Name

	if isKnownChain(do.Name) {
		logger.Infof("Renaming chain =>\t\t%s:%s\n", do.Name, do.NewName)

		logger.Debugf("Loading Chain Def File =>\t%s\n", do.Name)
		chainDef, err := loaders.LoadChainDefinition(do.Name, 1) // TODO:CNUM
		if err != nil {
			return err
		}

		if !transformOnly {
			logger.Debugln("Embarking on DockerRename.")
			err = perform.DockerRename(chainDef.Service, chainDef.Operations, do.Name, newNameBase)
			if err != nil {
				return err
			}
		}

		oldFile := findChainDefinitionFile(do.Name)
		if err != nil {
			return err
		}

		if filepath.Base(oldFile) == do.NewName {
			logger.Infoln("Those are the same file. Not renaming")
			return nil
		}

		logger.Debugln("Renaming Chain Definition File.")
		var newFile string
		if filepath.Ext(do.NewName) == "" {
			newFile = strings.Replace(oldFile, do.Name, do.NewName, 1)
		} else {
			newFile = filepath.Join(BlockchainsPath, do.NewName)
		}

		chainDef.Name = newNameBase
		chainDef.Service.Name = ""
		chainDef.Service.Image = ""
		err = WriteChainDefinitionFile(chainDef, newFile)
		if err != nil {
			return err
		}

		if !transformOnly {
			logger.Infof("Renaming DataC (fm ChainRaw) =>\t%s:%s\n", do.Name, do.NewName)
			do.Operations.ContainerNumber = chainDef.Operations.ContainerNumber
			logger.Debugf("\twith ContainerNumber =>\t%d\n", do.Operations.ContainerNumber)
			err = data.RenameDataRaw(do)
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

func UpdateChainRaw(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}

	// DockerRebuild is built for services, adding false to the final
	//   variable will mean it pulls. But we want the opposite default
	//   behaviour for chains as we do for services in this regard
	//   so we flip the variable.
	err = perform.DockerRebuild(chain.Service, chain.Operations, do.SkipPull)
	if err != nil {
		return err
	}
	return nil
}

func RmChainRaw(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}
	if IsChainExisting(chain) {
		if err = perform.DockerRemove(chain.Service, chain.Operations, do.RmD); err != nil {
			return err
		}
	}

	if do.File {
		oldFile := findChainDefinitionFile(do.Name)
		if err != nil {
			return err
		}
		if err := os.Remove(oldFile); err != nil {
			return err
		}
	}
	return nil
}

func GraduateChainRaw(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name, 1)
	if err != nil {
		return err
	}

	serv := loaders.ServiceDefFromChain(chain, ErisChainStart)
	if err := services.WriteServiceDefinitionFile(serv, path.Join(ServicesPath, chain.ChainID+".toml")); err != nil {
		return err
	}
	return nil
}

func CatChainRaw(do *definitions.Do) error {
	cat, err := ioutil.ReadFile(path.Join(BlockchainsPath, do.Name+".toml"))
	if err != nil {
		return err
	}
	logger.Println(string(cat))
	return nil

}

//------------------------------------------------------------------------

// the main function for setting up a chain container
// handles both "new" and "fetch" - most of the differentiating logic is in the container
func setupChain(do *definitions.Do, cmd string) (err error) {
	// XXX: if do.Name is unique, we can safely assume (and we probably should) that do.Operations.ContainerNumber = 1

	// do.Name is mandatory
	if do.Name == "" {
		return fmt.Errorf("setupChain requires a chainame")
	}
	containerName := util.ChainContainersName(do.Name, do.Operations.ContainerNumber)
	if do.ChainID == "" {
		do.ChainID = do.Name
	}

	// do.Run containers and exit (creates data container)
	if !data.IsKnown(containerName) {
		if err := perform.DockerCreateDataContainer(do.Name, do.Operations.ContainerNumber); err != nil {
			return fmt.Errorf("Error creating data containr =>\t%v", err)
		}
	}

	// if something goes wrong, cleanup
	defer func() {
		if err != nil {
			logger.Infof("Error on setupChain =>\t\t%v\n", err)
			logger.Infoln("Cleaning up...")
			if err2 := RmChainRaw(do); err2 != nil {
				err = fmt.Errorf("Tragic! Our marmots encountered an error during setupChain for %s.\nThey also failed to cleanup after themselves (remove containers) due to another error.\nFirst error =>\t\t%v\nCleanup error =>\t%v\n", containerName, err, err2)
			}
		}
	}()

	// copy do.DirToCopy, do.GenesisFile, config into container
	containerDst := path.Join("blockchains", do.Name)           // path in container
	dst := path.Join(DataContainersPath, do.Name, containerDst) // path on host
	// TODO: deal with do.Operations.ContainerNumbers ....!
	// we probably need to update Import

	if err = os.MkdirAll(dst, 0700); err != nil {
		return fmt.Errorf("Error making data directory: %v", err)
	}

	if do.DirToCopy != "" {
		if err = Copy(do.DirToCopy, dst); err != nil {
			return err
		}
	}
	if do.GenesisFile != "" {
		if err = Copy(do.GenesisFile, path.Join(dst, "genesis.json")); err != nil {
			return err
		}
	} else {
		// TODO: do.Run mintgen and open the do.GenesisFile in editor
	}

	if do.ConfigFile != "" {
		if err = Copy(do.ConfigFile, path.Join(dst, "config."+path.Ext(do.ConfigFile))); err != nil {
			return err
		}
	}

	// copy from host to container
	if err = data.ImportDataRaw(do); err != nil {
		return err
	}

	chain := loaders.MockChainDefinition(do.Name, do.ChainID, do.Operations.ContainerNumber)

	// write the chain definition file ...
	fileName := filepath.Join(BlockchainsPath, do.Name) + ".toml"
	if _, err = os.Stat(fileName); err != nil {
		if err = WriteChainDefinitionFile(chain, fileName); err != nil {
			return fmt.Errorf("error writing chain definition to file: %v", err)
		}
	}

	chain, err = loaders.LoadChainDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}
	logger.Debugf("Chain Loaded. Image =>\t\t%v\n", chain.Service.Image)
	chain.Operations.PublishAllPorts = do.PublishAllPorts // TODO: remove this and marshall into struct from cli directly

	// cmd should be "new" or "install"
	chain.Service.Command = cmd

	// do we need to create our own do.GenesisFile?
	var genGen bool
	if do.GenesisFile == "" {
		genGen = true
	}

	// set chainid and other vars
	envVars := []string{
		"CHAIN_ID=" + do.ChainID,
		"CONTAINER_NAME=" + containerName,
		fmt.Sprintf("RUN=%v", do.Run),
		fmt.Sprintf("GENERATE_GENESIS=%v", genGen),
	}

	logger.Debugf("Set env vars from setupChain =>\t%v\n", envVars)
	for _, eV := range envVars {
		chain.Service.Environment = append(chain.Service.Environment, eV)
	}
	// TODO mint vs. erisdb (in terms of rpc)

	chain.Operations.DataContainerName = util.DataContainersName(do.Name, do.Operations.ContainerNumber)

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		chain.Operations.Remove = true
	}

	logger.Debugf("Starting chain via Docker =>\t%s\n", chain.Service.Name)
	logger.Debugf("\twith Image =>\t\t%s\n", chain.Service.Image)
	err = perform.DockerRun(chain.Service, chain.Operations)
	// this err is caught in the defer above

	return
}

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

func exportFile(chainName string) (string, error) {
	fileName := findChainDefinitionFile(chainName)

	var hash string
	var err error
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
