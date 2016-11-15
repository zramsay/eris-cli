package pkgs

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"
)

func RunPackage(do *definitions.Do) error {
	var err error

	// Load the package if it doesn't exist
	if do.Package == nil {
		do.Package, err = LoadPackage(do.YAMLPath)
		if err != nil {
			return err
		}
	}

	// boot the chain

	switch do.ChainName { // switch on the flag
	case "", "$chain":
		head, _ := util.GetHead() // checks the checkedout chain
		if head != "" {           // used checked out chain
			log.WithField("=>", head).Info("No chain flag or in package file. Booting chain from checked out chain")
			err = bootChain(head, do)
		} else { // if no chain is checked out and no --chain given, default to a throwaway
			log.Warn("No chain was given, please start a chain")
			err = fmt.Errorf("no more throwaway chains")
		}
	default:
		log.WithField("=>", do.ChainName).Info("No chain flag used. Booting chain from package file")
		err = bootChain(do.ChainName, do)
	}

	if err != nil {
		return err
	}
	linkAppToChain(do)
	//do.ChainName = fmt.Sprintf("tcp://%s:%s", do.ChainName, do.ChainPort)
	return perform.RunJobs(do)
}
