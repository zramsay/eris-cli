package actions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/common/go/ipfs"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
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
		return err
	}
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
		return fmt.Errorf("Cannot rename to same name")
	}

	do.Name = strings.Replace(do.Name, " ", "_", -1)
	do.NewName = strings.Replace(do.NewName, " ", "_", -1)
	act, _, err := LoadActionDefinition(do.Name)
	if err != nil {
		log.WithFields(log.Fields{
			"from": do.Name,
			"to":   do.NewName,
		}).Debug("Failed renaming action")
		return err
	}

	do.Name = strings.Replace(do.Name, " ", "_", -1)
	log.WithField("file", do.Name).Debug("Finding action definition file")
	oldFile := util.GetFileByNameAndType("actions", do.Name)
	if oldFile == "" {
		return fmt.Errorf("Could not find that action definition file.")
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
	err = WriteActionDefinitionFile(act, newFile)
	if err != nil {
		return err
	}

	log.WithField("file", oldFile).Debug("Removing old file")

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
			return err
		}
	}
	return nil
}

func exportFile(actionName string) (string, error) {
	fileName := util.GetFileByNameAndType("actions", actionName)
	if fileName == "" {
		return "", fmt.Errorf("no file to export")
	}

	return ipfs.SendToIPFS(fileName, "")
}
