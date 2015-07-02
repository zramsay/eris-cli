package data

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"
)

func RenameDataRaw(oldName, newName string, containerNumber int) error {
	if parseKnown(oldName, containerNumber) {
		logger.Infoln("Renaming data container", oldName, "to", newName)
		srv, ops := MockService(oldName, containerNumber)
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
	if parseKnown(name, containerNumber) {
		logger.Infoln("Inspecting data container" + name)

		srv, ops := MockService(name, containerNumber)
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
	if parseKnown(name, containerNumber) {
		logger.Infoln("Removing data container " + name)

		srv, ops := MockService(name, containerNumber)
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
	return util.ParseContainerNames("data", true), nil
}
