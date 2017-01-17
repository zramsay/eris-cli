package maker

import (
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
)

func MakeErisDBChain(name string, seeds []string, accounts []*definitions.ErisDBAccount, chainImageName string,
	useDataContainer bool, exportedPorts []string, containerEntrypoint string) error {
	genesis := definitions.BlankGenesis()
	genesis.ChainID = name
	for _, account := range accounts {
		log.WithFields(log.Fields{
			"name":    account.Name,
			"address": account.Address,
			"tokens":  account.Amount,
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
		if err := WritePrivVals(genesis.ChainID, account); err != nil {
			return err
		}
		if err := WriteGenesisFile(genesis.ChainID, genesis, account); err != nil {
			return err
		}
		// TODO: [ben] we can expose seeds to be written into the configuration file
		// here, but currently not used and we'll overwrite the configuration file
		// with flag or environment variable in eris-db container
		theSeeds := strings.Join(seeds, ",") // format for config file (if len>1)
		if err := WriteConfigurationFile(genesis.ChainID, account.Name, theSeeds,
			chainImageName, useDataContainer, exportedPorts, containerEntrypoint); err != nil {
			return err
		}
	}
	return nil
}

func MakeErisDBAccount(account *definitions.ErisDBAccount) *definitions.ErisDBAccount {
	mintAct := &definitions.ErisDBAccount{}
	mintAct.Address = account.Address
	mintAct.Amount = account.Amount
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
