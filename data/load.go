package data

import (
	"fmt"

	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

func PretendToBeAService(serviceYourPretendingToBe string, cNum ...int) *def.ServiceDefinition {
	srv := def.BlankServiceDefinition()
	srv.Name = serviceYourPretendingToBe

	if len(cNum) == 0 || cNum[0] == 0 {
		log.WithField("=>", fmt.Sprintf("%s:1", serviceYourPretendingToBe)).Debug("Loading service definition (autoassigned)")
		// TODO: findNextContainerIndex => util/container_operations.go
		if len(cNum) == 0 {
			cNum = append(cNum, 1)
		} else {
			cNum[0] = 1
		}
	} else {
		log.WithField("=>", fmt.Sprintf("%s:%d", serviceYourPretendingToBe, cNum[0])).Debug("Loading service definition")
	}

	srv.Operations.ContainerNumber = cNum[0]

	giveMeAllTheNames(serviceYourPretendingToBe, srv)
	return srv
}

func giveMeAllTheNames(name string, srv *def.ServiceDefinition) {
	log.WithField("=>", name).Debug("Giving myself all the names")
	srv.Name = name
	srv.Service.Name = name
	srv.Operations.SrvContainerName = util.DataContainersName(srv.Name, srv.Operations.ContainerNumber)
	srv.Operations.DataContainerName = util.DataContainersName(srv.Name, srv.Operations.ContainerNumber)
	log.WithFields(log.Fields{
		"data container":    srv.Operations.DataContainerName,
		"service container": srv.Operations.SrvContainerName,
	}).Debug("Using names")
}

func checkServiceGiven(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("No Data Container Given. Please rerun command with a known data container.")
	}
	return nil
}
