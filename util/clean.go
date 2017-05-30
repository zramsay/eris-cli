package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"

	docker "github.com/fsouza/go-dockerclient"
)

func Clean(toClean map[string]bool) error {
	if toClean["yes"] {
		if err := cleanHandler(toClean); err != nil {
			return err
		}
	} else {
		if canWeRemove(toClean) {
			if err := cleanHandler(toClean); err != nil {
				return err
			}
		} else {
			return nil
		}
	}
	return nil
}

func cleanHandler(toClean map[string]bool) error {
	if toClean["containers"] {
		log.Debug("Removing all monax containers")
		if err := RemoveAllMonaxContainers(); err != nil {
			return err
		}
	}

	if toClean["chains"] {
		log.Debug("Removing latent chains data directories")
		if err := cleanLatentChainData(); err != nil {
			return err
		}
	}

	if toClean["scratch"] {
		log.Debug("Removing containers' scratch data")
		if err := cleanScratchData(); err != nil {
			return err
		}
	}

	if toClean["root"] {
		log.Debug("Removing Monax root directory")
		if err := os.RemoveAll(config.MonaxRoot); err != nil {
			return err
		}
	}

	if toClean["images"] {
		log.Debug("Removing all Monax Docker images")
		if err := RemoveMonaxImages(); err != nil {
			return err
		}
	}
	return nil
}

// stops and removes containers and their volumes
func RemoveAllMonaxContainers() error {
	contns, err := DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return fmt.Errorf("Error listing containers: %v", DockerError(err))
	}

	for _, container := range contns {
		// [pv]: Make sure legacy data containers are removed as well.
		// The prefix bit is to be removed in 0.12.
		if container.Labels[definitions.LabelMonax] == "true" ||
			strings.HasPrefix(strings.TrimLeft(container.Names[0], "/"), "monax_") {

			if err := removeContainer(container.ID); err != nil {
				return fmt.Errorf("Error removing container: %v", DockerError(err))
			}
		}
	}

	return nil
}

func removeContainer(containerID string) error {
	removeOpts := docker.RemoveContainerOptions{
		ID:            containerID,
		RemoveVolumes: true,
		Force:         true,
	}

	if err := DockerClient.RemoveContainer(removeOpts); err != nil {
		// In 1.10.1 there is a weird EOF error which occurs here
		// even though the container is removed. ignoring that.
		if fmt.Sprintf("%v", err) == "EOF" {
			log.Debug("Weird EOF error. Not reaping")
			return nil
		}
		return err
	}
	return nil
}

func cleanLatentChainData() error {
	// get everything in ~/.monax/chains
	files, err := ioutil.ReadDir(config.ChainsPath)
	if err != nil {
		return err
	}

	// leave these files/dirs alone
	dontDelete := map[string]bool{
		"account-types": true,
		"chain-types":   true,
		"HEAD":          true,
	}

	// remove everything else
	for _, f := range files {
		if !dontDelete[f.Name()] {
			if err := os.RemoveAll(filepath.Join(config.ChainsPath, f.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}

func cleanScratchData() error {
	if err := os.RemoveAll(config.DataContainersPath); err != nil {
		return err
	}
	if err := os.Mkdir(config.DataContainersPath, 0777); err != nil {
		return err
	}
	return nil
}

func RemoveMonaxImages() error {
	images, err := DockerClient.ListImages(docker.ListImagesOptions{All: true})
	if err != nil {
		return DockerError(err)
	}

	for _, i := range images {
		if len(i.RepoTags) == 0 {
			continue
		}

		if !strings.Contains(i.RepoTags[0], "monax/") {
			continue
		}
		log.WithFields(log.Fields{
			"image": i.RepoTags[0],
		}).Debug("Removing image")
		if err := DockerClient.RemoveImageExtended(i.ID, docker.RemoveImageOptions{Force: true, NoPrune: true}); err != nil {
			return DockerError(err)
		}
	}

	return nil
}

func canWeRemove(toClean map[string]bool) bool {
	var toWarn = map[string]string{
		"containers": "all",
		"chains":     fmt.Sprintf("%s/.monax/chains", config.HomeDir()),
		"scratch":    fmt.Sprintf("%s/.monax/scratch/data", config.HomeDir()),
		"root":       fmt.Sprintf("%s/.monax", config.HomeDir()),
		"images":     "all",
	}

	if !toClean["all"] {
		log.Warn("The marmots are about to remove the following")
		if toClean["containers"] {
			log.WithField("containers", toWarn["containers"]).Warn()
		}
		if toClean["chains"] {
			log.WithField("chains", toWarn["chains"]).Warn()
		}
		if toClean["scratch"] {
			log.WithField("scratch", toWarn["scratch"]).Warn()
		}
		if toClean["root"] {
			log.WithField("root", toWarn["root"]).Warn()
		}
		if toClean["images"] {
			log.WithField("images", toWarn["images"]).Warn()
		}
	} else {
		log.WithFields(log.Fields{
			"containers": toWarn["containers"],
			"chains":     toWarn["chains"],
			"scratch":    toWarn["scratch"],
			"root":       toWarn["root"],
			"images":     toWarn["images"],
		}).Warn("The marmots are about to remove the following")
	}

	if QueryYesOrNo("Please confirm") == Yes {
		log.Warn("Authorization given, removing")
		return true
	}
	log.Warn("Authorization not given, exiting")
	return false
}
