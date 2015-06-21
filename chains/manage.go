package chains

import (
  "fmt"
  "path/filepath"
  "regexp"
  "strings"

  "github.com/eris-ltd/eris-cli/perform"
  "github.com/eris-ltd/eris-cli/services"
  "github.com/eris-ltd/eris-cli/util"

  def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"
  dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
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
  checkChainGiven(args)
  ConfigureRaw(args[0])
}


func Inspect(cmd *cobra.Command, args []string) {
  checkChainGiven(args)
  if len(args) == 1 {
    args = append(args, "all")
  }
  chain := LoadChainDefinition(args[0])
  if IsChainExisting(chain) {
    services.InspectServiceByService(chain.Service, args[1], cmd.Flags().Lookup("verbose").Changed)
  }
}

func ListKnown() {
  chains := ListKnownRaw()
  for _, s := range chains {
    fmt.Println(s)
  }
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

func ListKnownRaw() []string {
  chns := []string{}
  fileTypes := []string{}
  for _, t := range []string{"*.json", "*.yaml", "*.toml"} {
    fileTypes = append(fileTypes, filepath.Join(dir.BlockchainsPath, t))
  }
  for _, t := range fileTypes {
    s, _ := filepath.Glob(t)
    for _, s1 := range s {
      s1 = strings.Split(filepath.Base(s1), ".")[0]
      chns = append(chns, s1)
    }
  }
  return chns
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

func ConfigureRaw(chainName string) {
  chainConf := loadChainDefinition(chainName)
  filePath := chainConf.ConfigFileUsed()
  dir.Editor(filePath)
}

func ListRunningRaw() []string {
  return listChains(false)
}

func ListExistingRaw() []string {
  return listChains(true)
}

func IsChainExisting(chain *def.Chain) bool {
  return parseChains(util.FullNameToShort(chain.Service.Name), true)
}

func IsChainRunning(chain *def.Chain) bool {
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