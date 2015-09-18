package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	//chn "github.com/eris-ltd/eris-cli/chains"
	//"github.com/eris-ltd/eris-cli/data"
	//"github.com/eris-ltd/eris-cli/definitions"
	//srv "github.com/eris-ltd/eris-cli/services"
	//"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

func Clean(prompt, all, rmd, images bool) error {

	//do this always
	if err := defaultClean(prompt); err != nil {
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

	if prompt || canWeRemove([]string{}, "all") {
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
		logger.Println("permission to remove eris root directory not given, continuing with clean")
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
				logger.Printf("regexp error: %v\n", err)
			}

			if r.MatchString(rt) == true {
				erisImages = append(erisImages, rt)
				erisImageIDs = append(erisImageIDs, imageIDs[i])
			}
		}
	}

	if !prompt || canWeRemove(erisImages, "images") {
		for i, imageID := range erisImageIDs {
			logger.Debugf("removing image: %s with ID\t%s ", erisImages[i], imageID)
			if err := DockerClient.RemoveImage(imageID); err != nil {
				return err
			}
		}
	} else {
		logger.Println("permission to remove images not given, continuing with clean")
	}
	return nil
}

func canWeRemove(removing []string, what string) bool {
	//if nothing in removing, say so and return, or something
	var input string
	if what == "all" {
		logger.Printf("The marmots are about to forcefully remove all running and existing eris containers with corresponding volumes. Please confirm (y/Y):\n")
	} else {
		logger.Printf("The marmots are about to remove %s:\n%sPlease confirm (y/Y):", what, strings.Join(removing, "\n"))
	}

	fmt.Scanln(&input)
	if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
		logger.Printf("authorization given, removing %s\n", what)
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
