package maker

import (
	"encoding/json"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/eris-ltd/eris/definitions"
	"github.com/eris-ltd/eris/keys"
	"github.com/eris-ltd/eris/log"

	"github.com/eris-ltd/eris-db/genesis"
	ptypes "github.com/eris-ltd/eris-db/permission/types"
)

// ErisDBAccountConstructor contains different views on a single account
// for the purpose of constructing the configuration, genesis, and private
// validator file.
// Note that the generation of key pairs for the private validator is only
// for development purposes and that under 
type ErisDBAccountConstructor struct {
	genesisAccount          *genesis.GenesisAccount          `json:"genesis_account"`
	genesisValidator        *genesis.GenesisValidator        `json:"genesis_validator"`
	genesisPrivateValidator *genesis.GenesisPrivateValidator `json:"genesis_private_validator"`
}

// MakeAccounts specifies the chaintype and chain name and creates the constructors for generating
// configuration, genesis and private validator files (the latter if required - for development purposes)
func MakeAccounts(name, chainType string, accountTypes []*definitions.ErisDBAccountType) ([]*ErisDBAccountConstructor, error) {

	accountConstructors := []*ErisDBAccountConstructor{}

	switch chainType {
	// NOTE: [ben] "mint" is a legacy differentiator that refers to the consensus engine that eris-db uses
	// and currently Tendermint is the only consensus engine (chain) that is supported.  As such the variable
	// "chainType" can be misleading.
	case "mint":

		for _, accountType := range accountTypes {
			log.WithField("type", accountType.Name).Info("Making Account Type")
			var err error
			for i := 0; i < accountType.Number; i++ {
				// account names are formatted <ChainName_AccountTypeName_nnn>
				accountName := strings.ToLower(fmt.Sprintf(
					"%s_%s_%03d", name, accountType.Name, i))
				log.WithField("name", accountName).Debug("Making Account")

				// NOTE: [ben] for v0.16 we get the private validator file if `ToBond` > 0
				// For v0.17 we will default to all validators only using remote signing,
				// and then we should block by default extraction of private validator file.
				// NOTE: [ben] currently we default to ed25519/SHA512 for PKI and ripemd16
				// for address calculation.
				accountConstructor, err := newErisDBAccountConstructor(accountName, "ed25519,ripemd160", 
					accountType, false)
				if err != nil {
					return nil, fmt.Errorf("Failed construct account %s for %s", accountName, name)
				}
				// add the account constructor to the return slice
				accountConstructors = append(accountConstructors, accountConstructor)
			}
		}
		return accountConstructors, nil
	default:
		return nil, fmt.Errorf("Unknown chain type specifier (chainType: %s)", chainType)	
	}
}

//-----------------------------------------------------------------------------------------------------
// helper functions for MakeAccounts

// newErisDBAccountConstructor returns an ErisDBAccountConstructor that has a GenesisAccount
// and depending on the AccountType returns a GenesisValidator.  If a private validator file
// is needed for a validating account, it will pull the private key, unless this is
// explicitly blocked.
func newErisDBAccountConstructor(accountName string, keyAddressType string, 
	accountType *definitions.ErisDBAccountType, blockPrivateValidator bool)	(*ErisDBAccountConstructor, error) {

	var err error
	accountConstructor := &ErisDBAccountConstructor{} 
	permissions := &ptypes.AccountPermissions{}
	// TODO: expose roles
	// convert the permissions map of string-integer pairs to an
	// AccountPermissions type.
	if permissions, err = ptypes.ConvertPermissionsMapAndRolesToAccountPermissions(
		accountType.Perms, []string{}); err != nil {
		return nil, err
	}
	var address, publicKeyBytes []byte
	switch keyAddressType {
	// use ed25519/SHA512 for PKI and ripemd160 for Address
	case "ed25519,ripemd160":
		address, publicKeyBytes, genesisPrivateValidator, err := generateAddressAndKey(
			keyAddressType, blockPrivateValidator)
	default:
		// the other code paths in eris-keys are currently not tested for;
		return nil, fmt.Errorf("Currently only supported ed265519/ripemd160: unknown key type (%s)",
			keyAddressType)
	}

	accountConstructor.genesisAccount = genesis.NewGenesisAccount(
		// Genesis address
		address,
		// Genesis amount
		int64(accountType.Tokens),
		// Genesis name
		accountName,
		// Genesis permissions
		permissions)

	// Define this account as a bonded validator in genesis.
	if accountType.ToBond > 0 && accountType.Tokens >= accountType.ToBond {
		accountConstructor.genesisValidator, err = genesis.NewGenesisValidator(
			// Genesis validator amount
			int64(accountType.Tokens),
			// Genesis validator name
			accountName,
			// Genesis validator unbond to address
			address,
			// Genesis validator bond amount
			int64(accountType.ToBond),
			// Genesis validator public key type string
			"ed25519",
			// Genesis validator public key bytes
			publicKeyBytes)
		if err != nil {
			// CONTINUE
		}
	}

	return accountConstructor, nil
}

//----------------------------------------------------------------------------------------------------
// helper functions with eris-keys

// generateAddressAndKey returns an address, public key and if requested the JSON bytes of a
// private validator structure.
func generateAddressAndKey(keyAddressType string, blockPrivateValidator bool) (address []byte, publicKey []byte,
	genesisPrivateValidator *genesis.GenesisPrivateValidator, err error) {
	addressString, publicKeyString, privateValidatorJson, err := makeKey(keyAddressType, blockPrivateValidator)
	if err != nil {
		return
	}

	if address, err = hex.DecodeString(addressString); err != nil {
		return
	}

	if publicKey, err = hex.DecodeString(publicKeyString); err != nil {
		return
	}

	if !blockPrivateValidator {
		if err = json.Unmarshal(privateValidatorJson, genesisPrivateValidator); err != nil {
			log.Error(string(privateValidatorJson))
			return
		}
	}

	return
}

// ugh. TODO: further clean up eris-keys.
func makeKey(keyType string, blockPrivateValidator bool) (address string, publicKey string, privateValidatorJson []byte, err error) {
	log.WithFields(log.Fields{
		"type": keyType,
	}).Debug("Sending Call to eris-keys server")

	keyClient, err := keys.InitKeyClient()
	if err != nil {
		return
	}

	// note, for now we use no password to lock/unlock keys
	if address, err = keyClient.GenerateKey(false, true, keyType, ""); err != nil {
		return
	}

	if publicKey, err = keyClient.PubKey(address, ""); err != nil {
		return
	}

	if !blockPrivateValidator {
		if privateValidatorJson, err = keyClient.Convert(address, ""); err != nil {
			return
		}
	} else {
		privateValidatorJson = []byte{}
	}

	return
}
