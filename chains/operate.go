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

	. "github.com/eris-ltd/common/go/common"

	log "github.com/eris-ltd/eris-logger"
	"github.com/pborman/uuid"
)

func StartChain(do *definitions.Do) error {
	chainExists := true // if can't load chain definition, flip to false

	_, err := loaders.LoadChainDefinition(do.Name)
	if err != nil {
		chainExists = false
	}

	if chainExists {
		log.Warn("chain exists, starting it")
		if do.Path != "" {
			return fmt.Errorf("chain already exists, cannot start existing chain with --init-dir")
		}

		if !util.IsData(do.Name) {
			return fmt.Errorf("no data container found, start a chain with [--init-dir]")
		}
		_, err := startChain(do, false) // [zr] why are we ignoring the buffer?
		return err

	} else if do.Path == "" {
		log.Warn("--init-dir left empty & chain does not exist")
		assumePath := filepath.Join(ChainsPath, do.Name)
		runThisCommand := fmt.Sprintf("[eris chains start %s --init-dir %s]", do.Name, assumePath)
		if util.DoesDirExist(assumePath) {
			return fmt.Errorf("a directory of chain name was found. re-run this command:\n%s", runThisCommand)

		}

		doWeMakeYouAChain := fmt.Sprintf("would you like the marmots to make you a simplechain?\nthis is normally done by running\n%s", runThisCommand)

		if QueryYesOrNo(doWeMakeYouAChain) == Yes {
			// chains make --chain-type=simplechain // name is what ?!?
			doMake := definitions.NowDo()
			doMake.Name = do.Name
			doMake.ChainType = "simplechain"
			if err := MakeChain(doMake); err != nil {
				return err
			}
		}

		do.Path = assumePath
		log.WithField("=>", do.Name).Warn("Setting up chain")
		return setupChain(do, loaders.ErisChainNew)

	} else {
		log.Warn("chain does not exist, new-ing it")
		if err := cleanChainData(do.Name); err != nil {
			return err
		}
		if !util.DoesDirExist(do.Path) {
			return fmt.Errorf("path specified on --init-dir (%s) is not a directory", do.Path)
		}

		// todo: fix this hack
		// for now we just let setupChain force do.ChainID = do.Name
		// and we overwrite using jq in the container
		log.WithField("=>", do.Name).Debug("Setting up chain")
		return setupChain(do, loaders.ErisChainNew)
	}
	return nil
}

func KillChain(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name)
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

	if do.Rm {
		if err := perform.DockerRemove(chain.Service, chain.Operations, do.RmD, do.Volumes, do.Force); err != nil {
			return err
		}
	}

	return nil
}

func ExecChain(do *definitions.Do) (buf *bytes.Buffer, err error) {
	return startChain(do, true)
}

// Throw away chains are used for eris contracts
// [zr] this should either be refactored & better documented or deprecated
// I favour the latter
func ThrowAwayChain(do *definitions.Do) error {
	do.Name = do.Name + "_" + strings.Split(uuid.New(), "-")[0]
	do.Path = filepath.Join(ChainsPath, "default")
	log.WithFields(log.Fields{
		"=>":   do.Name,
		"path": do.Path,
	}).Debug("Making a throaway chain")

	StartChain(do)
	log.WithField("=>", do.Name).Debug("Throwaway chain started")
	return nil
}

