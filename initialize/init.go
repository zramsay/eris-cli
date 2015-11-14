package initialize

import (
	"fmt"
	"os"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func Initialize(do *definitions.Do) error {
	logger.Printf("The marmots have connected to Docker successfully.\nThey will now will install a few default services and actions for your use.\n\n")

	if _, err := os.Stat(common.ErisRoot); os.IsNotExist(err) {
		logger.Printf("Eris Root Directory does not exist. The marmots will initialize this directory for you.\n")
		if err := common.InitErisDir(); err != nil {
			return fmt.Errorf("Error:\tcould not Initialize the Eris Root Directory.\n%s\n", err)
		}
	} else if do.Yes {
		logger.Debugf("Not requiring input. Proceeding.\n")
	} else {
		var input string
		logger.Printf("Eris Root Directory (%s) already exists.\nContinuing may overwrite files in:\n%s\n%s\nDo you wish to continue? (y/n): ", common.ErisRoot, common.ServicesPath, common.ActionsPath)
		fmt.Scanln(&input)
		if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
			logger.Debugf("Confirmation verified. Proceeding.\n")
		} else {
			logger.Printf("\nThe marmots will not proceed without your permission to overwrite.\nPlease backup your files and try again.\n")
			return fmt.Errorf("Error:\tno permission given to overwrite services and actions.\n")
		}
	}

	//goes and does everything
	if err := InitDefaultServices(do); err != nil {
		return fmt.Errorf("Error:\tcould not Instantiate default services.\n%s\n", err)
	}

	var prompt bool
	if do.Yes || os.Getenv("ERIS_MIGRATE_APPROVE") == "true" {
		prompt = false
	} else {
		prompt = true
	}
	if err := util.MigrateDeprecatedDirs(common.DirsToMigrate, prompt); err != nil {
		return fmt.Errorf("Error:\tcould not migrate directories.\n%s\n", err)
	}
	return nil
}

func InitDefaultServices(do *definitions.Do) error {
	logger.Debugf("Adding default files\n")

	//do
	if !do.Pull {
		if err := cloneRepo(do.Services, "eris-services.git", common.ServicesPath); err != nil {
			logger.Errorf("Error cloning default services repository.\n%v\nTrying default defs.\n", err)
		}
		if err2 := dropDefaults(); err2 != nil {
			return fmt.Errorf("Error:\tcannot clone services.\nError:\tcannot drop default services.\n%v", err2)
		}
		if err3 := cloneRepo(do.Actions, "eris-actions.git", common.ActionsPath); err3 != nil {
			return fmt.Errorf("Error:\tcannot clone actions.\n%v", err3)
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
