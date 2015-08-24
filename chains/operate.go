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
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
)

func NewChain(do *definitions.Do) error {
	// read chainID from genesis. genesis may be in dir
	// if no genesis or no genesis.chain_id, chainID = name
	/*var err error
	if do.GenesisFile = resolveGenesisFile(do.GenesisFile, do.Path); do.GenesisFile == "" {
		do.ChainID = do.Name
	} else {
		do.ChainID, err = getChainIDFromGenesis(do.GenesisFile, do.Name)
		if err != nil {
			return err
		}
	}*/

	// for now we just let setupChain force do.ChainID = do.Name
	// and we overwrite using jq in the container
	logger.Debugf("Starting Setup for ChnID =>\t%s\n", do.ChainID)
	return setupChain(do, loaders.ErisChainNew)
}

func InstallChain(do *definitions.Do) error {
	return setupChain(do, loaders.ErisChainInstall)
}

func StartChain(do *definitions.Do) error {
	logger.Infoln("Ensuring Key Server is Started.")
	//should it take a flag? keys server may be running another cNum
	// XXX: currently we don't use or need a key server.
	// plus this should be specified in a service def anyways
	keysService, err := loaders.LoadServiceDefinition("keys", false, 1)
	if err != nil {
		return err
	}

	err = perform.DockerRun(keysService.Service, keysService.Operations)
	if err != nil {
		return err
	}

	chain, err := loaders.LoadChainDefinition(do.Name, false, do.Operations.ContainerNumber)
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
	if do.Run {
		chain.Service.Command = loaders.ErisChainStartApi
	}
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
	chain, err := loaders.LoadChainDefinition(do.Name, false, do.Operations.ContainerNumber)
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

// Throw away chains are used for eris contracts
func ThrowAwayChain(do *definitions.Do) error {
	do.Name = do.Name + "_" + strings.Split(uuid.New(), "-")[0]
	do.Path = filepath.Join(ChainsConfigPath, "default")
	logger.Debugf("Making a ThrowAwayChain =>\t%s:%s\n", do.Name, do.Path)

	if err := NewChain(do); err != nil {
		return err
	}

	logger.Debugf("ThrowAwayChain created =>\t%s\n", do.Name)

	logger.Debugf("Starting a ThrowAwayChain =>\t%s\n", do.Name)
	do.Operations.Remove = true
	StartChain(do)

	logger.Debugf("ThrowAwayChain started =>\t%s\n", do.Name)
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

	// ensure/create data container
	if !data.IsKnown(containerName) {
		if err := perform.DockerCreateDataContainer(do.Name, do.Operations.ContainerNumber); err != nil {
			return fmt.Errorf("Error creating data containr =>\t%v", err)
		}
	} else {
		logger.Debugln("Data container already exists for", do.Name)
	}

	logger.Debugf("Chain's Data Contain Built =>\t%s\n", do.Name)

	// if something goes wrong, cleanup
	defer func() {
		if err != nil {
			logger.Infof("Error on setupChain =>\t\t%v\n", err)
			logger.Infoln("Cleaning up...")
			if err2 := RmChain(do); err2 != nil {
				err = fmt.Errorf("Our marmots encountered an error during setupChain for %s.\nThey also failed to cleanup after themselves (remove containers) due to another error.\nFirst error =>\t\t\t%v\nCleanup error =>\t\t%v\n", containerName, err, err2)
			}
		}
	}()

	// copy do.Path, do.GenesisFile, do.ConfigFile, do.Priv, do.CSV into container
	containerDst := path.Join("blockchains", do.Name)           // path in container
	dst := path.Join(DataContainersPath, do.Name, containerDst) // path on host
	// TODO: deal with do.Operations.ContainerNumbers ....!
	// we probably need to update Import

	logger.Debugf("Container destination =>\t%s\n", containerDst)
	logger.Debugf("Local destination =>\t\t%s\n", dst)

	if err = os.MkdirAll(dst, 0700); err != nil {
		return fmt.Errorf("Error making data directory: %v", err)
	}

	// var csvFile string
	// if do.CSV != "" {
	// 	csvFile = "genesis.csv"
	// }

	// if err := copyFiles(dst, []stringPair{
	// 	{do.Path, ""},
	// 	{do.GenesisFile, "genesis.json"},
	// 	{do.ConfigFile, "config.toml"},
	// 	{do.Priv, "priv_validator.json"},
	// 	{do.CSV, csvFile},
	// }); err != nil {
	// 	return err
	// }

	// // copy from host to container
	// logger.Debugf("Copying Files into DataCont =>\t%s:%s\n", dst, containerDst)
	// importDo := definitions.NowDo()
	// importDo.Name = do.Name
	// importDo.Operations = do.Operations
	// if err = data.ImportData(importDo); err != nil {
	// 	return err
	// }

	chain := loaders.MockChainDefinition(do.Name, do.ChainID, false, do.Operations.ContainerNumber)

	//set maintainer info
	chain.Maintainer.Name, chain.Maintainer.Email, err = util.GitConfigUser()
	if err != nil {
		logger.Debugf(err.Error())
	}

	// write the chain definition file ...
	fileName := filepath.Join(BlockchainsPath, do.Name) + ".toml"
	if _, err = os.Stat(fileName); err != nil {
		if err = WriteChainDefinitionFile(chain, fileName); err != nil {
			return fmt.Errorf("error writing chain definition to file: %v", err)
		}
	}

	chain, err = loaders.LoadChainDefinition(do.Name, false, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}
	logger.Debugf("Chain Loaded. Image =>\t\t%v\n", chain.Service.Image)
	logger.Debugf("\tBooting =>\t\t%v:%v\n", chain.Service.EntryPoint, chain.Service.Command)
	chain.Operations.PublishAllPorts = do.Operations.PublishAllPorts // TODO: remove this and marshall into struct from cli directly

	// cmd should be "new" or "install"
	chain.Service.Command = cmd

	// do we need to create our own do.GenesisFile?
	var genGen bool
	if do.GenesisFile == "" {
		genGen = true
	}

	// write the list of <key>:<value> config options as flags
	buf := new(bytes.Buffer)
	for _, cv := range do.ConfigOpts {
		spl := strings.Split(cv, "=")
		if len(spl) != 2 {
			return fmt.Errorf("Config options should be <key>=<value> pairs. Got %s", cv)
		}
		buf.WriteString(fmt.Sprintf(" --%s=%s", spl[0], spl[1]))
	}
	configOpts := buf.String()

	// set chainid and other vars
	envVars := []string{
		fmt.Sprintf("CHAIN_ID=%s", do.ChainID),
		fmt.Sprintf("CONTAINER_NAME=%s", containerName),
		fmt.Sprintf("RUN=%v", do.Run),
		fmt.Sprintf("GENERATE_GENESIS=%v", genGen),
		// fmt.Sprintf("CSV=/home/eris/.eris/blockchains/%s/%s", do.ChainID, csvFile),
		fmt.Sprintf("CONFIG_OPTS=%s", configOpts),
	}

	if do.Path != "" {
		if err := resolveFilesFromPath(do); err != nil {
			return err
		}
	}

	if do.GenesisFile == "" && do.ConfigFile == "" { // XXX these are the minimums we need, no?
		do.Path = path.Join(BlockchainsPath, "config", "default")
		if err := resolveFilesFromPath(do); err != nil {
			return err
		}
	}

	if err := envVarsFromFiles(chain, []stringPair{
		{"GENESIS", do.GenesisFile},
		{"CHAIN_CONFIG", do.ConfigFile},
		{"KEY", do.Priv},
		{"GENESIS_CSV", do.CSV},
		{"SERVER_CONFIG", do.ServerConf},
	}); err != nil {
		return err
	}

	logger.Debugf("Set env vars from setupChain =>\t%v\n", envVars)
	for _, eV := range envVars {
		chain.Service.Environment = append(chain.Service.Environment, eV)
	}

	// TODO: if do.N > 1 ...

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

func resolveFilesFromPath(do *definitions.Do) error {
	do.GenesisFile = path.Join(do.Path, "genesis.json")
	do.ConfigFile = path.Join(do.Path, "config.toml")
	do.Priv = path.Join(do.Path, "priv_validator.json")
	do.ServerConf = path.Join(do.Path, "server_conf.toml")
	do.CSV = path.Join(do.Path, "genesis.csv")
	return nil // returning an error in case we want to do more error checking here
}

type stringPair struct {
	key   string
	value string
}

func envVarsFromFiles(chain *definitions.Chain, pairs []stringPair) error {
	var contents []byte
	for _, e := range pairs {
		if _, err := os.Stat(e.value); os.IsNotExist(err) {
			contents = []byte{}
		} else {
			contents, err = ioutil.ReadFile(e.value)
			if err != nil {
				return fmt.Errorf("Could not read file (%s).\n%s", e.value, err)
			}
		}
		eV := fmt.Sprintf("%s=%s", e.key, contents)
		chain.Service.Environment = append(chain.Service.Environment, eV)
	}
	return nil
}
