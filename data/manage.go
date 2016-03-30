package data

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func RenameData(do *definitions.Do) error {
	log.WithFields(log.Fields{
		"from": do.Name,
		"to":   do.NewName,
	}).Info("Renaming data container")

	if util.IsData(do.Name) {
		ops := loaders.LoadDataDefinition(do.Name)
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
	if util.IsData(do.Name) {
		log.WithField("=>", do.Name).Info("Inspecting data container")

		srv := definitions.BlankServiceDefinition()
		srv.Operations.SrvContainerName = util.ContainerName(definitions.TypeData, do.Name)

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
		if util.IsData(do.Name) {
			log.WithField("=>", do.Name).Info("Removing data container")

			srv := definitions.BlankServiceDefinition()
			srv.Operations.SrvContainerName = util.ContainerName("data", do.Name)

			if err = perform.DockerRemove(srv.Service, srv.Operations, false, do.Volumes, false); err != nil {
				log.Errorf("Error removing %s: %v", do.Name, err)
				return err
			}

		} else {
			err = fmt.Errorf("I cannot find that data container for %s. Please check the data container name you sent me.", do.Name)
			log.Error(err)
			return err
		}

		if do.RmHF {
			log.WithField("=>", do.Name).Warn("Removing host directory")
			if err = os.RemoveAll(filepath.Join(DataContainersPath, do.Name)); err != nil {
				return err
			}
		}
	}

	do.Result = "success"
	return err
}
