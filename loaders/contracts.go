package loaders

import (
	"fmt"
	"regexp"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func LoadContractPackage(path, chainName, command, typ string) (*definitions.Contracts, error) {
	contConf, err := loadContractPackage(path)
	if err != nil {
		// return a custom error message because util.LoadViperConfig's message
		// will be unhelpful in this context.
		return nil, fmt.Errorf("The marmots could not read that package.json.")
	}

	// marshal chain and always reset the operational requirements
	// this will make sure to sync with docker so that if changes
	// have occured in the interim they are caught.
	dapp, err := marshalContractPackage(contConf)
	if err != nil {
		return nil, err
	}

	if err := setDappType(dapp, chainName, command, typ); err != nil {
		return nil, err
	}

	if err := checkDappAndChain(dapp, chainName); err != nil {
		return nil, err
	}

	return dapp, nil
}

// read the config file into viper
func loadContractPackage(path string) (*viper.Viper, error) {
	return util.LoadViperConfig(path, "package", "contracts")
}

func marshalContractPackage(contConf *viper.Viper) (*definitions.Contracts, error) {
	pkg := definitions.BlankPackage()
	err := contConf.Marshal(pkg)
	dapp := pkg.Contracts
	dapp.Name = pkg.Name

	logger.Debugf("Package marshalled, testType =>\t%s\n", dapp.TestType)
	logger.Debugf("\tdeployType =>\t\t%s\n", dapp.DeployType)
	logger.Debugf("\ttestTask =>\t\t%s\n", dapp.TestTask)
	logger.Debugf("\tdeployTask =>\t\t%s\n", dapp.DeployTask)

	if err != nil {
		// Viper's error messages are atrocious.
		return nil, fmt.Errorf("Sorry, the marmots could not figure that package.json out.\nPlease check your package.json is properly formatted.\n")
	}

	return dapp, nil
}

func setDappType(dapp *definitions.Contracts, name, command, typ string) error {
	var t string

	logger.Debugf("Setting Dapp Type. Task =>\t%s\n", command)
	logger.Debugf("\tType =>\t\t%s\n", typ)
	logger.Debugf("\tChainName =>\t\t%s\n", name)

	if typ != "" {
		t = typ
	} else {
		switch command {
		case "test":
			t = dapp.TestType
		case "deploy":
			t = dapp.DeployType
		}
	}

	switch t {
	case "embark":
		dapp.DappType = definitions.EmbarkDapp()
	case "pyepm":
		dapp.DappType = definitions.PyEpmDapp()
	case "sunit":
		dapp.DappType = definitions.SUnitDapp()
	case "manual":
		dapp.DappType = definitions.GulpDapp()
	default:
		return fmt.Errorf("Unregistered DappType.\nUnfortunately our marmots cannot deal with that.\nPlease ensure that the dapp type is set in the package.json.")
	}

	logger.Debugf("\tDapp Type =>\t\t%s\n", dapp.DappType.Name)

	return nil
}

func checkDappAndChain(dapp *definitions.Contracts, name string) error {
	var chain string

	// name is pulled in from the do struct. need to work with both
	// it (as an override) and the dapp.ChainName
	switch name {
	case "":
		if dapp.ChainName == "" {
			return nil
		} else {
			chain = dapp.ChainName
		}
	case "t", "tmp", "temp":
		return nil
	default:
		chain = name
	}

	// this is hacky.... at best.
	if len(dapp.DappType.ChainTypes) == 1 && dapp.DappType.ChainTypes[0] == "eth" {
		if r := regexp.MustCompile("eth"); r.MatchString(chain) {
			return nil
		} else {
			return fmt.Errorf("The marmots detected a disturbance in the force.\n\nYou asked them to run the Dapp Type: (%s).\nBut the chainName (%s) doesn't contain the name (%s).\nPlease rename the chain or service to contain the name (%s)", dapp.DappType.Name, chain, "eth", "eth")
		}
	}

	return nil
}
