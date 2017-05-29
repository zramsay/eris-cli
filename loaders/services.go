package loaders

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/util"

	"github.com/spf13/viper"
)

// LoadServiceDefinition reads a service definition specified by a service
// name from the config.ServicesPath directory and returns the corresponding
// service definition file.
// LoadServiceDefinition can return missing file or definition file bad format
// errors.
func LoadServiceDefinition(servName string) (*definitions.ServiceDefinition, error) {
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

	if err = checkImage(srv.Service); err != nil {
		return nil, err
	}

	addDependencyVolumesAndLinks(srv.Dependencies, srv.Service, srv.Operations)

	ServiceFinalizeLoad(srv)
	return srv, nil
}

// MockServiceDefinition returns a service definition structure with
// necessary fields already filled in (with an exception of the Image field).
func MockServiceDefinition(servName string) *definitions.ServiceDefinition {
	srv := definitions.BlankServiceDefinition()
	srv.Name = servName

	log.WithField("=>", servName).Debug("Mocking service definition")

	srv.Operations.ContainerType = definitions.TypeService
	srv.Operations.Labels = util.Labels(servName, srv.Operations)

	ServiceFinalizeLoad(srv)
	return srv
}

// MarshalServiceDefinition converts a Viper configuration structure to a
// service definition one; it can return marshalling errors.
func MarshalServiceDefinition(serviceConf *viper.Viper, srv *definitions.ServiceDefinition) error {
	if err := serviceConf.Unmarshal(srv); err != nil {
		// [zr] this error to deduplicate with config/config.go:103 in #468
		return fmt.Errorf("Formatting error with your definition file:\n\n%v", err)
	}

	// toml bools don't really marshal well
	if serviceConf.GetBool("service.data_container") {
		srv.Service.AutoData = true
	}

	return nil
}

// ServiceFinalizeLoad performs sanity checks on the most import fields of the
// service definition structure by filling in missing ones and avoiding
// duplication. ServiceFinalizeLoad panics if all necessary fields are empty.
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

	srv.Operations.SrvContainerName = util.ServiceContainerName(srv.Name)
	srv.Operations.DataContainerName = util.ContainerName(definitions.TypeData, srv.Name)
}

// ConnectToAService operates in two ways
//  - if link is true, sets srv.Links to point to a service container specifiend by name:internalName
//  - if mount is true, sets srv.VolumesFrom to point to a service container specified by name
func ConnectToAService(srv *definitions.Service, ops *definitions.Operation, name, internalName string, link, mount bool) {
	connectToAService(srv, ops, definitions.TypeService, name, internalName, link, mount)
}

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
	return config.LoadViper(filepath.Join(config.ServicesPath), servName)
}

func checkImage(srv *definitions.Service) error {
	// Services must be given an image. Flame out if they do not.
	if srv.Image == "" {
		return fmt.Errorf(`An "image" field is required in the service definition file`)
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
