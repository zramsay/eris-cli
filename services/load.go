package services

import (
	"errors"
	"os"

	"github.com/eris-ltd/eris-cli/definitions"
	. "github.com/eris-ltd/eris-cli/errors"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/util"
	log "github.com/eris-ltd/eris-logger"
)

//checks that a service is running. if not, tells user to start it
// TODO return ErisError
func EnsureRunning(do *definitions.Do) error {
	if os.Getenv("ERIS_SKIP_ENSURE") != "" {
		return nil
	}

	srv, err := loaders.LoadServiceDefinition(do.Name)
	if err != nil {
		return err
	}

	if !util.IsService(srv.Service.Name, true) {
		return errors.New(ErrServiceNotRunning)
	} else {
		log.WithField("=>", do.Name).Info("Service is running")
	}
	return nil
}

func IsServiceKnown(service *definitions.Service, ops *definitions.Operation) bool {
	return parseKnown(service.Name)
}

func FindServiceDefinitionFile(name string) string {
	return util.GetFileByNameAndType("services", name)
}

func parseKnown(name string) bool {
	known := util.GetGlobalLevelConfigFilesByType("services", false)
	if len(known) != 0 {
		for _, srv := range known {
			if srv == name {
				return true
			}
		}
	}
	return false
}
