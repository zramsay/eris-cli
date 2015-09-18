package clean

import (
	"fmt"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
)

func Clean(do *definitions.Do) error {
	fmt.Printf("yes %v\n", do.Yes)
	fmt.Printf("all %v\n", do.All)
	fmt.Printf("rmd %v\n", do.RmD)
	fmt.Printf("img %v\n", do.Images)
	if err := util.Clean(do.Yes, do.All, do.RmD, do.Images); err != nil {
		return err
	}
	return nil
}
