package jobs

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/monax/cli/log"

	"github.com/monax/burrow/client/rpc"
)

// ------------------------------------------------------------------------
// Transaction Jobs
// ------------------------------------------------------------------------

type Send struct {
	// (Optional, if account job or global account set) address of the account from which to send (the
	// public key for the account must be available to keys)
	Source string `mapstructure:"source" yaml:"source"`
	// (Required) address of the account to send the tokens
	Destination string `mapstructure:"destination" yaml:"destination"`
	// (Required) amount of tokens to send from the `source` to the `destination`
	Amount string `mapstructure:"amount" yaml:"amount"`
	// (Optional, advanced only) nonce to use when keys signs the transaction (do not use unless you
	// know what you're doing)
	Nonce string `mapstructure:"nonce" yaml:"nonce"`
}

func (send *Send) PreProcess(jobs *Jobs) (err error) {
	send.Source, _, err = preProcessString(send.Source, jobs)
	if err != nil {
		return err
	}
	send.Destination, _, err = preProcessString(send.Destination, jobs)
	if err != nil {
		return err
	}
	send.Amount, _, err = preProcessString(send.Amount, jobs)
	if err != nil {
		return err
	}
	send.Nonce, _, err = preProcessString(send.Nonce, jobs)
	if err != nil {
		return err
	}
	log.Debug("Default job account: ", jobs.Account)
	send.Source = useDefault(send.Source, jobs.Account)
	send.Amount = useDefault(send.Amount, jobs.DefaultAmount)
	return nil
}

func (send *Send) Execute(jobs *Jobs) (*JobResults, error) {
	// Use Default

	// Don't use pubKey if account override
	var oldKey string
	if send.Source != jobs.Account {
		oldKey = jobs.PublicKey
		jobs.PublicKey = ""
	}

	// Formulate tx
	log.WithFields(log.Fields{
		"source":      send.Source,
		"destination": send.Destination,
		"amount":      send.Amount,
	}).Info("Sending Transaction")

	tx, err := rpc.Send(jobs.NodeClient, jobs.KeyClient, jobs.PublicKey, send.Source, send.Destination, send.Amount, send.Nonce)
	if err != nil {
		return MintChainErrorHandler(jobs, err)
	}

	// Don't use pubKey if account override
	if send.Source != jobs.Account {
		jobs.PublicKey = oldKey
	}

	// Sign, broadcast, display
	return txFinalize(tx, jobs)
}

type RegisterName struct {
	// (Optional, if account job or global account set) address of the account from which to send (the
	// public key for the account must be available to keys)
	Source string `mapstructure:"source" yaml:"source"`
	// (Required) name which will be registered
	Name string `mapstructure:"name" yaml:"name"`
	// (Optional, if data_file is used; otherwise required) data which will be stored at the `name` key
	Data string `mapstructure:"data" yaml:"data"`
	// (Optional) csv file in the form (name,data[,amount]) which can be used to bulk register names
	DataFile string `mapstructure:"data_file" json:"data_file" yaml:"data_file" toml:"data_file"`
	// (Optional) amount of blocks which the name entry will be reserved for the registering user
	Amount string `mapstructure:"amount" yaml:"amount"`
	// (Optional) validators' fee
	Fee string `mapstructure:"fee" yaml:"fee"`
	// (Optional, advanced only) nonce to use when keys signs the transaction (do not use unless you
	// know what you're doing)
	Nonce string `mapstructure:"nonce" yaml:"nonce"`
}

func (name *RegisterName) PreProcess(jobs *Jobs) (err error) {
	name.Source, _, err = preProcessString(name.Source, jobs)
	if err != nil {
		return err
	}
	name.Name, _, err = preProcessString(name.Name, jobs)
	if err != nil {
		return err
	}
	name.Data, _, err = preProcessString(name.Data, jobs)
	if err != nil {
		return err
	}
	name.DataFile, _, err = preProcessString(name.DataFile, jobs)
	if err != nil {
		return err
	}
	name.Amount, _, err = preProcessString(name.Amount, jobs)
	if err != nil {
		return err
	}
	name.Fee, _, err = preProcessString(name.Fee, jobs)
	if err != nil {
		return err
	}
	name.Nonce, _, err = preProcessString(name.Nonce, jobs)
	if err != nil {
		return err
	}

	name.Source = useDefault(name.Source, jobs.Account)
	name.Fee = useDefault(name.Fee, jobs.DefaultFee)
	name.Amount = useDefault(name.Amount, jobs.DefaultAmount)

	if name.DataFile != "" && name.Data != "" {
		return fmt.Errorf("Cannot have both data and datafile field populated.")
	}
	return
}

