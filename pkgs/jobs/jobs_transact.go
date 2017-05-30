package jobs

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/util"

	"github.com/hyperledger/burrow/client"
	"github.com/hyperledger/burrow/client/rpc"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/txs"
)

func SendJob(send *definitions.Send, do *definitions.Do) (string, error) {

	// Process Variables
	send.Source, _ = util.PreProcess(send.Source, do)
	send.Destination, _ = util.PreProcess(send.Destination, do)
	send.Amount, _ = util.PreProcess(send.Amount, do)

	// Use Default
	send.Source = useDefault(send.Source, do.Package.Account)

	// Don't use pubKey if account override
	var oldKey string
	if send.Source != do.Package.Account {
		oldKey = do.PublicKey
		do.PublicKey = ""
	}

	// Formulate tx
	log.WithFields(log.Fields{
		"source":      send.Source,
		"destination": send.Destination,
		"amount":      send.Amount,
	}).Info("Sending Transaction")

	monaxNodeClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
	monaxKeyClient := keys.NewBurrowKeyClient(do.Signer, loggers.NewNoopInfoTraceLogger())
	tx, err := rpc.Send(monaxNodeClient, monaxKeyClient, do.PublicKey, send.Source, send.Destination, send.Amount, send.Nonce)
	if err != nil {
		return util.MintChainErrorHandler(do, err)
	}

	// Don't use pubKey if account override
	if send.Source != do.Package.Account {
		do.PublicKey = oldKey
	}

	// Sign, broadcast, display
	return txFinalize(do, tx)
}

func RegisterNameJob(name *definitions.RegisterName, do *definitions.Do) (string, error) {
	// Process Variables
	name.DataFile, _ = util.PreProcess(name.DataFile, do)

	// If a data file is given it should be in csv format and
	// it will be read first. Once the file is parsed and sent
	// to the chain then a single nameRegTx will be sent if that
	// has been populated.
	if name.DataFile != "" {
		// open the file and use a reader
		fileReader, err := os.Open(name.DataFile)
		if err != nil {
			return "", err
		}

		defer fileReader.Close()
		r := csv.NewReader(fileReader)

		// loop through the records
		for {
			// Read the record
			record, err := r.Read()

			// Catch the errors
			if err == io.EOF {
				break
			}
			if err != nil {
				return "", err
			}

			// Sink the Amount into the third slot in the record if
			// it doesn't exist
			if len(record) <= 2 {
				record = append(record, name.Amount)
			}

			// Send an individual Tx for the record
			// [TODO]: move these to async using goroutines?
			r, err := registerNameTx(&definitions.RegisterName{
				Source: name.Source,
				Name:   record[0],
				Data:   record[1],
				Amount: record[2],
				Fee:    name.Fee,
				Nonce:  name.Nonce,
			}, do)

			if err != nil {
				return "", err
			}

			n := fmt.Sprintf("%s:%s", record[0], record[1])

			// TODO: write smarter
			if err = WriteJobResultCSV(n, r); err != nil {
				return "", err
			}
		}
	}

	// If the data field is populated then there is a single
	// nameRegTx to send. So do that *now*.
	if name.Data != "" {
		return registerNameTx(name, do)
	} else {
		return "data_file_parsed", nil
	}
}

// Runs an individual nametx.
func registerNameTx(name *definitions.RegisterName, do *definitions.Do) (string, error) {
	// Process Variables
	name.Source, _ = util.PreProcess(name.Source, do)
	name.Name, _ = util.PreProcess(name.Name, do)
	name.Data, _ = util.PreProcess(name.Data, do)
	name.Amount, _ = util.PreProcess(name.Amount, do)
	name.Fee, _ = util.PreProcess(name.Fee, do)

	// Set Defaults
	name.Source = useDefault(name.Source, do.Package.Account)
	name.Fee = useDefault(name.Fee, do.DefaultFee)
	name.Amount = useDefault(name.Amount, do.DefaultAmount)

	// Don't use pubKey if account override
	var oldKey string
	if name.Source != do.Package.Account {
		oldKey = do.PublicKey
		do.PublicKey = ""
	}

	// Formulate tx
	log.WithFields(log.Fields{
		"name":   name.Name,
		"data":   name.Data,
		"amount": name.Amount,
	}).Info("NameReg Transaction")

	monaxNodeClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
	monaxKeyClient := keys.NewBurrowKeyClient(do.Signer, loggers.NewNoopInfoTraceLogger())
	tx, err := rpc.Name(monaxNodeClient, monaxKeyClient, do.PublicKey, name.Source, name.Amount, name.Nonce, name.Fee, name.Name, name.Data)
	if err != nil {
		return util.MintChainErrorHandler(do, err)
	}

	// Don't use pubKey if account override
	if name.Source != do.Package.Account {
		do.PublicKey = oldKey
	}

	// Sign, broadcast, display
	return txFinalize(do, tx)
}

