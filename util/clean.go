package util

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
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

	if toClean["scratch"] {
		log.Debug("Removing contents of DataContainersPath")
		if err := cleanScratchData(); err != nil {
			return err
		}
	}

	if toClean["rmd"] {
		log.Debug("Removing Eris Root Directory")
		if err := os.RemoveAll(common.ErisRoot); err != nil {
			return err
		}
	}

	if toClean["images"] {
		log.Debug("Removing all Eris docker images")
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
		return fmt.Errorf("error listing containers: %v\n", err)
	}

	for _, container := range contns {
		if container.Labels["eris:ERIS"] == "true" {
			if err := removeContainer(container.ID); err != nil {
				return fmt.Errorf("error removing container: %v\n", err)
			}
		}
	}

	// above doesn't catch them all ... ?
	dataConts := DataContainers()
	if len(dataConts) != 0 {
		log.Warn("dangling data containers found, removing them")
		for _, dCont := range dataConts {
			if err := removeContainer(dCont.ContainerID); err != nil {
				return fmt.Errorf("error removing container: %v\n", err)
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
		// in 1.10.1 there is a weird EOF error which occurs here even though the container is removed. ignoring that.
		if fmt.Sprintf("%v", err) == "EOF" {
			log.Debug("Weird EOF error. Not reaping.")
			return nil
		}
		return err
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
	opts := docker.ListImagesOptions{
		All:     true,
		Filters: nil,
		Digests: false,
	}
	allTheImages, err := DockerClient.ListImages(opts)
	if err != nil {
		return err
	}

	//get all repo tags & IDs
	repoTags := make(map[int][]string)
	imageIDs := make(map[int]string)
	for i, image := range allTheImages {
		repoTags[i] = image.RepoTags
		imageIDs[i] = image.ID
	}

	erisImages := []string{}
	erisImageIDs := []string{}

	//searches through repo tags for eris images & "maps" to ID
	for i, repoTag := range repoTags {
		for _, rt := range repoTag {
			r, err := regexp.Compile(`eris`)
			if err != nil {
				log.Errorf("Regexp error: %v", err)
			}

			if r.MatchString(rt) == true {
				erisImages = append(erisImages, rt)
				erisImageIDs = append(erisImageIDs, imageIDs[i])
			}
		}
	}

	for i, imageID := range erisImageIDs {
		log.WithFields(log.Fields{
			"=>": erisImages[i],
			"id": imageID,
		}).Debug("Removing image")
		if err := DockerClient.RemoveImage(imageID); err != nil {
			return err
		}
	}
	return nil
}

func canWeRemove(toClean map[string]bool) bool {
	var input string
	var canWe bool

	var toWarn = map[string]string{
		"containers": "All existing Eris containers",
		"scratch":    "The contents of ($HOME/eris/scratch/data)",
		"rmd":        "The Eris Root Directory ($HOME/.eris)",
		"images":     "All Eris docker images",
	}

	if toClean["all"] != true {
		log.Warn("The marmots are about to remove")
		if toClean["containers"] {
			log.WithField("containers", toWarn["containers"]).Warn("")
		}
		if toClean["scratch"] {
			log.WithField("scratch", toWarn["scratch"]).Warn("")
		}
		if toClean["rmd"] {
			log.WithField("rmd", toWarn["rmd"]).Warn("")
		}
		if toClean["images"] {
			log.WithField("images", toWarn["images"]).Warn("")
		}
	} else {
		log.WithFields(log.Fields{
			"containers": toWarn["containers"],
			"scratch":    toWarn["scratch"],
			"rmd":        toWarn["rmd"],
			"images":     toWarn["images"],
		}).Warn("The marmots are about to remove")
	}

	fmt.Print("Please confirm (y/n): ")

	fmt.Scanln(&input)
	if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
		log.Warn("Authorization given, removing")
		canWe = true
	} else {
		log.Warn("Authorization not given, exiting")
		canWe = false
	}
	return canWe
}

func TrimString(strang string) string {
	return strings.TrimSpace(strings.Trim(strang, "\n"))
}

/* TODO
if do.Volumes {
	if err := removeOrphanedVolumes(yes); err != nil {
		return err
	}
}

if do.Uninstall {
	if err := uninstallEris(yes); err != nil { //will also removeErisDir
		return err
	}
}

func removeOrphanedVolumes() error {
	return nil
}

func uninstallEris() error {
	return nil
}*/
