package loaders

import (
	"fmt"
	"path/filepath"

	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"

	"github.com/spf13/viper"
)

func LoadPackage(fileName string) (*definitions.Package, error) {
	log.Info("Loading monax Jobs Definition File.")
	var pkg = definitions.BlankPackage()
	var epmJobs = viper.New()

	// setup file
	abs, err := filepath.Abs(fileName)
	if err != nil {
		return nil, fmt.Errorf("Sorry, the marmots were unable to find the absolute path to the monax jobs file.")
	}

	path := filepath.Dir(abs)
	file := filepath.Base(abs)
	extName := filepath.Ext(file)
	bName := file[:len(file)-len(extName)]
	log.WithFields(log.Fields{
		"path": path,
		"name": bName,
	}).Debug("Loading monax jobs file")

	epmJobs.SetConfigType("yaml")
	epmJobs.AddConfigPath(path)
	epmJobs.SetConfigName(bName)

	// load file
	if err := epmJobs.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Sorry, the marmots were unable to load the monax jobs file. Please check your path: %v", err)
	}

	// marshall file
	if err := epmJobs.Unmarshal(pkg); err != nil {
		return nil, fmt.Errorf(`Sorry, the marmots could not figure that monax jobs file out. 
			Please check that your epm.yaml is properly formatted: %v`, err)
	}

	// TODO more file sanity check (fail before running)

	return pkg, nil
}
