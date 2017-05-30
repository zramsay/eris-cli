package util

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"

	"github.com/hyperledger/burrow/client"
	"github.com/hyperledger/burrow/logging/loggers"
)

// Maximum entries in the HEAD file
var MaxHead = 100

// Change the head to null (no head)
func NullHead() error {
	return ChangeHead("")
}

// Get the current active chain (top of the HEAD file)
// Returns chain name
func GetHead() (string, error) {
	// TODO: only read the one line!
	f, err := ioutil.ReadFile(config.HEAD)
	if os.IsNotExist(err) {
		if _, err := os.Create(config.HEAD); err != nil {
			return "", err
		}
	} else {
		return "", err
	}

	fspl := strings.Split(string(f), "\n")
	head := fspl[0]

	if head == "" {
		return "", fmt.Errorf("There is no chain checked out")
	}

	return head, nil
}

// Add a new entry (name) to the top of the HEAD file
// Expects the chain type and head (id) to be full (already resolved)
func ChangeHead(name string) error {
	if !IsChain(name, false) && name != "" {
		log.Debug("Chain name not known. Not saving")
		return nil
	}

	log.Debug("Chain name known (or blank). Saving to head file")
	// read in the entire head file and clip
	// if we have reached the max length
	b, err := ioutil.ReadFile(config.HEAD)
	if os.IsNotExist(err) {
		if _, err := os.Create(config.HEAD); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	bspl := strings.Split(string(b), "\n")
	var bsp string
	if len(bspl) >= MaxHead {
		bsp = strings.Join(bspl[:MaxHead-1], "\n")
	} else {
		bsp = string(b)
	}

	s := name + "\n" + bsp
	if err := ioutil.WriteFile(config.HEAD, []byte(s), 0666); err != nil {
		return err
	}

	log.Debug("Head file saved")
	return nil
}

func GetBlockHeight(do *definitions.Do) (latestBlockHeight int, err error) {
	nodeClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
	// NOTE: NodeInfo is no longer exposed through Status();
	// other values are currently not use by the package manager
	_, _, _, latestBlockHeight, _, err = nodeClient.Status()
	if err != nil {
		return 0, err
	}
	// set return values
	return
}

// TODO: it is unpreferable to mix static and non-static use of Do
func GetChainID(do *definitions.Do) error {
	if do.ChainID == "" {
		nodeClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
		_, chainId, _, err := nodeClient.ChainId()
		if err != nil {
			return err
		}
		do.ChainID = chainId
		log.WithField("=>", do.ChainID).Info("Using ChainID from Node")
	}

	return nil
}

func AccountsInfo(account, field string, do *definitions.Do) (string, error) {

	addrBytes, err := hex.DecodeString(account)
	if err != nil {
		return "", fmt.Errorf("Account Addr %s is improper hex: %v", account, err)
	}
	nodeClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
	r, err := nodeClient.GetAccount(addrBytes)
	if err != nil {
		return "", err
	}
	if r == nil {
		return "", fmt.Errorf("Account %s does not exist", account)
	}

	var s string
	if strings.Contains(field, "permissions") {
		// TODO: [ben] resolve conflict between explicit types and json better

		fields := strings.Split(field, ".")

		if len(fields) > 1 {
			switch fields[1] {
			case "roles":
				s = strings.Join(r.Permissions.Roles, ",")
			case "base", "perms":
				s = strconv.Itoa(int(r.Permissions.Base.Perms))
			case "set":
				s = strconv.Itoa(int(r.Permissions.Base.SetBit))
			}
		}
	} else if field == "balance" {
		s = strconv.Itoa(int(r.Balance))
	}

	if err != nil {
		return "", err
	}

	return s, nil
}

func NamesInfo(name, field string, do *definitions.Do) (string, error) {
	nodeClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
	owner, data, expirationBlock, err := nodeClient.GetName(name)
	if err != nil {
		return "", err
	}

	switch strings.ToLower(field) {
	case "name":
		return name, nil
	case "owner":
		return string(owner), nil
	case "data":
		return data, nil
	case "expires":
		return strconv.Itoa(expirationBlock), nil
	default:
		return "", fmt.Errorf("Field %s not recognized", field)
	}
}

func ValidatorsInfo(field string, do *definitions.Do) (string, error) {
	nodeClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
	_, bondedValidators, unbondingValidators, err := nodeClient.ListValidators()
	if err != nil {
		return "", err
	}

	vals := []string{}
	switch strings.ToLower(field) {
	case "bonded_validators":
		for _, v := range bondedValidators {
			vals = append(vals, string(v.Address()))
		}
	case "unbonding_validators":
		for _, v := range unbondingValidators {
			vals = append(vals, string(v.Address()))
		}
	default:
		return "", fmt.Errorf("Field %s not recognized", field)
	}
	return strings.Join(vals, ","), nil
}
