package services

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
)

func ImportServiceRaw(do *definitions.Do) error {
	fileName := filepath.Join(ServicesPath, do.Name)
	if filepath.Ext(fileName) == "" {
		fileName = fileName + ".toml"
	}

	s := strings.Split(do.Path, ":")
	if s[0] == "ipfs" {

		var err error
		if logger.Level > 0 {
			err = util.GetFromIPFS(s[1], fileName, logger.Writer)
		} else {
			err = util.GetFromIPFS(s[1], fileName, bytes.NewBuffer([]byte{}))
		}

		if err != nil {
			return err
		}
		return nil
	}

	if strings.Contains(s[0], "github") {
		logger.Errorln("https://twitter.com/ryaneshea/status/595957712040628224")
		return nil
	}
	do.Result = "success"
	return fmt.Errorf("I do not know how to get that file. Sorry.")
}

func NewServiceRaw(do *definitions.Do) error {
	srv := definitions.BlankServiceDefinition()
	srv.Name = do.Name
	srv.Service.Name = do.Name
	srv.Service.Image = do.Args[0]
	srv.Service.AutoData = true

	logger.Debugf("Creating a new srv def file =>\t%s:%s\n", srv.Service.Name, srv.Service.Image)
	err := WriteServiceDefinitionFile(srv, path.Join(ServicesPath, do.Name+".toml"))
	if err != nil {
		return err
	}
	do.Result = "success"
	return nil
}

func EditServiceRaw(do *definitions.Do) error {
	servDefFile := FindServiceDefinitionFile(do.Name)
	do.Result = "success"
	return Editor(servDefFile)
}

func RenameServiceRaw(do *definitions.Do) error {
	logger.Infof("Renaming Service =>\t\t%s:%s:%d\n", do.Name, do.NewName, do.Operations.ContainerNumber)

	if do.Name == do.NewName {
		return fmt.Errorf("Cannot rename to same name")
	}

	newNameBase := strings.Replace(do.NewName, filepath.Ext(do.NewName), "", 1)
	transformOnly := newNameBase == do.Name

	if parseKnown(do.Name) {
		serviceDef, err := loaders.LoadServiceDefinition(do.Name, do.Operations.ContainerNumber)
		if err != nil {
			return err
		}

		if !transformOnly {
			err = perform.DockerRename(serviceDef.Service, serviceDef.Operations, do.Name, do.NewName)
			if err != nil {
				return err
			}
		}

		oldFile := FindServiceDefinitionFile(do.Name)

		if filepath.Base(oldFile) == do.NewName {
			logger.Infoln("Those are the same file. Not renaming")
			return nil
		}

		var newFile string
		if filepath.Ext(do.NewName) == "" {
			newFile = strings.Replace(oldFile, do.Name, do.NewName, 1)
		} else {
			newFile = filepath.Join(ServicesPath, do.NewName)
		}

		serviceDef.Service.Name = strings.Replace(do.NewName, filepath.Ext(do.NewName), "", 1)
		err = WriteServiceDefinitionFile(serviceDef, newFile)
		if err != nil {
			return err
		}

		if !transformOnly {
			err = data.RenameDataRaw(do)
			if err != nil {
				return err
			}
		}

		os.Remove(oldFile)
	} else {
		return fmt.Errorf("I cannot find that service. Please check the service name you sent me.")
	}
	do.Result = "success"
	return nil
}

func InspectServiceRaw(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}
	err = InspectServiceByService(service.Service, service.Operations, do.Args[0])
	if err != nil {
		return err
	}
	return nil
}

func LogsServiceRaw(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}
	return LogsServiceByService(service.Service, service.Operations, do.Follow, do.Tail)
}

func ExecServiceRaw(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}

	if IsServiceExisting(service.Service, service.Operations) {
		return ExecServiceByService(service.Service, service.Operations, do.Args, do.Interactive)
	} else {
		return fmt.Errorf("Services does not exist. Please start the service container with eris services start %s.\n", do.Name)
	}

	return nil
}

func ExportServiceRaw(do *definitions.Do) error {
	if parseKnown(do.Name) {
		ipfsService, err := loaders.LoadServiceDefinition("ipfs", 1)
		if err != nil {
			return err
		}

		if IsServiceRunning(ipfsService.Service, ipfsService.Operations) {
			logger.Infoln("IPFS is running. Adding now.")

			hash, err := exportFile(do.Name)
			if err != nil {
				return err
			}

			logger.Println(hash)
		} else {
			logger.Infoln("IPFS is not running. Starting now.")

			if err := perform.DockerRun(ipfsService.Service, ipfsService.Operations); err != nil {
				return err
			}

			hash, err := exportFile(do.Name)
			if err != nil {
				return err
			}

			do.Result = hash
			logger.Println(hash)
		}

	} else {
		return fmt.Errorf(`I don't known of that service.
Please retry with a known service.
To find known services use: eris services known`)
	}
	return nil
}

