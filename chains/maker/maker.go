package maker

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/monax/monax/definitions"
	"github.com/monax/monax/loaders"
	"github.com/monax/monax/log"
	"github.com/monax/monax/util"

	"github.com/hyperledger/burrow/genesis"
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
		// TODO: [ben] if do.CSV is a dead-code path, then realign it with
		// do.Known
		log.Info("Making chain using csv type paradigm.")
		return makeRaw(do, "csv")
	case do.Wizard:
		log.Info("Making chain using wizard paradigm.")
		return makeWizard(do)
	case do.Known:
		// TODO: [ben] do.Known needs to become the default way to generate the config
		// and genesis files from a set of known public keys; so follow the same logic
		// and go over makeRaw(do, "known"); and construct the account constructors
		// from the known csv files.
		log.Info("Making chain from known accounts and validators")
		log.WithField("=>", do.ChainMakeActs).Info("Accounts path")
		log.WithField("=>", do.ChainMakeVals).Info("Validators path")

		genesisFileString, err := genesis.GenerateKnown(do.Name, do.ChainMakeActs, do.ChainMakeVals)
		if err != nil {
			return err
		}
		fmt.Println(genesisFileString)
		// write to assumed location (maybe check if one is there?)
		// there's nothing else to do, since all the accounts/vals
		// were already generated
		return nil
	default:
		// TODO: [ben] construct the switch statement to be logically complete.
		return fmt.Errorf("Unexpected configuration encountered while attempting to make a chain. Please contact the support@monax.io.")
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

	accountTypes, err := loaders.LoadAccountTypes()
	if err != nil {
		return err
	}

	for _, accountT := range accountTypes {
		if err := assembleTypesWizard(accountT, prelims["tokens"]); err != nil {
			return err
		}
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
	accountTypes, err := loaders.LoadAccountTypes()
	if err != nil {
		return err
	}
	log.Debug("Account types loaded.")

	if err := assembleTypesRaw(accountTypes, do, typ); err != nil {
		return err
	}

	return maker(do, "mint", accountTypes)
}

func maker(do *definitions.Do, consensusType string, accountTypes []*definitions.MonaxDBAccountType) error {
	var err error
	// make the accountConstructor slice bases on the accountTypes
	accounts, err := MakeAccounts(do.Name, consensusType, accountTypes, do.Unsafe)
	if err != nil {
		return err
	}

	// use the accountConstructors to write the necessary files (config, genesis and private validator) per node
	if err = MakeMonaxDBNodes(do.Name, do.SeedsIP, accounts, do.ChainImageName,
		do.UseDataContainer, do.ExportedPorts, do.ContainerEntrypoint); err != nil {
		return err
	}

	// write out the overview files on the host: accounts.csv, validators.csv (and *unsafe* accounts.json)
	if err = SaveAccountResults(do, accounts); err != nil {
		return err
	}

	return nil
}

func assembleTypesWizard(accountT *definitions.MonaxDBAccountType, tokenIze bool) error {
	var err error
	accountT.DefaultNumber, err = util.GetIntResponse(AccountTypeIntro(accountT), accountT.DefaultNumber, reader)
	log.WithField("=>", accountT.DefaultNumber).Debug("What the marmots heard")
	if err != nil {
		return err
	}

	if tokenIze && accountT.DefaultNumber > 0 {
		accountT.DefaultTokens, err = util.GetIntResponse(AccountTypeTokens(accountT), accountT.DefaultTokens, reader)
		log.WithField("=>", accountT.DefaultTokens).Debug("What the marmots heard")
		if err != nil {
			return err
		}
	}

	if accountT.Perms["bond"] && accountT.DefaultNumber > 0 {
		accountT.DefaultBond, err = util.GetIntResponse(AccountTypeToBond(accountT), accountT.DefaultBond, reader)
		log.WithField("=>", accountT.DefaultBond).Debug("What the marmots heard")
		if err != nil {
			return err
		}
	} else {
		log.Info("Setting accountType.DefaultBond to 0")
		log.WithField("=>", accountT.Name).Debug("No bond permissions")
		accountT.DefaultBond = 0
	}

	return nil
}

