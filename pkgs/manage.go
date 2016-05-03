package pkgs

import (
	"fmt"
	"os"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/files"

	log "github.com/Sirupsen/logrus"
	//. "github.com/eris-ltd/common/go/common"
)

func ImportPackage(do *definitions.Do) error {
	// do.Path
	// do.Name -> ~/.eris/apps/do.Name

	return nil
}

func ExportPackage(do *definitions.Do) error {

	//ensure path is dir
	f, err := os.Stat(do.Name)
	if err != nil {
		return err
	}

	if !f.IsDir() {
		return fmt.Errorf("path (%s) is not a directory; please provide a path to a directory")
	}

	doPut := definitions.NowDo()
	doPut.Name = do.Name
	if err := files.PutFiles(doPut); err != nil {
		return err
	}

	log.Warn("output from PutFiles & instructions")
	return nil
}