func ListKnownRaw(do *definitions.Do) error {
	srvs := util.GetGlobalLevelConfigFilesByType("services", false)
	do.Result = strings.Join(srvs, "\n")
	return nil
}

func ListRunningRaw(do *definitions.Do) error {
	logger.Debugln("Asking Docker Client for the Running Containers. Quiet? %v", do.Quiet)
	if do.Quiet {
		do.Result = strings.Join(util.ServiceContainerNames(false), "\n")
		if len(do.Args) != 0 && do.Args[0] != "testing" {
			logger.Printf("%s\n", "\n")
		}
	} else {
		perform.PrintTableReport("service", false) // TODO: return this as a string.
	}
	return nil
}

func ListExistingRaw(do *definitions.Do) error {
	logger.Debugln("Asking Docker Client for the Existing Containers.")
	if do.Quiet {
		do.Result = strings.Join(util.ServiceContainerNames(true), "\n")
		if len(do.Args) != 0 && do.Args[0] != "testing" {
			logger.Printf("%s\n", "\n")
		}
	} else {
		perform.PrintTableReport("service", false) // TODO: return this as a string.
	}
	return nil
}

func UpdateServiceRaw(do *definitions.Do) error {
	service, err := loaders.LoadServiceDefinition(do.Name, do.Operations.ContainerNumber)
	if err != nil {
		return err
	}
	err = perform.DockerRebuild(service.Service, service.Operations, do.SkipPull)
	if err != nil {
		return err
	}
	do.Result = "success"
	return nil
}

func RmServiceRaw(do *definitions.Do) error {
	for _, servName := range do.Args {
		service, err := loaders.LoadServiceDefinition(servName, do.Operations.ContainerNumber)
		if err != nil {
			return err
		}
		if IsServiceExisting(service.Service, service.Operations) {
			err = perform.DockerRemove(service.Service, service.Operations, do.RmD)
			if err != nil {
				return err
			}
		}

		if do.File {
			oldFile := util.GetFileByNameAndType("services", servName)
			if err != nil {
				return err
			}
			oldFile = path.Join(ServicesPath, oldFile)+".toml"
			logger.Printf("Removing file =>\t\t%s\n", oldFile)
			if err := os.Remove(oldFile); err != nil {
				return err
			}
		}
	}
	do.Result = "success"
	return nil
}

func CatServiceRaw(do *definitions.Do) error {
	dir, err := ioutil.ReadDir(ServicesPath)
	if err != nil {
		return err
	}

	var valid bool
	for i := 0; i < len(dir); i++ {
		fi := dir[i].Name()

		switch fi {
		case do.Name + ".toml":
			prettyCat(do.Name, ".toml")
			valid = true
		case do.Name + ".yaml":
			prettyCat(do.Name, ".yaml")
			valid = true
		case do.Name + ".yml":
			prettyCat(do.Name, ".yml")
			valid = true
		case do.Name + ".json":
			prettyCat(do.Name, ".json")
			valid = true
		default:
		}
	}
	if valid != true {
		logger.Println("Invalid file format. Only '.toml', '.yaml', '.yml', and '.json' are allowed")
	}
	return nil
}

func prettyCat(servName string, ext string) error {
	cat, err := ioutil.ReadFile(path.Join(ServicesPath, do.Name+ext))
	if err != nil {
		return err
	}
	do.Result = string(cat)
	return nil
}

func InspectServiceByService(srv *definitions.Service, ops *definitions.Operation, field string) error {
	if IsServiceExisting(srv, ops) {
		err := perform.DockerInspect(srv, ops, field)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("No service matching that name.\n")
	}
	return nil
}

func exportFile(servName string) (string, error) {
	fileName := FindServiceDefinitionFile(servName)

	var hash string
	var err error
	if logger.Level > 0 {
		hash, err = util.SendToIPFS(fileName, logger.Writer)
	} else {
		hash, err = util.SendToIPFS(fileName, bytes.NewBuffer([]byte{}))
	}

	if err != nil {
		return "", err
	}

	return hash, nil
}
