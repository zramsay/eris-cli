package actions

import (
	"fmt"
	"strings"

	dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func LoadActionDefinition(act []string) (*def.Action, []string, error) {
	var action def.Action

	act, actionVars := fixVars(act)
	actionConf, err := readActionDefinition(strings.Join(act, "_"))
	if err != nil {
		return &action, actionVars, err
	}

	err = marshalActionDefinition(actionConf, &action)
	if err != nil {
		return &action, actionVars, err
	}

	return &action, actionVars, nil
}

func fixVars(act []string) ([]string, []string) {
	var actionVars []string
	var varList bool

	for n, a := range act {
		if strings.Contains(a, ":") {
			actionVars = append(actionVars, strings.Replace(a, ":", "=", 1))
			if varList {
				continue
			}
			act = append(act[:n])
			varList = true
		}
	}
	return act, actionVars
}

func readActionDefinition(actionName string) (*viper.Viper, error) {
	var actionConf = viper.New()

	actionConf.AddConfigPath(dir.ActionsPath)
	actionConf.SetConfigName(actionName)
	actionConf.ReadInConfig()
	// if err != nil {
	//   return nil, err
	// }

	return actionConf, nil
}

// marshal from viper to definitions struct
func marshalActionDefinition(actionConf *viper.Viper, action *def.Action) error {
	err := actionConf.Marshal(action)
	if err != nil {
		return fmt.Errorf("Error marshalling from viper to action def: %v", err)
	}
	return nil
}

// get the config file's path from the chain name
func configFileNameFromActionName(actionName string) (string, error) {
	actionConf, err := readActionDefinition(actionName)
	if err != nil {
		return "", err
	}
	return actionConf.ConfigFileUsed(), nil
}

func checkActionGiven(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("No Service Given. Please rerun command with a known service.")
	}
	return nil
}
