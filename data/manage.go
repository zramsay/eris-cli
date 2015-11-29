package data

import (
	"fmt"
	"os"
	"path"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func RenameData(do *definitions.Do) error {
	logger.Infof("Renaming Data =>\t\t%s:%s\n", do.Name, do.NewName)
	logger.Debugf("\twith ContainerNumber =>\t%d\n", do.Operations.ContainerNumber)

	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {
		ops := loaders.LoadDataDefinition(do.Name, do.Operations.ContainerNumber)
		util.Merge(ops, do.Operations)

		err := perform.DockerRename(ops, do.NewName)
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
		srv.Operations.SrvContainerName = util.ContainersName(definitions.TypeData, do.Name, do.Operations.ContainerNumber)

		err := perform.DockerInspect(srv.Service, srv.Operations, do.Operations.Args[0])
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}
	do.Result = "success"
	return nil
}

// TODO: skip errors flag
func RmData(do *definitions.Do) (err error) {
	if len(do.Operations.Args) == 0 {
		do.Operations.Args = []string{do.Name}
	}
	for _, name := range do.Operations.Args {
		do.Name = name
		if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {
			logger.Infoln("Removing data container " + do.Name)

			srv := definitions.BlankServiceDefinition()
			srv.Operations.SrvContainerName = util.ContainersName("data", do.Name, do.Operations.ContainerNumber)

			if err = perform.DockerRemove(srv.Service, srv.Operations, false, do.Volumes); err != nil {
				logger.Errorf("Error removing %s: %v", do.Name, err)
				return err
			}

		} else {
			err = fmt.Errorf("I cannot find that data container for %s. Please check the data container name you sent me.", do.Name)
			logger.Errorln(err)
			return err
		}

		if do.RmHF {
			logger.Println("Removing host folder " + do.Name)
			if err = os.RemoveAll(path.Join(DataContainersPath, do.Name)); err != nil {
				return err
			}
		}
	}

	do.Result = "success"
	return err
}

func IsKnown(name string) bool {
	return _parseKnown(name)
}
