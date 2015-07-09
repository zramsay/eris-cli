package actions

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
)

func NewActionRaw(do *definitions.Do) error {
	do.Name = strings.Join(do.Args, "_")
	path := filepath.Join(ActionsPath, do.Name)
	logger.Debugf("NewActionRaw to MockAction =>\t%v:%s\n", do.Name, path)
	act, _ := MockAction(do.Name)
	if err := WriteActionDefinitionFile(act, path); err != nil {
		return err
	}
	return nil
}

func ImportActionRaw(do *definitions.Do) error {
	if do.Name == "" {
		do.Name = strings.Join(do.Args, "_")
	}
	fileName := filepath.Join(ActionsPath, strings.Join(do.Args, " "))
	if filepath.Ext(fileName) == "" {
		fileName = fileName + ".toml"
	}

	s := strings.Split(do.Path, ":")
	if s[0] == "ipfs" {

		var err error

		ipfsService, err := loaders.LoadServiceDefinition("ipfs", 1)
		if err != nil {
			return err
		}

		err = perform.DockerRun(ipfsService.Service, ipfsService.Operations)
		if err != nil {
			return err
		}

		if logger.Level > 0 {
			err = util.GetFromIPFS(s[1], fileName, logger.Writer)
		} else {
			err = util.GetFromIPFS(s[1], fileName, bytes.NewBuffer([]byte{}))
		}

		if err != nil {
			return err
		}
		return nil
	}

	if strings.Contains(s[0], "github") {
		logger.Println("https://twitter.com/ryaneshea/status/595957712040628224")
		return nil
	}

	fmt.Println("I do not know how to get that file. Sorry.")
	return nil
}

func ExportActionRaw(do *definitions.Do) error {
	_, _, err := LoadActionDefinition(do.Name)
	if err != nil {
		return err
	}

	ipfsService, err := loaders.LoadServiceDefinition("ipfs", 1)
	if err != nil {
		return err
	}

	err = perform.DockerRun(ipfsService.Service, ipfsService.Operations)
	if err != nil {
		return err
	}

	hash, err := exportFile(do.Name)
	if err != nil {
		return err
	}
	do.Result = hash
	logger.Println(hash)
	return nil
}

func EditActionRaw(do *definitions.Do) error {
	f := filepath.Join(ActionsPath, do.Name) + ".toml"
	Editor(f)
	return nil
}

func RenameActionRaw(do *definitions.Do) error {
	if do.Name == do.NewName {
		return fmt.Errorf("Cannot rename to same name")
	}

	do.Name = strings.Replace(do.Name, " ", "_", -1)
	do.NewName = strings.Replace(do.NewName, " ", "_", -1)
	act, _, err := LoadActionDefinition(do.Name)
	if err != nil {
		logger.Debugf("About to fail. Name:NewName =>\t%s:%s", do.Name, do.NewName)
		return err
	}

	do.Name = strings.Replace(do.Name, " ", "_", -1)
	logger.Debugf("About to find defFile =>\t%s\n", do.Name)
	oldFile := util.GetFileByNameAndType("actions", do.Name)
	if oldFile == "" {
		return fmt.Errorf("Could not find that action definition file.")
	}
	logger.Debugf("Found defFile at =>\t\t%s\n", oldFile)

	if !strings.Contains(oldFile, ActionsPath) {
		oldFile = filepath.Join(ActionsPath, oldFile)+".toml"
	}

	var newFile string
	newNameBase := strings.Replace(strings.Replace(do.NewName, " ", "_", -1), filepath.Ext(do.NewName), "", 1)

	if newNameBase == do.Name {
		newFile = strings.Replace(oldFile, filepath.Ext(oldFile), filepath.Ext(do.NewName), 1)
	} else {
		newFile = strings.Replace(oldFile, do.Name, do.NewName, 1)
		newFile = strings.Replace(newFile, " ", "_", -1)
	}

	if newFile == oldFile {
		logger.Infoln("Those are the same file. Not renaming")
		return nil
	}

	act.Name = strings.Replace(newNameBase, "_", " ", -1)

	logger.Debugf("About to write new def file =>\t%s:%s\n", act.Name, newFile)
	err = WriteActionDefinitionFile(act, newFile)
	if err != nil {
		return err
	}

	logger.Debugf("Removing old file =>\t\t%s\n", oldFile)
	os.Remove(oldFile)

	return nil
}

func ListKnownRaw(do *definitions.Do) error {
	chns := util.GetGlobalLevelConfigFilesByType("actions", false)
	do.Result = strings.Join(chns, "\n")
	return nil
}

func RmActionRaw(do *definitions.Do) error {
	do.Name = strings.Join(do.Args, "_")
	if do.File {
		oldFile := util.GetFileByNameAndType("actions", do.Name)
		if oldFile == "" {
			return nil
		}

		if !strings.Contains(oldFile, ActionsPath) {
			oldFile = filepath.Join(ActionsPath, oldFile)+".toml"
		}

		logger.Debugf("Removing file =>\t\t%s\n", oldFile)
		os.Remove(oldFile)
	}
	return nil
}

func exportFile(actionName string) (string, error) {
	var err error
	fileName := util.GetFileByNameAndType("actions", actionName)
	if fileName == "" {
		return "", fmt.Errorf("no file to export")
	}

	var hash string
	if logger.Level > 0 {
		hash, err = util.SendToIPFS(fileName, logger.Writer)
	} else {
		hash, err = util.SendToIPFS(fileName, bytes.NewBuffer([]byte{}))
	}

	if err != nil {
		return "", err
	}

	return hash, nil
}
