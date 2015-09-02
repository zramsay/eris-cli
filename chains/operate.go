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

	"github.com/eris-ltd/eris-cli/config"
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
	chain.Service.Environment = append(chain.Service.Environment, do.Env...)
	chain.Service.Links = append(chain.Service.Links, do.Links...)

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

func ExecChain(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name, false, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}

	if IsChainExisting(chain) {
		logger.Infoln("Chain exists.")
		return perform.DockerExec(chain.Service, chain.Operations, do.Args, do.Interactive)
	} else {
		return fmt.Errorf("Chain does not exist. Please start the chain container with eris chains start %s.\n", do.Name)
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

	// fakeId := strings.Split(uuid.New(), "-")[0]
	// srv, err := loaders.MockChainDefinition(do.Name, fakeId, true, 1)
	// if err != nil {
	// 	return err
	// }

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
				// maybe be less dramatic
				err = fmt.Errorf("Tragic! Our marmots encountered an error during setupChain for %s.\nThey also failed to cleanup after themselves (remove containers) due to another error.\nFirst error =>\t\t\t%v\nCleanup error =>\t\t%v\n", containerName, err, err2)
			}
		}
	}()

	// copy do.Path, do.GenesisFile, do.ConfigFile, do.Priv, do.CSV into container
	containerDst := path.Join("blockchains", do.Name)           // path in container
	dst := path.Join(DataContainersPath, do.Name, containerDst) // path on host
	// TODO: deal with do.Operations.ContainerNumbers ....!
	// we probably need to update Import

	logger.Debugln("container destination:", containerDst)
	logger.Debugln("local destination:", dst)

	if err = os.MkdirAll(dst, 0700); err != nil {
		return fmt.Errorf("Error making data directory: %v", err)
	}

	// we accept two csvs: one for validators, one for accounts
	// if there's only one, its for validators and accounts
	var csvFiles []string
	var csvPaths string
	if do.CSV != "" {
		csvFiles = strings.Split(do.CSV, ",")
		if len(csvFiles) > 1 {
			csvPath1 := fmt.Sprintf("/home/eris/.eris/blockchains/%s/%s", do.ChainID, "validators.csv")
			csvPath2 := fmt.Sprintf("/home/eris/.eris/blockchains/%s/%s", do.ChainID, "accounts.csv")
			csvPaths = fmt.Sprintf("%s,%s", csvPath1, csvPath2)
		} else {
			csvPaths = fmt.Sprintf("/home/eris/.eris/blockchains/%s/%s", do.ChainID, "genesis.csv")
		}
	}

	filesToCopy := []stringPair{
		{do.Path, ""},
		{do.GenesisFile, "genesis.json"},
		{do.ConfigFile, "config.toml"},
		{do.Priv, "priv_validator.json"},
	}

	if len(csvFiles) == 1 {
		filesToCopy = append(filesToCopy, stringPair{csvFiles[0], "genesis.csv"})
	} else if len(csvFiles) > 1 {
		filesToCopy = append(filesToCopy, stringPair{csvFiles[0], "validators.csv"})
		filesToCopy = append(filesToCopy, stringPair{csvFiles[1], "accounts.csv"})
	}

	if err := copyFiles(dst, filesToCopy); err != nil {
		return err
	}

	// copy from host to container
	logger.Debugf("Copying Files into DataCont =>\t%s:%s\n", dst, containerDst)
	importDo := definitions.NowDo()
	importDo.Name = do.Name
	importDo.Operations = do.Operations
	if err = data.ImportData(importDo); err != nil {
		return err
	}

	chain := loaders.MockChainDefinition(do.Name, do.ChainID, false, do.Operations.ContainerNumber)

	//set maintainer info
	chain.Maintainer.Name, chain.Maintainer.Email, err = config.GitConfigUser()
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
	chain.Operations.PublishAllPorts = do.Operations.PublishAllPorts // TODO: remove this and marshall into struct from cli directly

	// cmd should be "new" or "install"
	chain.Service.Command = cmd

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
		fmt.Sprintf("CSV=%v", csvPaths),               // for mintgen
		fmt.Sprintf("CONFIG_OPTS=%s", configOpts),     // config.toml
		fmt.Sprintf("REGISTER_PUBKEY=%s", do.Address), // use to sign registration txs for etcb
	}
	envVars = append(envVars, do.Env...)

	logger.Debugf("Set env vars from setupChain =>\t%v\n", envVars)
	chain.Service.Environment = append(chain.Service.Environment, envVars...)
	logger.Debugf("Set links from setupChain =>\t%v\n", do.Links)
	chain.Service.Links = append(chain.Service.Links, do.Links...)

	// TODO: if do.N > 1 ...

	chain.Operations.DataContainerName = util.DataContainersName(do.Name, do.Operations.ContainerNumber)

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

type stringPair struct {
	key   string
	value string
}

func copyFiles(dst string, files []stringPair) error {
	for _, f := range files {
		if f.key != "" {
			logger.Debugf("\tcopying files from %s to %s\n", f.key, path.Join(dst, f.value))
			if err := Copy(f.key, path.Join(dst, f.value)); err != nil {
				return err
			}
		}
	}
	return nil
}
