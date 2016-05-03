package pkgs

import (
	"os"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/definitions"
	. "github.com/eris-ltd/eris-cli/errors"
	"github.com/eris-ltd/eris-cli/files"

	log "github.com/eris-ltd/eris-logger"
	"github.com/eris-ltd/common/go/common"
)

func ImportPackage(do *definitions.Do) error {

	doGet := definitions.NowDo()
	doGet.Hash = do.Hash
	doGet.Path = filepath.Join(common.AppsPath, do.Name)
	if err := files.GetFiles(doGet); err != nil {
		return err // returns an ErisError
	}
	log.WithField("path", doGet.Path).Warn("Your package has been succesfully added to")

	return nil
}

func ExportPackage(do *definitions.Do) error {

	// ensure path is dir
	f, err := os.Stat(do.Name)
	if err != nil {
		return err
	}
	if !f.IsDir() {
		return &ErisError{ErrGo, BaseErrorES(ErrPathIsNotDirectory, do.Name), "ensure the path provided is a directory"}
	}

	doPut := definitions.NowDo()
	doPut.Name = do.Name
	if err := files.PutFiles(doPut); err != nil {
		return err // returns an ErisError
	}

	log.Warn("The last entry in the list above is the hash required for [eris pkgs import HASH]. Save it.")

	return nil
}
