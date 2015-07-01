package services

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/data"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// install
func Import(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	if len(args) != 2 {
		cmd.Help()
		return
	}
	IfExit(ImportServiceRaw(args[0], args[1]))
}

func New(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	if len(args) != 2 {
		cmd.Help()
		return
	}
	IfExit(NewServiceRaw(args[0], args[1]))
}

func Edit(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	IfExit(EditServiceRaw(args[0]))
}

func Rename(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	if len(args) != 2 {
		cmd.Help()
		return
	}
	IfExit(RenameServiceRaw(args[0], args[1]))
}

func Inspect(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	if len(args) == 1 {
		args = append(args, "all")
	}
	IfExit(InspectServiceRaw(args[0], args[1]))
}

func Export(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	IfExit(ExportServiceRaw(args[0]))
}

// Updates an installed service, or installs it if it has not been installed.
func Update(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	IfExit(UpdateServiceRaw(args[0], cmd.Flags().Lookup("skip-pull").Changed))
}

// list known
func ListKnown() {
	services := ListKnownRaw()
	for _, s := range services {
		fmt.Println(s)
	}
}

func ListRunning() {
	services := ListRunningRaw()
	for _, s := range services {
		fmt.Println(s)
	}
}

func ListExisting() {
	services := ListExistingRaw()
	for _, s := range services {
		fmt.Println(s)
	}
}

func Rm(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	IfExit(RmServiceRaw(args, cmd.Flags().Lookup("force").Changed))
}

func ImportServiceRaw(servName, servPath string) error {
	fileName := filepath.Join(ServicesPath, servName)
	if filepath.Ext(fileName) == "" {
		fileName = fileName + ".toml"
	}

	s := strings.Split(servPath, ":")
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

	return fmt.Errorf("I do not know how to get that file. Sorry.")
}

func NewServiceRaw(servName, imageName string) error {
	srv := &def.Service{
		Name:  servName,
		Image: imageName,
	}
	srvDef := &def.ServiceDefinition{
		Service:    srv,
		Maintainer: &def.Maintainer{},
		Location:   &def.Location{},
		Machine:    &def.Machine{},
	}

	err := WriteServiceDefinitionFile(srvDef, "")
	if err != nil {
		return err
	}
	return nil
}

func EditServiceRaw(servName string) error {
	servDefFile, err := servDefFileByServName(servName)
	if err != nil {
		return err
	}
	return Editor(servDefFile)
}

func RenameServiceRaw(oldName, newName string) error {
	if oldName == newName {
		return fmt.Errorf("Cannot rename to same name")
	}

	newNameBase := strings.Replace(newName, filepath.Ext(newName), "", 1)
	transformOnly := newNameBase == oldName

	if parseKnown(oldName) {
		logger.Infoln("Renaming service", oldName, "to", newName)

		serviceDef, err := LoadServiceDefinition(oldName)
		if err != nil {
			return err
		}

		if !transformOnly {
			err = perform.DockerRename(serviceDef.Service, serviceDef.Operations, oldName, newName)
			if err != nil {
				return err
			}
		}

		oldFile, err := servDefFileByServName(oldName)
		if err != nil {
			return err
		}

		if filepath.Base(oldFile) == newName {
			logger.Infoln("Those are the same file. Not renaming")
			return nil
		}

		var newFile string
		if filepath.Ext(newName) == "" {
			newFile = strings.Replace(oldFile, oldName, newName, 1)
		} else {
			newFile = filepath.Join(ServicesPath, newName)
		}

		serviceDef.Service.Name = strings.Replace(newName, filepath.Ext(newName), "", 1)
		err = WriteServiceDefinitionFile(serviceDef, newFile)
		if err != nil {
			return err
		}

		if !transformOnly {
			err = data.RenameDataRaw(oldName, newName)
			if err != nil {
				return err
			}
		}

		os.Remove(oldFile)
	} else {
		return fmt.Errorf("I cannot find that service. Please check the service name you sent me.")
	}

	return nil
}

func InspectServiceRaw(servName, field string) error {
	service, err := LoadServiceDefinition(servName)
	if err != nil {
		return err
	}
	err = InspectServiceByService(service.Service, service.Operations, field)
	if err != nil {
		return err
	}
	return nil
}

func ExportServiceRaw(servName string) error {
	if parseKnown(servName) {
		ipfsService, err := LoadServiceDefinition("ipfs")
		if err != nil {
			return err
		}

		if IsServiceRunning(ipfsService.Service) {
			logger.Infoln("IPFS is running. Adding now.")

			hash, err := exportFile(servName)
			if err != nil {
				return err
			}

			logger.Errorln(hash)
		} else {
			logger.Infoln("IPFS is not running. Starting now.")
			err := StartServiceByService(ipfsService.Service, ipfsService.Operations)
			if err != nil {
				return err
			}

			hash, err := exportFile(servName)
			if err != nil {
				return err
			}

			logger.Errorln(hash)
		}

	} else {
		return fmt.Errorf(`I don't known of that service.
Please retry with a known service.
To find known services use: eris services known
`)
	}
	return nil
}

func InspectServiceByService(srv *def.Service, ops *def.ServiceOperation, field string) error {
	if IsServiceExisting(srv) {
		err := perform.DockerInspect(srv, ops, field)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("No service matching that name.\n")
	}
	return nil
}

func ListKnownRaw() []string {
	srvs := []string{}
	fileTypes := []string{}
	for _, t := range []string{"*.json", "*.yaml", "*.toml"} {
		fileTypes = append(fileTypes, filepath.Join(ServicesPath, t))
	}
	for _, t := range fileTypes {
		s, _ := filepath.Glob(t)
		for _, s1 := range s {
			s1 = strings.Split(filepath.Base(s1), ".")[0]
			srvs = append(srvs, s1)
		}
	}
	return srvs
}

func ListRunningRaw() []string {
	return listServices(false)
}

func ListExistingRaw() []string {
	return listServices(true)
}

func UpdateServiceRaw(servName string, skipPull bool) error {
	service, err := LoadServiceDefinition(servName)
	if err != nil {
		return err
	}
	err = perform.DockerRebuild(service.Service, service.Operations, skipPull)
	if err != nil {
		return err
	}
	return nil
}

func RmServiceRaw(servNames []string, force bool) error {
	for _, servName := range servNames {
		service, err := LoadServiceDefinition(servName)
		if err != nil {
			return err
		}
		err = perform.DockerRemove(service.Service, service.Operations)
		if err != nil {
			return err
		}
		if force {
			oldFile, err := servDefFileByServName(servName)
			if err != nil {
				return err
			}
			if err := os.Remove(oldFile); err != nil {
				return err
			}
		}
	}
	return nil
}

func exportFile(servName string) (string, error) {
	fileName, err := servDefFileByServName(servName)
	if err != nil {
		return "", err
	}

	var hash string
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
