package loaders

import (
	"fmt"
	"path"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

// viper read config file, marshal to definition struct,
// load service, validate name and data container
func LoadChainDefinition(chainName string, cNum ...int) (*definitions.Chain, error) {
	if len(cNum) == 0 || cNum[0] == 0 {
		logger.Debugf("Loading Service Definition =>\t%s:1 (autoassigned)\n", chainName)
		// TODO: findNextContainerIndex => util/container_operations.go
		if len(cNum) == 0 {
			cNum = append(cNum, 1)
		} else {
			cNum[0] = 1
		}
	} else {
		logger.Debugf("Loading Service Definition =>\t%s:%d\n", chainName, cNum[0])
	}

	chain := definitions.BlankChain()
	chain.Name = chainName
	setChainDefaults(chain)

	chainConf, err := loadChainDefinition(chainName)
	if err != nil {
		return nil, err
	}

	// marshal chain and always reset the operational requirements
	// this will make sure to sync with docker so that if changes
	// have occured in the interim they are caught.
	if err = MarshalChainDefinition(chainConf, chain); err != nil {
		return nil, err
	}

	chain.Operations.ContainerNumber = cNum[0]

	checkChainNames(chain)
	logger.Debugf("Chain Loader. ContNumber =>\t%d\n", chain.Operations.ContainerNumber)
	logger.Debugf("\twith Environment =>\t%d\n", chain.Service.Environment)
	return chain, nil
}

func ServiceDefFromChain(chain *definitions.Chain, cmd string) *definitions.ServiceDefinition {
	chainID := chain.ChainID

	chain.Service.Name = chain.Name // this let's the data containers flow thru
	chain.Service.Image = "eris/erisdb:" + version.VERSION
	chain.Service.AutoData = true // default. they can turn it off. it's like BarBri
	chain.Service.Command = cmd
	chain.Service.Environment = append(chain.Service.Environment, "CHAIN_ID="+chainID)

	// TODO mint vs. erisdb (in terms of rpc) --> think we default them to erisdb's REST/Stream API

	return &definitions.ServiceDefinition{
		Name:        chain.Name,
		ServiceID:   chain.ChainID,
		ServiceDeps: []string{"keys"},
		Service:     chain.Service,
		Operations:  chain.Operations,
		Maintainer:  chain.Maintainer,
		Location:    chain.Location,
		Machine:     chain.Machine,
	}
}

func MockChainDefinition(chainName, chainID string, cNum ...int) *definitions.Chain {
	chn := definitions.BlankChain()
	chn.Name = chainName
	chn.ChainID = chainID
	chn.Service.AutoData = true

	if len(cNum) == 0 {
		// TODO: findNextContainerIndex => util/container_operations.go
		chn.Operations.ContainerNumber = 1
	} else {
		chn.Operations.ContainerNumber = cNum[0]
	}

	checkChainNames(chn)
	return chn
}

// marshal from viper to definitions struct
func MarshalChainDefinition(chainConf *viper.Viper, chain *definitions.Chain) error {
	chnTemp := definitions.BlankChain()
	// logger.Debugf("Loader.Chain: ChainID =>\t\t%v\n", chain.ChainID)
	// logger.Debugf("Loader.Chain. Conf =>\t\t%v\n", chainConf)
	err := chainConf.Marshal(chnTemp)
	if err != nil {
		return fmt.Errorf("The marmots coult not marshal from viper to chain def: %v", err)
	}
	// logger.Debugf("Loader.Chain.Marshal: ChanID =>\t%v\n", chnTemp.ChainID)

	mergeChainAndService(chain, chnTemp.Service)
	chain.ChainID = chnTemp.ChainID

	// toml bools don't really marshal well
	// data_container can be in the chain or
	// in the service layer. this is very
	// opinionated. we know.
	for _, s := range []string{"", "service."} {
		if chainConf.GetBool(s + "data_container") {
			logger.Debugln("Loader.Chain.Marshal: Data Containers Turned On.")
			chain.Service.AutoData = true
		}
	}

	return nil
}

func setChainDefaults(chain *definitions.Chain) {
	ver := strings.Join(strings.Split(version.VERSION, ".")[:2], ".") // only need 0.10 not full ver => 0.10.0
	chain.Service.Image = "eris/erisdb:" + ver
	chain.Service.AutoData = true
	logger.Debugf("Chain Defaults Set. Image =>\t%s\n", chain.Service.Image)
}

// read the config file into viper
func loadChainDefinition(chainName string) (*viper.Viper, error) {
	return util.LoadViperConfig(path.Join(BlockchainsPath), chainName, "chain")
}

//----------------------------------------------------------------------
// validation funcs
func checkChainNames(chain *definitions.Chain) {
	chain.Service.Name = chain.Name
	chain.Operations.SrvContainerName = util.ChainContainersName(chain.Name, chain.Operations.ContainerNumber)
	chain.Operations.DataContainerName = util.DataContainersName(chain.Name, chain.Operations.ContainerNumber)
}

// overwrite service attributes with chain config
// TODO: remove this in favor of using Viper's SetDefaults functionality.
func mergeChainAndService(chain *definitions.Chain, service *definitions.Service) {
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