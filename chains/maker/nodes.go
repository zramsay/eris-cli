package maker

import (
	"fmt"
	"strings"

	"github.com/monax/monax/log"

	"github.com/hyperledger/burrow/genesis"
)

// MakeMonaxDBNode writes the chain name folder with a folder for every account.
// In each folder is the genesis file and the configuration file written
func MakeMonaxDBNodes(chainName string, seeds []string, accounts []*MonaxDBAccountConstructor, chainImageName string,
	useDataContainer bool, exportedPorts []string, containerEntrypoint string) error {
	var genesisAccounts []*genesis.GenesisAccount
	var genesisValidators []*genesis.GenesisValidator

	// regroup the accounts and validators for the creation of the genesis file.
	for _, accountConstructor := range accounts {
		if accountConstructor.genesisAccount != nil {
			genesisAccounts = append(genesisAccounts, accountConstructor.genesisAccount)

			// if the GenesisAccount exists; check if GenesisValidator is defined
			if accountConstructor.genesisValidator != nil {
				genesisValidators = append(genesisValidators, accountConstructor.genesisValidator)
				log.WithFields(log.Fields{
					"name":        accountConstructor.genesisAccount.Name,
					"address":     accountConstructor.genesisAccount.Address,
					"tokens":      accountConstructor.genesisAccount.Amount,
					"permissions": accountConstructor.genesisAccount.Permissions,
				}).Debug("Adding validator account.")
			} else {
				log.WithFields(log.Fields{
					"name":        accountConstructor.genesisAccount.Name,
					"address":     accountConstructor.genesisAccount.Address,
					"tokens":      accountConstructor.genesisAccount.Amount,
					"permissions": accountConstructor.genesisAccount.Permissions,
				}).Debug("Adding genesis account.")
			}
		} else {
			return fmt.Errorf("Unexpected behavior: constructor contains nil-GenesisAccount information.")
		}
	}

	// generate the json bytes for the genesis file which are unique for all nodes.
	genesisFileBytes, err := genesis.GenerateGenesisFileBytes(chainName, genesisAccounts, genesisValidators)
	if err != nil {
		return err
	}
	seedsString := strings.Join(seeds, ",") // format for config file (if len>1)

	// now write the files to disk for each node.
	for _, accountConstructor := range accounts {
		if accountConstructor.genesisAccount != nil {
			accountName := accountConstructor.genesisAccount.Name
			if err := WriteGenesisFile(chainName, accountName, genesisFileBytes); err != nil {
				return err
			}
			if err := WriteConfigurationFile(chainName, accountName, seedsString,
				chainImageName, useDataContainer, exportedPorts, containerEntrypoint); err != nil {
				return err
			}
			// only if the GenesisPrivateValidator has been created in the constructor
			// write the priv_validator.json file to disk
			if accountConstructor.genesisPrivateValidator != nil {
				if err := WritePrivateValidatorFile(chainName, accountName,
					accountConstructor.genesisPrivateValidator); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
