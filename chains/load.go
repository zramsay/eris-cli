package chains

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func LoadChainDefinition(chainName string) *def.Chain {
	var chain def.Chain
	chainConf := loadChainDefinition(chainName)

	// marshal chain and always reset the operational requirements
	// this will make sure to sync with docker so that if changes
	// have occured in the interim they are caught.
	marshalChainDefinition(chainConf, &chain)
	chain.Operations = &def.ServiceOperation{}

	serv := services.LoadService(chain.Type)
	mergeChainAndService(&chain, serv)

	checkChainHasUniqueName(&chain)
	checkDataContainerTurnedOn(&chain, chainConf)
	checkDataContainerHasName(chain.Operations)

	return &chain
}

func IsChainExisting(chain *def.Chain) bool {
	return parseChains(chain.Service.Name, true)
}

func IsChainRunning(chain *def.Chain) bool {
	return parseChains(chain.Service.Name, false)
}

func loadChainDefinition(chainName string) *viper.Viper {
	var chainConf = viper.New()

	chainConf.AddConfigPath(dir.BlockchainsPath)
	chainConf.SetConfigName(chainName)
	chainConf.ReadInConfig()

	return chainConf
}

func marshalChainDefinition(chainConf *viper.Viper, chain *def.Chain) {
	err := chainConf.Marshal(chain)
	if err != nil {
		// TODO: error handling
		fmt.Println(err)
		os.Exit(1)
	}
}

func chainDefFileByChainName(chainName string) string {
	chainConf := loadChainDefinition(chainName)
	return chainConf.ConfigFileUsed()
}

func mergeChainAndService(chain *def.Chain, service *def.Service) {
	chain.Service.Name = chain.Name
	chain.Service.Image = overWriteString(chain.Service.Image, service.Image)
	chain.Service.Command = overWriteString(chain.Service.Command, service.Command)
	chain.Service.ServiceDeps = overWriteSlice(chain.Service.ServiceDeps, service.ServiceDeps)
	chain.Service.Labels = mergeMap(chain.Service.Labels, service.Labels)
	chain.Service.Links = overWriteSlice(chain.Service.Links, service.Links)
	chain.Service.Ports = overWriteSlice(chain.Service.Ports, service.Ports)
	chain.Service.Expose = overWriteSlice(chain.Service.Expose, service.Expose)
	chain.Service.Volumes = overWriteSlice(chain.Service.Volumes, service.Volumes)
	chain.Service.VolumesFrom = overWriteSlice(chain.Service.VolumesFrom, service.VolumesFrom)
	chain.Service.Environment = mergeSlice(chain.Service.Environment, service.Environment)
	chain.Service.EnvFile = overWriteSlice(chain.Service.EnvFile, service.EnvFile)
	chain.Service.Net = overWriteString(chain.Service.Net, service.Net)
	chain.Service.PID = overWriteString(chain.Service.PID, service.PID)
	chain.Service.CapAdd = overWriteSlice(chain.Service.CapAdd, service.CapAdd)
	chain.Service.CapDrop = overWriteSlice(chain.Service.CapDrop, service.CapDrop)
	chain.Service.DNS = overWriteSlice(chain.Service.DNS, service.DNS)
	chain.Service.DNSSearch = overWriteSlice(chain.Service.DNSSearch, service.DNSSearch)
	chain.Service.CPUShares = overWriteInt64(chain.Service.CPUShares, service.CPUShares)
	chain.Service.WorkDir = overWriteString(chain.Service.WorkDir, service.WorkDir)
	chain.Service.EntryPoint = overWriteString(chain.Service.EntryPoint, service.EntryPoint)
	chain.Service.HostName = overWriteString(chain.Service.HostName, service.HostName)
	chain.Service.DomainName = overWriteString(chain.Service.DomainName, service.DomainName)
	chain.Service.User = overWriteString(chain.Service.User, service.User)
	chain.Service.MemLimit = overWriteInt64(chain.Service.MemLimit, service.MemLimit)
}

func overWriteBool(trumpEr, toOver bool) bool {
	if trumpEr {
		return trumpEr
	}
	return toOver
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

func checkChainGiven(args []string) {
	if len(args) == 0 {
		fmt.Println("No ChainName Given. Please rerun command with a known chain.")
		os.Exit(1)
	}
}

func checkChainHasUniqueName(chain *def.Chain) {
	containerNumber := 1 // tmp
	chain.Operations.SrvContainerName = "eris_chain_" + chain.Name + "_" + strconv.Itoa(containerNumber)
}

func checkDataContainerTurnedOn(chain *def.Chain, chainConf *viper.Viper) {
	// toml bools don't really marshal well
	if chainConf.GetBool("service.data_container") {
		chain.Service.AutoData = true
		chain.Operations.DataContainer = true
	}
}

func checkDataContainerHasName(ops *def.ServiceOperation) {
	if ops.DataContainer {
		ops.DataContainerName = ""
		if ops.DataContainer {
			dataSplit := strings.Split(ops.SrvContainerName, "_")
			dataSplit[1] = "data"
			ops.DataContainerName = strings.Join(dataSplit, "_")
		}
	}
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

func parseKnown(name string) bool {
	known := ListKnownRaw()
	if len(known) != 0 {
		for _, srv := range known {
			if srv == name {
				return true
			}
		}
	}
	return false
}
