package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/eris-ltd/common/go/common"
	def "github.com/eris-ltd/eris-cli/definitions"

	log "github.com/eris-ltd/eris-logger"
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
		log.Debug("Removing all eris containers")
		if err := RemoveAllErisContainers(); err != nil {
			return err
		}
	}

	if toClean["chn-dirs"] {
		log.Debug("Removing latent chains data in ChainsPath")
		if err := cleanLatentChainData(); err != nil {
			return err
		}
	}

	if toClean["scratch"] {
		log.Debug("Removing contents of DataContainersPath")
		if err := cleanScratchData(); err != nil {
			return err
		}
	}

	if toClean["root"] {
		log.Debug("Removing Eris root directory")
		if err := os.RemoveAll(common.ErisRoot); err != nil {
			return err
		}
	}

	if toClean["images"] {
		log.Debug("Removing all Eris Docker images")
		if err := RemoveErisImages(); err != nil {
			return err
		}
	}
	return nil
}

// stops and removes containers and their volumes
func RemoveAllErisContainers() error {
	contns, err := DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return fmt.Errorf("Error listing containers: %v", DockerError(err))
	}

	for _, container := range contns {
		// [pv]: Make sure legacy data containers are removed as well.
		// The prefix bit is to be removed in 0.12.
		if container.Labels[def.LabelEris] == "true" ||
			strings.HasPrefix(strings.TrimLeft(container.Names[0], "/"), "eris_") {

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
	// get everything in ~/.eris/chains
	files, err := ioutil.ReadDir(common.ChainsPath)
	if err != nil {
		return err
	}

	// leave these files/dirs alone
	dontDelete := map[string]bool{
		"account-types": true,
		"chain-types":   true,
		"default":       true,
		"default.toml":  true,
		"HEAD":          true,
	}

	// remove everything else
	for _, f := range files {
		if !dontDelete[f.Name()] {
			if err := os.RemoveAll(filepath.Join(common.ChainsPath, f.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}

func cleanScratchData() error {
	if err := os.RemoveAll(common.DataContainersPath); err != nil {
		return err
	}
	if err := os.Mkdir(common.DataContainersPath, 0777); err != nil {
		return err
	}
	return nil
}

func RemoveErisImages() error {
	images, err := DockerClient.ListImages(docker.ListImagesOptions{All: true})
	if err != nil {
		return DockerError(err)
	}

	for _, i := range images {
		if !strings.Contains(i.RepoTags[0], "eris/") {
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
	var home string
	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
	} else {
		home = os.Getenv("HOME")
	}
	var toWarn = map[string]string{
		"containers": "all",
		"chn-dirs":   "latent files & dirs from ~/.eris/chains",
		"scratch":    fmt.Sprintf("%s/.eris/scratch/data", home),
		"root":       fmt.Sprintf("%s/.eris", home),
		"images":     "all",
	}

	if toClean["all"] != true {
		log.Warn("The marmots are about to remove the following")
		if toClean["containers"] {
			log.WithField("containers", toWarn["containers"]).Warn("")
		}
		if toClean["chn-dirs"] {
			log.WithField("chn-dirs", toWarn["chn-dirs"]).Warn("")
		}
		if toClean["scratch"] {
			log.WithField("scratch", toWarn["scratch"]).Warn("")
		}
		if toClean["root"] {
			log.WithField("root", toWarn["root"]).Warn("")
		}
		if toClean["images"] {
			log.WithField("images", toWarn["images"]).Warn("")
		}
	} else {
		log.WithFields(log.Fields{
			"containers": toWarn["containers"],
			"chn-dirs":   toWarn["chn-dirs"],
			"scratch":    toWarn["scratch"],
			"root":       toWarn["root"],
			"images":     toWarn["images"],
		}).Warn("The marmots are about to remove the following")
	}

	if common.QueryYesOrNo("Please confirm") == common.Yes {
		log.Warn("Authorization given, removing")
		return true
	}
	log.Warn("Authorization not given, exiting")
	return false
}

func TrimString(strang string) string {
	return strings.TrimSpace(strings.Trim(strang, "\n"))
}
