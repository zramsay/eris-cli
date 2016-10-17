package chains

import (
	"bytes"
	"encoding/csv"
	"fmt"
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

	. "github.com/eris-ltd/common/go/common"

	log "github.com/eris-ltd/eris-logger"
)

func checkIfPathIsInChainsDirAndDoSomeCopying(name, path string) (string, bool, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", false, err
	}

	// check that common.ChainsPath is not within the
	// absolute path of do.Path given
	if strings.Contains(abs, ChainsPath) {
		log.Warn("strings contains")
		return path, false, nil
	}

	dest := filepath.Join(ChainsPath, name)

	if util.DoesDirExist(dest) {
		// TODO provide y/n option to overwrite
		return "", true, fmt.Errorf("Directory of chain name already exists in the chains directory. Please remove before continuing")
	}

	if err := os.Mkdir(dest, 0666); err != nil {
		log.Warn("making dir")
		return "", true, err
	}

	if err := util.CopyTree(abs, dest); err != nil {
		log.Warn("copy tree")
		return "", true, err
	}

	return abs, true, nil
}

func StartChain(do *definitions.Do) error {
	chainConfigExists, chainDataExists, chainContainerExists := whatChainStuffExists(do.Name)

	if do.Path != "" {
		// [eris chains start whatever --init-dir ~/.eris/chains/whatever]
		var err error
		var someBool bool

		do.Path, someBool, err = checkIfPathIsInChainsDirAndDoSomeCopying(do.Name, do.Path)
		if err != nil {
			return err
		}

		// skip this if not using common.ChainsPath
		if !someBool {
			do.Path, err = chainsPathSimplifier(do.Name, do.Path)
			if err != nil {
				return err
			}
		}

		if chainDataExists {
			return fmt.Errorf("Data container exists, re-run without [--init-dir]")
		}

		log.WithField("=>", do.Name).Debug("Data container does not exist, initializing it")
		return setupChain(do)

	} else {
		// [eris chains start whatever] (without --init-dir)
		if !chainConfigExists {
			log.Info("Could not find the local config file for that chain, checking for existence of chain data container")
		}

		if !chainDataExists {
			return fmt.Errorf("No data container found, please start a chain with [--init-dir]")
		}

		if !chainContainerExists {
			log.Info("Chain process container does not exist, creating it")
		}

		_, err := startChain(do, false)
		return err
	}
}

func StopChain(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name)
	if err != nil {
		return err
	}

	if do.Force {
		// Overrides the default.
		do.Timeout = 0
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
	chain, err := loaders.LoadChainDefinition(do.Name)
	if err != nil {
		log.Error("Cannot start a chain I cannot find")
		return nil, nil
	}

	if chain.Name == "" {
		log.Error("Cannot start a chain without a name")
		return nil, nil
	}

	// boot the dependencies (eg. keys, logrotate)
	if err := bootDependencies(chain, do); err != nil {
		return nil, err
	}

	util.Merge(chain.Operations, do.Operations)

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
		chain.Service.EntryPoint = do.Service.EntryPoint
		chain.Service.Command = do.Service.Command

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
		return buf, err
	}

	return buf, nil
}