func (name *RegisterName) Execute(jobs *Jobs) (*JobResults, error) {
	// Don't use pubKey if account override
	var oldKey string
	swapKeyOut := func() {
		if name.Source != jobs.Account {
			oldKey = jobs.PublicKey
			jobs.PublicKey = ""
		}
	}
	swapKeyIn := func() {
		// don't use pubKey if account override
		if name.Source != jobs.Account {
			jobs.PublicKey = oldKey
		}
	}

	switch {
	// If a data file is given it should be in csv format and
	// it will be read first. Once the file is parsed and sent
	// to the chain then a single nameRegTx will be sent if that
	// has been populated.
	case name.DataFile != "":
		// open the file and use a reader
		fileReader, err := os.Open(name.DataFile)
		if err != nil {
			return nil, err
		}

		defer fileReader.Close()
		r := csv.NewReader(fileReader)
		swapKeyOut()
		// loop through the records
		for {
			// Read the record
			record, err := r.Read()

			// Catch the errors
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			// Sink the Amount into the third slot in the record if
			// it doesn't exist
			if len(record) <= 2 {
				record = append(record, name.Amount)
			}

			// Send an individual Tx for the record
			tx, err := rpc.Name(jobs.NodeClient, jobs.KeyClient, jobs.PublicKey, useDefault(record[0], jobs.Account),
				useDefault(record[2], jobs.DefaultAmount), name.Nonce, name.Fee, record[0], record[1])
			if err != nil {
				return nil, err
			}

			resp, err := txFinalize(tx, jobs)
			if err != nil {
				return nil, err
			}

			n := fmt.Sprintf("%s:%s", record[0], record[1])
			// TODO: fix this... simple and naive result just now.
			if err = WriteJobResultCSV(n, resp.FullResult.StringResult); err != nil {
				return nil, err
			}
		}
		swapKeyIn()
		return &JobResults{Type{"data_file_parsed", "data_file_parsed"}, nil}, nil
	case name.Data != "":
		swapKeyOut()
		// Formulate tx
		log.WithFields(log.Fields{
			"name":   name.Name,
			"data":   name.Data,
			"amount": name.Amount,
		}).Info("NameReg Transaction")

		tx, err := rpc.Name(jobs.NodeClient, jobs.KeyClient, jobs.PublicKey, name.Source, name.Amount, name.Nonce, name.Fee, name.Name, name.Data)
		if err != nil {
			return MintChainErrorHandler(jobs, err)
		}
		// Sign, broadcast, display

		res, err := txFinalize(tx, jobs)
		swapKeyIn()
		return res, err
	default:
		return nil, fmt.Errorf("Missing required fields. Please fill data or datafile field.")
	}

}

type Permission struct {
	// (Optional, if account job or global account set) address of the account from which to send (the
	// public key for the account must be available to keys)
	Source string `mapstructure:"source" yaml:"source"`
	// (Required) actions must be in the set ["set_base", "unset_base", "set_global", "add_role" "rm_role"]
	Action string `mapstructure:"action" yaml:"action"`
	// (Required, unless add_role or rm_role action selected) the name of the permission flag which is to
	// be updated
	PermissionFlag string `mapstructure:"permission" yaml:"permission"`
	// (Required) the value of the permission or role which is to be updated
	Value string `mapstructure:"value" yaml:"value"`
	// (Required) the target account which is to be updated
	Target string `mapstructure:"target" yaml:"target"`
	// (Required, if add_role or rm_role action selected) the role which should be given to the account
	Role string `mapstructure:"role" yaml:"role"`
	// (Optional, advanced only) nonce to use when keys signs the transaction (do not use unless you
	// know what you're doing)
	Nonce string `mapstructure:"nonce" yaml:"nonce"`
}

func (perm *Permission) PreProcess(jobs *Jobs) (err error) {
	perm.Source, _, err = preProcessString(perm.Source, jobs)
	if err != nil {
		return err
	}
	perm.Action, _, err = preProcessString(perm.Action, jobs)
	if err != nil {
		return err
	}
	perm.PermissionFlag, _, err = preProcessString(perm.PermissionFlag, jobs)
	if err != nil {
		return err
	}
	perm.Value, _, err = preProcessString(perm.Value, jobs)
	if err != nil {
		return err
	}
	perm.Target, _, err = preProcessString(perm.Target, jobs)
	if err != nil {
		return err
	}
	perm.Role, _, err = preProcessString(perm.Role, jobs)
	if err != nil {
		return err
	}
	// Set defaults
	perm.Source = useDefault(perm.Source, jobs.Account)
	return nil
}

