package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/monax/monax/log"

	"github.com/hyperledger/burrow/client/rpc"
)

// This is a closer function which is called by most of the tx_run functions
func ReadTxSignAndBroadcast(result *rpc.TxResult, err error) error {
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

func ReadAbi(root, contract string) (string, error) {
	p := path.Join(root, stripHex(contract))
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("Abi doesn't exist for =>\t%s", p)
	}

	b, err := ioutil.ReadFile(p)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// TODO use go-ethereum/common
func stripHex(s string) string {
	if len(s) > 1 {
		if s[:2] == "0x" {
			s = s[2:]
			if len(s)%2 != 0 {
				s = "0" + s
			}
			return s
		}
	}
	return s
}
