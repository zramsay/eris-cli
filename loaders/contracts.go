package loaders

import (
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

func LoadContractPackage(path, chainName, command, typ string) (*definitions.Contracts, error) {
	var app *definitions.Contracts
	contConf, err := loadContractPackage(path)

	if err != nil {
		log.Info("The marmots could not read that package.json. Will use defaults")
		app, _ = DefaultContractPackage()

		_, err := LoadEPMInstructions(path)
		if err != nil {
			// TODO [csk]: rework this logic
		}
	} else {
		// marshal chain and always reset the operational requirements
		// this will make sure to sync with docker so that if changes
		// have occured in the interim they are caught.
		app, err = marshalContractPackage(contConf)
		if err != nil {
			return nil, err
		}
	}

	log.WithFields(log.Fields{
		"test task":   app.TestTask,
		"test type":   app.TestType,
		"deploy type": app.DeployType,
		"deploy task": app.DeployTask,
	}).Debug("Loading package")

	if err := setAppType(app, chainName, command, typ); err != nil {
		return nil, err
	}

	if err := checkAppAndChain(app, chainName); err != nil {
		return nil, err
	}

	return app, nil
}

// read the config file into viper
func loadContractPackage(path string) (*viper.Viper, error) {
	return config.LoadViperConfig(path, "package", "contracts")
}

// read the epm file into viper. probably only need to check the error here.... still WIP
func LoadEPMInstructions(path string) (*viper.Viper, error) {
	// return config.LoadViperConfig(path, "eris", "epm")
	return nil, nil
}

// set's the defaults
func DefaultContractPackage() (*definitions.Contracts, error) {
	pkg := definitions.BlankPackage()
	app := pkg.Contracts
	// we don't catch the directory error here as it should be caught prior to
	//   calling this function.
	pwd, _ := os.Getwd()
	app.Name = path.Base(pwd)
	app.ChainName = ""
	app.TestType = "epm"
	app.DeployType = "epm"
	app.TestTask = "default"
	app.DeployTask = "default"
	return app, nil
}

func marshalContractPackage(contConf *viper.Viper) (*definitions.Contracts, error) {
	pkg := definitions.BlankPackage()
	err := contConf.Marshal(pkg)
	app := pkg.Contracts
	app.Name = pkg.Name

	if err != nil {
		// Viper's error messages are atrocious.
		return nil, fmt.Errorf("Sorry, the marmots could not figure that package.json out.\nPlease check your package.json is properly formatted.\n")
	}

	return app, nil
}

func setAppType(app *definitions.Contracts, name, command, typ string) error {
	var t string

	log.WithFields(log.Fields{
		"task": command,
		"type": typ,
		"app":  name,
	}).Debug("Setting app type")

	if typ != "" {
		t = typ
	} else {
		switch command {
		case "test":
			t = app.TestType
		case "deploy":
			t = app.DeployType
		}
	}

	switch t {
	case "embark":
		app.AppType = definitions.EmbarkApp()
	case "sunit":
		app.AppType = definitions.SUnitApp()
	case "manual":
		app.AppType = definitions.GulpApp()
	default:
		app.AppType = definitions.EPMApp()
	}

	log.WithField("app type", app.AppType.Name).Debug()

	return nil
}

func checkAppAndChain(app *definitions.Contracts, name string) error {
	var chain string

	// name is pulled in from the do struct. need to work with both
	// it (as an override) and the app.ChainName
	switch name {
	case "":
		if app.ChainName == "" {
			return nil
		} else {
			chain = app.ChainName
		}
	case "t", "tmp", "temp":
		return nil
	default:
		chain = name
	}

	// this is hacky.... at best.
	if len(app.AppType.ChainTypes) == 1 && app.AppType.ChainTypes[0] == "eth" {
		if r := regexp.MustCompile("eth"); r.MatchString(chain) {
			return nil
		} else {
			return fmt.Errorf("The marmots detected a disturbance in the force.\n\nYou asked them to run the App Type: (%s).\nBut the chainName (%s) doesn't contain the name (%s).\nPlease rename the chain or service to contain the name (%s)", app.AppType.Name, chain, "eth", "eth")
		}
	}

	return nil
}
