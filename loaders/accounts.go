package loaders

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"

	"github.com/spf13/viper"
)

func LoadAccountTypes() ([]*definitions.ErisDBAccountType, error) {
	loadedAccounts := []*definitions.ErisDBAccountType{}
	accounts, err := AccountTypes(config.AccountsTypePath)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		thisAct, err := LoadAccountType(account)
		if err != nil {
			return nil, err
		}
		log.WithField("=>", thisAct.Name).Debug("Loaded Account Named")
		loadedAccounts = append(loadedAccounts, thisAct)
	}
	return loadedAccounts, nil
}

func LoadAccountType(fileName string) (*definitions.ErisDBAccountType, error) {
	log.WithField("=>", fileName).Debug("Loading Account Type")
	var accountType = viper.New()
	typ := definitions.BlankAccountType()

	if err := getSetup(fileName, accountType); err != nil {
		return nil, err
	}

	// marshall file
	if err := accountType.Unmarshal(typ); err != nil {
		return nil, fmt.Errorf(`Sorry, the account type file %v confused the marmots.
			Please check that your account type definition file is properly formatted: %v`, fileName, err)
	}

	return typ, nil
}

// returns a list of filenames which are the account_types files
// these *should be* absolute paths, but this is not a contract
// with calling functions.
func AccountTypes(erisPath string) ([]string, error) {
	haveTyps, err := filepath.Glob(filepath.Join(erisPath, "*.toml"))
	if err != nil {
		return []string{}, err
	}
	return haveTyps, nil
}

func AccountTypesNames(erisPath string, withExt bool) ([]string, error) {
	files, err := AccountTypes(erisPath)
	if err != nil {
		return []string{}, err
	}
	names := []string{}
	for _, file := range files {
		names = append(names, filepath.Base(file))
	}
	if !withExt {
		for e, name := range names {
			names[e] = strings.Replace(name, ".toml", "", 1)
		}
	}
	return names, nil
}
