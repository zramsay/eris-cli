package initialize

import (
	"fmt"
	"os"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func Initialize(skipPull bool) error {
	logger.Printf("The marmots have connected to Docker successfully.\nThey will now will install a few default services and actions for your use.\n\n")

	if _, err := os.Stat(common.ErisRoot); err != nil {
		if err := common.InitErisDir(); err != nil {
			return fmt.Errorf("Error:\tcould not Initialize the Eris Root Directory.\n%s\n", err)
		}
	} else {
		logger.Infof("Root eris directory (%s) already exists. Please type `eris` to see the help.\n", common.ErisRoot)
	}

	if err := InitDefaultServices(skipPull); err != nil {
		return fmt.Errorf("Error:\tcould not Instantiate default services.\n%s\n", err)
	}
	logger.Infof("Initialized eris root directory (%s) with default actions and service files.\n", common.ErisRoot)

	// todo: when called from cli provide option to go on tour, like `ipfs tour`
	logger.Printf("\nThe marmots have everything set up for you.\nIf you are just getting started please type [eris] to get an overview of the tool.\n")
	return nil
}

func InitDefaultServices(skipPull bool) error {
	logger.Debugf("Adding default files\n")
	if err := dropChainDefaults(); err != nil {
		return err
	}
	logger.Debugf("Chain defaults written.\n")

	if !skipPull {
		if err := cloneRepo("eris-services", common.ServicesPath); err != nil {
			logger.Errorf("Error cloning default services repository.\n%v\nTrying default defs.\n", err)
			if err2 := dropDefaults(); err2 != nil {
				return fmt.Errorf("Error:\tcannot clone services.\n%v\n%v", err, err2)
			}
		} else {
			if err2 := cloneRepo("eris-actions", common.ActionsPath); err2 != nil {
				return fmt.Errorf("Error:\tcannot clone actions.\n%v", err2)
			}
		}
	} else {
		logger.Debugf("Skip pull param given. Complying.\n")
		if err := dropDefaults(); err != nil {
			return err
		}
	}

	logger.Debugf("Service defaults written.\n")
	logger.Debugf("Action defaults written.\n")
	return nil
}
