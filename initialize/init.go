package initialize

import (
	"fmt"
	"os"

	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func Initialize(skipPull, verbose bool) error { // todo: remove the verbose here

	if _, err := os.Stat(common.ErisRoot); err != nil {
		if err := common.InitErisDir(); err != nil {
			return fmt.Errorf("Could not Initialize the Eris Root Directory.\n%s\n", err)
		}
	} else {
		logger.Infof("Root eris directory (%s) already exists. Please type `eris` to see the help.\n", common.ErisRoot)
	}

	logger.Debugf("Checking connection to Docker....\n")
	if err := util.CheckDockerClient(); err != nil {
		return err
	}
	logger.Infof("Docker Connection OK.\n")

	if err := InitDefaultServices(skipPull); err != nil {
		return fmt.Errorf("Could not instantiate default services.\n%s\n", err)
	}
	logger.Infof("Initialized eris root directory (%s) with default actions and service files.\n", common.ErisRoot)

	// todo: when called from cli provide option to go on tour, like `ipfs tour`
	logger.Printf("The marmots have everything set up for you.\n")

	return nil
}

func InitDefaultServices(skipPull bool) error {
	logger.Debugf("Adding default files\n")
	if err := dropChainDefaults(); err != nil {
		return err
	}
	logger.Debugf("Chain defaults written.\n")

	if !skipPull {
		if err := pullRepo("eris-services", common.ServicesPath); err != nil {
			logger.Debugf("Using default defs.")
			if err2 := dropDefaults(); err2 != nil {
				return fmt.Errorf("Cannot pull: %s. %s.\n", err, err2)
			}
		} else {
			if err2 := pullRepo("eris-actions", common.ActionsPath); err2 != nil {
				return fmt.Errorf("Cannot pull actions: %s.\n", err2)
			}
		}
	} else {
		if err := dropDefaults(); err != nil {
			return err
		}
	}

	logger.Debugf("Service defaults written.\n")
	logger.Debugf("Action defaults written.\n")
	return nil
}
