package util

import (
	"fmt"
	"regexp"

	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
)

func MintChainErrorHandler(do *definitions.Do, err error) (string, error) {
	log.WithFields(log.Fields{
		"defAddr":   do.Package.Account,
		"chainID":   do.ChainID,
		"chainNAME": do.ChainName,
		"rawErr":    err,
	}).Error("")

	return "", fmt.Errorf(`
There has been an error talking to your monax chain.

%v

Debugging this error is tricky, but don't worry the marmot recovery checklist is...
  * is the %s account right?
  * is the account you want to use in your keys service: monax keys ls ?
  * is the account you want to use in your genesis.json: see http://localhost:46657/genesis
  * is your chain making blocks: monax chains logs -f %s ?
  * do you have permissions to do what you're trying to do on the chain?
`, err, do.Package.Account, do.ChainName)
}

func KeysErrorHandler(do *definitions.Do, err error) (string, error) {
	log.WithFields(log.Fields{
		"defAddr": do.Package.Account,
	}).Error("")

	r := regexp.MustCompile(fmt.Sprintf("open /home/monax/.monax/keys/data/%s/%s: no such file or directory", do.Package.Account, do.Package.Account))
	if r.MatchString(fmt.Sprintf("%v", err)) {
		return "", fmt.Errorf(`
Unfortunately the marmots could not find the key you are trying to use in the keys service.

There is one way to fix this.
  * Import your keys from your host: monax keys import %s

Now, run  monax keys ls  to check that the keys are available. If they are not there
then change the account. Once you have verified that the keys for account

%s

are in the keys service, then rerun me.
`, do.Package.Account, do.Package.Account)
	}

	return "", fmt.Errorf(`
There has been an error talking to your monax keys service.

%v

Debugging this error is tricky, but don't worry the marmot recovery checklist is...
  * is your %s account right?
  * is the key for %s in your keys service: monax keys ls ?
`, err, do.Package.Account, do.Package.Account)
}

func ABIErrorHandler(do *definitions.Do, err error, call *definitions.Call, query *definitions.QueryContract) (string, error) {
	switch {
	case call != nil:
		log.WithFields(log.Fields{
			"data":   call.Data,
			"abi":    call.ABI,
			"dest":   call.Destination,
			"rawErr": err,
		}).Error("ABI Error")
	case query != nil:
		log.WithFields(log.Fields{
			"data":   query.Data,
			"abi":    query.ABI,
			"dest":   query.Destination,
			"rawErr": err,
		}).Error("ABI Error")
	}

	return "", fmt.Errorf(`
There has been an error in finding or in using your ABI. ABI's are "Application Binary
Interface" and they are what let us know how to talk to smart contracts.

These little json files can be read by a variety of things which need to talk to smart
contracts so they are quite necessary to be able to find and use properly.

The ABIs are saved after the deploy events. So if there was a glitch in the matrix,
we apologize in advance.

The marmot recovery checklist is...
  * ensure your chain is running and you have enough validators online
  * ensure that your contracts successfully deployed
  * if you used imports or have multiple contracts in one file check the instance
    variable in the deploy and the abi variable in the call/query-contract
  * make sure you're calling or querying the right function
  * make sure you're using the correct variables for job results
`)
}
