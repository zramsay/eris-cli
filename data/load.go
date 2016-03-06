package data

import (
	"fmt"

	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

func PretendToBeAService(serviceYourPretendingToBe string) *def.ServiceDefinition {
	srv := def.BlankServiceDefinition()
	srv.Name = serviceYourPretendingToBe

	giveMeAllTheNames(serviceYourPretendingToBe, srv)
	return srv
}

func giveMeAllTheNames(name string, srv *def.ServiceDefinition) {
	log.WithField("=>", name).Debug("Giving myself all the names")
	srv.Name = name
	srv.Service.Name = name
	srv.Operations.SrvContainerName = util.DataContainersName(srv.Name)
	srv.Operations.DataContainerName = util.DataContainersName(srv.Name)
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
