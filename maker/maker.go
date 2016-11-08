package maker

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	definitions "github.com/eris-ltd/eris-cli/definitions/maker"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/util"
)

var (
	// TODO: [csk] move to a global config struct
	reader *os.File = os.Stdin
)

func MakeChain(do *definitions.Do) error {
	switch {
	case len(do.AccountTypes) != 0:
		log.Info("Making chain using account type paradigm.")
		return makeRaw(do, "accounttype")
	case do.ChainType != "":
		log.Info("Making chain using chain type paradigm.")
		return makeRaw(do, "chaintype")
	case do.CSV != "":
		log.Info("Making chain using csv type paradigm.")
		return makeRaw(do, "csv")
	default:
		log.Info("Making chain using wizard paradigm.")
		return makeWizard(do)
	}
	return nil
}

func makeWizard(do *definitions.Do) error {
	proceed, err := util.GetBoolResponse(ChainsMakeWelcome(), true, os.Stdin)
	log.WithField("=>", proceed).Debug("What the marmots heard")
	if err != nil {
		return err
	}

	if !proceed {
		log.Warn("The marmots will not proceed without your authorization. Exiting.")
		return nil
	}

	prelims := make(map[string]bool)
	for e, q := range ChainsMakePrelimQuestions() {
		prelims[e], err = util.GetBoolResponse(q, false, os.Stdin)
		log.WithField("=>", prelims[e]).Debug("What the marmots heard")
		if err != nil {
			return err
		}
	}

	accountTypes, err := LoadAccountTypes()
	if err != nil {
		return err
	}

	for _, accountT := range accountTypes {
		if err := assembleTypesWizard(accountT, prelims["tokens"]); err != nil {
			return err
		}
	}

	if prelims["dryrun"] {
		// todo check if procede or return....
	}

	if prelims["manual"] {
		var err error
		accountTypes, err = addManualAccountType(accountTypes, 0)
		if err != nil {
			return err
		}
	}

	return maker(do, "mint", accountTypes)
}

func makeRaw(do *definitions.Do, typ string) error {
	accountTypes, err := LoadAccountTypes()
	if err != nil {
		return err
	}
	log.Debug("Account types loaded.")

	if err := assembleTypesRaw(accountTypes, do, typ); err != nil {
		return err
	}

	return maker(do, "mint", accountTypes)
}

func maker(do *definitions.Do, consensus_type string, accountTypes []*definitions.AccountType) error {
	var err error
	do.Accounts, err = MakeAccounts(do.Name, consensus_type, accountTypes)
	if err != nil {
		return err
	}

	return MakeMintChain(do.Name, do.Accounts, do.ChainImageName,
		do.UseDataContainer, do.ExportedPorts, do.ContainerEntrypoint)
}

func assembleTypesWizard(accountT *definitions.AccountType, tokenIze bool) error {
	var err error
	accountT.Number, err = util.GetIntResponse(AccountTypeIntro(accountT), accountT.Number, reader)
	log.WithField("=>", accountT.Number).Debug("What the marmots heard")
	if err != nil {
		return err
	}

	if tokenIze && accountT.Number > 0 {
		accountT.Tokens, err = util.GetIntResponse(AccountTypeTokens(accountT), accountT.Tokens, reader)
		log.WithField("=>", accountT.Tokens).Debug("What the marmots heard")
		if err != nil {
			return err
		}
	}

	if accountT.Perms["bond"] == 1 && accountT.Number > 0 {
		accountT.ToBond, err = util.GetIntResponse(AccountTypeToBond(accountT), accountT.ToBond, reader)
		log.WithField("=>", accountT.ToBond).Debug("What the marmots heard")
		if err != nil {
			return err
		}
	} else {
		log.Info("Setting accountType.ToBond to 0")
		log.WithField("=>", accountT.Name).Debug("No bond permissions")
		accountT.ToBond = 0
	}

	return nil
}

