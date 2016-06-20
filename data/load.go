package data

import (
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-logger"
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
	srv.Operations.SrvContainerName = util.DataContainerName(srv.Name)
	srv.Operations.DataContainerName = util.DataContainerName(srv.Name)
	log.WithFields(log.Fields{
		"data container":    srv.Operations.DataContainerName,
		"service container": srv.Operations.SrvContainerName,
	}).Debug("Using names")
}