func PermissionJob(perm *definitions.Permission, do *definitions.Do) (string, error) {
	// Process Variables
	perm.Source, _ = util.PreProcess(perm.Source, do)
	perm.Action, _ = util.PreProcess(perm.Action, do)
	perm.PermissionFlag, _ = util.PreProcess(perm.PermissionFlag, do)
	perm.Value, _ = util.PreProcess(perm.Value, do)
	perm.Target, _ = util.PreProcess(perm.Target, do)
	perm.Role, _ = util.PreProcess(perm.Role, do)

	// Set defaults
	perm.Source = useDefault(perm.Source, do.Package.Account)

	log.Debug("Target: ", perm.Target)
	log.Debug("Marmots Deny: ", perm.Role)
	log.Debug("Action: ", perm.Action)
	// Populate the transaction appropriately
	var args []string
	switch perm.Action {
	case "setGlobal":
		args = []string{perm.PermissionFlag, perm.Value}
	case "setBase":
		args = []string{perm.Target, perm.PermissionFlag, perm.Value}
	case "unsetBase":
		args = []string{perm.Target, perm.PermissionFlag}
	case "addRole", "removeRole":
		args = []string{perm.Target, perm.Role}
	}

	// Don't use pubKey if account override
	var oldKey string
	if perm.Source != do.Package.Account {
		oldKey = do.PublicKey
		do.PublicKey = ""
	}

	// Formulate tx
	//arg := fmt.Sprintf("%s:%s", args[0], args[1])
	//log.WithField(perm.Action, arg).Info("Setting Permissions")

	monaxNodeClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
	monaxKeyClient := keys.NewBurrowKeyClient(do.Signer, loggers.NewNoopInfoTraceLogger())
	tx, err := rpc.Permissions(monaxNodeClient, monaxKeyClient, do.PublicKey, perm.Source, perm.Nonce, perm.Action, args)
	if err != nil {
		return util.MintChainErrorHandler(do, err)
	}

	log.Debug("What are the args returned in transaction: ", tx.PermArgs)

	// Don't use pubKey if account override
	if perm.Source != do.Package.Account {
		do.PublicKey = oldKey
	}

	// Sign, broadcast, display
	return txFinalize(do, tx)
}

func BondJob(bond *definitions.Bond, do *definitions.Do) (string, error) {
	// Process Variables
	bond.Account, _ = util.PreProcess(bond.Account, do)
	bond.Amount, _ = util.PreProcess(bond.Amount, do)
	bond.PublicKey, _ = util.PreProcess(bond.PublicKey, do)

	// Use Defaults
	bond.Account = useDefault(bond.Account, do.Package.Account)
	do.PublicKey = useDefault(do.PublicKey, bond.PublicKey)

	// Formulate tx
	log.WithFields(log.Fields{
		"public key": do.PublicKey,
		"amount":     bond.Amount,
	}).Infof("Bond Transaction")

	monaxNodeClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
	monaxKeyClient := keys.NewBurrowKeyClient(do.Signer, loggers.NewNoopInfoTraceLogger())
	tx, err := rpc.Bond(monaxNodeClient, monaxKeyClient, do.PublicKey, bond.Account, bond.Amount, bond.Nonce)
	if err != nil {
		return util.MintChainErrorHandler(do, err)
	}

	// Sign, broadcast, display
	return txFinalize(do, tx)
}

func UnbondJob(unbond *definitions.Unbond, do *definitions.Do) (string, error) {
	// Process Variables
	var err error
	unbond.Account, err = util.PreProcess(unbond.Account, do)
	if err != nil {
		return "", err
	}
	unbond.Height, err = util.PreProcess(unbond.Height, do)
	if err != nil {
		return "", err
	}

	// Use defaults
	unbond.Account = useDefault(unbond.Account, do.Package.Account)

	// Don't use pubKey if account override
	var oldKey string
	if unbond.Account != do.Package.Account {
		oldKey = do.PublicKey
		do.PublicKey = ""
	}

	// Formulate tx
	log.WithFields(log.Fields{
		"account": unbond.Account,
		"height":  unbond.Height,
	}).Info("Unbond Transaction")

	tx, err := rpc.Unbond(unbond.Account, unbond.Height)
	if err != nil {
		return util.MintChainErrorHandler(do, err)
	}

	// Don't use pubKey if account override
	if unbond.Account != do.Package.Account {
		do.PublicKey = oldKey
	}

	// Sign, broadcast, display
	return txFinalize(do, tx)
}

func RebondJob(rebond *definitions.Rebond, do *definitions.Do) (string, error) {
	// Process Variables
	var err error
	rebond.Account, err = util.PreProcess(rebond.Account, do)
	if err != nil {
		return "", err
	}
	rebond.Height, err = util.PreProcess(rebond.Height, do)
	if err != nil {
		return "", err
	}

	// Use defaults
	rebond.Account = useDefault(rebond.Account, do.Package.Account)

	// Don't use pubKey if account override
	var oldKey string
	if rebond.Account != do.Package.Account {
		oldKey = do.PublicKey
		do.PublicKey = ""
	}

	// Formulate tx
	log.WithFields(log.Fields{
		"account": rebond.Account,
		"height":  rebond.Height,
	}).Info("Rebond Transaction")

	tx, err := rpc.Rebond(rebond.Account, rebond.Height)
	if err != nil {
		return util.MintChainErrorHandler(do, err)
	}

	// Don't use pubKey if account override
	if rebond.Account != do.Package.Account {
		do.PublicKey = oldKey
	}

	// Sign, broadcast, display
	return txFinalize(do, tx)
}

func txFinalize(do *definitions.Do, tx interface{}) (string, error) {
	var result string

	monaxNodeClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
	monaxKeyClient := keys.NewBurrowKeyClient(do.Signer, loggers.NewNoopInfoTraceLogger())
	res, err := rpc.SignAndBroadcast(do.ChainID, monaxNodeClient, monaxKeyClient, tx.(txs.Tx), true, true, true)
	if err != nil {
		return util.MintChainErrorHandler(do, err)
	}

	if err := util.ReadTxSignAndBroadcast(res, err); err != nil {
		return "", err
	}

	result = fmt.Sprintf("%X", res.Hash)
	return result, nil
}

func useDefault(thisOne, defaultOne string) string {
	if thisOne == "" {
		return defaultOne
	}
	return thisOne
}
