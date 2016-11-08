package chains

import (
	"fmt"
	//"io"
	//"io/ioutil"
	//"path"
	//"path/filepath"
	// "strings"

	"github.com/eris-ltd/eris-cli/config"
	// "github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	//"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	//"github.com/eris-ltd/eris-cli/util"

	cm_maker "github.com/eris-ltd/eris-cli/maker"
	cm_definitions "github.com/eris-ltd/eris-cli/maker_definitions"
	cm_util "github.com/eris-ltd/eris-cli/util"
	keys "github.com/eris-ltd/eris-keys/eris-keys"
)

// MakeChain runs the `eris-cm make` command in a Docker container.
// It returns an error. Note that if do.Known, do.AccountTypes
// or do.ChainType are not set the command will run via interactive
// shell.
//
//  do.Name          - name of the chain to be created (required)
//  do.Known         - will use the mintgen tool to parse csv's and create a genesis.json (requires do.ChainMakeVals and do.ChainMakeActs) (optional)
//  do.Output  	     - unclear what this does XXX
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
	doKeys := definitions.NowDo()
	doKeys.Name = "keys"
	if err := services.EnsureRunning(doKeys); err != nil {
		return err
	}

	// loop through chains directories to make sure they exist & are appropriately populated
	//for _, d := range ChainsDirs {
	//	if _, err := os.Stat(d); os.IsNotExist(err) {
	//		os.MkdirAll(d, 0755)
	//	}
	//}
	if err := cm_util.CheckDefaultTypes(config.AccountsTypePath, "account-types"); err != nil {
		return err
	}
	if err := cm_util.CheckDefaultTypes(config.ChainTypePath, "chain-types"); err != nil {
		return err
	}

	// announce.
	log.Info("Hello! I'm the marmot who makes eris chains.")
	maker := cm_definitions.NowDo()
	keys.DaemonAddr = "http://172.17.0.2:4767" // tmp

	// todo. clean this up... struct merge them or something
	maker.Name = do.Name
	maker.Verbose = do.Verbose
	maker.Debug = do.Debug
	maker.ChainType = do.ChainType
	maker.AccountTypes = do.AccountTypes
	maker.Zip = do.ZipFile
	maker.Tarball = do.Tarball
	maker.Output = do.Output
	if do.Known {
		maker.CSV = fmt.Sprintf("%s,%s", do.ChainMakeVals, do.ChainMakeActs)
	}

	// make it
	if err := cm_maker.MakeChain(maker); err != nil {
		return err
	}

	// cm currently is not opinionated about its writers.
	if maker.Tarball {
		if err := cm_util.Tarball(maker); err != nil {
			return err
		}
	} else if maker.Zip {
		if err := cm_util.Zip(maker); err != nil {
			return err
		}
	}
	if maker.Output {
		if err := cm_util.SaveAccountResults(maker); err != nil {
			return err
		}
	}

	return nil
}
