package chains

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

func NewChain(do *definitions.Do) error {
	//overwrites directory if --force
	dir := filepath.Join(DataContainersPath, do.Name)
	if _, err := os.Stat(dir); err == nil {
		log.WithField("dir", dir).Debug("Chain data already exists in")
		if do.Force {
			log.Debug("Overwriting with new data")
			if os.RemoveAll(dir); err != nil {
				return err
			}
		} else {
			log.Debug("Using existing data; `--force` flag not given")
		}
	}

	// for now we just let setupChain force do.ChainID = do.Name
	// and we overwrite using jq in the container
	log.WithField("=>", do.Name).Debug("Setting up chain")
	return setupChain(do, loaders.ErisChainNew)
}

func InstallChain(do *definitions.Do) error {
	return setupChain(do, loaders.ErisChainInstall)
}

func KillChain(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name, false, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}

	if do.Force {
		do.Timeout = 0 //overrides 10 sec default
	}

	if IsChainRunning(chain) {
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

func StartChain(do *definitions.Do) error {
	_, err := startChain(do, false)

	return err
}

func ExecChain(do *definitions.Do) (buf *bytes.Buffer, err error) {
	return startChain(do, true)
}

// Throw away chains are used for eris contracts
func ThrowAwayChain(do *definitions.Do) error {
	do.Name = do.Name + "_" + strings.Split(uuid.New(), "-")[0]
	do.Path = filepath.Join(ChainsPath, "default")
	log.WithFields(log.Fields{
		"=>":   do.Name,
		"path": do.Path,
	}).Debug("Making a throaway chain")

	if err := NewChain(do); err != nil {
		return err
	}

	log.WithField("=>", do.Name).Debug("Throwaway chain created")
	do.Run = true  // turns on edb api
	StartChain(do) // XXX [csk]: may not need to do this now that New starts....
	log.WithField("=>", do.Name).Debug("Throwaway chain started")
	return nil
}

//------------------------------------------------------------------------
func startChain(do *definitions.Do, exec bool) (buf *bytes.Buffer, err error) {
	chain, err := loaders.LoadChainDefinition(do.Name, false, do.Operations.ContainerNumber)
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

	// boot the dependencies (eg. keys, logsrotate)
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

		// always link the chain to the exec container when doing chains exec
		// so that there is never any problems with sending info to the service (chain) container
		chain.Service.Links = append(chain.Service.Links, fmt.Sprintf("%s:%s", util.ContainersName("chain", chain.Name, 1), "chain"))

		buf, err = perform.DockerExecService(chain.Service, chain.Operations)
	} else {
		err = perform.DockerRunService(chain.Service, chain.Operations)
	}
	if err != nil {
		do.Result = "error"
		return nil, err
	}

	return buf, nil
}

