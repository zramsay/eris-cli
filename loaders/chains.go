package loaders

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
	// "github.com/eris-ltd/eris-cli/version"

	"github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"

	"github.com/spf13/viper"
)

// Standard package commands.
const (
	ErisChainStart = "run"
	ErisChainNew   = "new"
)

// LoadChainConfigFile reads the "default" then chainName definition files
// from the common.ChainsPath directory and returns a chain structure set
// accordingly. LoadChainConfigFile also returns missing files or definition
// reading errors, if any.

// TODO refactor to find the config.toml given from the "CONFIG_PATH"
// in whichever subdirectory
func LoadChainConfigFile(chainName string) (*definitions.Chain, error) {
	chain := definitions.BlankChain()
	chain.Name = chainName
	chain.Operations.ContainerType = definitions.TypeChain
	chain.Operations.Labels = util.Labels(chain.Name, chain.Operations)
	//if err := setChainDefaults(chain); err != nil {
	//	return nil, err
	//}

	//definition, err := config.LoadViper(filepath.Join(common.ChainsPath), chainName)
	// XXX this need to append config.toml to do.Path somehow
	// i.e., currently only works for --chain-type=simplechain
	// pathToConfig can be read in as a string from  ~/.eris/chains/NAME/CONFIG_PATH
	// like checkout out chains does it

	whereIsTheConfigFile := filepath.Join(common.ChainsPath, chainName, "CONFIG_PATH")
	definition := viper.New()
	var err error
	pathToConfig, err := ioutil.ReadFile(whereIsTheConfigFile)
	if err != nil {
		return nil, err
	}

	definition, err = config.LoadViper(string(pathToConfig), "config")
	if err != nil {
		return nil, err
		// try in chain dir
		definition, err = config.LoadViper(string(pathToConfig), "config")
		if err != nil {
			return nil, err
		}
	} else {

		log.Warn("MARMOTpathConfig")
		log.Warn(string(pathToConfig))

		definition, err = config.LoadViper(string(pathToConfig), "config")
		if err != nil {
			return nil, err
		}
	}

	// Overwrite chain.ChainID and chain.Service according from
	// the definition.

	// TODO remove this overwrite... ?
	// chain.ChainID = chainName

	if err = MarshalChainDefinition(definition, chain); err != nil {
		return nil, err
	}

	if chain.Dependencies != nil {
		addDependencyVolumesAndLinks(chain.Dependencies, chain.Service, chain.Operations)
	}

	checkChainNames(chain)
	log.WithFields(log.Fields{
		"chain name":  chain.Name,
		"chain ID":    chain.ChainID,
		"environment": chain.Service.Environment,
		"entrypoint":  chain.Service.EntryPoint,
		"cmd":         chain.Service.Command,
	}).Warn("Chain definition loaded") // use warn temporarily.
	return chain, nil
}

// ChainsAsAService convert the chain definition to a service one
// and set the CHAIN_ID environment variable. ChainsAsAService
// can return config load errors.
func ChainsAsAService(chainName string) (*definitions.ServiceDefinition, error) {
	chain, err := LoadChainConfigFile(chainName)
	if err != nil {
		return nil, err
	}

	//setChainDefaults(chain) => now gotten from config.toml
	// TODO read these in from chain def (or don't overwrite...?)
	chain.Service.Name = chain.Name
	chain.Service.Image = path.Join(config.Global.DefaultRegistry, config.Global.ImageDB)
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
// XXX [zr] LoadChainConfigFile should do this, no?
// remove this function....
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

	if err := definition.Unmarshal(chnTemp); err != nil {
		return fmt.Errorf("The marmots coult not read the chain definition: %v", err)
	}

	util.Merge(chain.Service, chnTemp.Service) // [zr] we shouldn't need anymore ...?
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

//----------------------------------------------------------------------
// validation funcs
func checkChainNames(chain *definitions.Chain) {
	chain.Service.Name = chain.Name
	chain.Operations.SrvContainerName = util.ChainContainerName(chain.Name)
	chain.Operations.DataContainerName = util.DataContainerName(chain.Name)
}
