package jobs

import (
	"encoding/hex"
	"strings"

	"github.com/monax/cli/log"
)

// ------------------------------------------------------------------------
// Util Jobs
// ------------------------------------------------------------------------

type Account struct {
	// (Required) address of the account which should be used as the default (if source) is
	// not given for future transactions. Will make sure the monax-keys has the public key
	// for the account. Generally account should be the first job called unless it is used
	// via a flag or environment variables to establish what default to use.
	Address string `mapstructure:"address" yaml:"address"`
}

func (account *Account) PreProcess(jobs *Jobs) (err error) {
	if account.Address, _, err = preProcessString(account.Address, jobs); err != nil {
		return err
	}
	if strings.HasPrefix(account.Address, "0x") {
		account.Address = strings.TrimPrefix(account.Address, "0x")
	}
	account.Address = strings.ToUpper(account.Address)
	return nil
}

func (account *Account) Execute(jobs *Jobs) (*JobResults, error) {
	log.WithField("=>", account.Address).Debug("Establishing Account")
	jobs.Account = account.Address
	// Set the public key from eris-keys
	addr, err := hex.DecodeString(jobs.Account)
	if err != nil {
		return nil, err
	}
	pubkey, err := jobs.KeyClient.PublicKey(addr)
	if err != nil {
		return KeysErrorHandler(jobs, err)
	}
	jobs.PublicKey = hex.EncodeToString(pubkey)
	return &JobResults{
		FullResult:   Type{account.Address, account.Address},
		NamedResults: nil,
	}, nil
}

type Set struct {
	// (Required) value which should be saved along with the jobName (which will be the key)
	// this is useful to set variables which can be used throughout the epm definition file.
	// It should be noted that arrays and bools must be defined using strings as such "[1,2,3]"
	// if they are intended to be used further in a assert job.
	Value interface{} `mapstructure:"val" yaml:"val"`
}

func (set *Set) PreProcess(jobs *Jobs) (err error) {
	if set.Value, err = preProcessInterface(set.Value, jobs); err != nil {
		return err
	}
	return nil
}

func (set *Set) Execute(jobs *Jobs) (*JobResults, error) {
	log.WithField("=>", set.Value.(Type).StringResult).Debug("Setting Value")
	return &JobResults{
		FullResult:   set.Value.(Type),
		NamedResults: nil,
	}, nil
}