// boot chain dependencies
// TODO: this currently only supports simple services (with no further dependencies)
func bootDependencies(chain *definitions.Chain, do *definitions.Do) error {
	if do.Logsrotate {
		chain.Dependencies.Services = append(chain.Dependencies.Services, "logsrotate")
	}
	if chain.Dependencies != nil {
		name := do.Name
		log.WithFields(log.Fields{
			"services": chain.Dependencies.Services,
			"chains":   chain.Dependencies.Chains,
		}).Info("Booting chain dependencies")
		for _, srvName := range chain.Dependencies.Services {
			do.Name = srvName
			srv, err := loaders.LoadServiceDefinition(do.Name, false, do.Operations.ContainerNumber)
			if err != nil {
				return err
			}

			// Start corresponding service.
			if !services.IsServiceRunning(srv.Service, srv.Operations) {
				name := strings.ToUpper(do.Name)
				log.WithField("=>", name).Info("Dependency not running. Starting now")
				if err = perform.DockerRunService(srv.Service, srv.Operations); err != nil {
					return err
				}
			}

		}
		do.Name = name // undo side effects

		for _, chainName := range chain.Dependencies.Chains {
			chn, err := loaders.LoadChainDefinition(chainName, false, do.Operations.ContainerNumber)
			if err != nil {
				return err
			}
			if !IsChainRunning(chn) {
				return fmt.Errorf("chain %s depends on chain %s but %s is not running", chain.Name, chainName, chainName)
			}
		}
	}
	return nil
}

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

	//if given path does not exist, see if its a reference to something in ~/.eris/chains/chainName
	if do.Path != "" {
		src, err := os.Stat(do.Path)
		if err != nil || !src.IsDir() {
			log.WithField("path", do.Path).Info("Path does not exist or not a directory")
			log.WithField("path", "$HOME/.eris/chains/"+do.Path).Info("Trying")
			do.Path, err = util.ChainsPathChecker(do.Path)
			if err != nil {
				return err
			}
		}
	} else if do.GenesisFile == "" && do.CSV == "" && len(do.ConfigOpts) == 0 {
		// NOTE: this expects you to have ~/.eris/chains/default/ (ie. to have run `eris init`)
		do.Path, err = util.ChainsPathChecker("default")
		if err != nil {
			return err
		}
	}

	// ensure/create data container
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {
		log.WithField("=>", do.Name).Debug("Chain data container already exists")
	} else {
		ops := loaders.LoadDataDefinition(do.Name, do.Operations.ContainerNumber)
		if err := perform.DockerCreateData(ops); err != nil {
			return fmt.Errorf("Error creating data container =>\t%v", err)
		}
	}

	log.WithField("=>", do.Name).Debug("Chain data container built")

	// if something goes wrong, cleanup
	defer func() {
		if err != nil {
			log.Infof("Error on setting up chain: %v", err)
			log.Info("Cleaning up")
			if err2 := RmChain(do); err2 != nil {
				// maybe be less dramatic
				err = fmt.Errorf("Tragic! Our marmots encountered an error during setupChain for %s.\nThey also failed to cleanup after themselves (remove containers) due to another error.\nFirst error =>\t\t\t%v\nCleanup error =>\t\t%v\n", containerName, err, err2)
			}
		}
	}()

	// copy do.Path, do.GenesisFile, do.ConfigFile, do.Priv, do.CSV into container
	containerDst := filepath.Join("chains", do.Name)                // path in container
	dst := filepath.Join(DataContainersPath, do.Name, containerDst) // path on host
	// TODO: deal with do.Operations.ContainerNumbers ....!
	// we probably need to update Import

	log.WithFields(log.Fields{
		"container path": containerDst,
		"local path":     dst,
	}).Debug()

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
			csvPath1 := fmt.Sprintf("%s/%s/%s/%s", ErisContainerRoot, "chains", do.ChainID, "validators.csv")
			csvPath2 := fmt.Sprintf("%s/%s/%s/%s", ErisContainerRoot, "chains", do.ChainID, "accounts.csv")
			csvPaths = fmt.Sprintf("%s,%s", csvPath1, csvPath2)
		} else {
			csvPaths = fmt.Sprintf("%s/%s/%s/%s", ErisContainerRoot, "chains", do.ChainID, "genesis.csv")
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
	importDo.Destination = ErisContainerRoot
	importDo.Source = filepath.Join(DataContainersPath, do.Name)
	if err = data.ImportData(importDo); err != nil {
		return err
	}

	chain := loaders.MockChainDefinition(do.Name, do.ChainID, false, do.Operations.ContainerNumber)

	//set maintainer info
	chain.Maintainer.Name, chain.Maintainer.Email, err = config.GitConfigUser()
	if err != nil {
		log.Debug(err.Error())
	}

	// write the chain definition file ...
	fileName := filepath.Join(ChainsPath, do.Name) + ".toml"
	if _, err = os.Stat(fileName); err != nil {
		if err = WriteChainDefinitionFile(chain, fileName); err != nil {
			return fmt.Errorf("error writing chain definition to file: %v", err)
		}
	}

	chain, err = loaders.LoadChainDefinition(do.Name, false, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}
	log.WithField("image", chain.Service.Image).Debug("Chain loaded")
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
		fmt.Sprintf("CSV=%v", csvPaths),                                          // for mintgen
		fmt.Sprintf("CONFIG_OPTS=%s", configOpts),                                // for config.toml
		fmt.Sprintf("NODE_ADDR=%s", do.Gateway),                                  // etcb host
		fmt.Sprintf("DOCKER_FIX=%s", "                                        "), // https://github.com/docker/docker/issues/14203
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

	// TODO: if do.N > 1 ...

	chain.Operations.DataContainerName = util.DataContainersName(do.Name, do.Operations.ContainerNumber)

	if err := bootDependencies(chain, do); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"=>":    chain.Service.Name,
		"image": chain.Service.Image,
	}).Debug("Performing chain container start")

	err = perform.DockerRunService(chain.Service, chain.Operations)
	// this err is caught in the defer above

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
