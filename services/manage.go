package services

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"
	log "github.com/eris-ltd/eris-logger"

	. "github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/common/go/ipfs"
)

func MakeService(do *definitions.Do) error {
	srv := definitions.BlankServiceDefinition()
	srv.Name = do.Name
	srv.Service.Name = do.Name
	srv.Service.Image = do.Operations.Args[0]
	srv.Service.AutoData = true

	var err error
	//get maintainer info
	srv.Maintainer.Name, srv.Maintainer.Email, err = config.GitConfigUser()
	if err != nil {
		// don't return -> field not required
		log.Debug(err.Error())
	}

	log.WithFields(log.Fields{
		"service": srv.Service.Name,
		"image":   srv.Service.Image,
	}).Debug("Creating a new service definition file")
	err = WriteServiceDefinitionFile(srv, filepath.Join(ServicesPath, do.Name+".toml"))
	if err != nil {
		return err
	}
	return nil
}

func EditService(do *definitions.Do) error {
	servDefFile := FindServiceDefinitionFile(do.Name)
	log.WithField("=>", servDefFile).Info("Editing service")
	return Editor(servDefFile)
}

func RenameService(do *definitions.Do) error {
	log.WithFields(log.Fields{
		"from": do.Name,
		"to":   do.NewName,
	}).Info("Renaming service")

	if do.Name == do.NewName {
		return fmt.Errorf("Cannot rename to same name")
	}

	newNameBase := strings.Replace(do.NewName, filepath.Ext(do.NewName), "", 1)
	transformOnly := newNameBase == do.Name

	if parseKnown(do.Name) {
		serviceDef, err := loaders.LoadServiceDefinition(do.Name)
		if err != nil {
			return err
		}

		if !transformOnly {
			log.WithFields(log.Fields{
				"from": do.Name,
				"to":   do.NewName,
			}).Debug("Performing container rename")
			err = perform.DockerRename(serviceDef.Operations, do.NewName)
			if err != nil {
				return err
			}
		} else {
			log.Info("Changing service definition file only. Not renaming container")
		}

		oldFile := FindServiceDefinitionFile(do.Name)

		if filepath.Base(oldFile) == do.NewName {
			log.Info("Those are the same file. Not renaming")
			return nil
		}

		var newFile string
		if filepath.Ext(do.NewName) == "" {
			newFile = strings.Replace(oldFile, do.Name, do.NewName, 1)
		} else {
			newFile = filepath.Join(ServicesPath, do.NewName)
		}

		serviceDef.Service.Name = strings.Replace(do.NewName, filepath.Ext(do.NewName), "", 1)
		serviceDef.Name = serviceDef.Service.Name
		err = WriteServiceDefinitionFile(serviceDef, newFile)
		if err != nil {
			return err
		}

		if !transformOnly {
			log.WithFields(log.Fields{
				"from": do.Name,
				"to":   do.NewName,
			}).Debug("Performing data container rename")
			err = data.RenameData(do)
			if err != nil {
				return err
			}
		}

		os.Remove(oldFile)
	} else {
		return fmt.Errorf("I cannot find that service. Please check the service name you sent me.")
	}
	return nil
}

func InspectService(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name)
	if err != nil {
		return err
	}
	err = InspectServiceByService(service.Service, service.Operations, do.Operations.Args[0])
	if err != nil {
		return err
	}
	return nil
}

func PortsService(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name)
	if err != nil {
		return err
	}

	if util.IsService(service.Service.Name, false) {
		log.Debug("Service exists, getting port mapping")
		return util.PrintPortMappings(service.Operations.SrvContainerName, do.Operations.Args)
	}

	return nil
}

func LogsService(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name)
	if err != nil {
		return err
	}
	return perform.DockerLogs(service.Service, service.Operations, do.Follow, do.Tail)
}

func UpdateService(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name)
	if err != nil {
		return err
	}
	service.Service.Environment = append(service.Service.Environment, do.Env...)
	service.Service.Links = append(service.Service.Links, do.Links...)
	err = perform.DockerRebuild(service.Service, service.Operations, do.Pull, do.Timeout)
	if err != nil {
		return err
	}
	return nil
}

func RmService(do *definitions.Do) error {
	for _, servName := range do.Operations.Args {
		service, err := loaders.LoadServiceDefinition(servName)
		if err != nil {
			return err
		}
		if util.IsService(service.Service.Name, false) {
			if err := perform.DockerRemove(service.Service, service.Operations, do.RmD, do.Volumes, do.Force); err != nil {
				return err
			}
		}

		if do.RmImage {
			if err := perform.DockerRemoveImage(service.Service.Image, true); err != nil {
				return err
			}
		}

		if do.File {
			oldFile := util.GetFileByNameAndType("services", servName)
			if err != nil {
				return err
			}
			log.WithField("file", oldFile).Warn("Removing file")
			if err := os.Remove(oldFile); err != nil {
				return err
			}
		}
	}
	return nil
}

func CatService(do *definitions.Do) (string, error) {
	configs := util.GetGlobalLevelConfigFilesByType("services", true)
	for _, c := range configs {
		cName := strings.Split(filepath.Base(c), ".")[0]
		if cName == do.Name {
			cat, err := ioutil.ReadFile(c)
			if err != nil {
				return "", err
			}
			return string(cat), nil
		}
	}
	return "", fmt.Errorf("Unknown service %s or invalid file extension", do.Name)
}

func InspectServiceByService(srv *definitions.Service, ops *definitions.Operation, field string) error {
	err := perform.DockerInspect(srv, ops, field)
	if err != nil {
		return err
	}

	return nil
}

func exportFile(servName string) (string, error) {
	return ipfs.SendToIPFS(FindServiceDefinitionFile(servName), "", "")
}
