package chains

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/monax/cli/config"
	"github.com/monax/cli/data"
	"github.com/monax/cli/definitions"
	"github.com/monax/cli/loaders"
	"github.com/monax/cli/log"
	"github.com/monax/cli/perform"
	"github.com/monax/cli/services"
	"github.com/monax/cli/util"
)

func StartChain(do *definitions.Do) error {
	// Start an already set up chain.
	if util.IsChain(do.Name, false) && util.IsData(do.Name) && !do.Force {
		_, err := startChain(do, false)
		return err
	}

	// Default [--init-dir] value is chain's root.
	if do.Path == "" {
		do.Path = filepath.Join(config.ChainsPath, do.Name)
	}

	// Resolve chain's path.
	var err error
	do.Path, err = resolveChainsPath(do.Name, do.Path)
	if err != nil {
		return err
	}

	if do.Force {
		// Chain is reinitialized upon request.
		log.WithField("=>", do.Name).Debug("Initializing the chain: [--force] flag given")
	} else {
		// Chain is broken (either chain or data chain container doesn't exist),
		// initialize the chain.
		log.WithField("=>", do.Name).Debug("Initializing the chain: chain or data container doesn't exist")
	}

	return setupChain(do)
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

// InspectChain is Monax' version of [docker inspect]. It returns
// an error.
//
//  do.Name            - name of the chain to inspect (required)
//  do.Operations.Args - fields to inspect in the form Major.Minor or "all" (required)
//
func InspectChain(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name)
	if err != nil {
		return err
	}

	if util.IsChain(chain.Name, false) {
		log.WithField("=>", chain.Service.Name).Debug("Inspecting chain")
		if err := services.InspectServiceByService(chain.Service, chain.Operations, do.Operations.Args[0]); err != nil {
			return err
		}
	}

	return nil
}

// LogsChain returns the logs of a chains' service container
// for display by the user.
//
//  do.Name    - name of the chain (required)
//  do.Follow  - follow the logs until the user sends SIGTERM (optional)
//  do.Tail    - number of lines to display (can be "all") (optional)
//
func LogsChain(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name)
	if err != nil {
		return err
	}

	err = perform.DockerLogs(chain.Service, chain.Operations, do.Follow, do.Tail)
	if err != nil {
		return err
	}

	return nil
}

// CheckoutChain writes to the ChainPath/HEAD file the name
// of the chain to be "checked out". It returns an error. This
// operates similar to git branches and is predominantly a
// scoping function which is used by other portions of the
// platform where a --chain flag may otherwise be used.
//
//  do.Name - the name of the chain to checkout; if blank will "uncheckout" current chain (optional)
//
func CheckoutChain(do *definitions.Do) error {
	if do.Name == "" {
		return util.NullHead()
	}

	curHead, _ := util.GetHead()
	if do.Name == curHead {
		return nil
	}

	return util.ChangeHead(do.Name)
}

// CurrentChain displays the currently in scope (or checked out) chain. It
// returns an error (which should never be triggered)
//
func CurrentChain(do *definitions.Do) (string, error) {
	head, _ := util.GetHead()

	if head == "" {
		head = "There is no chain checked out"
	}

	return head, nil
}

// CatChain displays chain information. It returns nil on success, or input/output
// errors otherwise.
//
//  do.Name - chain name
//  do.Type - "genesis", "config"
//
func CatChain(do *definitions.Do) error {
	if do.Name == "" {
		return fmt.Errorf("a chain name is required")
	}
	rootDir := path.Join(config.MonaxContainerRoot, "chains", do.Name)

	doCat := definitions.NowDo()
	doCat.Name = do.Name
	doCat.Operations.SkipLink = true

	switch do.Type {
	case "genesis":
		doCat.Operations.Args = []string{"cat", path.Join(rootDir, "genesis.json")}
	case "config":
		doCat.Operations.Args = []string{"cat", path.Join(rootDir, "config.toml")}
	// TODO re-implement with monax-client ... mintinfo was remove from container (and write tests for these cmds)
	// case "status":
	//	doCat.Operations.Args = []string{"mintinfo", "--node-addr", "http://chain:46657", "status"}
	// case "validators":
	//	doCat.Operations.Args = []string{"mintinfo", "--node-addr", "http://chain:46657", "validators"}
	default:
		return fmt.Errorf("unknown cat subcommand %q", do.Type)
	}
	// edb docker image is (now) properly formulated with entrypoint && cmd
	// so the entrypoint must be overwritten.
	log.WithField("args", do.Operations.Args).Debug("Executing command")

	buf, err := ExecChain(doCat)

	if buf != nil {
		io.Copy(config.Global.Writer, buf)
	}

	return err
}

// PortsChain displays the port mapping for a particular chain.
// It returns an error.
//
//  do.Name - name of the chain to display port mappings for (required)
//
func PortsChain(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name)
	if err != nil {
		return err
	}

	if util.IsChain(chain.Name, false) {
		log.WithField("=>", chain.Name).Debug("Getting chain port mapping")
		return util.PrintPortMappings(chain.Operations.SrvContainerName, do.Operations.Args)
	}

	return nil
}

