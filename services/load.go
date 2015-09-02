package services

import (
	"strings"
	"time"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"
)

// EnsureRunning checks if a service is running and starts it if not
// TODO: ping all exposed ports until at least one is available (issue #149)
// NOTE: does not accept ENV vars
func EnsureRunning(do *definitions.Do) error {
	srv, err := loaders.LoadServiceDefinition(do.Name, false, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}

	if !IsServiceRunning(srv.Service, srv.Operations) {
		name := strings.ToUpper(do.Name)
		logger.Infof("%s is not running. Starting now. Waiting for %s to become available \n", name, name)
		err := perform.DockerRun(srv.Service, srv.Operations)
		if err != nil {
			return err
		}
		time.Sleep(time.Second * 5)
	} else {
		logger.Infof("%s is running.\n", strings.ToUpper(do.Name))
	}
	return nil

}

func IsServiceExisting(service *definitions.Service, ops *definitions.Operation) bool {
	logger.Debugf("Is Service Existing? =>\t\t%s:%d\n", service.Name, ops.ContainerNumber)
	return parseContainers(service, ops, true)
}

func IsServiceRunning(service *definitions.Service, ops *definitions.Operation) bool {
	logger.Debugf("Is Service Running? =>\t\t%s:%d\n", service.Name, ops.ContainerNumber)
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
	cName := util.FindServiceContainer(service.Name, ops.ContainerNumber, all)
	if cName == nil {
		return false
	}
	ops.SrvContainerName = cName.DockersName
	ops.SrvContainerID = cName.ContainerID

	// populate data container specifics
	if service.AutoData && ops.DataContainerID == "" {
		dName := util.FindDataContainer(service.Name, ops.ContainerNumber)
		if dName != nil {
			ops.DataContainerName = dName.DockersName
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
