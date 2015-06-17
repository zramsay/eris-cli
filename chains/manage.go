package chains

import (
  "fmt"
  "regexp"

  "github.com/eris-ltd/eris-cli/util"
  "github.com/eris-ltd/eris-cli/perform"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Install(cmd *cobra.Command, args []string) {

}

func New(cmd *cobra.Command, args []string) {

}

func Add(cmd *cobra.Command, args []string) {

}

func Config(cmd *cobra.Command, args []string) {

}


func Inspect(cmd *cobra.Command, args []string) {

}

func ListKnown() {

}

func ListInstalled() {
  listChains(true)
}

func ListChains() {
  chains := ListExistingRaw()
  for _, s := range chains {
    fmt.Println(s)
  }
}

func ListRunning() {
  chains := ListRunningRaw()
  for _, s := range chains {
    fmt.Println(s)
  }
}

func Rename(cmd *cobra.Command, args []string) {

}

func Remove(cmd *cobra.Command, args []string) {

}

func Update(cmd *cobra.Command, args []string) {
  checkChainGiven(args)
  UpdateChainRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func Rm(cmd *cobra.Command, args []string) {

}

func ListRunningRaw() []string {
  return listChains(false)
}

func ListExistingRaw() []string {
  return listChains(true)
}

func IsChainExisting(chain *util.Chain) bool {
  return parseChains(util.FullNameToShort(chain.Service.Name), true)
}

func IsChainRunning(chain *util.Chain) bool {
  return parseChains(util.FullNameToShort(chain.Service.Name), false)
}

func UpdateChainRaw(chainName string, verbose bool) {
  chain := LoadChainDefinition(chainName)
  perform.DockerRebuild(chain.Service, verbose)
}

func listChains(running bool) []string {
  chains := []string{}
  r := regexp.MustCompile(`\/eris_chain_(.+)_\d`)

  contns, _ := util.DockerClient.ListContainers(docker.ListContainersOptions{All: running})
  for _, con := range contns {
    for _, c := range con.Names {
      match := r.FindAllStringSubmatch(c, 1)
      if len(match) != 0 {
        chains = append(chains, r.FindAllStringSubmatch(c, 1)[0][1])
      }
    }
  }

  return chains
}

func parseChains(name string, all bool) bool {
  running := listChains(all)
  if len(running) != 0 {
    for _, srv := range running {
      if srv == name {
        return true
      }
    }
  }
  return false
}