func RemoveChain(do *definitions.Do) error {
	chain, err := loaders.LoadChainDefinition(do.Name)
	if err != nil {
		return err
	}

	if util.IsChain(chain.Name, false) {
		if err = perform.DockerRemove(chain.Service, chain.Operations, do.RmD, do.Volumes, do.Force); err != nil {
			return err
		}
	} else {
		log.Info("Chain container does not exist")
	}

	if do.RmHF {
		dirPath := filepath.Join(config.ChainsPath, do.Name) // the dir

		log.WithField("directory", dirPath).Warn("Removing directory")
		if err := os.RemoveAll(dirPath); err != nil {
			return err
		}
	}

	return nil
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

		// This override is necessary because monaxdb uses an entryPoint and
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
				return nil, fmt.Errorf("chain %v has failed to start. You may want to check the [monax chains logs %[1]s] command output", chain.Name)
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

// setupChain is invoked on [monax chains start CHAIN_NAME] command and
// creates chain and (if they're missing) keys containers.
func setupChain(do *definitions.Do) (err error) {
	// do.Name is mandatory.
	if do.Name == "" {
		return fmt.Errorf("Setting up chain without a chain name. Aborting")
	}

	containerName := util.ChainContainerName(do.Name)
	containerDst := path.Join(config.MonaxContainerRoot, "chains", do.Name)
	hostSrc := do.Path

	chain, err := loaders.LoadChainDefinition(do.Name, filepath.Join(do.Path, "config"))
	if err != nil {
		do.RmD = true
		RemoveChain(do)
		return fmt.Errorf("Failed to load chain config: %v", err)
	}
	log.WithField("image", chain.Service.Image).Debug("Chain loaded")

	chain.Service.Name = do.Name
	util.Merge(chain.Operations, do.Operations)

	// Set chain name and other vars.
	envVars := []string{
		// TODO remove CHAIN_ID once the fix in edb is merged
		fmt.Sprintf("CHAIN_ID=%s", chain.Name),
		// [zr] replacement for CHAIN_ID is CHAIN_NAME
		fmt.Sprintf("CHAIN_NAME=%s", chain.Name),
		fmt.Sprintf("BURROW_WORKDIR=%s", containerDst),
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

	// Ensure/create data container.
	if util.IsData(do.Name) {
		log.WithField("=>", do.Name).Debug("Chain data container already exists")
	} else {
		ops := loaders.LoadDataDefinition(do.Name)
		if err := perform.DockerCreateData(ops); err != nil {
			return fmt.Errorf("Could not create data container: %v", err)
		}
		ops.Args = []string{"mkdir", "-p", path.Join(config.MonaxContainerRoot, "chains", do.Name)}
		if _, err := perform.DockerExecData(ops, nil); err != nil {
			return err
		}
	}
	log.WithField("=>", do.Name).Debug("Chain data container built")

	// copy from host to container
	log.WithFields(log.Fields{
		"from": hostSrc,
		"to":   containerDst,
	}).Debug("Copying files into data container")

	importDo := definitions.NowDo()
	importDo.Name = do.Name
	importDo.Operations = do.Operations
	importDo.Destination = containerDst
	importDo.Source = hostSrc
	if err = data.ImportData(importDo); err != nil {
		do.RmD = true
		RemoveChain(do)
		return fmt.Errorf("Could not import data: %v", err)
	}

	// mintkey has been removed from the monaxdb image. this functionality
	// needs to be wholesale refactored. For now we'll just run the keys
	// service (where mintkey is....)

	importKey := definitions.NowDo()
	importKey.Name = "keys"
	importKey.Destination = containerDst
	importKey.Source = filepath.Join(hostSrc, "priv_validator.json")
	if err = data.ImportData(importKey); err != nil {
		do.RmD = true
		RemoveChain(do)
		return fmt.Errorf("Could not import [priv_validator.json] to signer: %v", err)
	}

	if out, err := services.ExecHandler("keys", []string{"mintkey", "monax", fmt.Sprintf("%s/chains/%s/priv_validator.json", config.MonaxContainerRoot, do.Name)}); err != nil {
		log.Error(err)
		do.RmD = true
		RemoveChain(do)
		return fmt.Errorf("Failed to transliterate [priv_validator.json] to monax-key: %v", out)
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

func resolveChainsPath(chainName, pathGiven string) (string, error) {
	for _, path := range []string{
		// Absolute path.
		pathGiven,
		// Relative of chains root path.
		filepath.Join(config.ChainsPath, pathGiven),
		// Relative of chain's dir path.
		filepath.Join(config.ChainsPath, chainName, pathGiven),
	} {
		if util.DoesFileExist(filepath.Join(path, "config.toml")) {
			return path, nil
		}
	}

	// No "config.toml" in a dir.
	if util.DoesDirExist(pathGiven) {
		log.WithField("=>", pathGiven).Info("Failed to find config.toml in [--init-dir]")
		return "", fmt.Errorf("Missing config.toml in %v. Try [monax chains make] first", pathGiven)
	}

	log.WithField("=>", pathGiven).Info("Failed to find [--init-dir]")
	return "", fmt.Errorf("Directory given on [--init-dir] could not be determined")
}

func exportFile(chainName string) (string, error) {
	fileName := util.GetFileByNameAndType("chains", chainName)

	return util.SendToIPFS(fileName, "", "")
}
