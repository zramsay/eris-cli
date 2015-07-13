package chains

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
)

func NewChain(do *definitions.Do) error {
	// read chainID from genesis. genesis may be in dir
	// if no genesis or no genesis.chain_id, chainID = name
	var err error
	if do.GenesisFile = resolveGenesisFile(do.GenesisFile, do.Path); do.GenesisFile == "" {
		do.ChainID = do.Name
	} else {
		do.ChainID, err = getChainIDFromGenesis(do.GenesisFile, do.Name)
		if err != nil {
			return err
		}
	}

	return setupChain(do, loaders.ErisChainNew)
}

func InstallChain(do *definitions.Do) error {
	return setupChain(do, loaders.ErisChainInstall)
}

func StartChain(do *definitions.Do) error {
	logger.Infoln("Ensuring Key Server is Started.")
	//should it take a flag? keys server may be running another cNum
	keysService, err := loaders.LoadServiceDefinition("keys", 1)
	if err != nil {
		return err
	}

	err = perform.DockerRun(keysService.Service, keysService.Operations)
	if err != nil {
		return err
	}

	chain, err := loaders.LoadChainDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		logger.Infoln("Cannot start a chain I cannot find.")
		do.Result = "no file"
		return nil
	}

	if chain.Name == "" {
		logger.Infoln("Cannot start a chain without a name.")
		do.Result = "no name"
		return nil
	}

	chain.Service.Command = loaders.ErisChainStart
	util.OverWriteOperations(chain.Operations, do.Operations)
	chain.Service.Environment = append(chain.Service.Environment, "CHAIN_ID="+chain.ChainID)

	logger.Infof("StartChainRaw to DockerRun =>\t%s\n", chain.Service.Name)
	logger.Debugf("\twith ChainID =>\t\t%v\n", chain.ChainID)
	logger.Debugf("\twith Environment =>\t%v\n", chain.Service.Environment)
	logger.Debugf("\twith AllPortsPublshd =>\t%v\n", chain.Operations.PublishAllPorts)
	if err := perform.DockerRun(chain.Service, chain.Operations); err != nil {
		do.Result = "error"
		return err
	}

	return nil
}

func KillChain(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}

	if do.Force {
		if do.Timeout == 10 { // default set by flags
			do.Timeout = 0
		}
	}

	if IsChainRunning(chain) {
		if err := perform.DockerStop(chain.Service, chain.Operations, do.Timeout); err != nil {
			return err
		}
	} else {
		logger.Infoln("Chain not currently running. Skipping.")
	}

	if do.Rm {
		if err := perform.DockerRemove(chain.Service, chain.Operations, do.RmD); err != nil {
			return err
		}
	}

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
			if err2 := RmChain(do); err2 != nil {
				err = fmt.Errorf("Tragic! Our marmots encountered an error during setupChain for %s.\nThey also failed to cleanup after themselves (remove containers) due to another error.\nFirst error =>\t\t%v\nCleanup error =>\t%v\n", containerName, err, err2)
			}
		}
	}()

	// copy do.Path, do.GenesisFile, config into container
	containerDst := path.Join("blockchains", do.Name)           // path in container
	dst := path.Join(DataContainersPath, do.Name, containerDst) // path on host
	// TODO: deal with do.Operations.ContainerNumbers ....!
	// we probably need to update Import

	if err = os.MkdirAll(dst, 0700); err != nil {
		return fmt.Errorf("Error making data directory: %v", err)
	}

	if do.Path != "" {
		if err = Copy(do.Path, dst); err != nil {
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
	if err = data.ImportData(do); err != nil {
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
	chain.Operations.PublishAllPorts = do.Operations.PublishAllPorts // TODO: remove this and marshall into struct from cli directly

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
