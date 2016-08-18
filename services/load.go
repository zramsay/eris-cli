package services

import (
	"errors"
	"os"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/util"
	log "github.com/eris-ltd/eris-logger"
)

var (
	ErrServiceNotRunning = errors.New("The requested service is not running, start it with `eris services start [serviceName]`")
)

// Checks that a service is running and starts it if it isn't.
func EnsureRunning(do *definitions.Do) error {
	if os.Getenv("ERIS_SKIP_ENSURE") != "" {
		return nil
	}

	if _, err := loaders.LoadServiceDefinition(do.Name); err != nil {
		return err
	}

	if !util.IsService(do.Name, true) {
		log.WithField("=>", do.Name).Warn("Starting service")
		do.Operations.Args = []string{do.Name}
		StartService(do)
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
