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

	//"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/common/go/common"

	log "github.com/eris-ltd/eris-logger"
)

func chainsPathSimplifier(chainName, pathGiven string) (string, error) {
	if util.DoesDirExist(pathGiven) { // full path given, check that config.toml exists though
		if !doesConfigExist(pathGiven) {
			return "", fmt.Errorf("config.toml does not exists in %s", pathGiven)
		} else {
			return pathGiven, nil
		}
	} else { // figure out actual path

		chainDirPathSimple := filepath.Join(ChainsPath, pathGiven)               // if simplechain, pathGiven == chainName
		chainDirPathNotSimple := filepath.Join(ChainsPath, chainName, pathGiven) // ignored if simplechain

		if util.DoesDirExist(chainDirPathSimple) && doesConfigExist(chainDirPathSimple) {
			return chainDirPathSimple, nil
		} else if util.DoesDirExist(chainDirPathNotSimple) && doesConfigExist(chainDirPathNotSimple) { // try dir where !simplechain
			return chainDirPathNotSimple, nil
		} else { // can't find any of it
			log.WithField("=>", pathGiven).Info("Directory or config.toml does not exist")
			return "", fmt.Errorf("directory given on [--init-dir] could not be determined")
		}
	}
	return "", nil
}

func doesConfigExist(dirPath string) bool {
	var configExists bool
	pathToConfig := filepath.Join(dirPath, "config.toml")
	if _, err := os.Stat(pathToConfig); os.IsNotExist(err) {
		log.Warn("config does not exist")
		configExists = false
	} else {
		log.Warn("config does exist")
		configExists = true
	}
	return configExists
}

func StartChain(do *definitions.Do) error {
	chainDirExists, chainConfigExists, chainDataExists, chainContainerExists := whatChainStuffExists(do.Name)

	if do.Path != "" { // [eris chains start whatever --init-dir ~/.eris/chains/whatever]
		var err error
		do.Path, err = chainsPathSimplifier(do.Name, do.Path) // func is also called by whatChainStuffExists
		if err != nil {
			return err
		}

		if !chainDirExists { // config not yet written to "CONFIG_PATH"
			return fmt.Errorf("no dir")
		}
		if chainDataExists || chainContainerExists { // these ought not be existing if --init-dir given
			return fmt.Errorf("data container or chain  container exists")
		}

		// clean leftover data in ~/.eris/scratch/data
		if err := cleanChainData(do.Name); err != nil {
			return err
		}

		// todo: fix this hack
		// for now we just let setupChain force do.ChainID = do.Name
		// and we overwrite using jq in the container
		log.Warn("chain does not exist, new-ing it")
		log.WithField("=>", do.Name).Debug("Setting up chain")
		return setupChain(do, loaders.ErisChainNew)
		// TODO get rid of loaders.ErisChainNew [zr] to discuss with [ben]

	} else { // [eris chains start whatever] (without init-dir)
		if !chainDirExists || !chainConfigExists { // Config need to be in dir, but anyway
			return fmt.Errorf("")
		}

		if !chainDataExists {
			return fmt.Errorf("no data container found, start a chain with [--init-dir]")
		}
		if !chainContainerExists {
			// since normally the edb process container would exist
			// we can alert users and have them (y) to start one.
		}
		_, err := startChain(do, false) // [zr] why are we ignoring the buffer?
		return err

	}
	return nil
}

func StopChain(do *definitions.Do) error {
	chain, err := loaders.LoadChainConfigFile(do.Name)
	if err != nil {
		return err
	}

	if do.Force {
		do.Timeout = 0 //overrides 10 sec default
	}

	if util.IsChain(chain.Name, true) {
		if err := perform.DockerStop(chain.Service, chain.Operations, do.Timeout); err != nil {
			return err
		}
	} else {
		log.Info("Chain not currently running. Skipping")
	}

	return nil
}

func ExecChain(do *definitions.Do) (buf *bytes.Buffer, err error) {
	return startChain(do, true)
}

