package clean

import (
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
)

func Clean(do *definitions.Do) error {
	// in util so that other pkgs can import it easily
	toClean := map[string]bool{
		"yes":        do.Yes,
		"all":        do.All,
		"containers": do.Containers,
		"scratch":    do.Scratch,
		"root":       do.RmD,
		"images":     do.Images,
	}
	if err := util.Clean(toClean); err != nil {
		// returns an ErisError
		return err
	}
	return nil
}
