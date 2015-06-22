package services

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// install
func Install(cmd *cobra.Command, args []string) {
	checkServiceGiven(args)
	if len(args) != 2 {
		fmt.Println("Please give me: eris services install [name] [location]")
		return
	}
	err := InstallServiceRaw(args[0], args[1], cmd.Flags().Lookup("verbose").Changed)
	if err != nil {
		fmt.Println(err)
	}
}

func New(cmd *cobra.Command, args []string) {

}

func Edit(cmd *cobra.Command, args []string) {
	checkServiceGiven(args)
	EditServiceRaw(args[0])
}

func Rename(cmd *cobra.Command, args []string) {
	checkServiceGiven(args)
	if len(args) != 2 {
		fmt.Println("Please give me: eris services rename [oldName] [newName]")
		return
	}
	RenameServiceRaw(args[0], args[1], cmd.Flags().Lookup("verbose").Changed)
}

func Inspect(cmd *cobra.Command, args []string) {
	checkServiceGiven(args)
	if len(args) == 1 {
		args = append(args, "all")
	}
	InspectServiceRaw(args[0], args[1], cmd.Flags().Lookup("verbose").Changed)
}

func Export(cmd *cobra.Command, args []string) {
	checkServiceGiven(args)
	err := ExportServiceRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
	if err != nil {
		fmt.Println(err)
	}
}

// Updates an installed service, or installs it if it has not been installed.
func Update(cmd *cobra.Command, args []string) {
	checkServiceGiven(args)
	UpdateServiceRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
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
	checkServiceGiven(args)
	RmServiceRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func InstallServiceRaw(servName, servPath string, verbose bool) error {
	fileName := filepath.Join(ServicesPath, servName)
	if filepath.Ext(fileName) == "" {
		fileName = fileName + ".toml"
	}

	s := strings.Split(servPath, ":")
	if s[0] == "ipfs" {

		var err error
		if verbose {
			err = util.GetFromIPFS(s[1], fileName, os.Stdout)
		} else {
			err = util.GetFromIPFS(s[1], fileName, bytes.NewBuffer([]byte{}))
		}

		if err != nil {
			return err
		}
		return nil
	}

	if strings.Contains(s[0], "github") {
		fmt.Println("https://twitter.com/ryaneshea/status/595957712040628224")
		return nil
	}

	fmt.Println("I do not know how to get that file. Sorry.")
	return nil
}

func EditServiceRaw(servName string) {
	Editor(servDefFileByServName(servName))
}

func RenameServiceRaw(oldName, newName string, verbose bool) {
	if parseKnown(oldName) {
		if verbose {
			fmt.Println("Renaming service", oldName, "to", newName)
		}

		serviceDef := LoadServiceDefinition(oldName)
		perform.DockerRename(serviceDef.Service, serviceDef.Operations, oldName, newName, verbose)
		oldFile := servDefFileByServName(oldName)
		newFile := strings.Replace(oldFile, oldName, newName, 1)

		serviceDef.Service.Name = newName
		_ = WriteServiceDefinitionFile(serviceDef, newFile)

		data.RenameDataRaw(oldName, newName, verbose)

		os.Remove(oldFile)
	} else {
		if verbose {
			fmt.Println("I cannot find that service. Please check the service name you sent me.")
		}
	}
}

func InspectServiceRaw(servName, field string, verbose bool) {
	service := LoadServiceDefinition(servName)
	InspectServiceByService(service.Service, service.Operations, field, verbose)
}

func ExportServiceRaw(servName string, verbose bool) error {
	if parseKnown(servName) {
		ipfsService := LoadServiceDefinition("ipfs")

		if IsServiceRunning(ipfsService.Service) {
			if verbose {
				fmt.Println("IPFS is running. Adding now.")
			}

			hash, err := exportFile(servName, verbose)
			if err != nil {
				return err
			}
			fmt.Println(hash)
		} else {
			if verbose {
				fmt.Println("IPFS is not running. Starting now.")
			}
			StartServiceByService(ipfsService.Service, ipfsService.Operations, verbose)

			hash, err := exportFile(servName, verbose)
			if err != nil {
				return err
			}
			fmt.Println(hash)
		}

	} else {
		return fmt.Errorf(`I don't known of that service.
Please retry with a known service.
To find known services use: eris services known`)
	}
	return nil
}

func InspectServiceByService(srv *def.Service, ops *def.ServiceOperation, field string, verbose bool) {
	if IsServiceExisting(srv) {
		perform.DockerInspect(srv, ops, field, verbose)
	} else {
		if verbose {
			fmt.Println("No service matching that name.")
		}
	}
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

func UpdateServiceRaw(servName string, verbose bool) {
	service := LoadServiceDefinition(servName)
	perform.DockerRebuild(service.Service, service.Operations, true, verbose)
}

func RmServiceRaw(servName string, verbose bool) {
	service := LoadServiceDefinition(servName)
	perform.DockerRemove(service.Service, service.Operations, verbose)
	// oldFile := servDefFileByServName(servName)
	// os.Remove(oldFile)
}

func exportFile(servName string, verbose bool) (string, error) {
	fileName := servDefFileByServName(servName)

	var err error
	var hash string
	if verbose {
		hash, err = util.SendToIPFS(fileName, os.Stdout)
	} else {
		hash, err = util.SendToIPFS(fileName, bytes.NewBuffer([]byte{}))
	}

	if err != nil {
		return "", err
	}

	return hash, nil
}
