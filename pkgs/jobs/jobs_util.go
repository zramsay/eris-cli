package jobs

import (
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/util"

	keys "github.com/eris-ltd/eris-keys/eris-keys"
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
	keys.DaemonAddr = do.Signer
	log.WithField("from", keys.DaemonAddr).Info("Getting Public Key")
	do.PublicKey, err = keys.Call("pub", map[string]string{"addr": do.Package.Account, "name": ""})
	if _, ok := err.(keys.ErrConnectionRefused); ok {
		keys.ExitConnectErr(err)
	}

	if err != nil {
		return util.KeysErrorHandler(do, err)
	}

	// Set result and return
	result = account.Address
	return result, nil
}

func SetValJob(set *definitions.Set, do *definitions.Do) (string, error) {
	var result string
	set.Value, _ = util.PreProcess(set.Value, do)
	log.WithField("=>", set.Value).Info("Setting Variable")
	result = set.Value
	return result, nil
}
