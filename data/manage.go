package data

import (
	"fmt"
	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"
	"os"
	"path"
	"strings"
)

func RenameData(do *definitions.Do) error {
	logger.Infof("Renaming Data =>\t\t%s:%s\n", do.Name, do.NewName)
	logger.Debugf("\twith ContainerNumber =>\t%d\n", do.Operations.ContainerNumber)

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

func InspectData(do *definitions.Do) error {
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

func RmData(do *definitions.Do) error {
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

	if do.RmHF {
		logger.Println("Removing host folder " + do.Name)
		err := os.RemoveAll(path.Join(DataContainersPath, do.Name))
		if err != nil {
			return err
		}
	}

	do.Result = "success"
	return nil
}

func ListKnown(do *definitions.Do) error {
	do.Result = strings.Join(util.DataContainerNames(), "\n")
	return nil
}

func IsKnown(name string) bool {
	return _parseKnown(name)
}
