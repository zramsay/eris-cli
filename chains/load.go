package chains

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

//----------------------------------------------------------------------
// config

// viper read config file, marshal to definition struct,
// load service, validate name and data container
func LoadChainDefinition(chainName string, containerNumber int) (*def.Chain, error) {
	var chain def.Chain
	chainConf, err := loadChainDefinition(chainName)
	if err != nil {
		return nil, err
	}

	// marshal chain and always reset the operational requirements
	// this will make sure to sync with docker so that if changes
	// have occured in the interim they are caught.
	marshalChainDefinition(chainConf, &chain)
	chain.Operations = &def.ServiceOperation{}

	var serv *def.ServiceDefinition
	serv, err = services.LoadServiceDefinition(ErisChainType, containerNumber)
	if err != nil {
		return nil, err
	}

	if serv == nil {
		return nil, fmt.Errorf("I do not have the chain definition file available for that chain type.\n")
	}

	if chain.Service == nil {
		chain.Service = serv.Service
	} else {
		mergeChainAndService(&chain, serv.Service)
	}

	chain.Maintainer = serv.Maintainer
	chain.Location = serv.Location
	chain.Machine = serv.Machine

	chain.Operations.ContainerNumber = containerNumber
	checkChainHasUniqueName(&chain)
	checkDataContainerTurnedOn(&chain, chainConf)
	checkDataContainerHasName(chain.Operations)

	return &chain, nil
}

func IsChainExisting(chain *def.Chain) bool {
	return isRunningChain(chain.Service.Name, true)
}

func IsChainRunning(chain *def.Chain) bool {
	return isRunningChain(chain.Service.Name, false)
}

func ServiceDefFromChain(chain *def.Chain) *def.ServiceDefinition {
	chainID := chain.ChainID
	srv := chain.Service

	srv.Name = chainID
	// set the main command
	srv.Command = ErisChainInstall
	// TODO mint vs. erisdb (in terms of rpc)
	srv.Environment = append(chain.Service.Environment, "CHAIN_ID="+chainID)
	return &def.ServiceDefinition{
		Service:    srv,
		Maintainer: chain.Maintainer,
		Location:   chain.Location, // TODO
		Machine:    chain.Machine,
	}
}

// read the config file into viper
func loadChainDefinition(chainName string) (*viper.Viper, error) {
	return util.LoadViperConfig(path.Join(BlockchainsPath), chainName, "chain")
}

// marshal from viper to definitions struct
func marshalChainDefinition(chainConf *viper.Viper, chain *def.Chain) error {
	err := chainConf.Marshal(chain)
	return fmt.Errorf("Error marshalling from viper to chain def: %v", err)
}

// get the config file's path from the chain name
func configFileNameFromChainName(chainName string) (string, error) {
	chainConf, err := loadChainDefinition(chainName)
	if err != nil {
		return "", err
	}
	return chainConf.ConfigFileUsed(), nil
}

//----------------------------------------------------------------------
// chain defs overwrite service defs

// overwrite service attributes with chain config
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

//----------------------------------------------------------------------
// validation funcs

func checkChainHasUniqueName(chain *def.Chain) {
	chain.Operations.SrvContainerName = fmt.Sprintf("eris_chain_%s", util.NameAndNumber(chain.Name, chain.Operations.ContainerNumber))
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

//----------------------------------------------------------------------
// chain lists, lookups

// list all running eris_chain docker containers
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

// check if given chain is running
func isRunningChain(name string, all bool) bool {
	running := listChains(all)
	for _, srv := range running {
		if srv == name {
			return true
		}
	}
	return false
}

// check if given chain is known
func isKnownChain(name string) bool {
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
