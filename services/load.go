package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func LoadServiceDefinition(servName string) (*def.ServiceDefinition, error) {
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

	checkServiceHasName(service.Service, service.Operations)
	checkServiceHasDataContainer(serviceConf, service.Service, service.Operations)
	checkDataContainerHasName(service.Operations)

	return &service, nil
}

func LoadService(servName string) (*def.Service, error) {
	sd, err := LoadServiceDefinition(servName)
	return sd.Service, err
}

func LoadServiceOperation(servName string) (*def.ServiceOperation, error) {
	sd, err := LoadServiceDefinition(servName)
	return sd.Operations, err
}

func IsServiceExisting(service *def.Service) bool {
	return parseServices(service.Name, true)
}

func IsServiceRunning(service *def.Service) bool {
	return parseServices(service.Name, false)
}

func loadServiceDefinition(servName string) (*viper.Viper, error) {
	var serviceConf = viper.New()

	serviceConf.AddConfigPath(dir.ServicesPath)
	serviceConf.SetConfigName(servName)
	if err := serviceConf.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error loading service definition for %s: %v", servName, err)
	}

	return serviceConf, nil
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

func checkServiceGiven(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("No Service Given. Please rerun command with a known service.")
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
			service.Name = strings.Replace(service.Image, "/", "_", -1)
		}
	}

	containerNumber := 1 // tmp
	ops.SrvContainerName = "eris_service_" + service.Name + "_" + strconv.Itoa(containerNumber)
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

func parseServices(name string, all bool) bool {
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
	services := []string{}
	r := regexp.MustCompile(`\/eris_service_(.+?)_\d`)
	// docker has this weird thing where it returns links as individual
	// container (as in there is the container of two linked services and
	// the linkage between them is actually its own containers). this explains
	// the leading hash on containers. the q regexp is to filer out these
	// links from the return list as they are irrelevant to the information
	// desired by this function. and frankly they give false answers to
	// IsServiceRunning and ls,ps,known functions.
	q := regexp.MustCompile(`\A\/eris_service_(.+?)_\d/(.+?)\z`)

	running = false
	contns, _ := util.DockerClient.ListContainers(docker.ListContainersOptions{All: running})
	for _, con := range contns {
		for _, c := range con.Names {
			match := r.FindAllStringSubmatch(c, 1)
			m2 := q.FindAllStringSubmatch(c, 1)
			if len(m2) != 0 {
				continue
			}
			if len(match) != 0 {
				services = append(services, r.FindAllStringSubmatch(c, 1)[0][1])
			}
		}
	}

	return services
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
