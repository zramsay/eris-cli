package initialize

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"
)

func Initialize(do *definitions.Do) error {
	newDir, err := checkThenInitErisRoot(do.Quiet)
	if err != nil {
		return err
	}

	if err := overwriteErisToml(); err != nil {
		return err
	}

	if !newDir {
		if err := checkIfCanOverwrite(do.Yes); err != nil {
			return nil
		}

		log.Info("Checking if migration is required")
		if err := checkIfMigrationRequired(do.Yes); err != nil {
			return nil
		}

	}

	if do.Pull {
		if err := GetTheImages(do); err != nil {
			return err
		}
	}

	// Service definition defaults.
	log.Warn("Initializing default service definition files")
	if err := InitDefaults(do, newDir); err != nil {
		return fmt.Errorf("Error:\tcould not instantiate default services.\n%s\n", err)
	}

	if !do.Quiet {
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
`, "don't send", filepath.Join(config.ErisRoot, "eris.toml"))

		log.Warn("The marmots have everything set up for you. Type [eris] to get started")
	}
	return nil
}

func InitDefaults(do *definitions.Do, newDir bool) error {
	var srvPath string

	srvPath = config.ServicesPath

	if err := dropServiceDefaults(srvPath, do.ServicesSlice); err != nil {
		return err
	}

	log.WithField("root", config.ErisRoot).Warn("Initialized Eris root directory")

	return nil
}

// XXX all files in this sequence must be added to both
// the respective GH repo & mindy testnet (pinkpenguin.interblock.io:46657/list_names)
func dropServiceDefaults(dir string, services []string) error {
	if len(services) == 0 {
		services = version.SERVICE_DEFINITIONS
	}

	for _, service := range services {
		var err error

		switch service {
		case "keys":
			err = writeDefaultFile(config.ServicesPath, "keys.toml", defServiceKeys)
		case "ipfs":
			err = writeDefaultFile(config.ServicesPath, "ipfs.toml", defServiceIPFS)
		case "compilers":
			err = writeDefaultFile(config.ServicesPath, "compilers.toml", defServiceCompilers)
		default:
			err = drops([]string{service}, "services", dir)
		}
		if err != nil {
			return fmt.Errorf("Cannot add default %s: %v", service, err)
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
			"cm",
			"pm",
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
		"cm":        config.Global.ImageCM,
		"pm":        config.Global.ImagePM,
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

func drops(files []string, typ, dir string) error {
	//to get from github
	var repo string
	if typ == "services" {
		repo = "eris-services"
	} else if typ == "chains" {
		repo = "eris-chains"
	}
	// on different arch
	archPrefix := ""
	if runtime.GOARCH == "arm" {
		archPrefix = "arm/"
	}

	if !util.DoesDirExist(dir) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}

	for _, file := range files {
		log.WithField(file, dir).Debug("Getting file from GitHub, dropping into")
		if err := util.GetFromGithub("eris-ltd", repo, "master", archPrefix+file+".toml", dir, file+".toml"); err != nil {
			return err
		}
	}
	return nil
}

// TODO eventually eliminate this.
func writeDefaultFile(savePath, fileName string, toWrite func() string) error {
	if err := os.MkdirAll(savePath, 0777); err != nil {
		return err
	}
	pth := filepath.Join(savePath, fileName)
	writer, err := os.Create(pth)
	defer writer.Close()
	if err != nil {
		return err
	}
	writer.Write([]byte(toWrite()))
	return nil
}

func checkThenInitErisRoot(force bool) (bool, error) {
	var newDir bool
	if force {
		log.Info("Force initializing Eris root directory")
		if err := config.InitErisDir(); err != nil {
			return true, fmt.Errorf("Could not initialize Eris root directory: %v", err)
		}
		return true, nil
	}
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

//func askToPull removed since it's basically a duplicate of this
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

func GetTheImages(do *definitions.Do) error {
	if os.Getenv("ERIS_PULL_APPROVE") == "true" || do.Yes {
		if err := pullDefaultImages(do.ImagesSlice); err != nil {
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
	config.Global.ImagePM = version.ImagePM
	config.Global.ImageCM = version.ImageCM
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
