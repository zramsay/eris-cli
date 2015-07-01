package data

import (
	"fmt"
	"regexp"

	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

func RenameDataRaw(oldName, newName string, containerNumber int) error {
	if parseKnown(oldName) {
		logger.Infoln("Renaming data container", oldName, "to", newName)
		srv, ops := mockService(oldName, containerNumber)
		err := perform.DockerRename(srv, ops, oldName, newName)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}
	return nil
}

func InspectDataRaw(name, field string, containerNumber int) error {
	if parseKnown(name) {
		logger.Infoln("Inspecting data container" + name)

		srv, ops := mockService(name, containerNumber)
		err := perform.DockerInspect(srv, ops, field)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}
	return nil
}

func RmDataRaw(name string, containerNumber int) error {
	if parseKnown(name) {
		logger.Infoln("Removing data container" + name)

		srv, ops := mockService(name, containerNumber)
		err := perform.DockerRemove(srv, ops)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}

	return nil
}

func ListKnownRaw() ([]string, error) {
	dataCont := []string{}
	r := regexp.MustCompile(`\/eris_data_(.+)_\d`)

	contns, _ := util.DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	for _, con := range contns {
		for _, c := range con.Names {
			match := r.FindAllStringSubmatch(c, 1)
			if len(match) != 0 {
				dataCont = append(dataCont, r.FindAllStringSubmatch(c, 1)[0][1])
			}
		}
	}

	return dataCont, nil
}
