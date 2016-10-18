package chains

import (
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/common/go/common"
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
		do.Service.Command = fmt.Sprintf("known %s --csv=%s,%s > %s", do.Name, do.ChainMakeVals, do.ChainMakeActs, path.Join(common.ErisContainerRoot, "chains", do.Name, "genesis.json"))
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
		do.Service.WorkDir = path.Join(common.ErisContainerRoot, "chains", do.Name)
	}

	doData := definitions.NowDo()
	doData.Name = do.Name

	doData.Operations.DataContainerName = util.DataContainerName(do.Name)
	doData.Operations.ContainerType = "service"

	doData.Source = common.AccountsTypePath
	doData.Destination = path.Join(common.ErisContainerRoot, "chains", "account-types")
	if err := data.ImportData(doData); err != nil {
		return fmt.Errorf("Cannot import account-types into container: %v", err)
	}

	doData.Source = common.ChainTypePath
	doData.Destination = path.Join(common.ErisContainerRoot, "chains", "chain-types")
	if err := data.ImportData(doData); err != nil {
		return fmt.Errorf("Cannot import chain-types into container: %v", err)
	}

	chnPath := filepath.Join(common.ChainsPath, do.Name)
	doData.Source = chnPath
	doData.Destination = path.Join(common.ErisContainerRoot, "chains", do.Name)
	if util.DoesDirExist(doData.Source) {
		if err := data.ImportData(doData); err != nil {
			return fmt.Errorf("Cannot import chain directory into container: %v", err)
		}
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

	doData.Source = path.Join(common.ErisContainerRoot, "chains")
	doData.Destination = common.ErisRoot
	if err := data.ExportData(doData); err != nil {
		return fmt.Errorf("Cannot copy chain directory back to host: %v", err)
	}

	if !do.RmD {
		return data.RmData(doData)
	}

	return nil
}
