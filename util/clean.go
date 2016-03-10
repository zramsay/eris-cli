package util

import (
	"fmt"
	//"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

var toWarn = map[string]string{
	"containers": "All Eris Containers",
	"scratch":    "The contents of $HOME/eris/scratch",
	"rmd":        "The Eris Root Directory ($HOME/.eris)",
	"images":     "All Eris docker images",
}

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
		if err := RemoveAllErisContainers(); err != nil {
			return err
		}
	}

	if toClean["scratch"] {
		if err := CleanScratchData(); err != nil {
			return err
		}
	}

	if toClean["rmd"] {
		if err := os.RemoveAll(common.ErisRoot); err != nil {
			return err
		}
	}
	if toClean["images"] {
		if err := RemoveErisImages(); err != nil {
			return err
		}
	}
	return nil
}

// stops and removes containers and their volumes
func RemoveAllErisContainers() error {
	// TODO catch all data conts
	// use our listing functions ... ?
	contns, err := DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return fmt.Errorf("error listing containers: %v\n", err)
	}

	for _, container := range contns {
		if container.Labels["eris:ERIS"] == "true" {

			removeOpts := docker.RemoveContainerOptions{
				ID:            container.ID,
				RemoveVolumes: true,
				Force:         true,
			}
			if err := DockerClient.RemoveContainer(removeOpts); err != nil {
				// in 1.10.1 there is a weird EOF error which occurs here even though the container is removed. ignoring that.
				if fmt.Sprintf("%v", err) == "EOF" {
					log.Debug("Weird EOF error. Not reaping.")
					continue
				}
				return err
			}
		}
	}
	return nil
}

func CleanScratchData() error {
	if DoesDirExist(common.DataContainersPath) {
		d, err := os.Open(common.DataContainersPath)
		if err != nil {
			return err
		}
		defer d.Close()

		names, err := d.Readdirnames(-1)
		if err != nil {
			return err
		}
		for _, name := range names {
			err = os.RemoveAll(filepath.Join(common.DataContainersPath, name))
			if err != nil {
				return err
			}
		}
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

	//toWarnSome := make(map[string]string)

	if toClean["all"] != true {
		for _, thing := range toWarn {
			//TODO .. !
			fmt.Printf("thing: %v\n", thing)
		}
	} else {
		log.WithFields(log.Fields{
			"containers": toWarn["containers"],
			"scratch":    toWarn["scratch"],
			"rmd":        toWarn["rmd"],
			"images":     toWarn["images"],
		}).Warn("The marmots are about to remove")
	}

	fmt.Print("Please confirm (y/Y): ")

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
