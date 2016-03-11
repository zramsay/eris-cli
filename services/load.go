package services

import (
	"errors"
	"fmt"
	"os"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

var (
	ErrServiceNotRunning = errors.New("The requested service is not running, start it with `eris services start [serviceName]`")
)

//checks that a service is running. if not, tells user to start it
func EnsureRunning(do *definitions.Do) error {
	if os.Getenv("ERIS_SKIP_ENSURE") != "" {
		return nil
	}

	srv, err := loaders.LoadServiceDefinition(do.Name, false)
	if err != nil {
		return err
	}

	if !IsServiceRunning(srv.Service, srv.Operations) {
		e := fmt.Sprintf("The requested service is not running, start it with `eris services start %s`", do.Name)
		return errors.New(e)
	} else {
		log.WithField("=>", do.Name).Info("Service is running")
	}
	return nil
}

func IsServiceExisting(service *definitions.Service, ops *definitions.Operation) bool {
	log.WithField("=>", service.Name).Debug("Checking service existing")
	return parseContainers(service, ops, true)
}

func IsServiceRunning(service *definitions.Service, ops *definitions.Operation) bool {
	log.WithField("=>", service.Name).Debug("Checking service running")
	return parseContainers(service, ops, false)
}

func IsServiceKnown(service *definitions.Service, ops *definitions.Operation) bool {
	return parseKnown(service.Name)
}

func FindServiceDefinitionFile(name string) string {
	return util.GetFileByNameAndType("services", name)
}

func parseContainers(service *definitions.Service, ops *definitions.Operation, all bool) bool {
	// populate service container specifics
	cName := util.FindServiceContainer(service.Name, all)
	if cName == nil {
		return false
	}
	ops.SrvContainerName = cName.FullName
	ops.SrvContainerID = cName.ContainerID

	// populate data container specifics
	if service.AutoData && ops.DataContainerID == "" {
		dName := util.FindDataContainer(service.Name)
		if dName != nil {
			ops.DataContainerName = dName.FullName
			ops.DataContainerID = dName.ContainerID
		}
	}

	return true
}

func parseKnown(name string) bool {
	known := util.GetGlobalLevelConfigFilesByType("services", false)
	if len(known) != 0 {
		for _, srv := range known {
			if srv == name {
				return true
			}
		}
	}
	return false
}
