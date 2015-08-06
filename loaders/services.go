package loaders

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func LoadServiceDefinition(servName string, newCont bool, cNum ...int) (*definitions.ServiceDefinition, error) {
	if len(cNum) == 0 {
		cNum = append(cNum, 0)
	}

	if cNum[0] == 0 {
		cNum[0] = util.AutoMagic(0, "service", newCont)
		logger.Debugf("Loading Service Definition =>\t%s:%d (autoassigned)\n", servName, cNum[0])
	} else {
		logger.Debugf("Loading Service Definition =>\t%s:%d\n", servName, cNum[0])
	}

	srv := definitions.BlankServiceDefinition()
	srv.Operations.ContainerNumber = cNum[0]
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

	if os.Getenv("TEST_IN_CIRCLE") != "true" { // this really should be docker version < 1.7...?
		addDependencyVolumesAndLinks(srv)
	}

	ServiceFinalizeLoad(srv)
	return srv, nil
}

func MockServiceDefinition(servName string, newCont bool, cNum ...int) *definitions.ServiceDefinition {
	srv := definitions.BlankServiceDefinition()
	srv.Name = servName

	if len(cNum) == 0 {
		srv.Operations.ContainerNumber = util.AutoMagic(cNum[0], "service", newCont)
		logger.Debugf("Mocking Service Definition =>\t%s:%d (autoassigned)\n", servName, cNum[0])
	} else {
		srv.Operations.ContainerNumber = cNum[0]
		logger.Debugf("Mocking Service Definition =>\t%s:%d\n", servName, cNum[0])
	}

	ServiceFinalizeLoad(srv)
	return srv
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

// These are things we want to *always* control. Should be last
// called before a return...
func ServiceFinalizeLoad(srv *definitions.ServiceDefinition) {
	// If no name use image name
	if srv.Name == "" {
		if srv.Service.Name != "" {
			logger.Debugf("Service definition has no name.\nDefaulting to service name =>\t%s\n", srv.Service.Name)
			srv.Name = srv.Service.Name
		} else {
			if srv.Service.Image != "" {
				srv.Name = strings.Replace(srv.Service.Image, "/", "_", -1)
				srv.Service.Name = srv.Name
				logger.Debugf("Service definition has no name.\nDefaulting to image name =>\t%s\n", srv.Name)
			} else {
				panic("Service's Image should have been set before reaching ServiceFinalizeLoad")
			}
		}
	}

	container := util.FindServiceContainer(srv.Name, srv.Operations.ContainerNumber, true)

	if container != nil {
		logger.Debugf("Setting SrvCont Names =>\t%s:%s\n", container.FullName, container.ContainerID)
		srv.Operations.SrvContainerName = container.FullName
		srv.Operations.SrvContainerID = container.ContainerID
	} else {
		srv.Operations.SrvContainerName = util.ServiceContainersName(srv.Name, srv.Operations.ContainerNumber)
		srv.Operations.DataContainerName = util.ServiceToDataContainer(srv.Operations.SrvContainerName)
	}
	if srv.Service.AutoData {
		dataContainer := util.FindDataContainer(srv.Name, srv.Operations.ContainerNumber)
		if dataContainer != nil {
			logger.Debugf("Setting DataCont Names =>\t%s:%s\n", dataContainer.FullName, dataContainer.ContainerID)
			srv.Operations.DataContainerName = dataContainer.FullName
			srv.Operations.DataContainerID = dataContainer.ContainerID
		} else {
			srv.Operations.SrvContainerName = util.ServiceContainersName(srv.Name, srv.Operations.ContainerNumber)
			srv.Operations.DataContainerName = util.ServiceToDataContainer(srv.Operations.SrvContainerName)
		}
	}
}

func ConnectToAService(srv *definitions.ServiceDefinition, dep string) {
	// Automagically provide links to serviceDeps so they can easily
	// find each other using Docker's automagical modifications to
	// /etc/hosts
	newLink := util.ServiceContainersName(dep, srv.Operations.ContainerNumber) + ":" + dep
	srv.Service.Links = append(srv.Service.Links, newLink)

	// Automagically mount VolumesFrom for serviceDeps so they can
	// easily pass files back and forth
	newVol := util.ServiceContainersName(dep, srv.Operations.ContainerNumber) + ":rw" // for now mounting as "rw"
	srv.Service.VolumesFrom = append(srv.Service.VolumesFrom, newVol)
}

// --------------------------------------------------------------------
// helpers

func loadServiceDefinition(servName string) (*viper.Viper, error) {
	return util.LoadViperConfig(path.Join(ServicesPath), servName, "service")
}

// Services must be given an image. Flame out if they do not.
func checkImage(srv *definitions.Service) error {
	if srv.Image == "" {
		return fmt.Errorf("An \"image\" field is required in the service definition file.")
	}

	return nil
}

func addDependencyVolumesAndLinks(srv *definitions.ServiceDefinition) {
	for _, dep := range srv.ServiceDeps {
		ConnectToAService(srv, dep)
	}
}
