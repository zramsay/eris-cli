package initialize

import (
	"fmt"
	"os"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func Initialize(do *definitions.Do) error {
	if _, err := os.Stat(common.ErisRoot); err != nil {
		if os.IsNotExist(err) {
			log.Info("Eris root directory does not exist. The marmots will initialize this directory for you")
			if err := common.InitErisDir(); err != nil {
				return fmt.Errorf("Error:\tcould not Initialize the Eris Root Directory.\n%s\n", err)
			}
		} else {
			panic(err)
		}
	}

	if do.Yes {
		log.Debug("Not requiring input. Proceeding")
	} else {
		var input string
		log.WithField("path", common.ErisRoot).Warn("Eris root directory already exists")
		log.WithFields(log.Fields{
			"services path": common.ServicesPath,
			"actions path":  common.ActionsPath,
		}).Warn("Continuing may overwrite files in")
		fmt.Print("Do you wish to continue? (y/n): ")
		if _, err := fmt.Scanln(&input); err != nil {
			return fmt.Errorf("Error reading from stdin: %v\n", err)
		}
		if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
			log.Debug("Confirmation verified. Proceeding")
		} else {
			log.Warn("The marmots will not proceed without your permission to overwrite")
			log.Warn("Please backup your files and try again")
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
	log.Debug("Adding default files")

	if !do.Pull {
		if err := cloneRepo(do.Services, "eris-services.git", common.ServicesPath); err != nil {
			log.Errorf("Error cloning default services repository: %v", err)
			log.Error("Trying defaults")
		}
		if err2 := dropDefaults(); err2 != nil {
			return fmt.Errorf("Error:\tcannot clone services.\nError:\tcannot drop default services.\n%v", err2)
		}
		if err3 := cloneRepo(do.Actions, "eris-actions.git", common.ActionsPath); err3 != nil {
			return fmt.Errorf("Error:\tcannot clone actions.\n%v", err3)
		}
	} else {
		log.Warn("Skip pull param given. Complying")
		if err := dropDefaults(); err != nil {
			return err
		}
	}

	if err := dropChainDefaults(); err != nil { //moved b/c cleaner stdout for user
		return err
	}

	log.WithField("root", common.ErisRoot).Info("Initialized eris root directory with default action, service, and chain files")

	// TODO: when called from cli provide option to go on tour, like `ipfs tour`
	log.Warn("The marmots have everything set up for you")
	log.Warn("If you are just getting started, please type [eris] to get an overview of the tool")
	return nil
}