func addManualAccountType(accountT []*definitions.AccountType, iterator int) ([]*definitions.AccountType, error) {
	var err error
	thisActT := &definitions.AccountType{}
	thisActT.Name = fmt.Sprintf("%s_%02d", "manual", iterator)
	iterator++

	thisActT.Number, err = util.GetIntResponse(AccountTypeManualIntro(), 1, reader)
	if err != nil {
		return nil, err
	}

	thisActT.Tokens, err = util.GetIntResponse(AccountTypeManualTokens(), 0, reader)
	if err != nil {
		return nil, err
	}

	thisActT.ToBond, err = util.GetIntResponse(AccountTypeManualToBond(), 0, reader)
	if err != nil {
		return nil, err
	}

	thisActT.Perms = make(map[string]int)
	for _, perm := range AccountTypeManualPerms() {
		thisActT.Perms[perm], err = util.GetIntResponse(AccountTypeManualPermsQuestion(perm), 0, reader)
	}

	name, err := util.GetStringResponse(AccountTypeManualSave(), "", reader)
	if name != "" {
		thisActT.Name = name
		if err := util.SaveAccountType(thisActT); err != nil {
			return nil, err
		}
	}
	accountT = append(accountT, thisActT)

	again, err := util.GetBoolResponse(AccountTypeManualAnother(), false, reader)
	if err != nil {
		return nil, err
	}
	if again {
		return addManualAccountType(accountT, iterator)
	}
	return accountT, nil
}

func assembleTypesRaw(accountT []*definitions.AccountType, do *definitions.Do, typ string) error {
	// TODO
	switch typ {
	case "accounttype":
		return assembleTypesFlags(accountT, do)
	case "chaintype":
		return assembleTypesChainsTypesDefs(accountT, do)
	case "csv":
		return assembleTypesCSV(accountT, do)
	}
	return nil
}

func assembleTypesCSV(accountT []*definitions.AccountType, do *definitions.Do) error {
	clearDefaultNumbers(accountT)

	csvfile, err := os.Open(do.CSV)
	if err != nil {
		return err
	}
	defer csvfile.Close()

	reader := csv.NewReader(csvfile)
	reader.TrimLeadingSpace = true

	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return err
	}
	log.WithField("rawCSVdata", rawCSVdata).Debug("Data read.")

	for _, record := range rawCSVdata {
		act, num, tokens, toBond, perms := record[0], record[1], record[2], record[3], record[4:]
		for _, thisActT := range accountT {
			if thisActT.Name == act {
				var err error
				thisActT.Number, err = strconv.Atoi(num)
				if err != nil {
					return err
				}
				thisActT.Tokens, err = strconv.Atoi(tokens)
				if err != nil {
					return err
				}
				thisActT.ToBond, err = strconv.Atoi(toBond)
				if err != nil {
					return err
				}
				permsPrime := make(map[string]int)
				for i := 0; i < len(perms); i++ {
					p, err := strconv.Atoi(perms[i+1])
					if err != nil {
						return err
					}
					permsPrime[perms[i]] = p
					i++
				}
				thisActT.Perms = permsPrime
				log.WithFields(log.Fields{
					"name":   thisActT.Name,
					"number": thisActT.Number,
					"tokens": thisActT.Tokens,
					"toBond": thisActT.ToBond,
					"perms":  thisActT.Perms,
				}).Debug("Setting Account Type Number")
			}
		}
	}
	return nil
}

func assembleTypesFlags(accountT []*definitions.AccountType, do *definitions.Do) error {
	clearDefaultNumbers(accountT)

	for _, acctT := range do.AccountTypes {
		tmp := strings.Split(acctT, ":")
		act := tmp[0]

		var (
			err error

			// If the number of account types is missing,
			// assuming 1.
			num int = 1
		)
		if len(tmp) > 1 {
			num, err = strconv.Atoi(tmp[1])
			if err != nil {
				return err
			}
		}

		for _, thisActT := range accountT {
			if thisActT.Name == act {
				thisActT.Number = num
				log.WithFields(log.Fields{
					"name":   thisActT.Name,
					"number": thisActT.Number,
				}).Debug("Setting Account Type Number")
			}
		}
	}
	return nil
}

func assembleTypesChainsTypesDefs(accountT []*definitions.AccountType, do *definitions.Do) error {
	clearDefaultNumbers(accountT)

	chainTypeAccounts, err := LoadChainTypes(do.ChainType)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"chainType": do.ChainType,
	}).Debug("Chain Type Loaded")

	for act, num := range chainTypeAccounts.AccountTypes {
		for _, thisActT := range accountT {
			// we match against the accountType we get from the chain-type file
			// which will be upper case, however the current yaml unmarshal sequence
			// seems to lower case this for some odd reason.
			// TODO: see if burntsushi's toml renderer will handle this better in the future
			if thisActT.Name == strings.Title(act) {
				thisActT.Number = num
				log.WithFields(log.Fields{
					"name":   thisActT.Name,
					"number": thisActT.Number,
				}).Debug("Setting Account Type Number")
			}
		}
	}
	return nil
}

func clearDefaultNumbers(accountT []*definitions.AccountType) {
	for _, acctT := range accountT {
		acctT.Number = 0
	}
}
