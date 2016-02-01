package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

func Clean(prompt, all, rmd, images bool) error {
	// [zr]: TODO better prompt solution

	//do this always
	if err := defaultClean(prompt); err != nil {
		return err
	}

	//also do this always

	if err := cleanScratchData(); err != nil {
		return err
	}

	if all {
		rmd = true
		images = true
		//do.Volumes = true
	}

	if rmd {
		if err := removeErisDir(prompt); err != nil {
			return err
		}
	}
	if images {
		if err := removeErisImages(prompt); err != nil {
			return err
		}
	}

	/* TODO [zr] see issue #295
	if do.Volumes {
		if err := removeOrphanedVolumes(prompt); err != nil {
			return err
		}
	}

	if do.Uninstall {
		if err := uninstallEris(prompt); err != nil { //will also removeErisDir
			return err
		}
	}*/

	return nil
}

// stops and removes containers and their volumes
func defaultClean(prompt bool) error {
	contns, err := DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		fmt.Printf("List containers error: %v\n", err)
	}

	if !prompt || canWeRemove([]string{}, "all") {
		for _, container := range contns {
			if container.Labels["eris:ERIS"] == "true" {

				removeOpts := docker.RemoveContainerOptions{
					ID:            container.ID,
					RemoveVolumes: true,
					Force:         true,
				}
				if err := DockerClient.RemoveContainer(removeOpts); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func cleanScratchData() error {
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
	return nil
}

func removeErisDir(prompt bool) error {
	erisRoot, err := ioutil.ReadDir(common.ErisRoot)
	if err != nil {
		return err
	}

	dirsInErisRoot := make([]string, len(erisRoot))
	for i, dir := range erisRoot {
		dirsInErisRoot[i] = dir.Name()
	}

	if !prompt || canWeRemove(dirsInErisRoot, common.ErisRoot) {
		if err := os.RemoveAll(common.ErisRoot); err != nil {
			return err
		}
	} else {
		log.Warn("Permission to remove eris root directory not given, continuing with clean")
	}
	return nil
}

func removeErisImages(prompt bool) error {
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
	// [zr] this could probably be cleaner
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

	if !prompt || canWeRemove(erisImages, "images") {
		for i, imageID := range erisImageIDs {
			log.WithFields(log.Fields{
				"=>": erisImages[i],
				"id": imageID,
			}).Debug("Removing image")
			if err := DockerClient.RemoveImage(imageID); err != nil {
				return err
			}
		}
	} else {
		log.Warn("Permission to remove images not given, continuing with clean")
	}
	return nil
}

func canWeRemove(removing []string, what string) bool {
	//if nothing in removing, say so and return, or something
	var input string
	if what == "all" {
		fmt.Print("The marmots are about to forcefully remove all running and existing eris containers with corresponding volumes. Please confirm (y/Y): ")
	} else {
		log.WithField("=>", what).Warn("The marmots are about to remove")
		log.Warn(strings.Join(removing, "\n"))
		fmt.Print("Please confirm (y/Y): ")
	}

	fmt.Scanln(&input)
	if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
		log.WithField("=>", what).Warn("Authorization given, removing")
		return true
	} else {
		return false
	}
	return false
}

//TODO: see bash script
func removeOrphanedVolumes(prompt bool) error {
	return nil
}

//TODO
func uninstallEris(prompt bool) error {
	if err := removeErisDir(prompt); err != nil { //and other things
		return err
	}
	return nil
}

func TrimString(strang string) string {
	return strings.TrimSpace(strings.Trim(strang, "\n"))
}