func startChain(do *definitions.Do, exec bool) (buf *bytes.Buffer, err error) {
	chain, err := loaders.LoadChainDefinition(do.Name)
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

	// boot the dependencies (eg. keys, logrotate)
	if err := bootDependencies(chain, do); err != nil {
		return nil, err
	}

	chain.Service.Command = loaders.ErisChainStart
	util.Merge(chain.Operations, do.Operations)
	chain.Service.Environment = append(chain.Service.Environment, "CHAIN_ID="+chain.ChainID)
	chain.Service.Environment = append(chain.Service.Environment, do.Env...)
	if do.Run {
		chain.Service.Environment = append(chain.Service.Environment, "ERISDB_API=true")
	}
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
			chn, err := loaders.LoadChainDefinition(chainName)
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
func setupChain(do *definitions.Do, cmd string) (err error) {
	// do.Name is mandatory
	if do.Name == "" {
		return fmt.Errorf("setupChain requires a chainame")
	}

	containerName := util.ChainContainerName(do.Name)
	if do.ChainID == "" {
		do.ChainID = do.Name
	}

	if do.Path != "" {
		src, errSaved := os.Stat(do.Path)
		if errSaved != nil || !src.IsDir() {
			log.WithField("path", do.Path).Info("Path does not exist or not a directory")
			log.WithField("path", "$HOME/.eris/chains/"+do.Path).Info("Trying")
			do.Path, err = util.ChainsPathChecker(do.Path)
			if err != nil {
				// Output the error of first attempt, not the second, because
				// this "stat /Users/peter/.eris/chains/Users/peter/.eris/simplechain:
				// no such file or directory" is ugly.
				return errSaved
			}
		}
	} else if do.GenesisFile == "" && len(do.ConfigOpts) == 0 {
		// NOTE: this expects you to have ~/.eris/chains/default/ (ie. to have run `eris init`)
		do.Path, err = util.ChainsPathChecker("default")
		if err != nil {
			return err
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
		ops.Args = []string{"mkdir", "-p", path.Join(ErisContainerRoot, "chains", do.ChainID)}
		if _, err := perform.DockerExecData(ops, nil); err != nil {
			return err
		}
	}
	log.WithField("=>", do.Name).Debug("Chain data container built")

	// Get the config file
	if do.ConfigFile == "" {
		do.ConfigFile = filepath.Join(ChainsPath, "default", "config.toml")
	}

	// Copy do.Path, do.GenesisFile, do.ConfigFile, do.Priv into container.
	containerDst := path.Join(ErisContainerRoot, "chains", do.ChainID) // path in container
	dst := filepath.Join(DataContainersPath, do.Name, containerDst)    // path on host

	log.WithFields(log.Fields{
		"container path": containerDst,
		"local path":     dst,
	}).Debug()

	if err = os.MkdirAll(dst, 0700); err != nil {
		return fmt.Errorf("Error making data directory: %v", err)
	}

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

	chain := loaders.MockChainDefinition(do.Name, do.ChainID)

	// Set maintainer info.
	chain.Maintainer.Name, chain.Maintainer.Email, err = config.GitConfigUser()
	if err != nil {
		log.Debug(err.Error())
	}

	// Write the chain definition file.
	// write in chains/chainName dir!
	fileName := filepath.Join(ChainsPath, do.Name, do.Name) + ".toml"
	if _, err = os.Stat(fileName); err != nil {
		if err = WriteChainDefinitionFile(chain, fileName); err != nil {
			return fmt.Errorf("error writing chain definition to file: %v", err)
		}
	}

	chain, err = loaders.LoadChainDefinition(do.Name)
	if err != nil {
		return err
	}
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
		fmt.Sprintf("CHAIN_ID=%s", do.ChainID),
		fmt.Sprintf("CONTAINER_NAME=%s", containerName),
		fmt.Sprintf("CONFIG_OPTS=%s", configOpts),
		fmt.Sprintf("NODE_ADDR=%s", do.Gateway), // etcb host.
	}
	envVars = append(envVars, do.Env...)

	if do.Run {
		// run erisdb instead of tendermint
		envVars = append(envVars, "ERISDB_API=true")
	}

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
	do.Force = true

	if do.Chain.ChainType == "throwaway" {
		log.WithField("=>", do.Chain.Name).Debug("Destroying throwaway chain")
		doRm := definitions.NowDo()
		doRm.Operations = do.Operations
		doRm.Name = do.Chain.Name
		doRm.Rm = true
		doRm.RmD = true
		doRm.Volumes = true
		KillChain(doRm)

		latentDir := filepath.Join(DataContainersPath, do.Chain.Name)
		latentFile := filepath.Join(ChainsPath, do.Chain.Name+".toml")

		if doRm.Name == "default" {
			log.WithField("dir", latentDir).Debug("Removing latent dir")
			os.RemoveAll(latentDir)
		} else {
			log.WithFields(log.Fields{
				"dir":  latentDir,
				"file": latentFile,
			}).Debug("Removing latent dir and file")
			os.RemoveAll(latentDir)
			os.Remove(latentFile)
		}

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
