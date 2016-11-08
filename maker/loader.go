package maker

import (
	"fmt"
	"path/filepath"
	"strings"

	definitions "github.com/eris-ltd/eris-cli/definitions/maker"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
	"github.com/spf13/viper"
)

func LoadAccountTypes() ([]*definitions.AccountType, error) {
	loadedAccounts := []*definitions.AccountType{}
	accounts, err := util.AccountTypes(AccountsTypePath)
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

func LoadAccountType(fileName string) (*definitions.AccountType, error) {
	log.WithField("=>", fileName).Debug("Loading Account Type")
	var accountType = viper.New()
	typ := definitions.BlankAccountType()

	if err := getSetup(fileName, accountType); err != nil {
		return nil, err
	}

	// marshall file
	if err := accountType.Unmarshal(typ); err != nil {
		return nil, fmt.Errorf("\nSorry, the marmots could not figure that account types file out.\nPlease check your account type definition file is properly formatted.\nERROR =>\t\t\t%v", err)
	}

	return typ, nil
}

func LoadChainTypes(fileName string) (*definitions.ChainType, error) {
	fileName = filepath.Join(ChainTypePath, fileName)
	log.WithField("file name", fileName).Info("Loading Chain Definition.")
	var typ = definitions.BlankChainType()
	var chainType = viper.New()

	if err := getSetup(fileName, chainType); err != nil {
		return nil, err
	}

	// marshall file
	if err := chainType.Unmarshal(typ); err != nil {
		return nil, fmt.Errorf("\nSorry, the marmots could not figure that chain types file out.\nPlease check your chain type definition file is properly formatted.\nERROR =>\t\t\t%v", err)
	}

	return typ, nil
}

func getSetup(fileName string, cfg *viper.Viper) error {
	// setup file
	abs, err := filepath.Abs(fileName)
	if err != nil {
		return fmt.Errorf("\nSorry, the marmots were unable to find the absolute path to the account types file.")
	}

	path := filepath.Dir(abs)
	file := filepath.Base(abs)
	extName := filepath.Ext(file)
	bName := file[:len(file)-len(extName)]

	cfg.AddConfigPath(path)
	cfg.SetConfigName(bName)
	cfg.SetConfigType(strings.Replace(extName, ".", "", 1))

	// load file
	if err := cfg.ReadInConfig(); err != nil {
		return fmt.Errorf("\nSorry, the marmots were unable to load the file: (%s). Please check your path.\nERROR =>\t\t\t%v", fileName, err)
	}

	return nil
}
