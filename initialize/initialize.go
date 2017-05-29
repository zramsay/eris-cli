package initialize

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/util"
	"github.com/monax/monax/version"
)

// The entrypoint for [monax init]
// - Is required to be run after every version upgrade
// - Writes some default service & chain definition files
// - Pulls the required docker images
// - Write the ~/.monax/monax.toml file.
func Initialize(do *definitions.Do) error {
	newDir, err := checkThenInitMonaxRoot()
	if err != nil {
		return err
	}

	// If this is the first installation of monax, skip these checks.
	if !newDir {
		// Early exit if overwriting is not allowed
		if err := checkIfCanOverwrite(do.Yes); err != nil {
			return err
		}

		log.Info("Checking if migration is required")
		if err := checkIfMigrationRequired(do.Yes); err != nil {
			return err
		}
	}

	// Either the monax root dir exists, but we are allowed to overwrite, or this
	// dir does not and this is a fresh init.
	// Write the ~/.monax/monax.toml file.
	if err := config.Save(&config.Global.Settings); err != nil {
		return err
	}

	// Pull the default docker images (wraps [docker pull]).
	// Can and will overwrite versions. Not usually a problem.
	if do.Pull {
		if err := getTheImages(do); err != nil {
			return err
		}
	}

	// Write the default files to services/ and chains/account-&-chain-types/ subdirectories.
	if err := initDefaultFiles(); err != nil {
		return fmt.Errorf("Could not instantiate defaults:\n\n%s", err)
	}

	// TODO: [Silas] um, perhaps we should actually ask the filesystem what we've
	// _really_ done rather than hard coding the output of tree...
	log.Warn(`
Directory structure initialized:

+-- .monax/
¦   +-- monax.toml
¦   +-- apps/
¦   +-- bundles/
¦   +-- chains/
¦       +-- account-types/
¦       +-- chain-types/
¦   +-- keys/
¦       +-- data/
¦       +-- names/
¦   +-- scratch/
¦       +-- data/
¦       +-- languages/
¦   +-- services/
¦       +-- keys.toml
¦       +-- compilers.toml
¦       +-- logrotate.toml

Consider running [docker images] to see the images that were added.`)

	log.Warn("The marmots have everything set up for you. Type [monax] to get started")
	return nil
}

func initDefaultFiles() error {

	for _, serviceName := range ServiceDefinitions {
		serviceDefinition := defaultServices(serviceName)
		if err := WriteServiceDefinitionFile(serviceName, serviceDefinition); err != nil {
			return err
		}
	}

	for _, accountType := range AccountTypeDefinitions {
		accountDefinition := defaultAccountTypes(accountType)
		if err := writeAccountTypeDefinitionFile(accountType, accountDefinition); err != nil {
			return err
		}
	}

	for _, chainType := range ChainTypeDefinitions {
		chainDefinition := defaultChainTypes(chainType)
		if err := writeChainTypeDefinitionFile(chainType, chainDefinition); err != nil {
			return err
		}
	}

	return nil
}

func pullDefaultImages(images []string) error {
	// Default images.
	if len(images) == 0 {
		images = []string{
			"data",
			"keys",
			"db",
			"compilers",
		}
	}

	// Rewrite with versioned image names (full names
	// without a registry prefix).
	versionedImageNames := map[string]string{
		"data":      version.ImageData,
		"keys":      version.ImageKeys,
		"db":        version.ImageDB,
		"compilers": version.ImageCompilers,
	}

	for i, image := range images {
		images[i] = versionedImageNames[image]

		// Attach default registry prefix.
		if !strings.HasPrefix(images[i], version.DefaultRegistry) {
			images[i] = path.Join(version.DefaultRegistry, images[i])
		}
	}

	// Spacer.
	log.Warn()

	log.Warn("Pulling default Docker images from " + version.DefaultRegistry)
	for i, image := range images {
		log.WithField("image", image).Warnf("Pulling image %d out of %d", i+1, len(images))

		if err := util.PullImage(image, os.Stdout); err != nil {
			if err == util.ErrImagePullTimeout {
				return fmt.Errorf(`
It looks like marmots are taking too long to download the necessary images...
Please, try restarting the [monax init] command one more time now or a bit later.
This is likely a network performance issue with our Docker hosting provider`)
			}
			return err
		}
	}
	return nil
}

func checkThenInitMonaxRoot() (bool, error) {
	var newDir bool

	if !util.DoesDirExist(config.MonaxRoot) || !util.DoesDirExist(config.ServicesPath) {
		log.Warn("Monax root directory doesn't exist. The marmots will initialize it for you")
		if err := config.InitMonaxDir(); err != nil {
			return true, fmt.Errorf("Could not initialize Monax root directory: %v", err)

		}
		newDir = true
	} else {
		newDir = false
	}
	return newDir, nil
}

func checkIfMigrationRequired(doYes bool) error {
	if err := util.MigrateDeprecatedDirs(config.DirsToMigrate, !doYes); err != nil {
		return fmt.Errorf("Could not migrate directories.\n%s", err)
	}
	return nil
}

func checkIfCanOverwrite(doYes bool) error {
	if doYes {
		return nil
	}
	log.WithField("path", config.MonaxRoot).Warn("Monax root directory")
	log.WithFields(log.Fields{
		"services path": config.ServicesPath,
	}).Warn("Continuing may overwrite files in")
	if util.QueryYesOrNo("Do you wish to continue?") == util.Yes {
		log.Debug("Confirmation verified. Proceeding")
	} else {
		log.Warn("The marmots will not proceed without your permission")
		log.Warn("Please backup your files and try again")
		return fmt.Errorf("Error: no permission given to overwrite services")
	}
	return nil
}

func getTheImages(do *definitions.Do) error {
	if os.Getenv("MONAX_PULL_APPROVE") == "true" || do.Yes {
		if err := pullDefaultImages(do.ImagesSlice); err != nil {
			return err
		}
		log.Warn("Successfully pulled default images")
	} else {
		log.Warn(`
WARNING: Approximately 200 mb of Docker images are about to be pulled
onto your host machine. Please ensure that you have sufficient bandwidth to
handle the download. For a remote Docker server this should only take a few
minutes but can sometimes take 10 or more. These times can double or triple
on local host machines. If you already have the images, they'll be updated.
`)
		log.WithField("MONAX_PULL_APPROVE", "true").Warn("Skip confirmation with")
		log.Warn()

		if util.QueryYesOrNo("Do you wish to continue?") == util.Yes {
			if err := pullDefaultImages(do.ImagesSlice); err != nil {
				return err
			}
			log.Warn("Successfully pulled default images")
		}
	}
	return nil
}
