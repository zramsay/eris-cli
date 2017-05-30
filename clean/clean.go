package clean

import (
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/util"
)

func Clean(do *definitions.Do) error {
	// in util so that other pkgs can import it easily
	toClean := map[string]bool{
		"yes":        do.Yes,
		"all":        do.All,
		"containers": do.Containers,
		"chains":     do.ChnDirs,
		"scratch":    do.Scratch,
		"root":       do.RmD,
		"images":     do.Images,
	}
	return util.Clean(toClean)
}
