package loaders

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/util"
	"github.com/monax/monax/version"

	"github.com/spf13/viper"
)

func LoadChainTypes(fileName string) (*definitions.ChainType, error) {
	fileName = filepath.Join(config.ChainTypePath, fileName)
	log.WithField("file name", fileName).Info("Loading Chain Definition.")
	var typ = definitions.BlankChainType()
	var chainType = viper.New()

	if err := getSetup(fileName, chainType); err != nil {
		return nil, err
	}

	// marshall file
	if err := chainType.Unmarshal(typ); err != nil {
		return nil, fmt.Errorf(`Sorry, your chain types file "%v" confused the marmots.
			Please check your chain type definition file is properly formatted: %v`, fileName, err)
	}

	return typ, nil
}

func getSetup(fileName string, cfg *viper.Viper) error {
	// setup file
	abs, err := filepath.Abs(fileName)
	if err != nil {
		return fmt.Errorf("Sorry, the marmots were unable to find the absolute path to the account types file.")
	}

	path := filepath.Dir(abs)
	file := filepath.Base(abs)
	extName := filepath.Ext(file)
	bName := file[:len(file)-len(extName)]

	cfg.AddConfigPath(path)
	cfg.SetConfigName(bName)
	cfg.SetConfigType(strings.Replace(extName, ".", "", 1))

	// load file
	if err := cfg.ReadInConfig(); err != nil {
		return fmt.Errorf(`Sorry, the marmots were unable to load the file: %s. Please check your path. 
			Error: %v`, fileName, err)
	}

	return nil
}

// LoadChainDefinition returns a ChainDefinition settings for the chainName
// chain. It also enriches the the chain settings by reading the definition
// file specified by the optional definiton parameter. It returns Viper package
// and parsing errors.
func LoadChainDefinition(chainName string, definition ...string) (*definitions.ChainDefinition, error) {
	chain := definitions.BlankChainDefinition()
	chain.Name = chainName
	chain.Operations.ContainerType = definitions.TypeChain
	chain.Operations.Labels = util.Labels(chain.Name, chain.Operations)
	chain.Service.Image = path.Join(version.DefaultRegistry, version.ImageDB)
	chain.Service.AutoData = true

	// Optionally load settings from the definition file.
	if len(definition) > 0 {
		definition, err := config.LoadViper(filepath.Dir(definition[0]), filepath.Base(definition[0]))
		if err != nil {
			return nil, err
		}

		if err = MarshalChainDefinition(definition, chain); err != nil {
			return nil, err
		}

		if chain.Dependencies != nil {
			addDependencyVolumesAndLinks(chain.Dependencies, chain.Service, chain.Operations)
		}
	}

	chain.Service.Name = chain.Name
	chain.Operations.SrvContainerName = util.ChainContainerName(chain.Name)
	chain.Operations.DataContainerName = util.DataContainerName(chain.Name)

	log.WithFields(log.Fields{
		"chain name":  chain.Name,
		"environment": chain.Service.Environment,
		"entrypoint":  chain.Service.EntryPoint,
		"cmd":         chain.Service.Command,
	}).Debug("Chain definition loaded")
	return chain, nil
}

// ConnectToAChain operates in two ways
//  - if link is true, sets srv.Links to point to a chain container specifiend by name:internalName
//  - if mount is true, sets srv.VolumesFrom to point to a chain container specified by name
func ConnectToAChain(srv *definitions.Service, ops *definitions.Operation, name, internalName string, link, mount bool) {
	connectToAService(srv, ops, definitions.TypeChain, name, internalName, link, mount)
}

// MarshalChainDefinition reads the definition file and sets chain.Service fields
// in the chain structure. Returns config read errors.
func MarshalChainDefinition(definition *viper.Viper, chain *definitions.ChainDefinition) error {
	log.Debug("Marshalling chain")

	if err := definition.Unmarshal(chain); err != nil {
		return fmt.Errorf("The marmots could not read the chain definition: %v", err)
	}

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
