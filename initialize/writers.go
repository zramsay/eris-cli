package initialize

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/util"

	ver "github.com/eris-ltd/eris-cli/version"

	"github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
)

// XXX all files in this sequence must be added to both
// the respective GH repo & mindy testnet (pinkpenguin.interblock.io:46657/list_names)
func dropServiceDefaults(dir string, services []string) error {
	if len(services) == 0 {
		services = ver.SERVICE_DEFINITIONS
	}

	for _, service := range services {
		var err error

		switch service {
		case "keys":
			err = writeDefaultFile(common.ServicesPath, "keys.toml", defServiceKeys)
		case "ipfs":
			err = writeDefaultFile(common.ServicesPath, "ipfs.toml", defServiceIPFS)
		case "compilers":
			err = writeDefaultFile(common.ServicesPath, "compilers.toml", defServiceCompilers)
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
