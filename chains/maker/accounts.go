package maker

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eris-ltd/eris/definitions"
	"github.com/eris-ltd/eris/keys"
	"github.com/eris-ltd/eris/log"
)

func MakeAccounts(name, chainType string, accountTypes []*definitions.ErisDBAccountType) ([]*definitions.ErisDBAccount, error) {
	accounts := []*definitions.ErisDBAccount{}

	for _, accountT := range accountTypes {
		log.WithField("type", accountT.Name).Info("Making Account Type")

		perms := &definitions.ErisDBAccountPermissions{}
		var err error
		if chainType == "mint" {
			perms, err = ErisDBAccountPermissions(accountT.Perms, []string{}) // TODO: expose roles
			if err != nil {
				return nil, err
			}
		}

		for i := 0; i < accountT.Number; i++ {
			thisAct := &definitions.ErisDBAccount{}
			thisAct.Name = fmt.Sprintf("%s_%s_%03d", name, accountT.Name, i)
			thisAct.Name = strings.ToLower(thisAct.Name)

			log.WithField("name", thisAct.Name).Debug("Making Account")

			thisAct.Amount = accountT.Tokens
			thisAct.ToBond = accountT.ToBond

			thisAct.PermissionsMap = accountT.Perms
			thisAct.Validator = false

			if thisAct.ToBond != 0 {
				thisAct.Validator = true
			}

			if chainType == "mint" {
				thisAct.ErisDBPermissions = &definitions.ErisDBAccountPermissions{}
				thisAct.ErisDBPermissions.ErisDBBase = &definitions.ErisDBBasePermissions{}
				thisAct.ErisDBPermissions.ErisDBBase.ErisDBPerms = perms.ErisDBBase.ErisDBPerms
				thisAct.ErisDBPermissions.ErisDBBase.ErisDBSetBit = perms.ErisDBBase.ErisDBSetBit
				thisAct.ErisDBPermissions.ErisDBRoles = perms.ErisDBRoles
				log.WithField("perms", thisAct.ErisDBPermissions.ErisDBBase.ErisDBPerms).Debug()

				if err := makeKey("ed25519,ripemd160", thisAct); err != nil {
					return nil, err
				}
			}

			accounts = append(accounts, thisAct)
		}
	}

	return accounts, nil
}

func makeKey(keyType string, account *definitions.ErisDBAccount) error {
	log.WithFields(log.Fields{
		"type": keyType,
	}).Debug("Sending Call to eris-keys server")

	keyClient, err := keys.InitKeyClient()

	account.Address, err = keyClient.GenerateKey(false, true, keyType, "") // note, for now we use not password to lock/unlock keys
	if err != nil {
		return err
	}

	account.PubKey, err = keyClient.PubKey(account.Address, "")
	if err != nil {
		return err
	}

	mint, err := keyClient.Convert(account.Address, "")
	if err != nil {
		return err
	}
	// [zr] leave MintKey / MintPrivValidator
	account.MintKey = &definitions.MintPrivValidator{}
	err = json.Unmarshal(mint, account.MintKey)
	if err != nil {
		log.Error(string(mint))
		log.Error(account.MintKey)
		return err
	}

	account.MintKey.Address = account.Address
	return nil
}
