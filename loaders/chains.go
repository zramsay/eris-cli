package loaders

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

	. "github.com/eris-ltd/common/go/common"

	log "github.com/eris-ltd/eris-logger"
	"github.com/spf13/viper"
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
func LoadChainDefinition(chainName string, newCont bool) (*definitions.Chain, error) {

	chain := definitions.BlankChain()
	chain.Name = chainName
	chain.Operations.ContainerType = definitions.TypeChain
	chain.Operations.Labels = util.Labels(chain.Name, chain.Operations)
	if err := setChainDefaults(chain); err != nil {
		return nil, err
	}

	chainConf, err := config.LoadViperConfig(filepath.Join(ChainsPath), chainName, "chain")
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
	if util.IsMinimalDockerClientVersion() {
		if chain.Dependencies != nil {
			addDependencyVolumesAndLinks(chain.Dependencies, chain.Service, chain.Operations)
		}
	}

	checkChainNames(chain)
	log.WithFields(log.Fields{
		"container number": 1,
		"environment":      chain.Service.Environment,
		"entrypoint":       chain.Service.EntryPoint,
		"cmd":              chain.Service.Command,
	}).Debug("Chain definition loaded")
	return chain, nil
}

// Convert the chain def to a service def but keep the "eris_chains" containers prefix and set the chain id
func ChainsAsAService(chainName string, newCont bool) (*definitions.ServiceDefinition, error) {
	chain, err := LoadChainDefinition(chainName, newCont)
	if err != nil {
		return nil, err
	}
	s, err := ServiceDefFromChain(chain, ErisChainStart), nil
	if err != nil {
		return nil, err
	}
	// we keep the "eris_chain" prefix and set the CHAIN_ID var.
	// the run command is set in ServiceDefFromChain
	s.Operations.SrvContainerName = util.ChainContainerName(chainName)
	s.Service.Environment = append(s.Service.Environment, "CHAIN_ID="+chainName)
	return s, nil
}

func ServiceDefFromChain(chain *definitions.Chain, cmd string) *definitions.ServiceDefinition {
	// chainID := chain.ChainID
	setChainDefaults(chain)
	chain.Service.Name = chain.Name // this let's the data containers flow thru
	chain.Service.Image = path.Join(version.ERIS_REG_DEF, version.ERIS_IMG_DB)
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
	connectToAService(srv, ops, definitions.TypeChain, name, internalName, link, mount)
}

func MockChainDefinition(chainName, chainID string, newCont bool) *definitions.Chain {
	chn := definitions.BlankChain()
	chn.Name = chainName
	chn.ChainID = chainID
	chn.Service.AutoData = true

	log.WithField("=>", chainName).Debug("Mocking chain definition")

	chn.Operations.ContainerType = definitions.TypeChain
	chn.Operations.Labels = util.Labels(chainName, chn.Operations)

	checkChainNames(chn)
	return chn
}

// marshal from viper to definitions struct
func MarshalChainDefinition(chainConf *viper.Viper, chain *definitions.Chain) error {
	log.Debug("Marshalling chain")
	chnTemp := definitions.BlankChain()
	err := chainConf.Unmarshal(chnTemp)
	if err != nil {
		return fmt.Errorf("The marmots coult not marshal from viper to chain def: %v", err)
	}

	util.Merge(chain.Service, chnTemp.Service)
	chain.ChainID = chnTemp.ChainID

	// toml bools don't really marshal well
	// data_container can be in the chain or
	// in the service layer. this is very
	// opinionated. we know.
	for _, s := range []string{"", "service."} {
		if chainConf.GetBool(s + "data_container") {
			chain.Service.AutoData = true
			log.WithField("autodata", chain.Service.AutoData).Debug()
		}
	}

	return nil
}

func setChainDefaults(chain *definitions.Chain) error {
	cfg, err := config.LoadViperConfig(filepath.Join(ChainsPath), "default", "chain")
	if err != nil {
		return err
	}
	if err := cfg.Unmarshal(chain); err != nil {
		return err
	}

	log.WithField("image", chain.Service.Image).Debug("Chain defaults set")
	return nil
}

//----------------------------------------------------------------------
// validation funcs
func checkChainNames(chain *definitions.Chain) {
	chain.Service.Name = chain.Name
	chain.Operations.SrvContainerName = util.ChainContainerName(chain.Name)
	chain.Operations.DataContainerName = util.DataContainerName(chain.Name)
}
