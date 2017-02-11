package initialize

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris/config"
	"github.com/eris-ltd/eris/definitions"
	"github.com/eris-ltd/eris/log"
	"github.com/eris-ltd/eris/util"
	"github.com/eris-ltd/eris/version"
)

// The entrypoint for [eris init]
// - Is required to be run after every version upgrade
// - Writes some default service & chain definition files
// - Pulls the required docker images
// - Overwrites the eris.toml config file
func Initialize(do *definitions.Do) error {

	// Create the directory structure.
	newDir, err := checkThenInitErisRoot()
	if err != nil {
		return err
	}

	// Overwrite the ~/.eris/eris.toml configuration file.
	if err := overwriteErisToml(); err != nil {
		return err
	}

	// If this is the first installation of eris, skip these checks.
	if !newDir {
		if err := checkIfCanOverwrite(do.Yes); err != nil {
			return err
		}

		log.Info("Checking if migration is required")
		if err := checkIfMigrationRequired(do.Yes); err != nil {
			return err
		}
	}

	// Pull the default docker images (wraps [docker pull]).
	if do.Pull {
		if err := getTheImages(do); err != nil {
			return err
		}
	}

	if err := initDefaultFiles(); err != nil {
		return fmt.Errorf("Could not instantiate defaults:\n\n%s", err)
	}

	log.Warn(`
Directory structure initialized:

+-- .eris/
¦   +-- eris.toml
¦   +-- apps/
¦   +-- bundles/
¦   +-- chains/
¦       +-- account-types/
¦       +-- chain-types/
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
¦       +-- keys.toml
¦       +-- ipfs.toml
¦       +-- compilers.toml
¦       +-- logrotate.toml

Consider running [docker images] to see the images that were added.`)

	log.Warnf(`
Eris sends crash reports to a remote server in case something goes completely
wrong. You may disable this feature by adding the CrashReport = %q
line to the %s definition file.
`, "don't send", filepath.Join(config.ErisRoot, "eris.toml"))

	log.Warn("The marmots have everything set up for you. Type [eris] to get started")
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
			"ipfs",
			"db",
			"compilers",
		}
	}

	// Rewrite with versioned image names (full names
	// without a registry prefix).
	versionedImageNames := map[string]string{
		"data":      config.Global.ImageData,
		"keys":      config.Global.ImageKeys,
		"ipfs":      config.Global.ImageIPFS,
		"db":        config.Global.ImageDB,
		"compilers": config.Global.ImageCompilers,
	}

	for i, image := range images {
		images[i] = versionedImageNames[image]

		// Attach default registry prefix.
		if !strings.HasPrefix(images[i], config.Global.DefaultRegistry) {
			images[i] = path.Join(config.Global.DefaultRegistry, images[i])
		}
	}

	// Spacer.
	log.Warn()

	log.Warn("Pulling default Docker images from " + config.Global.DefaultRegistry)
	for i, image := range images {
		log.WithField("image", image).Warnf("Pulling image %d out of %d", i+1, len(images))

		if err := util.PullImage(image, os.Stdout); err != nil {
			if err == util.ErrImagePullTimeout {
				return fmt.Errorf(`
It looks like marmots are taking too long to download the necessary images...
Please, try restarting the [eris init] command one more time now or a bit later.
This is likely a network performance issue with our Docker hosting provider`)
			}
			return err
		}
	}
	return nil
}

func checkThenInitErisRoot() (bool, error) {
	var newDir bool

	if !util.DoesDirExist(config.ErisRoot) || !util.DoesDirExist(config.ServicesPath) {
		log.Warn("Eris root directory doesn't exist. The marmots will initialize it for you")
		if err := config.InitErisDir(); err != nil {
			return true, fmt.Errorf("Could not initialize Eris root directory: %v", err)
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
	log.WithField("path", config.ErisRoot).Warn("Eris root directory")
	log.WithFields(log.Fields{
		"services path": config.ServicesPath,
		"chains path":   config.ChainsPath,
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
	if os.Getenv("ERIS_PULL_APPROVE") == "true" || do.Yes {
		if err := pullDefaultImages(do.ImagesSlice); err != nil {
			return err
		}
		log.Warn("Successfully pulled default images")
	} else {
		log.Warn(`
WARNING: Approximately 400 mb of Docker images are about to be pulled
onto your host machine. Please ensure that you have sufficient bandwidth to
handle the download. For a remote Docker server this should only take a few
minutes but can sometimes take 10 or more. These times can double or triple
on local host machines. If you already have the images, they'll be updated.
`)
		log.WithField("ERIS_PULL_APPROVE", "true").Warn("Skip confirmation with")
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

func overwriteErisToml() error {
	config.Global.DefaultRegistry = version.DefaultRegistry
	config.Global.BackupRegistry = version.BackupRegistry
	config.Global.ImageData = version.ImageData
	config.Global.ImageKeys = version.ImageKeys
	config.Global.ImageDB = version.ImageDB
	config.Global.ImageIPFS = version.ImageIPFS

	// Ensure the directory the file being saved to exists.
	if err := os.MkdirAll(config.ErisRoot, 0755); err != nil {
		return err
	}

	if err := config.Save(&config.Global.Settings); err != nil {
		return err
	}
	return nil
}
