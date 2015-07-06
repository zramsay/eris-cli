package data

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"
)

func RenameDataRaw(oldName, newName string, containerNumber int) error {
	if util.IsDataContainer(oldName, containerNumber) {
		logger.Infoln("Renaming data container", oldName, "to", newName)

		srv := definitions.BlankServiceDefinition()
		srv.Operations.SrvContainerName = util.ContainersName("data", oldName, containerNumber)

		err := perform.DockerRename(srv.Service, srv.Operations, oldName, newName)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}
	return nil
}

func InspectDataRaw(name, field string, containerNumber int) error {
	if util.IsDataContainer(name, containerNumber) {
		logger.Infoln("Inspecting data container" + name)

		srv := definitions.BlankServiceDefinition()
		srv.Operations.SrvContainerName = util.ContainersName("data", name, containerNumber)

		err := perform.DockerInspect(srv.Service, srv.Operations, field)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}
	return nil
}

func RmDataRaw(name string, containerNumber int) error {
	if util.IsDataContainer(name, containerNumber) {
		logger.Infoln("Removing data container " + name)

		srv := definitions.BlankServiceDefinition()
		srv.Operations.SrvContainerName = util.ContainersName("data", name, containerNumber)

		err := perform.DockerRemove(srv.Service, srv.Operations, false)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}

	return nil
}

func ListKnownRaw() ([]string, error) {
	return util.DataContainerNames(), nil
}
