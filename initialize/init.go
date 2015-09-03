package initialize

import (
	"fmt"
	"os"

	"github.com/eris-ltd/eris-cli/definitions"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func Initialize(do *definitions.Do) error {
	var input string
	logger.Printf("The marmots have connected to Docker successfully.\nThey will now will install a few default services and actions for your use.\n\n")
	_, err := os.Stat(common.ErisRoot)
	if err != nil {
		logger.Printf("Eris Root Directory does not exist, initializing it\n")
		if err := common.InitErisDir(); err != nil {
			return fmt.Errorf("Error:\tcould not Initialize the Eris Root Directory.\n%s\n", err)
		}
	} else if do.All || do.Yes {
		//goes and does everything
		if err := InitDefaultServices(do); err != nil {
			return fmt.Errorf("Error:\tcould not Instantiate default services.\n%s\n", err)
		}

	} else {
		logger.Printf("Eris Root Directory (%s) already exists.\nContinuing may overwrite files in:\n%s\n%s\nDo you wish to continue? (Y/n): ", common.ErisRoot, common.ServicesPath, common.ActionsPath)

		fmt.Scanln(&input)
		if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
			//goes and does everything
			if err := InitDefaultServices(do); err != nil {
				return fmt.Errorf("Error:\tcould not Instantiate default services.\n%s\n", err)
			}
		} else {
			logger.Printf("Cannot proceed without permission to overwrite. Backup your files and try again")
		}
	}
	if err := InitDefaultServices(do); err != nil {
		return fmt.Errorf("Error:\tcould not Instantiate default services.\n%s\n", err)
	}
	return nil
}

func InitDefaultServices(do *definitions.Do) error {
	logger.Debugf("Adding default files\n")
	//do
	if !do.Pull {
		if err := cloneRepo(do.Services, "eris-services.git", common.ServicesPath); err != nil {
			logger.Errorf("Error cloning default services repository.\n%v\nTrying default defs.\n", err)
			if err2 := dropDefaults(); err2 != nil {
				return fmt.Errorf("Error:\tcannot clone services.\n%v\n%v", err, err2)
			}
		} else {
			if err2 := cloneRepo(do.Actions, "eris-actions.git", common.ActionsPath); err2 != nil {
				return fmt.Errorf("Error:\tcannot clone actions.\n%v", err2)
			}
		}
	} else {
		logger.Printf("Skip pull param given. Complying.\n")
		if err := dropDefaults(); err != nil {
			return err
		}
	}

	logger.Printf("Service defaults written.\n")
	logger.Printf("Action defaults written.\n")
	if err := dropChainDefaults(); err != nil { //moved b/c cleaner stdout for user
		return err
	}
	logger.Printf("Chain defaults written.\n")

	logger.Printf("Initialized eris root directory (%s) with default actions and service files.\n", common.ErisRoot)

	//TODO: when called from cli provide option to go on tour, like `ipfs tour`
	logger.Printf("\nThe marmots have everything set up for you.\nIf you are just getting started please type [eris] to get an overview of the tool.\n")
	return nil
}
