package keys

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/services"

	"github.com/eris-ltd/eris-db/keys"
)

type KeyClient struct{}

// Returns an initialized key client to a docker container
// running the keys server
// Adding the Ip address is optional and should only be used
// for passing data
func InitKeyClient() (*KeyClient, error) {
	keys := &KeyClient{}
	err := keys.ensureRunning()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// Keyclient returns a list of keys that it is aware of.
// params:
// host - search for keys on the host
// container - search for keys on the container
// quiet - don't print output, just return the list you find
func (keys *KeyClient) ListKeys(host, container, quiet bool) ([]string, error) {
	var result []string
	if host {
		keysPath := filepath.Join(config.KeysPath, "data")
		addrs, err := ioutil.ReadDir(keysPath)
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			result = append(result, addr.Name())
		}
		if !quiet {
			if len(addrs) == 0 {
				log.Warn("No keys found on host")
			} else {
				// First key.
				log.WithField("=>", result[0]).Warn("The keys on your host kind marmot")
				// Subsequent keys.
				if len(result) > 1 {
					for _, addr := range result[1:] {
						log.WithField("=>", addr).Warn()
					}
				}
			}
		}
	}

	if container {
		err := keys.ensureRunning()
		if err != nil {
			return nil, err
		}

		keysOut, err := services.ExecHandler("keys", []string{"ls", "/home/eris/.eris/keys/data"})
		if err != nil {
			return nil, err
		}
		result = strings.Fields(keysOut.String())
		if !quiet {
			if len(result) == 0 || result[0] == "" {
				log.Warn("No keys found in container")
			} else {
				// First key.
				log.WithField("=>", result[0]).Warn("The keys in your container kind marmot")
				// Subsequent keys.
				if len(result) > 1 {
					for _, addr := range result[1:] {
						log.WithField("=>", addr).Warn()
					}
				}
			}
		}
	}
	return result, nil
}

//TODO: Need to add a "type" field
// Keyclient generates a key.
// params:
// save - whether or not to export it from container to host when we're done generating
// password - not implemented yet
func (keys *KeyClient) GenerateKey(save bool, password string) (string, error) {
	err := keys.ensureRunning()
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if password != "" {
		return "", fmt.Errorf("Password currently unimplemented. Marmots are confused at how you got here.")
	}
	buf, err = services.ExecHandler("keys", []string{"eris-keys", "gen", "--no-pass"})
	if err != nil {
		return "", err
	}

	address := strings.TrimSpace(buf.String())
	if save {
		log.WithField("=>", address).Warn("Saving key to host")
		if err := keys.ExportKey(address, false); err != nil {
			return "", err
		}
	} else {
		log.Warn(address)
	}
	return address, nil
}

// Keyclient exports keys from container to host.
// params:
// address - address to export single key
// all - bool that says export all the keys
func (keys *KeyClient) ExportKey(address string, all bool) error {
	err := keys.ensureRunning()
	if err != nil {
		return err
	}
	do := definitions.NowDo()
	do.Name = "keys"
	if all && address == "" {
		do.Destination = config.KeysPath
		do.Source = path.Join(config.KeysContainerPath)
		do.All = all
	} else if all && address != "" {
		return fmt.Errorf("Dev implementation error: Cannot import both all and a single address: %v", address)
	} else {
		do.Destination = config.KeysDataPath
		do.Address = address
		do.Source = path.Join(config.KeysContainerPath, do.Address)
	}
	return data.ExportData(do)
}

// Keyclient imports keys from host to container.
// params:
// address - address to import single key
// all - bool that says import all the keys
func (keys *KeyClient) ImportKey(address string, all bool) error {
	err := keys.ensureRunning()
	if err != nil {
		return err
	}

	do := definitions.NowDo()
	do.Name = "keys"
	if all && address == "" {
		// get all keys from host
		result, err := keys.ListKeys(true, false, true)
		if err != nil {
			return err
		}
		// flip them for the import
		do.Container = true
		do.Host = false
		do.Quiet = false
		do.All = all
		for _, addr := range result {
			do.Source = filepath.Join(config.KeysDataPath, addr)
			do.Destination = path.Join(config.KeysContainerPath, addr)
			if err := data.ImportData(do); err != nil {
				return err
			}
		}
	} else if all && address != "" {
		return fmt.Errorf("Dev implementation error: Cannot import both all and a single address: %v", address)
	} else {
		do.Source = filepath.Join(config.KeysDataPath, address)
		do.Destination = path.Join(config.KeysContainerPath, address)
		do.Address = address
		if err := data.ImportData(do); err != nil {
			return err
		}
	}

	return nil
}

// Helper function used to ensure the keys container is indeed running
func (keys *KeyClient) ensureRunning() error {
	doKeys := definitions.NowDo()
	doKeys.Name = "keys"
	return services.EnsureRunning(doKeys)
}

// Keyclient returns the public key of an address.
// params:
// address - address whose public key we want to know
func (keys *KeyClient) PubKey(address string) (string, error) {
	err := keys.ensureRunning()
	if err != nil {
		return "", err
	}

	addr := strings.TrimSpace(address)
	buf, err := services.ExecHandler("keys", []string{"eris-keys", "pub", "--addr", addr, "--name", ""})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

// Keyclient returns a signed message.
// params:
// address - address that we wish to sign with
// msg - the message we wish to sign
func (keys *KeyClient) SignMsg(address, msg string) (string, error) {
	err := keys.ensureRunning()
	if err != nil {
		return "", err
	}

	addr := strings.TrimSpace(address)
	buf, err := services.ExecHandler("keys", []string{"eris-keys", "sign", "--addr", addr, msg})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

// Keyclient verifies the validity of a signed message
// params:
// keyType - the key type we are working with. If left empty, defaults to ed25519,ripemd160 key/curve combination
// signature - the signed message we are verifying
// publicKey - the public key we are using to verify the signature
// msg - the original message
func (keys *KeyClient) Verify(keyType, signature, publicKey, msg string) (bool, error) {
	err := keys.ensureRunning()
	if err != nil {
		return false, err
	}

	buf := new(bytes.Buffer)
	if keyType == "" {
		buf, err = services.ExecHandler("keys", []string{"eris-keys", "verify", msg, signature, publicKey})
		if err != nil {
			return false, err
		}
	} else {
		buf, err = services.ExecHandler("keys", []string{"eris-keys", "verify", "--type", keyType, msg, signature, publicKey})
		if err != nil {
			return false, err
		}
	}

	return strconv.ParseBool(strings.TrimSpace(buf.String()))
}
