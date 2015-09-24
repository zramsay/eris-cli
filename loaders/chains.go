package loaders

import (
	"fmt"
	"path"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

const (
	ErisChainStart    = "run"
	ErisChainStartApi = "api"
	ErisChainInstall  = "install"
	ErisChainNew      = "new"
	ErisChainRegister = "register"
)

// viper read config file, marshal to definition struct,
// load service, validate name and data container
func LoadChainDefinition(chainName string, newCont bool, cNum ...int) (*definitions.Chain, error) {
	if len(cNum) == 0 {
		cNum = append(cNum, 0)
	}

	if cNum[0] == 0 {
		cNum[0] = util.AutoMagic(0, "chain", newCont)
		logger.Debugf("Loading Chain Definition =>\t%s:%d (autoassigned)\n", chainName, cNum[0])
	} else {
		logger.Debugf("Loading Chain Definition =>\t%s:%d\n", chainName, cNum[0])
	}

	chain := definitions.BlankChain()
	chain.Name = chainName
	chain.Operations.ContainerNumber = cNum[0]
	if err := setChainDefaults(chain); err != nil {
		return nil, err
	}

	chainConf, err := config.LoadViperConfig(path.Join(BlockchainsPath), chainName, "chain")
	if err != nil {
		return nil, err
	}

	// marshal chain and always reset the operational requirements
	// this will make sure to sync with docker so that if changes
	// have occured in the interim they are caught.
	if err = MarshalChainDefinition(chainConf, chain); err != nil {
		return nil, err
	}

	// Docker 1.6 (which eris doesn't support) had different linking mechanism.
	if ver, _ := util.DockerClientVersion(); ver >= 1.7 {
		if chain.Dependencies != nil {
			addDependencyVolumesAndLinks(chain.Dependencies, chain.Service, chain.Operations)
		}
	}

	checkChainNames(chain)
	logger.Debugf("Chain Loader. ContNumber =>\t%d\n", chain.Operations.ContainerNumber)
	logger.Debugf("\twith Environment =>\t%v\n", chain.Service.Environment)
	logger.Debugf("\tBooting =>\t\t%v:%v\n", chain.Service.EntryPoint, chain.Service.Command)
	return chain, nil
}

// Convert the chain def to a service def but keep the "eris_chains" containers prefix and set the chain id
func ChainsAsAService(chainName string, newCont bool, cNum ...int) (*definitions.ServiceDefinition, error) {
	chain, err := LoadChainDefinition(chainName, newCont, cNum...)
	if err != nil {
		return nil, err
	}
	s, err := ServiceDefFromChain(chain, ErisChainStart), nil
	if err != nil {
		return nil, err
	}
	// we keep the "eris_chain" prefix and set the CHAIN_ID var.
	// the run command is set in ServiceDefFromChain
	s.Operations.SrvContainerName = util.ChainContainersName(chainName, s.Operations.ContainerNumber)
	s.Service.Environment = append(s.Service.Environment, "CHAIN_ID="+chainName)
	return s, nil
}

func ServiceDefFromChain(chain *definitions.Chain, cmd string) *definitions.ServiceDefinition {
	// chainID := chain.ChainID
	vers := version.VERSION
	setChainDefaults(chain)
	chain.Service.Name = chain.Name // this let's the data containers flow thru
	chain.Service.Image = "quay.io/eris/erisdb:" + vers
	chain.Service.AutoData = true // default. they can turn it off. it's like BarBri
	chain.Service.Command = cmd

	srv := &definitions.ServiceDefinition{
		Name:         chain.Name,
		ServiceID:    chain.ChainID,
		Dependencies: &definitions.Dependencies{Services: []string{"keys"}},
		Service:      chain.Service,
		Operations:   chain.Operations,
		Maintainer:   chain.Maintainer,
		Location:     chain.Location,
		Machine:      chain.Machine,
	}
	ServiceFinalizeLoad(srv) // these are mostly operational considerations that we want to ensure are met

	return srv
}

func ConnectToAChain(srv *definitions.Service, ops *definitions.Operation, name, internalName string, link, mount bool) {
	connectToAService(srv, ops, "chain", name, internalName, link, mount)
}

func MockChainDefinition(chainName, chainID string, newCont bool, cNum ...int) *definitions.Chain {
	chn := definitions.BlankChain()
	chn.Name = chainName
	chn.ChainID = chainID
	chn.Service.AutoData = true

	if len(cNum) == 0 {
		chn.Operations.ContainerNumber = util.AutoMagic(cNum[0], "chain", newCont)
		logger.Debugf("Mocking Chain Definition =>\t%s:%d (autoassigned)\n", chainName, cNum[0])
	} else {
		chn.Operations.ContainerNumber = cNum[0]
		logger.Debugf("Mocking Chain Definition =>\t%s:%d\n", chainName, cNum[0])
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

	mergeDefaultsAndChain(chain, chnTemp.Service)
	chain.ChainID = chnTemp.ChainID

	// toml bools don't really marshal well
	// data_container can be in the chain or
	// in the service layer. this is very
	// opinionated. we know.
	for _, s := range []string{"", "service."} {
		if chainConf.GetBool(s + "data_container") {
			logger.Debugln("Loader.Chain.Marshal. Data Containers Turned On.")
			chain.Service.AutoData = true
		}
	}

	return nil
}

func setChainDefaults(chain *definitions.Chain) error {
	cfg, err := config.LoadViperConfig(path.Join(BlockchainsPath), "default", "chain")
	if err != nil {
		return err
	}
	if err := cfg.Marshal(chain); err != nil {
		return err
	}

	logger.Debugf("Chain Defaults Set. Image =>\t%s\n", chain.Service.Image)
	return nil
}

//----------------------------------------------------------------------
// validation funcs
func checkChainNames(chain *definitions.Chain) {
	chain.Service.Name = chain.Name
	chain.Operations.SrvContainerName = util.ChainContainersName(chain.Name, chain.Operations.ContainerNumber)
	chain.Operations.DataContainerName = util.DataContainersName(chain.Name, chain.Operations.ContainerNumber)
}

// overwrite service attributes with chain config
func mergeDefaultsAndChain(chain *definitions.Chain, service *definitions.Service) {
	chain.Service.Name = chain.Name
	chain.Service.Image = util.OverWriteString(chain.Service.Image, service.Image)
	chain.Service.Command = util.OverWriteString(chain.Service.Command, service.Command)
	chain.Service.Links = util.OverWriteSlice(chain.Service.Links, service.Links)
	chain.Service.Ports = util.OverWriteSlice(chain.Service.Ports, service.Ports)
	chain.Service.Expose = util.OverWriteSlice(chain.Service.Expose, service.Expose)
	chain.Service.Volumes = util.OverWriteSlice(chain.Service.Volumes, service.Volumes)
	chain.Service.VolumesFrom = util.OverWriteSlice(chain.Service.VolumesFrom, service.VolumesFrom)
	chain.Service.Environment = util.MergeSlice(chain.Service.Environment, service.Environment)
	chain.Service.EnvFile = util.OverWriteSlice(chain.Service.EnvFile, service.EnvFile)
	chain.Service.Net = util.OverWriteString(chain.Service.Net, service.Net)
	chain.Service.PID = util.OverWriteString(chain.Service.PID, service.PID)
	chain.Service.DNS = util.OverWriteSlice(chain.Service.DNS, service.DNS)
	chain.Service.DNSSearch = util.OverWriteSlice(chain.Service.DNSSearch, service.DNSSearch)
	chain.Service.CPUShares = util.OverWriteInt64(chain.Service.CPUShares, service.CPUShares)
	chain.Service.WorkDir = util.OverWriteString(chain.Service.WorkDir, service.WorkDir)
	chain.Service.EntryPoint = util.OverWriteString(chain.Service.EntryPoint, service.EntryPoint)
	chain.Service.HostName = util.OverWriteString(chain.Service.HostName, service.HostName)
	chain.Service.DomainName = util.OverWriteString(chain.Service.DomainName, service.DomainName)
	chain.Service.User = util.OverWriteString(chain.Service.User, service.User)
	chain.Service.MemLimit = util.OverWriteInt64(chain.Service.MemLimit, service.MemLimit)
}
