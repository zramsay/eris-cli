package loaders

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-logger"
	. "github.com/eris-ltd/common/go/common"
	def "github.com/eris-ltd/eris-cli/definitions"

	"github.com/spf13/viper"
)

func LoadServiceDefinition(servName string, newCont bool) (*definitions.ServiceDefinition, error) {

	log.WithField("=>", servName).Debug("Loading service definition")

	srv := definitions.BlankServiceDefinition()
	srv.Operations.ContainerType = definitions.TypeService
	srv.Operations.Labels = util.Labels(servName, srv.Operations)
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

	// Docker 1.6 (which eris doesn't support) had different linking mechanism.
	if util.IsMinimalDockerClientVersion() {
		addDependencyVolumesAndLinks(srv.Dependencies, srv.Service, srv.Operations)
	}

	ServiceFinalizeLoad(srv)
	return srv, nil
}

func MockServiceDefinition(servName string, newCont bool) *definitions.ServiceDefinition {
	srv := definitions.BlankServiceDefinition()
	srv.Name = servName

	log.WithField("=>", servName).Debug("Mocking service definition")

	srv.Operations.ContainerType = definitions.TypeService
	srv.Operations.Labels = util.Labels(servName, srv.Operations)

	ServiceFinalizeLoad(srv)
	return srv
}

func MarshalServiceDefinition(serviceConf *viper.Viper, srv *definitions.ServiceDefinition) error {
	err := serviceConf.Unmarshal(srv)
	if err != nil {
		// Vipers error messages are atrocious.
		return fmt.Errorf("Sorry, the marmots could not figure that service definition out.\nPlease check for known services with [eris services ls --known] and retry.\n")
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
	if srv.Name == "" && srv.Service.Name == "" && srv.Service.Image == "" { // If no name or image, panic
		panic("Service's Image should have been set before reaching ServiceFinalizeLoad")
	} else if srv.Name == "" && srv.Service.Name == "" && srv.Service.Image != "" { // If no name use image
		srv.Name = strings.Replace(srv.Service.Image, "/", "_", -1)
		srv.Service.Name = srv.Name
		log.WithField("image", srv.Name).Debug("Defaulting to image")
	} else if srv.Service.Name != "" && srv.Name == "" { // harmonize names
		srv.Name = srv.Service.Name
		log.WithField("service", srv.Service.Name).Debug("Defaulting to service")
	} else if srv.Service.Name == "" && srv.Name != "" {
		srv.Service.Name = srv.Name
		log.WithField("service", srv.Name).Debug("Defaulting to service")
	}

	container := util.ContainerName(def.TypeService, srv.Name)

	if util.IsService(srv.Name, true) {
		log.WithField("=>", container).Debug("Setting service container name")
		srv.Operations.SrvContainerName = container
	} else {
		srv.Operations.SrvContainerName = util.ServiceContainerName(srv.Name)
		srv.Operations.DataContainerName = util.ContainerName(def.TypeData, srv.Name)
	}
	if srv.Service.AutoData {
		dataContainer := util.ContainerName(def.TypeData, srv.Name)
		if util.IsData(srv.Name) {
			log.WithField("=>", dataContainer).Debug("Setting data container name")
			srv.Operations.DataContainerName = dataContainer
		} else {
			srv.Operations.SrvContainerName = util.ServiceContainerName(srv.Name)
			srv.Operations.DataContainerName = util.ContainerName(def.TypeData, srv.Name)
		}
	}
}

func ConnectToAService(srv *definitions.Service, ops *definitions.Operation, name, internalName string, link, mount bool) {
	connectToAService(srv, ops, definitions.TypeService, name, internalName, link, mount)
}

// --------------------------------------------------------------------
// helpers

// links and mounts for service dependencies
func connectToAService(srv *definitions.Service, ops *definitions.Operation, typ, name, internalName string, link, mount bool) {
	log.WithFields(log.Fields{
		"=>":            srv.Name,
		"type":          typ,
		"name":          name,
		"internal name": internalName,
		"link":          link,
		"volumes from":  mount,
	}).Debug("Connecting to service")
	containerName := util.ContainerName(typ, name)

	if link {
		newLink := containerName + ":" + internalName
		srv.Links = append(srv.Links, newLink)
	}

	if mount {
		// Automagically mount VolumesFrom for serviceDeps so they can
		// easily pass files back and forth. note that this is opinionated
		// and will mount as read-write. we can revisit this if read-only
		// mounting required for specific use cases
		newVol := containerName + ":rw"
		srv.VolumesFrom = append(srv.VolumesFrom, newVol)
	}
}

func loadServiceDefinition(servName string) (*viper.Viper, error) {
	return config.LoadViperConfig(filepath.Join(ServicesPath), servName, "service")
}

// Services must be given an image. Flame out if they do not.
func checkImage(srv *definitions.Service) error {
	if srv.Image == "" {
		return fmt.Errorf("An \"image\" field is required in the service definition file.")
	}

	return nil
}

func addDependencyVolumesAndLinks(deps *definitions.Dependencies, srv *definitions.Service, ops *definitions.Operation) {
	if deps != nil {
		for i, dep := range deps.Services {
			name, internalName, link, mount := util.ParseDependency(dep)
			ConnectToAService(srv, ops, name, internalName, link, mount)
			deps.Services[i] = name
		}

		for i, dep := range deps.Chains {
			name, internalName, link, mount := util.ParseDependency(dep)
			ConnectToAChain(srv, ops, name, internalName, link, mount)
			deps.Chains[i] = name
		}
	}
}
