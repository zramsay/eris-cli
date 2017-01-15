package jobs

import (
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/keys"
)

func SetAccountJob(account *definitions.Account, do *definitions.Do) (string, error) {
	var result string
	var err error

	// Preprocess
	account.Address, _ = util.PreProcess(account.Address, do)

	// Set the Account in the Package & Announce
	do.Package.Account = account.Address
	log.WithField("=>", do.Package.Account).Info("Setting Account")

	// Set the public key from eris-keys
	keyClient, err := keys.InitKeyClient()
	if err != nil {
		return util.KeysErrorHandler(do, err)
	}
	do.PublicKey, err = keyClient.PubKey(do.Package.Account)
	if err != nil {
		return util.KeysErrorHandler(do, err)
	}

	// Set result and return
	result = account.Address
	return result, nil
}

func SetValJob(set *definitions.SetJob, do *definitions.Do) (string, error) {
	var result string
	set.Value, _ = util.PreProcess(set.Value, do)
	log.WithField("=>", set.Value).Info("Setting Variable")
	result = set.Value
	return result, nil
}
