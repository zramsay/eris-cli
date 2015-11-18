package clean

import (
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
)

func Clean(do *definitions.Do) error {
	if err := util.Clean(do.Yes, do.All, do.RmD, do.Images); err != nil {
		return err
	}
	return nil
}
