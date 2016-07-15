package loaders

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

	"github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"

	"github.com/spf13/viper"
)

// Standard package commands.
const (
	ErisChainStart    = "run"
	ErisChainStartAPI = "api"
	ErisChainInstall  = "install"
	ErisChainNew      = "new"
	ErisChainRegister = "register"
)

// LoadChainDefinition reads the "default" then chainName definition files
// from the common.ChainsPath directory and returns a chain structure set
// accordingly. LoadChainDefinition also returns missing files or definition
// reading errors, if any.
func LoadChainDefinition(chainName string) (*definitions.Chain, error) {
	chain := definitions.BlankChain()
	chain.Name = chainName
	chain.Operations.ContainerType = definitions.TypeChain
	chain.Operations.Labels = util.Labels(chain.Name, chain.Operations)
	if err := setChainDefaults(chain); err != nil {
		return nil, err
	}

	definition, err := config.LoadViperConfig(filepath.Join(common.ChainsPath), chainName)
	if err != nil {
		return nil, err
	}

	// Overwrite chain.ChainID and chain.Service according from
	// the definition.
	if err = MarshalChainDefinition(definition, chain); err != nil {
		return nil, err
	}

	if chain.Dependencies != nil {
		addDependencyVolumesAndLinks(chain.Dependencies, chain.Service, chain.Operations)
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

// ChainsAsAService convert the chain definition to a service one
// and set the CHAIN_ID environment variable. ChainsAsAService
// can return config load errors.
func ChainsAsAService(chainName string) (*definitions.ServiceDefinition, error) {
	chain, err := LoadChainDefinition(chainName)
	if err != nil {
		return nil, err
	}

	setChainDefaults(chain)
	chain.Service.Name = chain.Name
	chain.Service.Image = path.Join(version.ERIS_REG_DEF, version.ERIS_IMG_DB)
	chain.Service.AutoData = true
	chain.Service.Command = ErisChainStart

	s := &definitions.ServiceDefinition{
		Name:         chain.Name,
		ServiceID:    chain.ChainID,
		Dependencies: &definitions.Dependencies{Services: []string{"keys"}},
		Service:      chain.Service,
		Operations:   chain.Operations,
		Maintainer:   chain.Maintainer,
		Location:     chain.Location,
		Machine:      chain.Machine,
	}

	// These are mostly operational considerations that we want to ensure are met.
	ServiceFinalizeLoad(s)

	s.Operations.SrvContainerName = util.ChainContainerName(chainName)
	s.Service.Environment = append(s.Service.Environment, "CHAIN_ID="+chainName)
	return s, nil
}

// ConnectToAChain operates in two ways
//  - if link is true, sets srv.Links to point to a chain container specifiend by name:internalName
//  - if mount is true, sets srv.VolumesFrom to point to a chain container specified by name
func ConnectToAChain(srv *definitions.Service, ops *definitions.Operation, name, internalName string, link, mount bool) {
	connectToAService(srv, ops, definitions.TypeChain, name, internalName, link, mount)
}

// MockChainDefinition creates a chain definition with necessary fields
// already filled in.
func MockChainDefinition(chainName, chainID string) *definitions.Chain {
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

// MarshalChainDefinition reads the definition file and sets the chain.ChainID
// and chain.Service fields in the chain structure. Returns config read errors.
func MarshalChainDefinition(definition *viper.Viper, chain *definitions.Chain) error {
	log.Debug("Marshalling chain")
	chnTemp := definitions.BlankChain()
	err := definition.Unmarshal(chnTemp)
	if err != nil {
		return fmt.Errorf("The marmots coult not read the chain definition: %v", err)
	}

	util.Merge(chain.Service, chnTemp.Service)
	if len(chnTemp.Service.Ports) != 0 {
		chain.Service.Ports = chnTemp.Service.Ports
	}
	chain.ChainID = chnTemp.ChainID

	// toml bools don't really marshal well "data_container". It can be
	// in the chain or in the service layer.
	for _, s := range []string{"", "service."} {
		if definition.GetBool(s + "data_container") {
			chain.Service.AutoData = true
			log.WithField("autodata", chain.Service.AutoData).Debug()
		}
	}

	return nil
}

func setChainDefaults(chain *definitions.Chain) error {
	cfg, err := config.LoadViperConfig(filepath.Join(common.ChainsPath), "default")
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
