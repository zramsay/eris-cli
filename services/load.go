package services

import (
	"fmt"
	"net"
	"net/http"
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

	var id string
	if !IsServiceRunning(srv.Service, srv.Operations) {
		name := strings.ToUpper(do.Name)
		logger.Infof("%s is not running. Starting now. Waiting for %s to become available \n", name, name)
		if id, err = perform.DockerRun(srv.Service, srv.Operations); err != nil {
			return err
		}
		// TODO: to do this right we have to get the bound port
		// which might be randomly assigned!

		cont, err := util.DockerClient.InspectContainer(id)
		if err != nil {
			return err
		}
		exposedPorts := cont.NetworkSettings.Ports

	MAIN_LOOP:
		for {
			// give it a half second and then try all the ports
			var endpoint string
			time.Sleep(500 * time.Millisecond)
			for _, ep := range exposedPorts {
				for _, p := range ep {
					endpoint = fmt.Sprintf("%s:%s", p.HostIP, p.HostPort)
					if _, err := net.Dial("tcp", endpoint); err != nil {
						time.Sleep(500 * time.Millisecond)
					}
				}
				if _, err := http.Post(endpoint, "", nil); err != nil {
					time.Sleep(500 * time.Millisecond)
				}
				if _, err := http.Get(endpoint); err != nil {
					time.Sleep(500 * time.Millisecond)
				}
			}
			time.Sleep(500 * time.Millisecond)
			break MAIN_LOOP
		}

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
