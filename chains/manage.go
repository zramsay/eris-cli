package chains

import (
  "fmt"
  "regexp"
  "strings"

  "github.com/eris-ltd/eris-cli/util"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Install(cmd *cobra.Command, args []string) {

}

func Add(cmd *cobra.Command, args []string) {

}

func ListChains() {

}

func ListKnown() {

}

func ListInstalled() {

}

func ListRunning() {
  chains := ListRunningRaw()
  for _, s := range chains {
    fmt.Println(s)
  }
}

func ListRunningRaw() []string {
  chains := []string{}
  r := regexp.MustCompile(`\/eris_chain_(.+)_\S+?_\d`)

  contns, _ := util.DockerClient.ListContainers(docker.ListContainersOptions{All: false})
  for _, con := range contns {
    for _, c := range con.Names {
      if strings.Contains(c, "/eris_chain_"){
        chains = append(chains, r.FindAllStringSubmatch(c, -1)[0][1])
      }
    }
  }

  return chains
}

func Rename(cmd *cobra.Command, args []string) {

}

func Remove(cmd *cobra.Command, args []string) {

}

func Update(cmd *cobra.Command, args []string) {

}

func Clean(cmd *cobra.Command, args []string) {

}