// boot chain dependencies
// TODO: this currently only supports simple services (with no further dependencies)
func bootDependencies(chain *definitions.ChainDefinition, do *definitions.Do) error {
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
// handles both "new" and "fetch" - most of the differentiating logic is in the container <= [zr] huh?
// should only be ever called on [eris chains start someChain --init-dir ~/.eris/chains/someChain/someChain_full_000]
// or without the last dir for a simplechain.
func setupChain(do *definitions.Do) (err error) {
	// do.Name is mandatory
	if do.Name == "" {
		return fmt.Errorf("setupChain requires a chainame")
	}

	containerName := util.ChainContainerName(do.Name)
	containerDst := path.Join(ErisContainerRoot, "chains", do.Name)
	hostSrc := do.Path

	// writes a pointer (similar to checked out chain) for do.Path in the chain main dir
	// this can then be read by loaders.LoadChainDefinition(), in order to get the
	// path to the config.toml that was written in each directory
	// this allows cli to keep track of a given config.toml (locally)

	fileName := filepath.Join(ChainsPath, "CONFIG_PATHS.csv")
	csvFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("err opening file", err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	reader.TrimLeadingSpace = true

	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return err
	}
	//log.WithField("rawCSVdata", rawCSVdata).Debug("Data read.")
	for _, chain := range rawCSVdata {
		if chain[0] == do.Name {
			return fmt.Errorf("chain name already exists with path: %s", chain[1])
		}
	}

	writer := csv.NewWriter(csvFile)
	if err := writer.Write([]string{do.Name, do.Path}); err != nil {
		log.Fatal("err writing csv", err)
	}
	writer.Flush()

	chain, err := loaders.LoadChainDefinition(do.Name)
	if err != nil {
		do.RmD = true
		RemoveChain(do)
		return fmt.Errorf("Error loading chain config: %v", err)
	}
	log.WithField("image", chain.Service.Image).Debug("Chain loaded")

	chain.Service.Name = do.Name
	util.Merge(chain.Operations, do.Operations)

	// set chain name and other vars
	envVars := []string{
		// TODO remove CHAIN_ID once the fix in edb is merged
		fmt.Sprintf("CHAIN_ID=%s", chain.Name),
		// [zr] replacement for CHAIN_ID is CHAIN_NAME
		fmt.Sprintf("CHAIN_NAME=%s", chain.Name),
		fmt.Sprintf("ERIS_DB_WORKDIR=%s", containerDst),
		fmt.Sprintf("CONTAINER_NAME=%s", containerName),
	}
	envVars = append(envVars, do.Env...)

	chain.Service.Environment = append(chain.Service.Environment, envVars...)
	chain.Service.Links = append(chain.Service.Links, do.Links...)
	log.WithFields(log.Fields{
		"environment": chain.Service.Environment,
		"links":       chain.Service.Links,
	}).Debug()

	if err := bootDependencies(chain, do); err != nil {
		do.RmD = true
		RemoveChain(do)
		return fmt.Errorf("Error booting dependencies: %v", err)
	}

	// ensure/create data container
	if util.IsData(do.Name) {
		log.WithField("=>", do.Name).Debug("Chain data container already exists")
	} else {
		ops := loaders.LoadDataDefinition(do.Name)
		if err := perform.DockerCreateData(ops); err != nil {
			return fmt.Errorf("Error creating data container =>\t%v", err)
		}
		ops.Args = []string{"mkdir", "-p", path.Join(ErisContainerRoot, "chains", do.Name)}
		if _, err := perform.DockerExecData(ops, nil); err != nil {
			return err
		}
	}
	log.WithField("=>", do.Name).Debug("Chain data container built")

	// copy from host to container
	log.WithFields(log.Fields{
		"from local path":   hostSrc,
		"to container path": containerDst,
	}).Debug("Copying files into data container")

	importDo := definitions.NowDo()
	importDo.Name = do.Name
	importDo.Operations = do.Operations
	importDo.Destination = containerDst
	importDo.Source = hostSrc
	if err = data.ImportData(importDo); err != nil {
		do.RmD = true
		RemoveChain(do)
		return fmt.Errorf("Error importing data: %v", err)
	}

	// mintkey has been removed from the erisdb image. this functionality
	// needs to be wholesale refactored. For now we'll just run the keys
	// service (where mintkey is....)
	log.Info("Moving priv_validator.json into eris-keys")
	importKey := definitions.NowDo()
	importKey.Name = "keys"
	importKey.Destination = containerDst
	importKey.Source = filepath.Join(hostSrc, "priv_validator.json")
	if err = data.ImportData(importKey); err != nil {
		do.RmD = true
		RemoveChain(do)
		return fmt.Errorf("Error importing priv_validator to signer: %v", err)
	}
	doKeys := definitions.NowDo()
	doKeys.Name = "keys"
	doKeys.Operations.Args = []string{"mintkey", "eris", fmt.Sprintf("%s/chains/%s/priv_validator.json", ErisContainerRoot, do.Name)}
	doKeys.Operations.SkipLink = true
	doKeys.Service.VolumesFrom = []string{util.DataContainerName(do.Name)}
	if out, err := services.ExecService(doKeys); err != nil {
		log.Error(err)
		do.RmD = true
		RemoveChain(do)
		return fmt.Errorf("Error transliterating priv_validator to eris-key: %v", out) // out is the buffer from the container; error is from docker
	}

	log.WithFields(log.Fields{
		"=>":              chain.Service.Name,
		"links":           chain.Service.Links,
		"volumes from":    chain.Service.VolumesFrom,
		"image":           chain.Service.Image,
		"ports":           chain.Service.Ports,
		"environment":     chain.Service.Environment,
		"chain container": chain.Operations.SrvContainerName,
		"ports published": chain.Operations.PublishAllPorts,
	}).Debug("Performing chain container start")

	if err := perform.DockerRunService(chain.Service, chain.Operations); err != nil {
		do.RmD = true
		RemoveChain(do)
		return fmt.Errorf("Error starting chain: %v", err)
	}
	return
}

func chainsPathSimplifier(chainName, pathGiven string) (string, error) {
	if util.DoesDirExist(pathGiven) { // full path given, check that config.toml exists though
		if !doesConfigExist(pathGiven) {
			return "", fmt.Errorf("config.toml does not exists in %s", pathGiven)
		} else {
			return pathGiven, nil
		}
	} else {
		chainDirPathSimple := filepath.Join(ChainsPath, pathGiven)               // if simplechain, pathGiven == chainName
		chainDirPathNotSimple := filepath.Join(ChainsPath, chainName, pathGiven) // ignored if simplechain

		if util.DoesDirExist(chainDirPathSimple) && doesConfigExist(chainDirPathSimple) {
			return chainDirPathSimple, nil
		} else if util.DoesDirExist(chainDirPathNotSimple) && doesConfigExist(chainDirPathNotSimple) {
			return chainDirPathNotSimple, nil
		} else {
			return "", fmt.Errorf("Directory given on [--init-dir] could not be determined")
		}
	}
}

func doesConfigExist(dirPath string) bool {
	var configExists bool
	pathToConfig := filepath.Join(dirPath, "config.toml")
	if _, err := os.Stat(pathToConfig); os.IsNotExist(err) {
		configExists = false
	} else {
		configExists = true
	}
	return configExists
}

func whatChainStuffExists(chainName string) (bool, bool, bool) {
	// var chainDirExists bool
	var chainConfigExists bool
	var chainDataExists bool
	var chainContainerExists bool

	// does the chain directory exist?
	// if util.DoesDirExist(filepath.Join(ChainsPath, chainName)) {
	//	chainDirExists = true
	// } else {
	//	chainDirExists = false
	//}

	// does the config file exist?
	_, err := loaders.LoadChainDefinition(chainName)
	if err == nil {
		chainConfigExists = true
	} else {
		chainConfigExists = false
	}

	// does the chain data container exist?
	if util.IsData(chainName) {
		chainDataExists = true
	} else {
		chainDataExists = false
	}

	// does the chain container exist?
	if util.IsChain(chainName, false) { // false checks if exists
		chainContainerExists = true
	} else {
		chainContainerExists = false
	}

	return chainConfigExists, chainDataExists, chainContainerExists
}
