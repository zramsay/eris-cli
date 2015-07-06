package actions

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
)

func NewActionRaw(actionName []string) error {
	name := strings.Join(actionName, "_")
	path := filepath.Join(ActionsPath, name)
	act, _ := MockAction(actionName)
	if err := WriteActionDefinitionFile(act, path); err != nil {
		return err
	}
	return nil
}

func ImportActionRaw(actionName string, servPath string) error {
	fileName := filepath.Join(ActionsPath, strings.Replace(actionName, " ", "_", -1))
	if filepath.Ext(fileName) == "" {
		fileName = fileName + ".toml"
	}

	s := strings.Split(servPath, ":")
	if s[0] == "ipfs" {

		var err error
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

func ExportActionRaw(actionName []string) error {
	action := strings.Join(actionName, "_")
	_, _, err := LoadActionDefinition(actionName)
	if err != nil {
		return err
	}

	ipfsService, err := services.LoadServiceDefinition("ipfs", 1)
	if err != nil {
		return err
	}

	if services.IsServiceRunning(ipfsService.Service, ipfsService.Operations) {
		logger.Infoln("IPFS is running. Adding now.")

		hash, err := exportFile(action)
		if err != nil {
			return err
		}
		logger.Println(hash)
	} else {
		logger.Infoln("IPFS is not running. Starting now.")
		err := services.StartServiceByService(ipfsService.Service, ipfsService.Operations, []string{})
		if err != nil {
			return err
		}

		hash, err := exportFile(action)
		if err != nil {
			return err
		}
		logger.Println(hash)
	}
	return nil
}

func EditActionRaw(actionName []string) error {
	f := strings.Join(actionName, "_")
	f = filepath.Join(ActionsPath, f) + ".toml"
	Editor(f)
	return nil
}

func RenameActionRaw(oldName, newName string) error {
	if oldName == newName {
		return fmt.Errorf("Cannot rename to same name")
	}

	oldAction := strings.Split(oldName, " ")
	act, _, err := LoadActionDefinition(oldAction)
	if err != nil {
		return err
	}

	oldName = strings.Replace(oldName, " ", "_", -1)
	oldFile, err := configFileNameFromActionName(oldName)
	if err != nil {
		return err
	}

	var newFile string
	newNameBase := strings.Replace(strings.Replace(newName, " ", "_", -1), filepath.Ext(newName), "", 1)

	if newNameBase == oldName {
		newFile = strings.Replace(oldFile, filepath.Ext(oldFile), filepath.Ext(newName), 1)
	} else {
		newFile = strings.Replace(oldFile, oldName, newName, 1)
		newFile = strings.Replace(newFile, " ", "_", -1)
	}

	if newFile == oldFile {
		logger.Infoln("Those are the same file. Not renaming")
		return nil
	}

	act.Name = strings.Replace(newNameBase, "_", " ", -1)

	err = WriteActionDefinitionFile(act, newFile)
	if err != nil {
		return err
	}

	os.Remove(oldFile)

	return nil
}

func ListKnownRaw() []string {
	acts := []string{}
	fileTypes := []string{}
	for _, t := range []string{"*.json", "*.yaml", "*.toml"} {
		fileTypes = append(fileTypes, filepath.Join(ActionsPath, t))
	}
	for _, t := range fileTypes {
		s, _ := filepath.Glob(t)
		for _, s1 := range s {
			s1 = strings.Split(filepath.Base(s1), ".")[0]
			acts = append(acts, s1)
		}
	}
	return acts
}

func RmActionRaw(act []string, force bool) error {
	// TODO: add interactive check if no force flag
	if force {
		actName := strings.Join(act, "_")
		oldFile, err := configFileNameFromActionName(actName)
		if err != nil {
			return err
		}

		os.Remove(oldFile)
	}
	return nil
}

func exportFile(actionName string) (string, error) {
	fileName, err := configFileNameFromActionName(actionName)
	if err != nil {
		return "", err
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
