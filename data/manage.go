package data

import (
	"os"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/definitions"
	. "github.com/eris-ltd/eris-cli/errors"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-logger"
	. "github.com/eris-ltd/common/go/common"
)

func RenameData(do *definitions.Do) error {
	log.WithFields(log.Fields{
		"from": do.Name,
		"to":   do.NewName,
	}).Info("Renaming data container")

	if util.IsData(do.Name) {
		ops := loaders.LoadDataDefinition(do.Name)
		util.Merge(ops, do.Operations)

		if err := perform.DockerRename(ops, do.NewName); err != nil {
			return &ErisError{ErrDocker, err, ""}
		}
	} else {
		return &ErisError{ErrEris, ErrCantFindData, ""}
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
			return &ErisError{ErrDocker, err, ""}
		}
	} else {
		return &ErisError{ErrEris, ErrCantFindData, ""}
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
				// TODO error
				return &ErisError{ErrDocker, err, ""}
			}
		} else {
			return &ErisError{ErrDocker, ErrCantFindData, ""}
		}

		if do.RmHF {
			log.WithField("=>", do.Name).Warn("Removing host directory")
			if err = os.RemoveAll(filepath.Join(DataContainersPath, do.Name)); err != nil {
				return &ErisError{ErrGo, err, "use force"}
			}
		}
	}

	do.Result = "success"
	return err
}
