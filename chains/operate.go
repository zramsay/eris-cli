package chains

import (
  "fmt"
  "os"
  "path"
  "strings"
  "strconv"

  "github.com/eris-ltd/eris-cli/perform"
  "github.com/eris-ltd/eris-cli/util"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func New(cmd *cobra.Command, args []string) {

}

func Checkout(cmd *cobra.Command, args []string) {

}

func Config(cmd *cobra.Command, args []string) {

}

func Start(cmd *cobra.Command, args []string) {
  if len(args) == 0 {
    fmt.Println("No ChainName Given. Please rerun command with a known chain.")
    os.Exit(1)
  }

  chain := loadChainConf(args[0])

  // If no name use image name
  err := os.Chdir(path.Join(util.BlockchainsPath, "raw", chain.Directory))
  if err != nil {
    // TODO: error handling
    fmt.Println(err)
    os.Exit(1)
  }

  perform.DockerRun(chain.Service, cmd.Flags().Lookup("verbose").Changed)
}

func Logs(cmd *cobra.Command, args []string) {

}

func Kill(cmd *cobra.Command, args []string) {
  if len(args) == 0 {
    fmt.Println("No ChainName Given. Please rerun command with a known chain.")
    os.Exit(1)
  }

  chain := loadChainConf(args[0])

  // If no name use image name
  err := os.Chdir(path.Join(util.BlockchainsPath, "raw", chain.Directory))
  if err != nil {
    // TODO: error handling
    fmt.Println(err)
    os.Exit(1)
  }

  perform.DockerStop(chain.Service, cmd.Flags().Lookup("verbose").Changed)
}

func loadChainConf(name string) *util.Chain {
  var chain util.Chain
  var chainConf = viper.New()

  chainConf.AddConfigPath(util.BlockchainsPath)
  chainConf.SetConfigName(name)
  chainConf.ReadInConfig()

  err := chainConf.Marshal(&chain)
  if err != nil {
    // TODO: error handling
    fmt.Println(err)
    os.Exit(1)
  }

  serv  := loadServiceConf(chain.Type)
  mergeChainAndService(&chain, serv)

  // Services must be given an image. Flame out if they do not.
  if chain.Service.Image == "" {
    fmt.Println("An \"image\" field is required in the service definition file in the chain_type.")
    os.Exit(1)
  }

  return &chain
}

func loadServiceConf(typ string) *util.Service {
  var service util.Service
  var srvConf = viper.New()

  srvConf.AddConfigPath(util.ServicesPath)
  srvConf.SetConfigName(typ)
  srvConf.ReadInConfig()

  err := srvConf.Marshal(&service)
  if err != nil {
    // TODO: error handling
    fmt.Println(err)
    os.Exit(1)
  }
  return &service
}

func mergeChainAndService(chain *util.Chain, service *util.Service) {
  containerNumber := 1 // tmp
  chain.Service.Name        = "eris_chain_" + chain.Name + "_" + strings.Split(uuid.New(), "-")[0] + "_" + strconv.Itoa(containerNumber)
  chain.Service.Image       = overWriteString(chain.Service.Image, service.Image)
  chain.Service.Command     = overWriteString(chain.Service.Command, service.Command)
  chain.Service.Labels      = mergeMap(chain.Service.Labels, service.Labels)
  chain.Service.Links       = overWriteSlice(chain.Service.Links, service.Links)
  chain.Service.Ports       = overWriteSlice(chain.Service.Ports, service.Ports)
  chain.Service.Expose      = overWriteSlice(chain.Service.Expose, service.Expose)
  chain.Service.Volumes     = overWriteSlice(chain.Service.Volumes, service.Volumes)
  chain.Service.VolumesFrom = overWriteSlice(chain.Service.VolumesFrom, service.VolumesFrom)
  chain.Service.Environment = mergeSlice(chain.Service.Environment, service.Environment)
  chain.Service.EnvFile     = overWriteSlice(chain.Service.EnvFile, service.EnvFile)
  chain.Service.Net         = overWriteString(chain.Service.Net, service.Net)
  chain.Service.PID         = overWriteString(chain.Service.PID, service.PID)
  chain.Service.CapAdd      = overWriteSlice(chain.Service.CapAdd, service.CapAdd)
  chain.Service.CapDrop     = overWriteSlice(chain.Service.CapDrop, service.CapDrop)
  chain.Service.DNS         = overWriteSlice(chain.Service.DNS, service.DNS)
  chain.Service.DNSSearch   = overWriteSlice(chain.Service.DNSSearch, service.DNSSearch)
  chain.Service.CPUShares   = overWriteInt64(chain.Service.CPUShares, service.CPUShares)
  chain.Service.WorkDir     = overWriteString(chain.Service.WorkDir, service.WorkDir)
  chain.Service.EntryPoint  = overWriteString(chain.Service.EntryPoint, service.EntryPoint)
  chain.Service.HostName    = overWriteString(chain.Service.HostName, service.HostName)
  chain.Service.DomainName  = overWriteString(chain.Service.DomainName, service.DomainName)
  chain.Service.User        = overWriteString(chain.Service.User, service.User)
  chain.Service.MemLimit    = overWriteInt64(chain.Service.MemLimit, service.MemLimit)
}

func overWriteString(trumpEr, toOver string) string {
  if trumpEr != "" {
    return trumpEr
  }
  return toOver
}

func overWriteInt64(trumpEr, toOver int64) int64 {
  if trumpEr != 0 {
    return trumpEr
  }
  return toOver
}

func overWriteSlice(trumpEr, toOver []string) []string {
  if len(trumpEr) != 0 {
    return trumpEr
  }
  return toOver
}

func mergeSlice(mapOne, mapTwo []string) []string {
  for _, v := range mapOne {
    mapTwo = append(mapTwo, v)
  }
  return mapTwo
}

func mergeMap(mapOne, mapTwo map[string]string) map[string]string {
  for k, v := range mapOne {
    mapTwo[k] = v
  }
  return mapTwo
}