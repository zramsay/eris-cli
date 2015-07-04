package services

import (
	"fmt"
	"strings"

	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func LoadServiceDefinition(servName string, containerNumber int) (*def.ServiceDefinition, error) {
	var service def.ServiceDefinition
	serviceConf, err := loadServiceDefinition(servName)
	if err != nil {
		return nil, err
	}

	// marshal service and always reset the operational requirements
	// this will make sure to sync with docker so that if changes
	// have occured in the interim they are caught.
	service.Operations = &def.ServiceOperation{}
	if err = marshalServiceDefinition(serviceConf, &service); err != nil {
		return &def.ServiceDefinition{}, err
	}

	if service.Service == nil {
		return &service, fmt.Errorf("No service given.")
	}

	err = checkServiceHasImage(service.Service)
	if err != nil {
		return &def.ServiceDefinition{}, err
	}

	// set container number and format names
	service.Operations.ContainerNumber = containerNumber
	checkServiceHasName(service.Service, service.Operations)
	checkServiceHasDataContainer(serviceConf, service.Service, service.Operations)
	checkDataContainerHasName(service.Operations)

	return &service, nil
}

// not currently used by anything
func LoadService(servName string) (*def.Service, error) {
	sd, err := LoadServiceDefinition(servName, 1)
	return sd.Service, err
}

func IsServiceExisting(service *def.Service, ops *def.ServiceOperation) bool {
	return parseServices(service.Name, ops.ContainerNumber, true)
}

func IsServiceRunning(service *def.Service, ops *def.ServiceOperation) bool {
	return parseServices(service.Name, ops.ContainerNumber, false)
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

func marshalServiceDefinition(serviceConf *viper.Viper, service *def.ServiceDefinition) error {
	err := serviceConf.Marshal(service)
	if err != nil {
		return err
	}

	return nil
}

// Services must be given an image. Flame out if they do not.
func checkServiceHasImage(service *def.Service) error {
	if service.Image == "" {
		return fmt.Errorf("An \"image\" field is required in the service definition file.")
	}

	return nil
}

func checkServiceHasName(service *def.Service, ops *def.ServiceOperation) {
	// If no name use image name
	if service.Name == "" {
		if service.Image != "" {
			logger.Debugln("Service definition has no name. Defaulting to image name")
			service.Name = strings.Replace(service.Image, "/", "_", -1)
		} else {
			panic("service.Image should have been set before reaching checkServiceHasName")
		}
	}
	ops.SrvContainerName = fmt.Sprintf("eris_service_%s", util.NameAndNumber(service.Name, ops.ContainerNumber))
}

func checkServiceHasDataContainer(serviceConf *viper.Viper, service *def.Service, ops *def.ServiceOperation) {
	// toml bools don't really marshal well
	if serviceConf.GetBool("service.data_container") {
		service.AutoData = true
		ops.DataContainer = true
	}
}

func checkDataContainerHasName(ops *def.ServiceOperation) {
	if ops.DataContainer {
		ops.DataContainerName = ""
		if ops.DataContainer {
			dataSplit := strings.Split(ops.SrvContainerName, "_")
			dataSplit[1] = "data"
			ops.DataContainerName = strings.Join(dataSplit, "_")
		}
	}
}

func parseServices(name string, number int, all bool) bool {
	name = util.NameAndNumber(name, number)
	running := listServices(all)
	if len(running) != 0 {
		for _, srv := range running {
			if srv == name {
				return true
			}
		}
	}
	return false
}

func listServices(running bool) []string {
	return util.ParseContainerNames("service", running)
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
