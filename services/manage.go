package services

import (
  "fmt"
  "regexp"
  // "strings"

  "github.com/eris-ltd/eris-cli/util"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// install
func Install(cmd *cobra.Command, args []string) {

}

// Updates an installed service, or installs it if it has not been installed.
func Update(cmd *cobra.Command, args []string) {

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

func ListRunningRaw() []string {
  services := []string{}
  r := regexp.MustCompile(`\/eris_service_(.+)_\S+?_\d`)

  contns, _ := util.DockerClient.ListContainers(docker.ListContainersOptions{All: false})
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

func ListInstalled() {

}

func Rm(cmd *cobra.Command, args []string) {

}

// endpoint := "tcp://[ip]:[port]"
// path := os.Getenv("DOCKER_CERT_PATH")
// ca := fmt.Sprintf("%s/ca.pem", path)
// cert := fmt.Sprintf("%s/cert.pem", path)
// key := fmt.Sprintf("%s/key.pem", path)
// client, _ := docker.NewTLSClient(endpoint, cert, key, ca)