func startChain(do *definitions.Do, exec bool) (buf *bytes.Buffer, err error) {
	chain, err := loaders.LoadChainConfigFile(do.Name)
	// ^ does not load chain.ChainID
	if err != nil {
		log.Error("Cannot start a chain I cannot find")
		do.Result = "no file"
		return nil, nil
	}

	if chain.Name == "" {
		log.Error("Cannot start a chain without a name")
		do.Result = "no name"
		return nil, nil
	}
	log.Warn("CHAINS_ID")
	log.Warn(chain.ChainID)

	chain.ChainID = do.Name
	log.Warn(chain.ChainID)

	// boot the dependencies (eg. keys, logrotate)
	if err := bootDependencies(chain, do); err != nil {
		return nil, err
	}
	log.Warn("CHAIN_ID")
	log.Warn(chain.ChainID)

	chain.Service.Command = loaders.ErisChainStart
	util.Merge(chain.Operations, do.Operations)
	chain.Service.Environment = append(chain.Service.Environment, "CHAIN_ID="+chain.ChainID)
	chain.Service.Environment = append(chain.Service.Environment, do.Env...)
	chain.Service.Links = append(chain.Service.Links, do.Links...)

	log.WithField("=>", chain.Service.Name).Info("Starting a chain")
	log.WithFields(log.Fields{
		"chain container": chain.Operations.SrvContainerName,
		"environment":     chain.Service.Environment,
		"ports published": chain.Operations.PublishAllPorts,
	}).Debug()

	if exec {
		if do.Image != "" {
			chain.Service.Image = do.Image
		}

		chain.Operations.Args = do.Operations.Args
		log.WithFields(log.Fields{
			"args":        chain.Operations.Args,
			"interactive": chain.Operations.Interactive,
		}).Debug()

		// This override is necessary because erisdb uses an entryPoint and
		// the perform package will respect the images entryPoint if it
		// exists.
		chain.Service.EntryPoint = ""
		chain.Service.Command = ""

		// there is literally never a reason not to randomize the ports.
		chain.Operations.PublishAllPorts = true

		// Link the chain to the exec container when doing chains exec so that there is
		// never any problems with sending info over network to the chain container.
		// Unless the variable SkipLink is set to true; in that case, don't link.
		if !do.Operations.SkipLink {
			// Check the chain is running.
			if !util.IsChain(chain.Name, true) {
				return nil, fmt.Errorf("chain %v has failed to start. You may want to check the [eris chains logs %[1]s] command output", chain.Name)
			}

			chain.Service.Links = append(chain.Service.Links, fmt.Sprintf("%s:%s", util.ContainerName("chain", chain.Name), "chain"))
		}

		buf, err = perform.DockerExecService(chain.Service, chain.Operations)
	} else {
		err = perform.DockerRunService(chain.Service, chain.Operations)
	}
	if err != nil {
		do.Result = "error"
		return buf, err
	}

	return buf, nil
}

// boot chain dependencies
// TODO: this currently only supports simple services (with no further dependencies)
func bootDependencies(chain *definitions.Chain, do *definitions.Do) error {
	if do.Logrotate {
		chain.Dependencies.Services = append(chain.Dependencies.Services, "logrotate")
	}
	if chain.Dependencies != nil {
		name := do.Name
		log.WithFields(log.Fields{
			"services": chain.Dependencies.Services,
			"chains":   chain.Dependencies.Chains,
		}).Info("Booting chain dependencies")
		for _, srvName := range chain.Dependencies.Services {
			do.Name = srvName
			srv, err := loaders.LoadServiceDefinition(do.Name)
			if err != nil {
				return err
			}

			// Start corresponding service.
			if !util.IsService(srv.Service.Name, true) {
				log.WithField("=>", do.Name).Info("Dependency not running. Starting now")
				if err = perform.DockerRunService(srv.Service, srv.Operations); err != nil {
					return err
				}
			}

		}
		do.Name = name // undo side effects

		for _, chainName := range chain.Dependencies.Chains {
			chn, err := loaders.LoadChainConfigFile(chainName)
			if err != nil {
				return err
			}
			if !util.IsChain(chn.Name, true) {
				return fmt.Errorf("chain %s depends on chain %s but %s is not running", chain.Name, chainName, chainName)
			}
		}
	}
	return nil
}

