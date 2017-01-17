package maker

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"

	log "github.com/eris-ltd/eris-cli/log"
	keys "github.com/eris-ltd/eris-keys/eris-keys"
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
		"path": keys.DaemonAddr,
		"type": keyType,
	}).Debug("Sending Call to eris-keys server")

	var err error
	log.WithField("endpoint", "gen").Debug()
	account.Address, err = keys.Call("gen", map[string]string{"auth": "", "type": keyType, "name": account.Name}) // note, for now we use not password to lock/unlock keys
	if _, ok := err.(keys.ErrConnectionRefused); ok {
		return fmt.Errorf("Could not connect to eris-keys server. Start it with `eris services start keys`. Error: %v", err)
	}
	if err != nil {
		return err
	}

	log.WithField("endpoint", "pub").Debug()
	account.PubKey, err = keys.Call("pub", map[string]string{"addr": account.Address, "name": account.Name})
	if _, ok := err.(keys.ErrConnectionRefused); ok {
		return fmt.Errorf("Could not connect to eris-keys server. Start it with `eris services start keys`. Error: %v", err)
	}
	if err != nil {
		return err
	}

	// log.WithField("endpoint", "to-mint").Debug()
	// mint, err := keys.Call("to-mint", map[string]string{"addr": account.Address, "name": account.Name})

	log.WithField("endpoint", "mint").Debug()
	mint, err := keys.Call("mint", map[string]string{"addr": account.Address, "name": account.Name})
	if _, ok := err.(keys.ErrConnectionRefused); ok {
		return fmt.Errorf("Could not connect to eris-keys server. Start it with `eris services start keys`. Error: %v", err)
	}
	if err != nil {
		return err
	}
	// [zr] leave MintKey / MintPrivValidator
	account.MintKey = &definitions.MintPrivValidator{}
	err = json.Unmarshal([]byte(mint), account.MintKey)
	if err != nil {
		log.Error(string(mint))
		log.Error(account.MintKey)
		return err
	}

	account.MintKey.Address = account.Address
	return nil
}
