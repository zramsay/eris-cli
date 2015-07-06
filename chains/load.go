package chains

import (
	"fmt"
	"path"

	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

//----------------------------------------------------------------------
// config

// viper read config file, marshal to definition struct,
// load service, validate name and data container
func LoadChainDefinition(chainName string, cNum ...int) (*def.Chain, error) {
	chain := def.BlankChain()
	chainConf, err := loadChainDefinition(chainName)
	if err != nil {
		return nil, err
	}

	// marshal chain and always reset the operational requirements
	// this will make sure to sync with docker so that if changes
	// have occured in the interim they are caught.
	if err = marshalChainDefinition(chainConf, chain); err != nil {
		return nil, err
	}

	serv := def.BlankServiceDefinition()
	serv, err = services.LoadServiceDefinition(ErisChainType, cNum...)
	if err != nil {
		return nil, err
	}

	if serv == nil {
		return nil, fmt.Errorf("I do not have the chain definition file available for that chain type.\n")
	}

	if chain.Service == nil {
		chain.Service = serv.Service
	} else {
		mergeChainAndService(chain, serv.Service)
	}

	// TODO -> pull these from tool level configs
	// chain.Maintainer = serv.Maintainer
	// chain.Location = serv.Location
	// chain.Machine = serv.Machine

	if len(cNum) == 0 {
		// TODO: findNextContainerIndex => util/container_operations.go
		chain.Operations.ContainerNumber = 1
	} else {
		chain.Operations.ContainerNumber = cNum[0]
	}

	checkNames(chain)

	return chain, nil
}

func MockChainDefinition(chainName, chainID string, cNum ...int) *def.Chain {
	chn := def.BlankChain()
	chn.Name = chainName
	chn.ChainID = chainID
	chn.Service.AutoData = true

	if len(cNum) == 0 {
		// TODO: findNextContainerIndex => util/container_operations.go
		chn.Operations.ContainerNumber = 1
	} else {
		chn.Operations.ContainerNumber = cNum[0]
	}

	checkNames(chn)
	return chn
}

func IsChainExisting(chain *def.Chain) bool {
	return util.IsChainContainer(chain.Name, chain.Operations.ContainerNumber, true)
}

func IsChainRunning(chain *def.Chain) bool {
	return util.IsChainContainer(chain.Name, chain.Operations.ContainerNumber, false)
}

func ServiceDefFromChain(chain *def.Chain) *def.ServiceDefinition {
	chainID := chain.ChainID
	srv := chain.Service

	srv.Name = chain.Name // this let's the data containers flow thru
	srv.Image = "eris/erisdb:" + version.VERSION
	srv.AutoData = true // default. they can turn it off. it's like BarBri
	srv.Command = ErisChainStart
	srv.Environment = append(chain.Service.Environment, "CHAIN_ID="+chainID)
	// TODO mint vs. erisdb (in terms of rpc) --> think we default them to erisdb's REST/Stream API

	return &def.ServiceDefinition{
		Name:        chain.Name,
		ServiceID:   chain.ChainID,
		ServiceDeps: []string{"keys"},
		Service:     srv,
		Maintainer:  chain.Maintainer,
		Location:    chain.Location,
		Machine:     chain.Machine,
	}
}

// read the config file into viper
func loadChainDefinition(chainName string) (*viper.Viper, error) {
	return util.LoadViperConfig(path.Join(BlockchainsPath), chainName, "chain")
}

// marshal from viper to definitions struct
func marshalChainDefinition(chainConf *viper.Viper, chain *def.Chain) error {
	err := chainConf.Marshal(chain)
	if err != nil {
		return fmt.Errorf("The marmots coult not marshal from viper to chain def: %v", err)
	}

	// toml bools don't really marshal well
	// data_container can be in the chain or
	// in the service layer. this is very
	// opinionated. we know.
	for _, s := range []string{"", "service."} {
		if chainConf.GetBool(s + "data_container") {
			logger.Debugln("Data Containers Turned On.")
			chain.Service.AutoData = true
		}
	}

	return nil
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
	chain.Service.Links = overWriteSlice(chain.Service.Links, service.Links)
	chain.Service.Ports = overWriteSlice(chain.Service.Ports, service.Ports)
	chain.Service.Expose = overWriteSlice(chain.Service.Expose, service.Expose)
	chain.Service.Volumes = overWriteSlice(chain.Service.Volumes, service.Volumes)
	chain.Service.VolumesFrom = overWriteSlice(chain.Service.VolumesFrom, service.VolumesFrom)
	chain.Service.Environment = mergeSlice(chain.Service.Environment, service.Environment)
	chain.Service.EnvFile = overWriteSlice(chain.Service.EnvFile, service.EnvFile)
	chain.Service.Net = overWriteString(chain.Service.Net, service.Net)
	chain.Service.PID = overWriteString(chain.Service.PID, service.PID)
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

func checkNames(chain *def.Chain) {
	chain.Operations.SrvContainerName = util.ChainContainersName(chain.Name, chain.Operations.ContainerNumber)
	chain.Operations.DataContainerName = util.DataContainersName(chain.Name, chain.Operations.ContainerNumber)
}

//----------------------------------------------------------------------
// chain lists, lookups

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
