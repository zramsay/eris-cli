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

package methods

import (
	"fmt"

	log "github.com/eris-ltd/eris-logger"

	"github.com/eris-ltd/eris-db/client"
	"github.com/eris-ltd/eris-db/definitions"
)

func Status(do *definitions.ClientDo) {
	erisNodeClient := client.NewErisNodeClient(do.NodeAddrFlag)
	genesisHash, validatorPublicKey, latestBlockHash, latestBlockHeight, latestBlockTime, err := erisNodeClient.Status()
	if err != nil {
		log.Errorf("Error requesting status from chain at (%s): %s", do.NodeAddrFlag, err)
		return
	}

	chainName, chainId, genesisHashfromChainId, err := erisNodeClient.ChainId()
	if err != nil {
		log.Errorf("Error requesting chainId from chain at (%s): %s", do.NodeAddrFlag, err)
		return
	}

	log.WithFields(log.Fields{
		"chain":                    do.NodeAddrFlag,
		"genesisHash":              fmt.Sprintf("%X", genesisHash),
		"chainName":                chainName,
		"chainId":                  chainId,
		"genesisHash from chainId": fmt.Sprintf("%X", genesisHashfromChainId),
		"validator public key":     fmt.Sprintf("%X", validatorPublicKey),
		"latest block hash":        fmt.Sprintf("%X", latestBlockHash),
		"latest block height":      latestBlockHeight,
		"latest block time":        latestBlockTime,
	}).Info("status")
}
