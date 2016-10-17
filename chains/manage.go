package chains

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
)

// MakeChain runs the `eris-cm make` command in a Docker container.
// It returns an error. Note that if do.Known, do.AccountTypes
// or do.ChainType are not set the command will run via interactive
// shell.
//
//  do.Name          - name of the chain to be created (required)
//  do.Known         - will use the mintgen tool to parse csv's and create a genesis.json (requires do.ChainMakeVals and do.ChainMakeActs) (optional)
//  do.ChainMakeVals - csv file to use for validators (optional)
//  do.ChainMakeActs - csv file to use for accounts (optional)
//  do.AccountTypes  - use eris-cm make account-types paradigm (example: Root:1,Participants:25,...) (optional)
//  do.ChainType     - use eris-cm make chain-types paradigm (example: simplechain) (optional)
//  do.Tarball       - instead of outputing raw files in directories, output packages of tarbals (optional)
//  do.ZipFile       - similar to do.Tarball except uses zipfiles (optional)
//  do.Verbose       - verbose output (optional)
//  do.Debug         - debug output (optional)
//
func MakeChain(do *definitions.Do) error {
	if err := checkKeysRunningOrStart(); err != nil {
		return err
	}

	do.Service.Name = do.Name
	do.Service.Image = path.Join(config.Global.DefaultRegistry, config.Global.ImageCM)
	do.Service.User = "eris"
	do.Service.AutoData = true
	do.Service.Links = []string{fmt.Sprintf("%s:%s", util.ServiceContainerName("keys"), "keys")}
	do.Service.Environment = []string{
		fmt.Sprintf("ERIS_KEYS_PATH=http://keys:%d", 4767), // note, needs to be made aware of keys port...
		fmt.Sprintf("ERIS_CHAINMANAGER_ACCOUNTTYPES=%s", strings.Join(do.AccountTypes, ",")),
		fmt.Sprintf("ERIS_CHAINMANAGER_CHAINTYPE=%s", do.ChainType),
		fmt.Sprintf("ERIS_CHAINMANAGER_TARBALLS=%v", do.Tarball),
		fmt.Sprintf("ERIS_CHAINMANAGER_ZIPFILES=%v", do.ZipFile),
		fmt.Sprintf("ERIS_CHAINMANAGER_OUTPUT=%v", do.Output),
		fmt.Sprintf("ERIS_CHAINMANAGER_VERBOSE=%v", do.Verbose),
		fmt.Sprintf("ERIS_CHAINMANAGER_DEBUG=%v", do.Debug),
	}

	do.Operations.ContainerType = definitions.TypeService
	do.Operations.SrvContainerName = util.ServiceContainerName(do.Name)
	do.Operations.DataContainerName = util.DataContainerName(do.Name)
	do.Operations.Labels = util.Labels(do.Name, do.Operations)
	if do.RmD {
		do.Operations.Remove = true
	}

	if do.Known {
		log.Debug("Using MintGen rather than eris:cm")
		do.Service.EntryPoint = "mintgen"
		do.Service.Command = fmt.Sprintf("known %s --csv=%s,%s > %s", do.Name, do.ChainMakeVals, do.ChainMakeActs, path.Join(ErisContainerRoot, "chains", do.Name, "genesis.json"))
	} else {
		log.Debug("Using eris:cm rather than MintGen")
		do.Service.EntryPoint = fmt.Sprintf("eris-cm make %s", do.Name)
	}

	if do.Wizard && len(do.AccountTypes) == 0 && do.ChainType == "" {
		do.Operations.Interactive = true
		do.Operations.Args = strings.Split(do.Service.EntryPoint, " ")
	}

	if do.Known {
		do.Operations.Args = append(do.Operations.Args, strings.Split(do.Service.Command, " ")...)
		do.Service.WorkDir = path.Join(ErisContainerRoot, "chains", do.Name)
	}

	doData := definitions.NowDo()
	doData.Name = do.Name

	doData.Operations.DataContainerName = util.DataContainerName(do.Name)
	doData.Operations.ContainerType = "service"

	doData.Source = AccountsTypePath
	doData.Destination = path.Join(ErisContainerRoot, "chains", "account-types")
	if err := data.ImportData(doData); err != nil {
		return fmt.Errorf("Cannot import account-types into container: %v", err)
	}

	doData.Source = ChainTypePath
	doData.Destination = path.Join(ErisContainerRoot, "chains", "chain-types")
	if err := data.ImportData(doData); err != nil {
		return fmt.Errorf("Cannot import chain-types into container: %v", err)
	}

	chnPath := filepath.Join(ChainsPath, do.Name)
	doData.Source = chnPath
	doData.Destination = path.Join(ErisContainerRoot, "chains", do.Name)
	if err := data.ImportData(doData); err != nil {
		return fmt.Errorf("Cannot import chain directory into container: %v", err)
	}

	buf, err := perform.DockerExecService(do.Service, do.Operations)
	if err != nil {
		// Log to both screen and logs for further analysis in Bugsnag.
		log.Debug("Dumping [eris-cm] output")
		log.Error(buf.String())

		// After all the imports are in place, [eris-cm] should not fail,
		// so it is worth investigating why it still failed.
		util.SendReport("`eris chains make` failed")

		return err
	}

	// TODO(pv): remove this line after `eris-cm` command line handling is fixed.
	// This line exists to capture `eris-cm` errors which return exit code 0.
	io.Copy(config.Global.Writer, buf)

	doData.Source = path.Join(ErisContainerRoot, "chains")
	doData.Destination = ErisRoot
	if err := data.ExportData(doData); err != nil {
		return fmt.Errorf("Cannot copy chain directory back to host: %v", err)
	}

	if !do.RmD {
		return data.RmData(doData)
	}

	return nil
}

// InspectChain is Eris' version of [docker inspect]. It returns
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
		err := services.InspectServiceByService(chain.Service, chain.Operations, do.Operations.Args[0])
		if err != nil {
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
	rootDir := path.Join(ErisContainerRoot, "chains", do.Name)

	doCat := definitions.NowDo()
	doCat.Name = do.Name
	doCat.Operations.SkipLink = true

	switch do.Type {
	case "genesis":
		doCat.Operations.Args = []string{"cat", path.Join(rootDir, "genesis.json")}
	case "config":
		doCat.Operations.Args = []string{"cat", path.Join(rootDir, "config.toml")}
	// TODO re-implement with eris-client ... mintinfo was remove from container (and write tests for these cmds)
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

// TODO remove pointer line in CONFIG_PATHS.csv
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
		dirPath := filepath.Join(ChainsPath, do.Name) // the dir

		log.WithField("directory", dirPath).Warn("Removing directory")
		if err := os.RemoveAll(dirPath); err != nil {
			return err
		}
	}

	return nil
}

func checkKeysRunningOrStart() error {
	srv, err := loaders.LoadServiceDefinition("keys")
	if err != nil {
		return err
	}

	if !util.IsService(srv.Service.Name, true) {
		do := definitions.NowDo()
		do.Operations.Args = []string{"keys"}
		if err := services.StartService(do); err != nil {
			return err
		}
	}
	return nil
}
