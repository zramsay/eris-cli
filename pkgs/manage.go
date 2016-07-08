package pkgs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/files"

	"github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
)

// Imports a package via ipfs hash.
//
//  do.Hash  - hash of IPFS object to acquire
//  do.Name  - location on host to write. note this is opinionated and uses ~/.eris/apps/NAME
//
func ImportPackage(do *definitions.Do) error {

	doGet := definitions.NowDo()
	doGet.Hash = do.Hash
	doGet.Path = filepath.Join(common.AppsPath, do.Name)
	if err := files.GetFiles(doGet); err != nil {
		return err
	}
	log.WithField("path", doGet.Path).Warn("Your package has been succesfully added to")

	return nil
}

// Exports a package to ipfs.
//
//  do.Name  - location on host to read and add to ipfs should be a relative or absolute path
//
func ExportPackage(do *definitions.Do) error {

	// ensure path is dir
	f, err := os.Stat(do.Name)
	if err != nil {
		return err
	}
	if !f.IsDir() {
		return fmt.Errorf("path (%s) is not a directory; please provide a path to a directory", do.Name)
	}

	doPut := definitions.NowDo()
	doPut.Name = do.Name
	if err := files.PutFiles(doPut); err != nil {
		return err
	}

	log.Warn("The last entry in the list above is the hash required for [eris pkgs import].")

	return nil
}