// the main function for setting up a chain container
// handles both "new" and "fetch" - most of the differentiating logic is in the container
// should only be ever call on [eris chains start someChain --init-dir ~/.eris/chains/someChain/someChain_full_000]
// or without the last dir for a simplechain
func setupChain(do *definitions.Do, cmd string) (err error) {
	// do.Name is mandatory
	if do.Name == "" {
		return fmt.Errorf("setupChain requires a chainame")
	}

	containerName := util.ChainContainerName(do.Name)
	//if do.ChainID == "" {
	//	do.ChainID = do.Name
	//}
	//log.Warn(fmt.Sprintf("CHAIN_ID: %s", do.ChainID))

	// writes a pointer (similar to checked out chain) for do.Path in the chain main dir
	// this can then be read by loaders.LoadChainConfigFile(), in order to get the
	// path to the config.toml that was written _in each directory_ (except for simplechain:( )
	// this allows cli to keep track of a given config.toml
	// [zr] this may conflict with how we use --machine ... ?
	fileName := filepath.Join(ChainsPath, do.Name, "CONFIG_PATH")
	if _, err = os.Stat(fileName); err != nil {
		if err = WriteChainDefinitionFile(do.Path, fileName); err != nil {
			return fmt.Errorf("error writing chain definition to file: %v", err)
		}
	}

	// ensure/create data container
	if util.IsData(do.Name) {
		log.WithField("=>", do.Name).Debug("Chain data container already exists")
	} else {
		ops := loaders.LoadDataDefinition(do.Name)
		if err := perform.DockerCreateData(ops); err != nil {
			return fmt.Errorf("Error creating data container =>\t%v", err)
		}
		ops.Args = []string{"mkdir", "-p", path.Join(ErisContainerRoot, "chains", do.Name)} // used to be do.ChainID
		if _, err := perform.DockerExecData(ops, nil); err != nil {
			return err
		}
	}
	log.WithField("=>", do.Name).Debug("Chain data container built")

	// set the config file (is this necessary?)
	do.ConfigFile = filepath.Join(do.Path, "config.toml")

	// Copy do.Path, do.GenesisFile, do.ConfigFile, do.Priv into container.
	// containerDst should not be do.ChainID but rather do.Name
	containerDst := path.Join(ErisContainerRoot, "chains", do.Name) // path in container [zr] used to be do.ChainID
	dst := filepath.Join(DataContainersPath, do.Name, containerDst) // path on host

	log.WithFields(log.Fields{
		"container path": containerDst,
		"local path":     dst,
	}).Debug()

	if err = os.MkdirAll(dst, 0700); err != nil {
		return fmt.Errorf("Error making data directory: %v", err)
	}

	//do.Config will need tuning
	filesToCopy := []stringPair{
		{do.Path, ""},
		{do.GenesisFile, "genesis.json"},
		{do.ConfigFile, "config.toml"},
		{do.Priv, "priv_validator.json"},
	}

	log.Info("Copying chain files into the correct location")
	if err := copyFiles(dst, filesToCopy); err != nil {
		return err
	}

	// copy from host to container
	log.WithFields(log.Fields{
		"from": dst,
		"to":   containerDst,
	}).Debug("Copying files into data container")
	importDo := definitions.NowDo()
	importDo.Name = do.Name
	importDo.Operations = do.Operations
	importDo.Destination = containerDst
	importDo.Source = dst
	if err = data.ImportData(importDo); err != nil {
		return err
	}

	// TODO get rid of this?
	//chain := loaders.MockChainDefinition(do.Name, do.ChainID)
	//log.Warn("chain id")
	//log.Warn(chain.ChainID)
	//log.Warn(do.ChainID)

	// need to have chainID written in here... ?
	chain, err := loaders.LoadChainConfigFile(do.Name)
	if err != nil {
		return err
	}

	chain.ChainID = do.Name // [zr] ugh

	log.Warn("chain ID")
	log.Warn(chain.ChainID)
	//log.Warn(do.ChainID)

	log.WithField("image", chain.Service.Image).Debug("Chain loaded")
	chain.Operations.PublishAllPorts = do.Operations.PublishAllPorts // TODO: remove this and marshall into struct from cli directly
	chain.Operations.Ports = do.Operations.Ports

	// Cmd should be "new" or "install".
	// [zr] install is basically deprecated. Can remove L20-26 in loaders/chains.go
	chain.Service.Command = cmd

	// Write the list of <key>:<value> config options as flags.
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
		fmt.Sprintf("CHAIN_ID=%s", chain.ChainID),
		fmt.Sprintf("CONTAINER_NAME=%s", containerName),
		fmt.Sprintf("CONFIG_OPTS=%s", configOpts),
		fmt.Sprintf("NODE_ADDR=%s", do.Gateway), // etcb host.
	}
	envVars = append(envVars, do.Env...)

	log.WithFields(log.Fields{
		"environment": envVars,
		"links":       do.Links,
	}).Debug()
	chain.Service.Environment = append(chain.Service.Environment, envVars...)
	chain.Service.Links = append(chain.Service.Links, do.Links...)
	chain.Operations.DataContainerName = util.DataContainerName(do.Name)

	if err := bootDependencies(chain, do); err != nil {
		return err
	}

	log.Info("Moving priv_validator.json into eris-keys")
	doKeys := definitions.NowDo()
	doKeys.Name = do.Name
	doKeys.Operations.Args = []string{"mintkey", "eris", fmt.Sprintf("%s/chains/%s/priv_validator.json", ErisContainerRoot, do.Name)}
	doKeys.Operations.SkipLink = true
	if out, err := ExecChain(doKeys); err != nil {
		if out != nil {
			log.Error(out)
		}
		return fmt.Errorf("Error moving keys: %v", err)
	}

	doChown := definitions.NowDo()
	doChown.Name = do.Name
	doChown.Operations.Args = []string{"chown", "--recursive", "eris", ErisContainerRoot}
	doChown.Operations.SkipLink = true
	if out, err := ExecChain(doChown); err != nil {
		if out != nil {
			log.Error(out)
		}
		return fmt.Errorf("Error changing owner: %v", err)
	}

	log.WithFields(log.Fields{
		"=>":           chain.Service.Name,
		"links":        chain.Service.Links,
		"volumes from": chain.Service.VolumesFrom,
		"image":        chain.Service.Image,
		"ports":        chain.Service.Ports,
	}).Debug("Performing chain container start")

	if err := perform.DockerRunService(chain.Service, chain.Operations); err != nil {
		return RemoveChain(do)
	}
	return
}

