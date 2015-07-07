package data

import (
	"fmt"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"
	"os"
	"path"
)

func RenameDataRaw(do *definitions.Do) error {
	logger.Infof("Renaming DataC (fm DataRaw) =>\t%s:%s\n", do.Name, do.NewName)

	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {

		srv := definitions.BlankServiceDefinition()
		srv.Operations.SrvContainerName = util.ContainersName("data", do.Name, do.Operations.ContainerNumber)

		err := perform.DockerRename(srv.Service, srv.Operations, do.Name, do.NewName)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}
	do.Result = "success"
	return nil
}

func InspectDataRaw(do *definitions.Do) error {
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {
		logger.Infoln("Inspecting data container" + do.Name)

		srv := definitions.BlankServiceDefinition()
		srv.Operations.SrvContainerName = util.ContainersName("data", do.Name, do.Operations.ContainerNumber)

		err := perform.DockerInspect(srv.Service, srv.Operations, do.Args[0])
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}
	do.Result = "success"
	return nil
}

func RmDataRaw(do *definitions.Do) error {
	if do.RmHF {
		logger.Println("Removing host folder " + do.Name)
		os.RemoveAll(path.Join(DataContainersPath, do.Name))
	}
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {
		logger.Infoln("Removing data container " + do.Name)

		srv := definitions.BlankServiceDefinition()
		srv.Operations.SrvContainerName = util.ContainersName("data", do.Name, do.Operations.ContainerNumber)

		err := perform.DockerRemove(srv.Service, srv.Operations, false)
		if err != nil {
			return err
		}

	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}
	do.Result = "success"
	return nil
}

func ListKnownRaw(do *definitions.Do) error {
	do.Result = strings.Join(util.DataContainerNames(), "\n")
	return nil
}

func IsKnown(name string) bool {
	return _parseKnown(name)
}
