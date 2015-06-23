package actions

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Get(cmd *cobra.Command, args []string) {
	checkActionGiven(args)
	if len(args) != 2 {
		fmt.Println("Please give me: eris actions get [name] [location]")
		return
	}
	err := ImportActionRaw(args[0], args[1], cmd.Flags().Lookup("verbose").Changed)
	if err != nil {
		fmt.Println(err)
	}
}

func New(cmd *cobra.Command, args []string) {
	checkActionGiven(args)
	err := EditActionRaw(args)
	if err != nil {
		fmt.Println(err)
	}
}

func ListGlobal() {

}

func ListProject() {

}

func ListKnown() {
	actions := ListKnownRaw()
	for _, s := range actions {
		fmt.Println(strings.Replace(s, "_", " ", -1))
	}
}

func Edit(args []string) {
	checkActionGiven(args)
	err := EditActionRaw(args)
	if err != nil {
		fmt.Println(err)
	}
}

func Rename(cmd *cobra.Command, args []string) {
	checkActionGiven(args)
	if len(args) != 2 {
		fmt.Println("Please give me: eris actions rename \"old action name\" \"new action name\"")
		return
	}
	err := RenameActionRaw(args[0], args[1], cmd.Flags().Lookup("verbose").Changed)
	if err != nil {
		fmt.Println(err)
	}
}

func Rm(cmd *cobra.Command, args []string) {
	checkActionGiven(args)
	err := RmActionRaw(args, cmd.Flags().Lookup("force").Changed, cmd.Flags().Lookup("verbose").Changed)
	if err != nil {
		fmt.Println(err)
	}
}

func ImportActionRaw(actionName, servPath string, verbose bool) error {
	fileName := filepath.Join(ActionsPath, actionName)
	if filepath.Ext(fileName) == "" {
		fileName = fileName + ".toml"
	}

	s := strings.Split(servPath, ":")
	if s[0] == "ipfs" {

		var err error
		if verbose {
			err = util.GetFromIPFS(s[1], fileName, os.Stdout)
		} else {
			err = util.GetFromIPFS(s[1], fileName, bytes.NewBuffer([]byte{}))
		}

		if err != nil {
			return err
		}
		return nil
	}

	if strings.Contains(s[0], "github") {
		fmt.Println("https://twitter.com/ryaneshea/status/595957712040628224")
		return nil
	}

	fmt.Println("I do not know how to get that file. Sorry.")
	return nil
}

func NewActionRaw(actionName []string) error {
	return nil
}

func EditActionRaw(actionName []string) error {
	f := strings.Join(actionName, "_")
	f = filepath.Join(ActionsPath, f) + ".toml"
	Editor(f)
	return nil
}

func RenameActionRaw(oldName, newName string, verbose bool) error {
	oldAction := strings.Split(oldName, " ")
	act, _, err := LoadActionDefinition(oldAction)
	if err != nil {
		return err
	}
	act.Name = newName
	oldName = strings.Replace(oldName, " ", "_", -1)
	oldFile, err := configFileNameFromActionName(oldName)
	if err != nil {
		return err
	}
	newFile := strings.Replace(oldFile, oldName, newName, 1)
	newFile = strings.Replace(newFile, " ", "_", -1)

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

func RmActionRaw(act []string, force, verbose bool) error {
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
