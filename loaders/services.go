package loaders


import (
	"fmt"
	"path"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func LoadServiceDefinition(servName string, cNum ...int) (*definitions.ServiceDefinition, error) {
	if len(cNum) == 0 || cNum[0] == 0 {
		logger.Debugf("Loading Service Definition =>\t%s:1 (autoassigned)\n", servName)
		// TODO: findNextContainerIndex => util/container_operations.go
		if len(cNum) == 0 {
			cNum = append(cNum, 1)
		} else {
			cNum[0] = 1
		}
	} else {
		logger.Debugf("Loading Service Definition =>\t%s:%d\n", servName, cNum[0])
	}

	srv := definitions.BlankServiceDefinition()
	serviceConf, err := loadServiceDefinition(servName)
	if err != nil {
		return nil, err
	}

	if err = MarshalServiceDefinition(serviceConf, srv); err != nil {
		return nil, err
	}

	if srv.Service == nil {
		return nil, fmt.Errorf("No service given.")
	}

	if err = checkImage(srv.Service); err != nil {
		return nil, err
	}

	srv.Operations.ContainerNumber = cNum[0]
	checkServiceNames(srv)
	addDependencyVolumesAndLinks(srv, cNum[0])

	return srv, nil
}

func MockServiceDefinition(servName string, cNum ...int) *definitions.ServiceDefinition {
	srv := definitions.BlankServiceDefinition()
	srv.Name = servName

	if len(cNum) == 0 {
		// TODO: findNextContainerIndex => util/container_operations.go
		srv.Operations.ContainerNumber = 1
	} else {
		srv.Operations.ContainerNumber = cNum[0]
	}

	checkServiceNames(srv)
	return srv
}

func loadServiceDefinition(servName string) (*viper.Viper, error) {
	return util.LoadViperConfig(path.Join(ServicesPath), servName, "service")
}

func MarshalServiceDefinition(serviceConf *viper.Viper, srv *definitions.ServiceDefinition) error {
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
func checkImage(srv *definitions.Service) error {
	if srv.Image == "" {
		return fmt.Errorf("An \"image\" field is required in the service definition file.")
	}

	return nil
}

func checkServiceNames(srv *definitions.ServiceDefinition) {
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
				panic("Service's Image should have been set before reaching checkServiceNames")
			}
		}
	}

	srv.Operations.SrvContainerName = util.ServiceContainersName(srv.Name, srv.Operations.ContainerNumber)
	srv.Operations.DataContainerName = util.ServiceToDataContainer(srv.Operations.SrvContainerName)
}

func addDependencyVolumesAndLinks(srv *definitions.ServiceDefinition, cNum int) {
	for _, dep := range srv.ServiceDeps {
		// Automagically provide links to serviceDeps so they can easily
		// find each other using Docker's automagical modifications to
		// /etc/hosts
		newLink := util.ServiceContainersName(dep, cNum) + ":" + dep
		srv.Service.Links = append(srv.Service.Links, newLink)

		// Automagically mount VolumesFrom for serviceDeps so they can
		// easily pass files back and forth
		newVol  := util.ServiceContainersName(dep, cNum) + ":rw" // for now mounting as "rw"
		srv.Service.VolumesFrom = append(srv.Service.VolumesFrom, newVol)
	}
}
