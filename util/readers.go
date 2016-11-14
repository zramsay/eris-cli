package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	ebi "github.com/eris-ltd/eris-cli/abi"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"

	"github.com/eris-ltd/eris-db/client/core"

	// TODO eliminate this dep
	"github.com/eris-ltd/common/go/common"
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

func ReadAbiFormulateCall(abiLocation string, funcName string, args []string, do *definitions.Do) ([]byte, error) {
	abiSpecBytes, err := readAbi(do.ABIPath, abiLocation)
	if err != nil {
		return []byte{}, err
	}
	log.WithField("=>", string(abiSpecBytes)).Debug("ABI Specification (Formulate)")
	log.WithFields(log.Fields{
		"function":  funcName,
		"arguments": fmt.Sprintf("%v", args),
	}).Debug("Packing Call via ABI")

	return ebi.Packer(abiSpecBytes, funcName, args...)
}

func ReadAndDecodeContractReturn(abiLocation, funcName string, resultRaw []byte, do *definitions.Do) ([]*definitions.Variable, error) {
	abiSpecBytes, err := readAbi(do.ABIPath, abiLocation)
	if err != nil {
		return nil, err
	}
	log.WithField("=>", abiSpecBytes).Debug("ABI Specification (Decode)")

	// Unpack the result
	return ebi.Unpacker(abiSpecBytes, funcName, resultRaw)
}

func readAbi(root, contract string) (string, error) {
	p := path.Join(root, common.StripHex(contract))
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("Abi doesn't exist for =>\t%s", p)
	}

	b, err := ioutil.ReadFile(p)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
