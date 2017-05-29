package maker

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/monax/monax/definitions"
	"github.com/monax/monax/keys"
	"github.com/monax/monax/log"

	"github.com/hyperledger/burrow/genesis"
	ptypes "github.com/hyperledger/burrow/permission/types"
)

// MonaxDBAccountConstructor contains different views on a single account
// for the purpose of constructing the configuration, genesis, and private
// validator file.
// Note that the generation of key pairs for the private validator is only
// for development purposes and that under
type MonaxDBAccountConstructor struct {
	genesisAccount          *genesis.GenesisAccount          `json:"genesis_account"`
	genesisValidator        *genesis.GenesisValidator        `json:"genesis_validator"`
	genesisPrivateValidator *genesis.GenesisPrivateValidator `json:"genesis_private_validator"`

	// NOTE: [ben] this is redundant information to preserve the current behaviour of
	// tooling to write the untyped public key for all accounts in accounts.csv
	untypedPublicKeyBytes []byte
	typeBytePublicKey     byte

	// NOTE: [ben] this is redundant information and unsafe but is put in place to
	// temporarily preserve the behaviour that the private keys of a *development*
	// chain can be written to the host
	// NOTE: [ben] because this is bad practice, it now requires explicit
	// flag `monax chains make --unsafe` (unsafe bool in signatures below)
	untypedPrivateKeyBytes []byte
}

// MakeAccounts specifies the chaintype and chain name and creates the constructors for generating
// configuration, genesis and private validator files (the latter if required - for development purposes)
// NOTE: [ben] if unsafe is set to true the private keys will be extracted from monax-keys and be written
// into accounts.json. This will be deprecated in v0.17
func MakeAccounts(name, chainType string, accountTypes []*definitions.MonaxDBAccountType, unsafe bool) ([]*MonaxDBAccountConstructor, error) {

	accountConstructors := []*MonaxDBAccountConstructor{}

	switch chainType {
	// NOTE: [ben] "mint" is a legacy differentiator that refers to the consensus engine that burrow uses
	// and currently Tendermint is the only consensus engine (chain) that is supported.  As such the variable
	// "chainType" can be misleading.
	case "mint":
		for _, accountType := range accountTypes {
			log.WithField("type", accountType.Name).Info("Making Account Type")
			for i := int64(0); i < accountType.DefaultNumber; i++ {
				// account names are formatted <ChainName_AccountTypeName_nnn>
				accountName := strings.ToLower(fmt.Sprintf(
					"%s_%s_%03d", name, accountType.Name, i))
				log.WithField("name", accountName).Debug("Making Account")

				// NOTE: [ben] for v0.16 we get the private validator file if `ToBond` > 0
				// For v0.17 we will default to all validators only using remote signing,
				// and then we should block by default extraction of private validator file.
				// NOTE: [ben] currently we default to ed25519/SHA512 for PKI and ripemd16
				// for address calculation.
				accountConstructor, err := newMonaxDBAccountConstructor(accountName, "ed25519,ripemd160",
					accountType, false, unsafe)
				if err != nil {
					return nil, fmt.Errorf("Failed to construct account %s for %s: %v", accountName, name, err)
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

// newMonaxDBAccountConstructor returns an MonaxDBAccountConstructor that has a GenesisAccount
// and depending on the AccountType returns a GenesisValidator.  If a private validator file
// is needed for a validating account, it will pull the private key, unless this is
// explicitly blocked.
func newMonaxDBAccountConstructor(accountName string, keyAddressType string,
	accountType *definitions.MonaxDBAccountType, blockPrivateValidator, unsafe bool) (*MonaxDBAccountConstructor, error) {

	var err error
	isValidator := (accountType.DefaultBond > 0 && accountType.DefaultTokens >= accountType.DefaultBond)
	accountConstructor := &MonaxDBAccountConstructor{}
	var genesisPrivateValidator *genesis.GenesisPrivateValidator
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
		if address, publicKeyBytes, genesisPrivateValidator, err = generateAddressAndKey(
			keyAddressType, blockPrivateValidator); err != nil {
			return nil, err
		}

		// NOTE: [ben] these auxiliary fields in the constructor are to be deprecated
		// but introduced to support current unsafe behaviour where all private keys
		// are extracted from monax-keys
		accountConstructor.untypedPublicKeyBytes = make([]byte, len(publicKeyBytes))
		copy(accountConstructor.untypedPublicKeyBytes[:], publicKeyBytes[:])
		// tendermint/go-crypto typebyte for ed25519
		accountConstructor.typeBytePublicKey = byte(0x01)

		if unsafe && genesisPrivateValidator != nil {
			// NOTE: [ben] this is a round-about way to not include tendermints crypto-types
			// before we deprecate private_validator files
			if privateKeyString, ok := genesisPrivateValidator.PrivKey[1].(string); ok {
				if privateKeyBytes, err := hex.DecodeString(privateKeyString); err != nil {
					return nil, err
				} else {
					accountConstructor.untypedPrivateKeyBytes = make([]byte, len(privateKeyBytes))
					copy(accountConstructor.untypedPrivateKeyBytes[:], privateKeyBytes[:])
				}
			}
		}
	default:
		// the other code paths in monax-keys are currently not tested for;
		return nil, fmt.Errorf("Currently only supported ed265519/ripemd160: unknown key type (%s)",
			keyAddressType)
	}

	accountConstructor.genesisAccount = genesis.NewGenesisAccount(
		// Genesis address
		address,
		// Genesis amount
		int64(accountType.DefaultTokens),
		// Genesis name
		accountName,
		// Genesis permissions
		permissions)

	// Define this account as a bonded validator in genesis.
	if isValidator {
		accountConstructor.genesisValidator, err = genesis.NewGenesisValidator(
			// Genesis validator amount
			int64(accountType.DefaultTokens),
			// Genesis validator name
			accountName,
			// Genesis validator unbond to address
			address,
			// Genesis validator bond amount
			int64(accountType.DefaultBond),
			// Genesis validator public key type string
			// Currently only ed22519 is exposed through the tooling
			"ed25519",
			// Genesis validator public key bytes
			publicKeyBytes)
		if err != nil {
			return nil, err
		}

		if genesisPrivateValidator != nil && !blockPrivateValidator {
			// explicitly copy genesis private validator for clarity
			accountConstructor.genesisPrivateValidator = genesisPrivateValidator
		}
	}

	return accountConstructor, nil
}

//----------------------------------------------------------------------------------------------------
// helper functions with monax-keys

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
		if privateValidatorJson != nil {
			genesisPrivateValidator = new(genesis.GenesisPrivateValidator)
			if err = json.Unmarshal(privateValidatorJson, genesisPrivateValidator); err != nil {
				log.Error(string(privateValidatorJson))
				return
			}
			// TODO: [ben] this is a hack, because the response from monax keys has a wrongly recoded
			// address that is provided as the original input; as we look to deprecate priv_validator
			// on v0.17 this can simply be patched for now.
			genesisPrivateValidator.Address = fmt.Sprintf("%X", address)
		} else {
			genesisPrivateValidator = nil
		}
	}

	return
}

// ugh. TODO: further clean up monax-keys.
func makeKey(keyType string, blockPrivateValidator bool) (address string, publicKey string, privateValidatorJson []byte, err error) {
	log.WithFields(log.Fields{
		"type": keyType,
	}).Debug("Sending Call to monax-keys server")

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
