package actions

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	. "github.com/eris-ltd/eris-cli/errors"
	//"github.com/eris-ltd/eris-cli/loaders"
	//"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/files"

	. "github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
	"github.com/eris-ltd/common/go/ipfs"

)

func NewAction(do *definitions.Do) error {
	do.Name = strings.Join(do.Operations.Args, "_")
	path := filepath.Join(ActionsPath, do.Name)
	log.WithFields(log.Fields{
		"action": do.Name,
		"file":   path,
	}).Debug("Creating new action (mocking)")
	act, _ := MockAction(do.Name)
	if err := WriteActionDefinitionFile(act, path); err != nil {
		return &ErisError{404, err, "check filesystem permissions"}
	}
	return nil
}

func ImportAction(do *definitions.Do) error {
	if do.Name == "" {
		do.Name = strings.Join(do.Operations.Args, "_")
	}
	fileName := filepath.Join(ActionsPath, strings.Join(do.Operations.Args, " "))
	if filepath.Ext(fileName) == "" {
		fileName = fileName + ".toml"
	}

	doGet := definitions.NowDo()
	doGet.Hash = do.Hash
	doGet.Path = fileName
	if err := files.GetFiles(doGet); err != nil {
		return err // returns an ErisError
	}
	log.WithField("path", doGet.Path).Warn("Your action has been succesfully added to")

	return nil
}

func ExportAction(do *definitions.Do) error {
	_, _, err := LoadActionDefinition(do.Name)
	if err != nil {
		return &ErisError{404, err, ""}
	}
	// ensure IPFS running?
	doPut := definitions.NowDo()
	doPut.Name = do.Name
	if err := files.PutFiles(doPut); err != nil {
		return err // returns an ErisError
	}
	do.Result = doPut.Result
	log.Warn(do.Result)
	return nil
}

func EditAction(do *definitions.Do) error {
	actDefFile := util.GetFileByNameAndType("actions", do.Name)
	log.WithField("file", actDefFile).Info("Editing action")
	do.Result = "success"
	return Editor(actDefFile)
}

func RenameAction(do *definitions.Do) error {
	if do.Name == do.NewName {
		return &ErisError{404, ErrRenaming, "use a different name"}
	}

	do.Name = strings.Replace(do.Name, " ", "_", -1)
	do.NewName = strings.Replace(do.NewName, " ", "_", -1)
	act, _, err := LoadActionDefinition(do.Name)
	if err != nil {
		log.WithFields(log.Fields{
			"from": do.Name,
			"to":   do.NewName,
		}).Debug("Failed renaming action")
		return &ErisError{404, err, ""}
	}

	do.Name = strings.Replace(do.Name, " ", "_", -1)
	log.WithField("file", do.Name).Debug("Finding action definition file")
	oldFile := util.GetFileByNameAndType("actions", do.Name)
	if oldFile == "" {
		return &ErisError{404, ErrCantFindAction, ""}
	}
	log.WithField("file", oldFile).Debug("Found action definition file")

	var newFile string
	newNameBase := strings.Replace(strings.Replace(do.NewName, " ", "_", -1), filepath.Ext(do.NewName), "", 1)

	if newNameBase == do.Name {
		newFile = strings.Replace(oldFile, filepath.Ext(oldFile), filepath.Ext(do.NewName), 1)
	} else {
		newFile = strings.Replace(oldFile, do.Name, do.NewName, 1)
		newFile = strings.Replace(newFile, " ", "_", -1)
	}

	if newFile == oldFile {
		log.Info("Not renaming the same file")
		return nil
	}

	act.Name = strings.Replace(newNameBase, "_", " ", -1)

	log.WithFields(log.Fields{
		"old": act.Name,
		"new": newFile,
	}).Debug("Writing new action definition file")

	if err = WriteActionDefinitionFile(act, newFile); err != nil {
		return &ErisError{404, err, ""}
	}

	log.WithField("file", oldFile).Debug("Removing old file")

	if err = os.Remove(oldFile); err != nil {
		return &ErisError{404, err, ""}
	}

	return os.Remove(oldFile)
}

func RmAction(do *definitions.Do) error {
	do.Name = strings.Join(do.Operations.Args, "_")
	if do.File {
		oldFile := util.GetFileByNameAndType("actions", do.Name)
		if oldFile == "" {
			return nil
		}
		log.WithField("file", oldFile).Debug("Removing file")
		if err := os.Remove(oldFile); err != nil {
			return &ErisError{404, err, ""}
		}
	}
	return nil
}

// TODO use files put
func exportFile(actionName string) (string, error) {
	var err error
	fileName := util.GetFileByNameAndType("actions", actionName)
	if fileName == "" {
		return "", ErrNoFileToExport
	}

	var hash string
	if log.GetLevel() > 0 {
		hash, err = ipfs.SendToIPFS(fileName, "", os.Stdout)
	} else {
		hash, err = ipfs.SendToIPFS(fileName, "", bytes.NewBuffer([]byte{}))
	}

	if err != nil {
		// TODO
		return "", err
	}

	return hash, nil
}
