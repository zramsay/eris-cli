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

	log.Warn("Checking for eris root directory")
	//do.Quiet forces a new dir, only used for testing
	newDir, err := checkThenInitErisRoot(do.Quiet)
	if err != nil {
		return err
	}

	if !newDir { //new ErisRoot won't have either...can skip
		if err := checkIfCanOverwrite(do.Yes); err != nil {
			return err
		}

		log.Warn("Checking if migration is required")
		if err := checkIfMigrationRequired(do.Yes); err != nil {
			return err
		}

	}

	if do.Pull { //true by default; if imgs already exist, will check for latest anyways
		if err := GetTheImages(do.Yes); err != nil {
			return err
		}
	}

	//drops: services, actions, & chain defaults from toadserver
	log.Warn("Initializing defaults")
	if err := InitDefaults(do, newDir); err != nil {
		return fmt.Errorf("Error:\tcould not instantiate default services.\n%s\n", err)
	}

	//TODO: when called from cli provide option to go on tour, like `ipfs tour`
	//[zr] this'll be cleaner with `make`
	log.Warn("The marmots have everything set up for you")
	log.Warn("If you are just getting started please type [eris] to get an overview of the tool")

	return nil
}

func InitDefaults(do *definitions.Do, newDir bool) error {
	var srvPath string
	var actPath string
	var chnPath string

	srvPath = common.ServicesPath
	actPath = common.ActionsPath
	chnPath = common.ChainsPath

	tsErrorFix := "toadserver may be down: re-run with `--source=rawgit`"

	if err := dropServiceDefaults(srvPath, do.Source); err != nil {
		return fmt.Errorf("%v\n%s\n", err, tsErrorFix)
	}

	if err := dropActionDefaults(actPath, do.Source); err != nil {
		return fmt.Errorf("%v\n%s\n", err, tsErrorFix)
	}

	if err := dropChainDefaults(chnPath, do.Source); err != nil {
		return fmt.Errorf("%v\n%s\n", err, tsErrorFix)
	}

	log.WithField("root", common.ErisRoot).Warn("Initialized eris root directory with default service, action, and chain files")

	return nil
}

func checkThenInitErisRoot(force bool) (bool, error) {
	var newDir bool
	if force { //for testing only
		log.Warn("Force initializing eris root directory")
		if err := common.InitErisDir(); err != nil {
			return true, fmt.Errorf("Error:\tcould not initialize the eris root directory.\n%s\n", err)
		}
		return true, nil
	}
	if !util.DoesDirExist(common.ErisRoot) {
		log.Warn("Eris root directory does not exist. The marmots will initialize this directory for you")
		if err := common.InitErisDir(); err != nil {
			return true, fmt.Errorf("Error:\tcould not initialize the eris root directory.\n%s\n", err)
		}
		newDir = true
	} else { // ErisRoot exists, prompt for overwrite
		newDir = false
	}
	return newDir, nil
}

func checkIfMigrationRequired(doYes bool) error {
	if err := util.MigrateDeprecatedDirs(common.DirsToMigrate, !doYes); err != nil {
		return fmt.Errorf("Error:\tcould not migrate directories.\n%s\n", err)
	}
	return nil
}

//func askToPull removed since it's basically a duplicate of this
func checkIfCanOverwrite(doYes bool) error {
	if doYes {
		return nil
	}
	var input string
	log.WithField("path", common.ErisRoot).Warn("Eris root directory already exists")
	log.WithFields(log.Fields{
		"services path": common.ServicesPath,
		"actions path":  common.ActionsPath,
		"chains path":   common.ChainsPath,
	}).Warn("Continuing may overwrite files in:")
	fmt.Print("Do you wish to continue? (y/n): ")
	if _, err := fmt.Scanln(&input); err != nil {
		return fmt.Errorf("Error reading from stdin: %v\n", err)
	}
	if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
		log.Debug("Confirmation verified. Proceeding")
	} else {
		log.Warn("The marmots will not proceed without your permission to overwrite")
		log.Warn("Please backup your files and try again")
		return fmt.Errorf("Error:\tno permission given to overwrite services and actions\n")
	}
	return nil
}

func GetTheImages(doYes bool) error {
	if os.Getenv("ERIS_PULL_APPROVE") == "true" || doYes {
		if err := pullDefaultImages(); err != nil {
			return err
		}
		log.Warn("Pulling of default images successful")
	} else {
		var input string
		log.Warn(`WARNING: Approximately 1 gigabyte of docker images are about to be pulled onto your host machine
Please ensure that you have sufficient bandwidth to handle the download
On a remote host in the cloud, this should only take a few minutes but can sometimes take 10 or more.
These times can double or triple on local host machines
If you already have these images, they will be updated`)

		log.WithField("ERIS_PULL_APPROVE", "true").Warn("To avoid this warning on all future pulls, set as an environment variable")

		fmt.Print("Do you wish to continue? (y/n): ")
		if _, err := fmt.Scanln(&input); err != nil {
			return fmt.Errorf("Error reading from stdin: %v\n", err)
		}
		if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
			if err := pullDefaultImages(); err != nil {
				return err
			}
			log.Warn("Pulling of default images successful")
		}
	}
	return nil
}
