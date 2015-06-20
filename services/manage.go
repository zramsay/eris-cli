package services

import (
  "fmt"
  "path/filepath"
  "regexp"
  "strings"

  "github.com/eris-ltd/eris-cli/perform"
  "github.com/eris-ltd/eris-cli/util"

  def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"
  dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// install
func Install(cmd *cobra.Command, args []string) {

}

func New(cmd *cobra.Command, args []string){

}

func Configure(cmd *cobra.Command, args []string) {

}

func Rename(cmd *cobra.Command, args []string) {

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

}

func InspectServiceRaw(servName, field string, verbose bool) {
  service := LoadServiceDefinition(servName)
  InspectServiceByService(service, field, verbose)
}

func InspectServiceByService(service *def.Service, field string, verbose bool) {
  if IsServiceExisting(service) {
    perform.DockerInspect(service, field, verbose)
  } else {
    if verbose {
      fmt.Println("No service matching that name.")
    }
  }
}

func ListRunningRaw() []string {
  return listServices(false)
}

func ListExistingRaw() []string {
  return listServices(true)
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

func IsServiceExisting(service *def.Service) bool {
  return parseServices(util.FullNameToShort(service.Name), true)
}

func IsServiceRunning(service *def.Service) bool {
  return parseServices(util.FullNameToShort(service.Name), false)
}

func UpdateServiceRaw(servName string, verbose bool) {
  service := LoadServiceDefinition(servName)
  perform.DockerRebuild(service, verbose)
}

func parseServices(name string, all bool) bool {
  running := listServices(all)
  if len(running) != 0 {
    for _, srv := range running {
      if srv == name {
        return true
      }
    }
  }
  return false
}

func listServices(running bool) []string {
  services := []string{}
  r := regexp.MustCompile(`\/eris_service_(.+)_\d`)

  contns, _ := util.DockerClient.ListContainers(docker.ListContainersOptions{All: running})
  for _, con := range contns {
    for _, c := range con.Names {
      match := r.FindAllStringSubmatch(c, 1)
      if len(match) != 0 {
        services = append(services, r.FindAllStringSubmatch(c, 1)[0][1])
      }
    }
  }

  return services
}