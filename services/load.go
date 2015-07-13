package services

import (
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
)

func IsServiceExisting(service *definitions.Service, ops *definitions.Operation) bool {
	logger.Debugf("Is Service Existing? =>\t\t%s:%d\n", service.Name, ops.ContainerNumber)
	cName := util.FindServiceContainer(service.Name, ops.ContainerNumber, true)
	if cName == nil {
		return false
	}
	ops.SrvContainerID = cName.ContainerID
	return true
}

func IsServiceRunning(service *definitions.Service, ops *definitions.Operation) bool {
	logger.Debugf("Is Service Running? =>\t\t%s:%d\n", service.Name, ops.ContainerNumber)
	cName := util.FindServiceContainer(service.Name, ops.ContainerNumber, false)
	if cName == nil {
		return false
	}
	ops.SrvContainerID = cName.ContainerID
	return true
}

func IsServiceKnown(service *definitions.Service, ops *definitions.Operation) bool {
	return parseKnown(service.Name)
}

func FindServiceDefinitionFile(name string) string {
	return util.GetFileByNameAndType("services", name)
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
