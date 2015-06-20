package services

import (
  "fmt"
  "regexp"

  "github.com/eris-ltd/eris-cli/perform"
  "github.com/eris-ltd/eris-cli/util"

  def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"
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
  imgs, _ := util.DockerClient.ListImages(docker.ListImagesOptions{All: false})
  for _, img := range imgs {
    fmt.Println("ID: ", img.ID)
    fmt.Println("RepoTags: ", img.RepoTags)
    fmt.Println("Created: ", img.Created)
    fmt.Println("Size: ", img.Size)
    fmt.Println("VirtualSize: ", img.VirtualSize)
    fmt.Println("ParentId: ", img.ParentID)
  }
}

// Updates an installed service, or installs it if it has not been installed.
func Update(cmd *cobra.Command, args []string) {
  checkServiceGiven(args)
  UpdateServiceRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

// list known
func ListKnown() {

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
  }}

func Rm(cmd *cobra.Command, args []string) {

}

func ListRunningRaw() []string {
  return listServices(false)
}

func ListExistingRaw() []string {
  return listServices(true)
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