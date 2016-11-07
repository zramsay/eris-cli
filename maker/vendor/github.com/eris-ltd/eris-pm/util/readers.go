package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/eris-ltd/eris-pm/definitions"

	"github.com/eris-ltd/common/go/common"
	ebi "github.com/eris-ltd/eris-abi/core"
	"github.com/eris-ltd/eris-db/client/core"
	log "github.com/eris-ltd/eris-logger"
)

// This is a closer function which is called by most of the tx_run functions
func ReadTxSignAndBroadcast(result *core.TxResult, err error) error {
	// if there's an error just return.
	if err != nil {
		return err
	}

	// if there is nothing to unpack then just return.
	if result == nil {
		return nil
	}

	// Unpack and display for the user.
	addr := fmt.Sprintf("%X", result.Address)
	hash := fmt.Sprintf("%X", result.Hash)
	blkHash := fmt.Sprintf("%X", result.BlockHash)
	ret := fmt.Sprintf("%X", result.Return)

	if result.Address != nil {
		log.WithField("addr", addr).Warn()
		log.WithField("txHash", hash).Info()
	} else {
		log.WithField("=>", hash).Warn("Transaction Hash")
		log.WithField("=>", blkHash).Debug("Block Hash")
		if len(result.Return) != 0 {
			if ret != "" {
				log.WithField("=>", ret).Warn("Return Value")
			} else {
				log.Debug("No return.")
			}
			log.WithField("=>", result.Exception).Debug("Exception")
		}
	}

	return nil
}

func ReadAbiFormulateCall(abiLocation, funcName string, dataRaw []string, do *definitions.Do) (string, error) {
	abiSpecBytes, err := readAbi(do.ABIPath, abiLocation)
	if err != nil {
		return "", err
	}
	log.WithField("=>", string(abiSpecBytes)).Debug("ABI Specification (Formulate)")
	totalArgs := make([]string, 0)
	totalArgs = append(totalArgs, funcName)
	// Process and Pack the Call
	totalArgs = append(totalArgs, dataRaw...)
	log.WithFields(log.Fields{
		"arguments": totalArgs,
	}).Debug("Packing Call via ABI")

	return ebi.Packer(abiSpecBytes, totalArgs...)
}

func ReadAndDecodeContractReturn(abiLocation string, funcName string, resultRaw string, do *definitions.Do) ([]*definitions.Variable, error) {
	abiSpecBytes, err := readAbi(do.ABIPath, abiLocation)
	if err != nil {
		return nil, err
	}
	log.WithField("=>", string(abiSpecBytes)).Debug("ABI Specification (Decode)")

	// Unpack the result
	res, err := ebi.UnPacker(abiSpecBytes, funcName, resultRaw, false)
	if err != nil {
		return nil, err
	}

	// Wrangle these returns
	type ContractReturn struct {
		Name  string `mapstructure:"," json:","`
		Type  string `mapstructure:"," json:","`
		Value string `mapstructure:"," json:","`
	}
	var resTotal []ContractReturn
	err = json.Unmarshal([]byte(res), &resTotal)
	if err != nil {
		return nil, err
	}

	// Get the values and put them into neat little structs
	result := make([]*definitions.Variable, len(resTotal))
	for index, i := range resTotal {
		if i.Name == "" {
			result[index] = &definitions.Variable{strconv.Itoa(index), i.Value}
		} else {
			result[index] = &definitions.Variable{i.Name, i.Value}
		}
	}

	return result, nil
}

func readAbi(root, contract string) ([]byte, error) {
	p := path.Join(root, common.StripHex(contract))
	if _, err := os.Stat(p); err != nil {
		return []byte{}, fmt.Errorf("Abi doesn't exist for =>\t%s", p)
	}

	b, err := ioutil.ReadFile(p)
	if err != nil {
		return []byte{}, err
	}

	return b, nil
}
