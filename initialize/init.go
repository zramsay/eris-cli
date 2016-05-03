package initialize

import (
	"os"
	"path/filepath"

	log "github.com/eris-ltd/eris-logger"
	"github.com/eris-ltd/eris-cli/definitions"
	. "github.com/eris-ltd/eris-cli/errors"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/common/go/common"
)

func Initialize(do *definitions.Do) error {
	newDir, err := checkThenInitErisRoot(do.Quiet)
	if err != nil {
		return &ErisError{ErrEris, err, ""}
	}

	if !newDir { //new ErisRoot won't have either...can skip
		if err := checkIfCanOverwrite(do.Yes); err != nil {
			return nil // [zr] why nil ?
		}

		log.Info("Checking if migration is required")
		if err := checkIfMigrationRequired(do.Yes); err != nil {
			return nil // [zr] why nil ?
		}

	}

	if do.Pull { //true by default; if imgs already exist, will check for latest anyways
		if err := GetTheImages(do.Yes); err != nil {
			return &ErisError{ErrDocker, err, "increase bandwidth or ensure you can access quay.io"}
		}
	}

	//drops: services, actions, & chain defaults from toadserver
	log.Warn("Initializing default service, action, and chain files")
	if err := InitDefaults(do, newDir); err != nil {
		return &ErisError{ErrEris, BaseError(ErrInitDefaults, err), "might need to hack around a proxy or get an internet connection"}
	}

	if !do.Quiet {
		log.Warn(`
Directory structure initialized:

+-- .eris/
¦   +-- eris.toml
¦   +-- actions/
¦   +-- apps/
¦   +-- bundles/
¦   +-- chains/
¦       +-- account-types
¦       +-- chain-types
¦       +-- default/config.toml
¦       +-- default.toml
¦   +-- keys/
¦       +-- data/
¦       +-- names/
¦   +-- remotes/
¦   +-- scratch/
¦       +-- data/
¦       +-- languages/
¦       +-- lllc/
¦       +-- ser/
¦       +-- sol/
¦   +-- services/
¦       +-- global/
¦       +-- btcd.toml
¦       +-- ipfs.toml
¦       +-- keys.toml

Several more services were also added; see them with:
[eris services ls --known]

Consider running [docker images] to see the images that were added.`)

		log.Warnf(`
Eris sends crash reports to a remote server in case something goes completely
wrong. You may disable this feature by adding the CrashReport = %q
line to the %s definition file.
`, "don't send", filepath.Join(common.ErisRoot, "eris.toml"))

		log.Warn("The marmots have everything set up for you. Type [eris] to get started")
	}
	return nil
}

func InitDefaults(do *definitions.Do, newDir bool) error {
	var srvPath string
	var actPath string
	var chnPath string

	srvPath = common.ServicesPath
	actPath = common.ActionsPath
	chnPath = common.ChainsPath

	if err := dropServiceDefaults(srvPath, do.Source); err != nil {
		return err
	}

	if err := dropActionDefaults(actPath, do.Source); err != nil {
		return err
	}

	if err := dropChainDefaults(chnPath, do.Source); err != nil {
		return err
	}

	log.WithField("root", common.ErisRoot).Warn("Initialized Eris root directory")

	return nil
}

func checkThenInitErisRoot(force bool) (bool, error) {
	var newDir bool
	if force { //for testing only
		log.Info("Force initializing Eris root directory")
		if err := common.InitErisDir(); err != nil {
			return true, BaseError(ErrInitErisRoot, err)
		}
		return true, nil
	}
	if !util.DoesDirExist(common.ErisRoot) {
		log.Warn("Eris root directory doesn't exist. The marmots will initialize it for you")
		if err := common.InitErisDir(); err != nil {
			return true, BaseError(ErrInitErisRoot, err)
		}
		newDir = true
	} else { // ErisRoot exists, prompt for overwrite
		newDir = false
	}
	return newDir, nil
}

func checkIfMigrationRequired(doYes bool) error {
	if err := util.MigrateDeprecatedDirs(common.DirsToMigrate, !doYes); err != nil {
		return BaseError(ErrMigratingDirs, err)
	}
	return nil
}

func checkIfCanOverwrite(doYes bool) error {
	if doYes {
		return nil
	}
	log.WithField("path", common.ErisRoot).Warn("Eris root directory")
	log.WithFields(log.Fields{
		"services path": common.ServicesPath,
		"actions path":  common.ActionsPath,
		"chains path":   common.ChainsPath,
	}).Warn("Continuing may overwrite files in")
	if util.QueryYesOrNo("Do you wish to continue?") == util.Yes {
		log.Debug("Confirmation verified. Proceeding")
	} else {
		log.Warn("The marmots will not proceed without your permission")
		log.Warn("Please backup your files and try again")
		//return &ErisError{404, BaseError(ErrNoPermGiven, err), "try again"}
		return ErrNoPermGiven
	}
	return nil
}

func GetTheImages(doYes bool) error {
	if os.Getenv("ERIS_PULL_APPROVE") == "true" || doYes {
		if err := pullDefaultImages(); err != nil {
			return err
		}
		log.Warn("Successfully pulled default images")
	} else {
		log.Warn(`
WARNING: Approximately 1 gigabyte of Docker images are about to be pulled
onto your host machine. Please ensure that you have sufficient bandwidth to
handle the download. For a remote Docker server this should only take a few
minutes but can sometimes take 10 or more. These times can double or triple
on local host machines. If you already have the images, they'll be updated.
`)
		log.WithField("ERIS_PULL_APPROVE", "true").Warn("Skip confirmation with")
		log.Warn()

		if util.QueryYesOrNo("Do you wish to continue?") == util.Yes {
			if err := pullDefaultImages(); err != nil {
				return err
			}
			log.Warn("Successfully pulled default images")
		}
	}
	return nil
}
