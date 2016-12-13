// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"encoding/hex"
	"fmt"
	"strconv"

	log "github.com/eris-ltd/eris-logger"

	"github.com/tendermint/go-crypto"

	acc "github.com/eris-ltd/eris-db/account"
	"github.com/eris-ltd/eris-db/client"
	"github.com/eris-ltd/eris-db/keys"
	ptypes "github.com/eris-ltd/eris-db/permission/types"
	"github.com/eris-ltd/eris-db/txs"
)

//------------------------------------------------------------------------------------
// sign and broadcast convenience

// tx has either one input or we default to the first one (ie for send/bond)
// TODO: better support for multisig and bonding
func signTx(keyClient keys.KeyClient, chainID string, tx_ txs.Tx) ([]byte, txs.Tx, error) {
	signBytesString := fmt.Sprintf("%X", acc.SignBytes(chainID, tx_))
	var inputAddr []byte
	var sigED crypto.SignatureEd25519
	switch tx := tx_.(type) {
	case *txs.SendTx:
		inputAddr = tx.Inputs[0].Address
		defer func(s *crypto.SignatureEd25519) { tx.Inputs[0].Signature = *s }(&sigED)
	case *txs.NameTx:
		inputAddr = tx.Input.Address
		defer func(s *crypto.SignatureEd25519) { tx.Input.Signature = *s }(&sigED)
	case *txs.CallTx:
		inputAddr = tx.Input.Address
		defer func(s *crypto.SignatureEd25519) { tx.Input.Signature = *s }(&sigED)
	case *txs.PermissionsTx:
		inputAddr = tx.Input.Address
		defer func(s *crypto.SignatureEd25519) { tx.Input.Signature = *s }(&sigED)
	case *txs.BondTx:
		inputAddr = tx.Inputs[0].Address
		defer func(s *crypto.SignatureEd25519) {
			tx.Signature = *s
			tx.Inputs[0].Signature = *s
		}(&sigED)
	case *txs.UnbondTx:
		inputAddr = tx.Address
		defer func(s *crypto.SignatureEd25519) { tx.Signature = *s }(&sigED)
	case *txs.RebondTx:
		inputAddr = tx.Address
		defer func(s *crypto.SignatureEd25519) { tx.Signature = *s }(&sigED)
	}
	sig, err := keyClient.Sign(signBytesString, inputAddr)
	if err != nil {
		return nil, nil, err
	}
	// TODO: [ben] temporarily address the type conflict here, to be cleaned up
	// with full type restructuring
	var sig64 [64]byte
	copy(sig64[:], sig)
	sigED = crypto.SignatureEd25519(sig64)
	return inputAddr, tx_, nil
}

func decodeAddressPermFlag(addrS, permFlagS string) (addr []byte, pFlag ptypes.PermFlag, err error) {
	if addr, err = hex.DecodeString(addrS); err != nil {
		return
	}
	if pFlag, err = ptypes.PermStringToFlag(permFlagS); err != nil {
		return
	}
	return
}

func checkCommon(nodeClient client.NodeClient, keyClient keys.KeyClient, pubkey, addr, amtS, nonceS string) (pub crypto.PubKey, amt int64, nonce int64, err error) {
	if amtS == "" {
		err = fmt.Errorf("input must specify an amount with the --amt flag")
		return
	}

	var pubKeyBytes []byte
	if pubkey == "" && addr == "" {
		err = fmt.Errorf("at least one of --pubkey or --addr must be given")
		return
	} else if pubkey != "" {
		if addr != "" {
			log.WithFields(log.Fields{
				"public key": pubkey,
				"address":    addr,
			}).Info("you have specified both a pubkey and an address. the pubkey takes precedent")
		}
		pubKeyBytes, err = hex.DecodeString(pubkey)
		if err != nil {
			err = fmt.Errorf("pubkey is bad hex: %v", err)
			return
		}
	} else {
		// grab the pubkey from eris-keys
		addressBytes, err2 := hex.DecodeString(addr)
		if err2 != nil {
			err = fmt.Errorf("Bad hex string for address (%s): %v", addr, err)
			return
		}
		pubKeyBytes, err2 = keyClient.PublicKey(addressBytes)
		if err2 != nil {
			err = fmt.Errorf("Failed to fetch pubkey for address (%s): %v", addr, err2)
			return
		}
	}

	if len(pubKeyBytes) == 0 {
		err = fmt.Errorf("Error resolving public key")
		return
	}

	amt, err = strconv.ParseInt(amtS, 10, 64)
	if err != nil {
		err = fmt.Errorf("amt is misformatted: %v", err)
	}

	var pubArray [32]byte
	copy(pubArray[:], pubKeyBytes)
	pub = crypto.PubKeyEd25519(pubArray)
	addrBytes := pub.Address()

	if nonceS == "" {
		if nodeClient == nil {
			err = fmt.Errorf("input must specify a nonce with the --nonce flag or use --node-addr (or ERIS_CLIENT_NODE_ADDR) to fetch the nonce from a node")
			return
		}
		// fetch nonce from node
		account, err2 := nodeClient.GetAccount(addrBytes)
		if err2 != nil {
			return pub, amt, nonce, err2
		}
		nonce = int64(account.Sequence) + 1
		log.WithFields(log.Fields{
			"nonce":           nonce,
			"account address": fmt.Sprintf("%X", addrBytes),
		}).Debug("Fetch nonce from node")
	} else {
		nonce, err = strconv.ParseInt(nonceS, 10, 64)
		if err != nil {
			err = fmt.Errorf("nonce is misformatted: %v", err)
			return
		}
	}

	return
}
