package maker

import (
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/util"
)

func MakeErisDBChain(name string, accounts []*definitions.ErisDBAccount, chainImageName string,
	useDataContainer bool, exportedPorts []string, containerEntrypoint string) error {
	genesis := definitions.BlankGenesis()
	genesis.ChainID = name
	for _, account := range accounts {
		log.WithFields(log.Fields{
			"name":    account.Name,
			"address": account.Address,
			"tokens":  account.Tokens,
			"perms":   account.ErisDBPermissions.ErisDBBase.ErisDBPerms,
		}).Debug("Making an ErisDB Account")

		thisAct := MakeErisDBAccount(account)
		genesis.Accounts = append(genesis.Accounts, thisAct)

		if account.Validator {
			thisVal := MakeErisDBValidator(account)
			genesis.Validators = append(genesis.Validators, thisVal)
		}
	}
	for _, account := range accounts {
		if err := util.WritePrivVals(genesis.ChainID, account, len(accounts) == 1); err != nil {
			return err
		}
		if err := util.WriteGenesisFile(genesis.ChainID, genesis, account, len(accounts) == 1); err != nil {
			return err
		}
		// TODO: [ben] we can expose seeds to be written into the configuration file
		// here, but currently not used and we'll overwrite the configuration file
		// with flag or environment variable in eris-db container
		if err := util.WriteConfigurationFile(genesis.ChainID, account.Name, "",
			len(accounts) == 1, chainImageName, useDataContainer, exportedPorts,
			containerEntrypoint); err != nil {
			return err
		}
	}
	return nil
}

func MakeErisDBAccount(account *definitions.ErisDBAccount) *definitions.ErisDBAccount {
	mintAct := &definitions.ErisDBAccount{}
	mintAct.Address = account.Address
	mintAct.Amount = account.Tokens
	mintAct.Name = account.Name
	mintAct.Permissions = account.ErisDBPermissions
	return mintAct
}

func MakeErisDBValidator(account *definitions.ErisDBAccount) *definitions.ErisDBValidator {
	mintVal := &definitions.ErisDBValidator{}
	mintVal.Name = account.Name
	mintVal.Amount = account.ToBond
	mintVal.UnbondTo = append(mintVal.UnbondTo, &definitions.ErisDBTxOutput{
		Address: account.Address,
		Amount:  account.ToBond,
	})
	mintVal.PubKey = append(mintVal.PubKey, 1)
	mintVal.PubKey = append(mintVal.PubKey, account.PubKey)
	return mintVal
}