// genesis file either given directly, in dir, or not found (empty)
func resolveGenesisFile(genesis, dir string) string {
	if genesis == "" {
		genesis = filepath.Join(dir, "genesis.json")
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
			log.WithFields(log.Fields{
				"from": f.key,
				"to":   filepath.Join(dst, f.value),
			}).Debug("Copying files")
			if err := Copy(f.key, filepath.Join(dst, f.value)); err != nil {
				log.Debugf("Error copying files: %v", err)
				return err
			}
		}
	}
	return nil
}

func CleanUp(do *definitions.Do) error {
	log.Info("Cleaning up")

	if do.Chain.ChainType == "throwaway" {
		log.WithField("=>", do.Chain.Name).Debug("Destroying throwaway chain")
		doRm := definitions.NowDo()
		doRm.Operations = do.Operations
		doRm.Name = do.Chain.Name

		doRm.File, doRm.RmD, doRm.RmHF, doRm.Force = true, true, true, true
		RemoveChain(doRm)

	} else {
		log.Debug("No throwaway chain to destroy")
	}

	if do.RmD {
		log.WithField("dir", filepath.Join(DataContainersPath, do.Service.Name)).Debug("Removing data dir on host")
		os.RemoveAll(filepath.Join(DataContainersPath, do.Service.Name))
	}

	if do.Rm {
		log.WithField("=>", do.Operations.SrvContainerName).Debug("Removing tmp service container")
		perform.DockerRemove(do.Service, do.Operations, true, true, false)
	}

	return nil
}

func whatChainStuffExists(chainName string) (bool, bool, bool, bool) {
	var chainDirExists bool
	var chainConfigExists bool
	var chainDataExists bool
	var chainContainerExists bool

	// does the chain directory exist?
	if util.DoesDirExist(filepath.Join(ChainsPath, chainName)) {
		chainDirExists = true
	} else {
		chainDirExists = false
		log.Debug("no chain dir exists")
	}

	// does the config file exist?
	_, err := loaders.LoadChainConfigFile(chainName) // TODO rename the func in loaders
	if err == nil {
		chainConfigExists = true
	} else {
		chainConfigExists = false
		log.Debug("no config file exists")
		//log.Warn(fmt.Sprintf("error: %v", err))
	}

	// does the chain data container exist?
	if util.IsData(chainName) {
		chainDataExists = true
	} else {
		chainDataExists = false
		log.Debug("no data container exists")
	}

	// does the chain container exist?
	if util.IsChain(chainName, false) { // false checks if exists
		chainContainerExists = true
	} else {
		chainContainerExists = false
		log.Debug("no chain container exists")
	}

	return chainDirExists, chainConfigExists, chainDataExists, chainContainerExists
}

func cleanChainData(name string) error {
	dir := filepath.Join(DataContainersPath, name)
	if util.DoesDirExist(dir) {
		log.WithField("dir", dir).Debug("Chain data already exists in")
		log.Debug("Overwriting with new data")
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	return nil
}
