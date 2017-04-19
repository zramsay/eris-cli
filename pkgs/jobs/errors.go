package jobs

import (
	"fmt"
	"regexp"

	"github.com/eris-ltd/eris/log"
)

func MintChainErrorHandler(jobs *Jobs, err error) (*JobResults, error) {
	log.WithFields(log.Fields{
		"defAddr": jobs.Account,
		"chainID": jobs.ChainID,
		//"chainURL": jobs.ChainName,
		"rawErr": err,
	}).Error("")

	return nil, fmt.Errorf(`
There has been an error talking to your eris chain.

%v

Debugging this error is tricky, but don't worry the marmot recovery checklist is...
  * is the %s account right?
  * is the account you want to use in your keys service: eris keys ls ?
  * is the account you want to use in your genesis.json: eris chains cat %s genesis ?
  * is your chain making blocks: eris chains logs -f %s ?
  * do you have permissions to do what you're trying to do on the chain?
`, err, jobs.Account, jobs.ChainID, jobs.ChainID)
}

func KeysErrorHandler(jobs *Jobs, err error) (*JobResults, error) {
	log.WithFields(log.Fields{
		"defAddr": jobs.Account,
	}).Error("")

	r := regexp.MustCompile(fmt.Sprintf("open /home/eris/.eris/keys/data/%s/%s: no such file or directory", jobs.Account, jobs.Account))
	if r.MatchString(fmt.Sprintf("%v", err)) {
		return nil, fmt.Errorf(`
Unfortunately the marmots could not find the key you are trying to use in the keys service.

There are two ways to fix this.
  1. Import your keys from your host: eris keys import %s
  2. Import your keys from your chain:

eris chains exec %s "mintkey eris chains/%s/priv_validator.json" && \
eris services exec keys "chown eris:eris -R /home/eris"

Now, run  eris keys ls  to check that the keys are available. If they are not there
then change the account. Once you have verified that the keys for account

%s

are in the keys service, then rerun me.
`, jobs.Account, jobs.ChainID, jobs.ChainID, jobs.Account)
	}

	return nil, fmt.Errorf(`
There has been an error talking to your eris keys service.

%v

Debugging this error is tricky, but don't worry the marmot recovery checklist is...
  * is your %s account right?
  * is the key for %s in your keys service: eris keys ls ?
`, err, jobs.Account, jobs.Account)
}

/*func ABIErrorHandler(do *definitions.Do, err error, call *definitions.Call, query *definitions.QueryContract) (string, error) {
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
}*/
