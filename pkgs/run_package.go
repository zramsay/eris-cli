package pkgs

import (
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/perform"
)

func RunPackage(do *definitions.Do) error {
	var err error

	// Load the package if it doesn't exist
	if do.Package == nil {
		do.Package, err = LoadPackage(do.YAMLPath)
		if err != nil {
			return err
		}
	}

	return perform.RunJobs(do)
}
