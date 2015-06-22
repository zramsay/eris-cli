package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/perform"

	dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
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
  InstallServiceRaw(args[0], args[1], cmd.Flags().Lookup("verbose").Changed)
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

func InstallServiceRaw(servName, servPath string, verbose bool) {

  // is it ipfs
  // is it https
  // cannot find fail
}

func EditServiceRaw(servName string) {
	dir.Editor(servDefFileByServName(servName))
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
		fileTypes = append(fileTypes, filepath.Join(dir.ServicesPath, t))
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
	oldFile := servDefFileByServName(servName)
	os.Remove(oldFile)
}
