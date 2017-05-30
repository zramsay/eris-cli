package services

import (
	"bytes"
	"fmt"
	"os"

	"github.com/monax/monax/definitions"
	"github.com/monax/monax/loaders"
	"github.com/monax/monax/log"
	"github.com/monax/monax/perform"
	"github.com/monax/monax/util"
)

func StartService(do *definitions.Do) (err error) {
	var services []*definitions.ServiceDefinition

	do.Operations.Args = append(do.Operations.Args, do.ServicesSlice...)
	log.WithField("args", do.Operations.Args).Info("Building services group")
	for _, srv := range do.Operations.Args {
		s, e := BuildServicesGroup(srv)
		if e != nil {
			return e
		}
		services = append(services, s...)
	}

	// [csk]: controls for ops reconciliation, overwrite will, e.g., merge the maps and stuff
	for _, s := range services {
		util.Merge(s.Operations, do.Operations)
	}

	log.Debug("Checking services")
	for _, s := range services {
		log.WithFields(log.Fields{
			"name":         s.Name,
			"dependencies": s.Dependencies,
			"links":        s.Service.Links,
			"volumes from": s.Service.VolumesFrom,
		}).Debug()

		// Spacer.
		log.Debug()
	}

	// NOTE: the top level service should be at the end of the list
	topService := services[len(services)-1]
	topService.Service.Environment = append(topService.Service.Environment, do.Env...)
	topService.Service.Links = append(topService.Service.Links, do.Links...)
	services[len(services)-1] = topService

	return StartGroup(services)
}

func KillService(do *definitions.Do) (err error) {
	var services []*definitions.ServiceDefinition

	log.WithField("args", do.Operations.Args).Info("Building services group")
	for _, servName := range do.Operations.Args {
		s, e := BuildServicesGroup(servName)
		if e != nil {
			return e
		}
		services = append(services, s...)
	}

	// if force flag given, this will override any timeout flag
	if do.Force {
		do.Timeout = 0
	}

	for _, service := range services {
		if util.IsService(service.Service.Name, true) {
			log.WithField("=>", service.Service.Name).Debug("Stopping service")
			if err := perform.DockerStop(service.Service, service.Operations, do.Timeout); err != nil {
				return err
			}

		} else {
			log.WithField("=>", service.Service.Name).Info("Service not currently running. Skipping")
		}

		if do.Rm {
			if err := perform.DockerRemove(service.Service, service.Operations, do.RmD, do.Volumes, do.Force); err != nil {
				return err
			}
		}
	}

	return nil
}

func ExecService(do *definitions.Do) (buf *bytes.Buffer, err error) {
	service, err := loaders.LoadServiceDefinition(do.Name)
	if err != nil {
		return nil, err
	}

	util.Merge(service.Operations, do.Operations)
	if do.Service.User != "" {
		service.Service.User = do.Service.User
	}

	// Get the main service container name, check if it's running.
	main := util.ServiceContainerName(do.Name)
	if util.IsService(do.Name, true) {
		if service.Service.ExecHost == "" {
			log.Info("exec_host not found in service definition file")
			log.WithField("service", do.Name).Info("May not be able to communicate with the service")
		} else {
			// Grab the Container to inspect the service's IP address
			cont, err := util.DockerClient.InspectContainer(main)
			if err != nil {
				return nil, util.DockerError(err)
			}
			service.Service.Environment = append(service.Service.Environment,
				fmt.Sprintf("%s=%s", service.Service.ExecHost,
					cont.NetworkSettings.IPAddress))
		}

		// Use service's short name as a link alias.
		service.Service.Links = append(service.Service.Links, fmt.Sprintf("%s:%s", main, do.Name))
	}

	// Override links on the command line.
	if len(do.Links) > 0 {
		service.Service.Links = do.Links
	}

	return perform.DockerExecService(service.Service, service.Operations)
}

// ExecHandler implemements ExecService for use within
// the cli for under the hood functionality
// (wrapping) calls to respective containers
func ExecHandler(srvName string, args []string) (buf *bytes.Buffer, err error) {
	do := definitions.NowDo()
	do.Name = srvName
	do.Operations.Interactive = false
	do.Operations.Args = args
	do.Operations.PublishAllPorts = true
	return ExecService(do)
}

// TODO: test this recursion and service deps generally
func BuildServicesGroup(srvName string, services ...*definitions.ServiceDefinition) ([]*definitions.ServiceDefinition, error) {
	log.WithFields(log.Fields{
		"=>":        srvName,
		"services#": len(services),
	}).Debug("Building services group for")
	srv, err := loaders.LoadServiceDefinition(srvName)
	if err != nil {
		return nil, err
	}
	if srv.Dependencies != nil {
		for _, sName := range srv.Dependencies.Services {
			log.WithField("=>", sName).Debug("Found service dependency")
			s, e := BuildServicesGroup(sName)
			if e != nil {
				return nil, e
			}
			services = append(services, s...)
		}
	}
	services = append(services, srv)
	return services, nil
}

// start a group of services. catch errors on a channel so we can stop as soon as something goes wrong
func StartGroup(group []*definitions.ServiceDefinition) error {
	log.WithField("services#", len(group)).Debug("Starting services group")
	for _, srv := range group {
		log.WithField("=>", srv.Name).Debug("Performing container start")
		if err := perform.DockerRunService(srv.Service, srv.Operations); err != nil {
			return fmt.Errorf("Error starting service %s: %v", srv.Name, err)
		}
	}
	return nil
}

// Checks that a service is running and starts it if it isn't.
func EnsureRunning(do *definitions.Do) error {
	if os.Getenv("MONAX_SKIP_ENSURE") != "" {
		return nil
	}

	if _, err := loaders.LoadServiceDefinition(do.Name); err != nil {
		return err
	}

	if !util.IsService(do.Name, true) {
		log.WithField("=>", do.Name).Info("Starting service")
		do.Operations.Args = []string{do.Name}
		StartService(do)
	} else {
		log.WithField("=>", do.Name).Info("Service is running")
	}
	return nil
}

func FindServiceDefinitionFile(name string) string {
	return util.GetFileByNameAndType("services", name)
}

// used by [services ip]
func InspectService(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name)
	if err != nil {
		return err
	}
	return InspectServiceByService(service.Service, service.Operations, do.Operations.Args[0])
}

func LogsService(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name)
	if err != nil {
		return err
	}
	return perform.DockerLogs(service.Service, service.Operations, do.Follow, do.Tail)
}

func RmService(do *definitions.Do) error {
	for _, servName := range do.Operations.Args {
		service, err := loaders.LoadServiceDefinition(servName)
		if err != nil {
			return err
		}
		if util.IsService(service.Service.Name, false) {
			if err := perform.DockerRemove(service.Service, service.Operations, do.RmD, do.Volumes, do.Force); err != nil {
				return err
			}
		}

		if do.RmImage {
			if err := perform.DockerRemoveImage(service.Service.Image, true); err != nil {
				return err
			}
		}

		if do.File {
			oldFile := util.GetFileByNameAndType("services", servName)
			if err != nil {
				return err
			}
			log.WithField("file", oldFile).Warn("Removing file")
			if err := os.Remove(oldFile); err != nil {
				return err
			}
		}
	}
	return nil
}

func InspectServiceByService(srv *definitions.Service, ops *definitions.Operation, field string) error {
	return perform.DockerInspect(srv, ops, field)
}
