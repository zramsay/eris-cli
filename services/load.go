package services

import (
	"fmt"
	"strings"

	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func LoadServiceDefinition(servName string, cNum ...int) (*def.ServiceDefinition, error) {
	srv := def.BlankServiceDefinition()
	serviceConf, err := loadServiceDefinition(servName)
	if err != nil {
		return nil, err
	}

	if err = marshalServiceDefinition(serviceConf, srv); err != nil {
		return nil, err
	}

	if srv.Service == nil {
		return nil, fmt.Errorf("No service given.")
	}

	if err = checkImage(srv.Service); err != nil {
		return nil, err
	}

	if len(cNum) == 0 {
		// TODO: findNextContainerIndex => util/container_operations.go
		srv.Operations.ContainerNumber = 1
	} else {
		srv.Operations.ContainerNumber = cNum[0]
	}

	checkNames(srv)

	return srv, nil
}

func MockServiceDefinition(servName string, cNum ...int) *def.ServiceDefinition {
	srv := def.BlankServiceDefinition()
	srv.Name = servName

	if len(cNum) == 0 {
		// TODO: findNextContainerIndex => util/container_operations.go
		srv.Operations.ContainerNumber = 1
	} else {
		srv.Operations.ContainerNumber = cNum[0]
	}

	checkNames(srv)
	return srv
}

func IsServiceExisting(service *def.Service, ops *def.Operation) bool {
	return util.IsServiceContainer(service.Name, ops.ContainerNumber, true)
}

func IsServiceRunning(service *def.Service, ops *def.Operation) bool {
	return util.IsServiceContainer(service.Name, ops.ContainerNumber, false)
}

func IsServiceKnown(service *def.Service, ops *def.Operation) bool {
	return parseKnown(service.Name)
}

func loadServiceDefinition(servName string) (*viper.Viper, error) {
	return util.LoadViperConfig(dir.ServicesPath, servName, "service")
}

func servDefFileByServName(servName string) (string, error) {
	serviceConf, err := loadServiceDefinition(servName)
	if err != nil {
		return "", err
	}
	return serviceConf.ConfigFileUsed(), nil
}

func marshalServiceDefinition(serviceConf *viper.Viper, srv *def.ServiceDefinition) error {
	err := serviceConf.Marshal(srv)
	if err != nil {
		// Vipers error messages are atrocious.
		return fmt.Errorf("Sorry, the marmots could not figure that service definition out.\nPlease check for known services with [eris services known] and retry.\n")
	}

	// toml bools don't really marshal well
	if serviceConf.GetBool("service.data_container") {
		srv.Service.AutoData = true
	}

	return nil
}

// Services must be given an image. Flame out if they do not.
func checkImage(srv *def.Service) error {
	if srv.Image == "" {
		return fmt.Errorf("An \"image\" field is required in the service definition file.")
	}

	return nil
}

func checkNames(srv *def.ServiceDefinition) {
	// If no name use image name
	if srv.Name == "" {
		if srv.Service.Name != "" {
			logger.Debugf("Service definition has no name. Defaulting to service name =>\t%s\n", srv.Service.Name)
			srv.Name = srv.Service.Name
		} else {
			if srv.Service.Image != "" {
				srv.Name = strings.Replace(srv.Service.Image, "/", "_", -1)
				srv.Service.Name = srv.Name
				logger.Debugf("Service definition has no name. Defaulting to image name =>\t%s\n", srv.Name)
			} else {
				panic("Service's Image should have been set before reaching checkNames")
			}
		}
	}

	srv.Operations.SrvContainerName = util.ServiceContainersName(srv.Name, srv.Operations.ContainerNumber)
	srv.Operations.DataContainerName = util.ServiceToDataContainer(srv.Operations.SrvContainerName)
}

func parseKnown(name string) bool {
	known := ListKnownRaw()
	if len(known) != 0 {
		for _, srv := range known {
			if srv == name {
				return true
			}
		}
	}
	return false
}