func addManualAccountType(accountT []*definitions.MonaxDBAccountType, iterator int) ([]*definitions.MonaxDBAccountType, error) {
	var err error
	thisActT := &definitions.MonaxDBAccountType{}
	thisActT.Name = fmt.Sprintf("%s_%02d", "manual", iterator)
	iterator++

	thisActT.DefaultNumber, err = util.GetIntResponse(AccountTypeManualIntro(), 1, reader)
	if err != nil {
		return nil, err
	}

	thisActT.DefaultTokens, err = util.GetIntResponse(AccountTypeManualTokens(), 0, reader)
	if err != nil {
		return nil, err
	}

	thisActT.DefaultBond, err = util.GetIntResponse(AccountTypeManualToBond(), 0, reader)
	if err != nil {
		return nil, err
	}

	thisActT.Perms = make(map[string]bool)
	for _, perm := range AccountTypeManualPerms() {
		thisActT.Perms[perm], err = util.GetBoolResponse(AccountTypeManualPermsQuestion(perm), false, reader)
		if err != nil {
			return nil, err
		}
	}

	name, err := util.GetStringResponse(AccountTypeManualSave(), "", reader)
	if err != nil {
		return nil, err
	}
	if name != "" {
		thisActT.Name = name
		if err := SaveAccountType(thisActT); err != nil {
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

func assembleTypesRaw(accountT []*definitions.MonaxDBAccountType, do *definitions.Do, typ string) error {
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

func assembleTypesCSV(accountT []*definitions.MonaxDBAccountType, do *definitions.Do) error {
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
				thisActT.DefaultNumber, err = strconv.ParseInt(num, 10, 64)
				if err != nil {
					return err
				}
				thisActT.DefaultTokens, err = strconv.ParseInt(tokens, 10, 64)
				if err != nil {
					return err
				}
				thisActT.DefaultBond, err = strconv.ParseInt(toBond, 10, 64)
				if err != nil {
					return err
				}
				permsPrime := make(map[string]bool)
				for i := 0; i < len(perms); i++ {
					p, err := strconv.ParseBool(perms[i+1])
					if err != nil {
						return err
					}
					permsPrime[perms[i]] = p
					i++
				}
				thisActT.Perms = permsPrime
				log.WithFields(log.Fields{
					"name":   thisActT.Name,
					"number": thisActT.DefaultNumber,
					"tokens": thisActT.DefaultTokens,
					"toBond": thisActT.DefaultBond,
					"perms":  thisActT.Perms,
				}).Debug("Setting Account Type Number")
			}
		}
	}
	return nil
}

func assembleTypesFlags(accountT []*definitions.MonaxDBAccountType, do *definitions.Do) error {
	clearDefaultNumbers(accountT)

	for _, acctT := range do.AccountTypes {
		tmp := strings.Split(acctT, ":")
		act := tmp[0]

		var (
			err error

			// If the number of account types is missing,
			// assuming 1.
			num int64 = 1
		)
		if len(tmp) > 1 {
			num, err = strconv.ParseInt(tmp[1], 10, 64)
			if err != nil {
				return err
			}
		}

		for _, thisActT := range accountT {
			if thisActT.Name == act {
				thisActT.DefaultNumber = num
				log.WithFields(log.Fields{
					"name":   thisActT.Name,
					"number": thisActT.DefaultNumber,
				}).Debug("Setting Account Type Number")
			}
		}
	}
	return nil
}

func assembleTypesChainsTypesDefs(accountT []*definitions.MonaxDBAccountType, do *definitions.Do) error {
	clearDefaultNumbers(accountT)

	chainTypeAccounts, err := loaders.LoadChainTypes(do.ChainType)
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
				thisActT.DefaultNumber = int64(num)
				log.WithFields(log.Fields{
					"name":   thisActT.Name,
					"number": thisActT.DefaultNumber,
				}).Debug("Setting Account Type Number")
			}
		}
	}
	return nil
}

func clearDefaultNumbers(accountT []*definitions.MonaxDBAccountType) {
	for _, acctT := range accountT {
		acctT.DefaultNumber = 0
	}
}
