package data

import (
	"fmt"

	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
)

func PretendToBeAService(serviceYourPretendingToBe string, cNum ...int) *def.ServiceDefinition {
	srv := def.BlankServiceDefinition()
	srv.Name = serviceYourPretendingToBe

	if len(cNum) == 0 || cNum[0] == 0 {
		logger.Debugf("Loading Service Definition =>\t%s:1 (autoassigned)\n", serviceYourPretendingToBe)
		// TODO: findNextContainerIndex => util/container_operations.go
		if len(cNum) == 0 {
			cNum = append(cNum, 1)
		} else {
			cNum[0] = 1
		}
	} else {
		logger.Debugf("Loading Service Definition =>\t%s:%d\n", serviceYourPretendingToBe, cNum[0])
	}

	srv.Operations.ContainerNumber = cNum[0]

	giveMeAllTheNames(serviceYourPretendingToBe, srv)
	return srv
}

func giveMeAllTheNames(name string, srv *def.ServiceDefinition) {
	logger.Debugf("Giving myself all the names =>\t%s\n", name)
	srv.Name = name
	srv.Service.Name = name
	srv.Operations.SrvContainerName = util.DataContainersName(srv.Name, srv.Operations.ContainerNumber)
	srv.Operations.DataContainerName = util.DataContainersName(srv.Name, srv.Operations.ContainerNumber)
	logger.Debugf("My service container name is =>\t%s\n", srv.Operations.SrvContainerName)
	logger.Debugf("My data container name is =>\t%s\n", srv.Operations.DataContainerName)
}

func checkServiceGiven(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("No Data Container Given. Please rerun command with a known data container.")
	}
	return nil
}

func parseKnown(name string, num int) bool {
	name = util.NameAndNumber(name, num)
	return _parseKnown(name)
}

func _parseKnown(name string) bool {
	do := def.NowDo()
	_ = ListKnown(do)
	if len(do.Args) != 0 {
		for _, srv := range do.Args {
			if srv == name {
				return true
			}
		}
	}
	return false
}
