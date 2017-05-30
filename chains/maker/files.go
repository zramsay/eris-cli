package maker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"

	configurationFile "github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/genesis"

	"github.com/BurntSushi/toml"
)

// XXX: this is temporary until legacy-keys.js is more tightly integrated with legacy-contracts.js
type accountInfo struct {
	Address string `mapstructure:"address" json:"address" yaml:"address" toml:"address"`
	PubKey  string `mapstructure:"pubKey" json:"pubKey" yaml:"pubKey" toml:"pubKey"`
	PrivKey string `mapstructure:"privKey" json:"privKey" yaml:"privKey" toml:"privKey"`
}

func SaveAccountResults(do *definitions.Do, accounts []*MonaxDBAccountConstructor) error {
	// Log a warning to users for the new behaviour:
	// if asked to output the accounts with do.Output, and `monax chains make --unsafe` is not
	// provided with the unsafe flag, then we no longer write the private keys in `accounts.json`
	if !do.Unsafe {
		log.Warn("The marmots care about your safety and no longer export the generated private keys onto your local host. " +
			"If you do want accounts.json to contain the private keys for use in a development environment, please make your " +
			"chain with `--unsafe` to write the private keys to disk.  This option will be deprecated once the javascript libraries " +
			"implements a remote signing path as tooling does.")
	}

	addrFile, err := os.Create(filepath.Join(config.ChainsPath, do.Name, "addresses.csv"))
	if err != nil {
		return fmt.Errorf("Error creating addresses file. This usually means that there was a problem with the chain making process.")
	}
	defer addrFile.Close()

	log.WithField("name", do.Name).Debug("Creating file")
	actFile, err := os.Create(filepath.Join(config.ChainsPath, do.Name, "accounts.csv"))
	if err != nil {
		return fmt.Errorf("Error creating accounts file.")
	}
	log.WithField("path", filepath.Join(config.ChainsPath, do.Name, "accounts.csv")).Debug("File successfully created")
	defer actFile.Close()

	log.WithField("name", do.Name).Debug("Creating file")
	actJSONFile, err := os.Create(filepath.Join(config.ChainsPath, do.Name, "accounts.json"))
	if err != nil {
		return fmt.Errorf("Error creating accounts file.")
	}
	log.WithField("path", filepath.Join(config.ChainsPath, do.Name, "accounts.json")).Debug("File successfully created")
	defer actJSONFile.Close()

	valFile, err := os.Create(filepath.Join(config.ChainsPath, do.Name, "validators.csv"))
	if err != nil {
		return fmt.Errorf("Error creating validators file.")
	}
	defer valFile.Close()

	accountJsons := make(map[string]*accountInfo)

	for _, accountConstructor := range accounts {
		if accountConstructor.genesisAccount != nil {
			address := fmt.Sprintf("%X", accountConstructor.genesisAccount.Address)
			name := accountConstructor.genesisAccount.Name
			// NOTE: [ben] this an untyped public key
			publicKey := fmt.Sprintf("%X", accountConstructor.untypedPublicKeyBytes)
			amount := accountConstructor.genesisAccount.Amount
			basePermissions := accountConstructor.genesisAccount.Permissions.Base
			accountJsons[name] = &accountInfo{
				Address: address,
				PubKey:  publicKey,
				// if do.Unsafe is not true, the private key bytes have not been copied
				PrivKey: fmt.Sprintf("%X", accountConstructor.untypedPrivateKeyBytes),
			}

			_, err := addrFile.WriteString(fmt.Sprintf("%s,%s\n", address, name))
			if err != nil {
				log.Error("Error writing addresses file.")
				return err
			}
			_, err = actFile.WriteString(fmt.Sprintf("%s,%d,%s,%d,%d\n", publicKey, amount,
				name, basePermissions.Perms, basePermissions.SetBit))
			if err != nil {
				log.Error("Error writing accounts file.")
				return err
			}
			if accountConstructor.genesisValidator != nil {
				_, err = valFile.WriteString(fmt.Sprintf("%s,%d,%s,%d,%d\n", publicKey,
					accountConstructor.genesisValidator.Amount, name, basePermissions.Perms, basePermissions.SetBit))
				if err != nil {
					log.Error("Error writing validators file.")
					return err
				}
			}
		}
	}
	addrFile.Sync()
	actFile.Sync()
	valFile.Sync()

	j, err := json.MarshalIndent(accountJsons, "", "  ")
	if err != nil {
		return err
	}

	_, err = actJSONFile.Write(j)
	if err != nil {
		return err
	}

	log.WithField("path", actJSONFile.Name()).Debug("Saving File.")
	log.WithField("path", addrFile.Name()).Debug("Saving File.")
	log.WithField("path", actFile.Name()).Debug("Saving File.")
	log.WithField("path", valFile.Name()).Debug("Saving File.")

	return nil
}

func WriteGenesisFile(chainName, accountName string, genesisFileBytes []byte) error {
	return writer(genesisFileBytes, chainName, accountName, "genesis.json")
}

func WritePrivateValidatorFile(chainName, accountName string,
	genesisPrivateValidator *genesis.GenesisPrivateValidator) error {
	privateValidatorFileBytes, err := json.MarshalIndent(genesisPrivateValidator, "", "  ")
	if err != nil {
		return err
	}
	return writer(privateValidatorFileBytes, chainName, accountName, "priv_validator.json")
}

func WriteConfigurationFile(chainName, accountName, seeds string, chainImageName string,
	useDataContainer bool, exportedPorts []string, containerEntrypoint string) error {
	if accountName == "" {
		return fmt.Errorf("No account name provided.")
	}
	if chainName == "" {
		return fmt.Errorf("No chain name provided.")
	}

	configurationFileBytes, err := configurationFile.GetConfigurationFileBytes(chainName,
		accountName, seeds, chainImageName, useDataContainer,
		convertExportPortsSliceToString(exportedPorts), containerEntrypoint)
	if err != nil {
		return err
	}

	return writer(configurationFileBytes, chainName, accountName, "config.toml")
}

func SaveAccountType(thisActT *definitions.MonaxDBAccountType) error {
	writer, err := os.Create(filepath.Join(config.AccountsTypePath, fmt.Sprintf("%s.toml", thisActT.Name)))
	if err != nil {
		return err
	}
	defer writer.Close()

	enc := toml.NewEncoder(writer)
	enc.Indent = ""
	return enc.Encode(thisActT)
}

func convertExportPortsSliceToString(exportPorts []string) string {
	if len(exportPorts) == 0 {
		return ""
	}
	return `[ "` + strings.Join(exportPorts[:], `", "`) + `" ]`
}

func writer(fileBytes []byte, chainName, accountName, fileBase string) error {
	file := filepath.Join(config.ChainsPath, chainName, accountName, fileBase)

	log.WithField("path", file).Debug("Saving File.")
	return writeFile(fileBytes, file)
}

func writeFile(data []byte, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0775); err != nil {
		return err
	}
	writer, err := os.Create(filepath.Join(path))
	if err != nil {
		return err
	}
	defer writer.Close()

	writer.Write(data)
	return nil
}
