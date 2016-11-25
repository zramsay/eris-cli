package loaders

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"

	"github.com/spf13/viper"
)

// LoadPackage loads a package definition specified by the directory or
// filename path and chainName and returns a package definition structure.
// LoadPackage can also return missing files or package loading errors.

// TODO [rj] deprecate
func LoadPackageOLD(path, chainName string) (*definitions.Package, error) {
	var name string
	var dir bool

	var err error
	f, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if f.IsDir() {
		name = filepath.Base(path)
		dir = true
	} else {
		name = filepath.Base(filepath.Dir(path))
		dir = false
	}

	var pkgConf *viper.Viper
	var pkg *definitions.Package
	if dir {
		pkgConf, err = loadPackage(path)
	} else {
		pkgConf, err = loadPackage(filepath.Dir(path))
	}

	if err != nil {
		log.Info("The marmots could not read that package.json. Will use defaults.")
		pkg = DefaultPackage(name, chainName)
	} else {
		// marshal chain and always reset the operational requirements
		// this will make sure to sync with docker so that if changes
		// have occurred in the interim they are caught.
		pkg, err = marshalPackage(pkgConf)
		if err != nil {
			return nil, err
		}
	}

	checkName(pkg, chainName)

	return pkg, nil
}

// read the config file into viper
func loadPackage(path string) (*viper.Viper, error) {
	return config.LoadViper(path, "package")
}

// DefaultPackage creates a package definition structure
// with some fields already filled in.
func DefaultPackage(name, chainName string) *definitions.Package {
	pkg := definitions.BlankPackage()
	pkg.Name = name
	pkg.ChainName = chainName
	pkg.PackageID = "" // TODO hash it. [pv]: would util.UniqueName(chainName) do?
	return pkg
}

func marshalPackage(pkgConf *viper.Viper) (*definitions.Package, error) {
	pkgDef := definitions.BlankPackageDefinition()
	err := pkgConf.Unmarshal(pkgDef)
	pkg := pkgDef.Package

	if pkgDef.Name != "" {
		pkg.Name = pkgDef.Name
	}
	if err != nil {
		return nil, fmt.Errorf(`Sorry, the marmots could not figure that package.json out: %v
Please check your package.json file is properly formatted`, err)
	}

	return pkg, nil
}

func checkName(pkg *definitions.Package, name string) {
	if strings.Contains(pkg.Name, " ") {
		newName := strings.Replace(pkg.Name, " ", "_", -1)
		log.WithFields(log.Fields{
			"old": pkg.Name,
			"new": newName,
		}).Debug("Correcting package name")
		pkg.Name = newName
	}
}