func (perm *Permission) Execute(jobs *Jobs) (*JobResults, error) {
	// Populate the transaction appropriately
	var args []string
	switch perm.Action {
	case "set_global":
		args = []string{perm.PermissionFlag, perm.Value}
	case "set_base":
		args = []string{perm.Target, perm.PermissionFlag, perm.Value}
	case "unset_base":
		args = []string{perm.Target, perm.PermissionFlag}
	case "add_role", "rm_role":
		args = []string{perm.Target, perm.Role}
	}

	// Don't use pubKey if account override
	var oldKey string
	if perm.Source != jobs.Account {
		oldKey = jobs.PublicKey
		jobs.PublicKey = ""
	}

	// Formulate tx
	arg := fmt.Sprintf("%s:%s", args[0], args[1])
	log.WithField(perm.Action, arg).Info("Setting Permissions")

	tx, err := rpc.Permissions(jobs.NodeClient, jobs.KeyClient, jobs.PublicKey, perm.Source, perm.Nonce, perm.Action, args)
	if err != nil {
		return MintChainErrorHandler(jobs, err)
	}

	// Don't use pubKey if account override
	if perm.Source != jobs.Account {
		jobs.PublicKey = oldKey
	}

	// Sign, broadcast, display
	return txFinalize(tx, jobs)
}

type Bond struct {
	// (Required) public key of the address which will be bonded
	PublicKey string `mapstructure:"pub_key" json:"pub_key" yaml:"pub_key" toml:"pub_key"`
	// (Required) address of the account which will be bonded
	Account string `mapstructure:"account" yaml:"account"`
	// (Required) amount of tokens which will be bonded
	Amount string `mapstructure:"amount" yaml:"amount"`
	// (Optional, advanced only) nonce to use when keys signs the transaction (do not use unless you
	// know what you're doing)
	Nonce string `mapstructure:"nonce" yaml:"nonce"`
}

func (bond *Bond) PreProcess(jobs *Jobs) (err error) {
	// Process Variables
	bond.Account, _, err = preProcessString(bond.Account, jobs)
	if err != nil {
		return err
	}
	bond.Amount, _, err = preProcessString(bond.Amount, jobs)
	if err != nil {
		return err
	}
	bond.PublicKey, _, err = preProcessString(bond.PublicKey, jobs)
	if err != nil {
		return err
	}
	bond.Nonce, _, err = preProcessString(bond.Nonce, jobs)
	if err != nil {
		return err
	}
	// Use Defaults
	bond.Account = useDefault(bond.Account, jobs.Account)
	bond.Amount = useDefault(bond.Amount, jobs.DefaultAmount)
	jobs.PublicKey = useDefault(jobs.PublicKey, bond.PublicKey)
	return nil
}

func (bond *Bond) Execute(jobs *Jobs) (*JobResults, error) {
	return nil, fmt.Errorf("Job bond currently unimplemented.")
}

type Unbond struct {
	// (Required) address of the account which to unbond
	Account string `mapstructure:"account" yaml:"account"`
	// (Required) block on which the unbonding will take place (users may unbond at any
	// time >= currentBlock)
	Height string `mapstructure:"height" yaml:"height"`
}

func (unbond *Unbond) PreProcess(jobs *Jobs) (err error) {
	unbond.Account, _, err = preProcessString(unbond.Account, jobs)
	if err != nil {
		return err
	}
	unbond.Height, _, err = preProcessString(unbond.Height, jobs)
	if err != nil {
		return err
	}
	// Use defaults
	unbond.Account = useDefault(unbond.Account, jobs.Account)
	return nil
}

func (unbond *Unbond) Execute(jobs *Jobs) (*JobResults, error) {
	return nil, fmt.Errorf("Job unbond currently unimplemented.")
}

type Rebond struct {
	// (Required) address of the account which to rebond
	Account string `mapstructure:"account" yaml:"account"`
	// (Required) block on which the rebonding will take place (users may rebond at any
	// time >= (unbondBlock || currentBlock))
	Height string `mapstructure:"height" yaml:"height"`
}

func (rebond *Rebond) PreProcess(jobs *Jobs) error {
	// Process Variables
	var err error
	rebond.Account, _, err = preProcessString(rebond.Account, jobs)
	rebond.Height, _, err = preProcessString(rebond.Height, jobs)
	if err != nil {
		return err
	}

	// Use defaults
	rebond.Account = useDefault(rebond.Account, jobs.Account)
	return nil
}

func (rebond *Rebond) Execute(jobs *Jobs) (*JobResults, error) {
	return nil, fmt.Errorf("Job rebond currently unimplemented.")
}
