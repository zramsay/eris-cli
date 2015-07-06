package actions

import (
	"fmt"
	"regexp"
	"strings"

	def "github.com/eris-ltd/eris-cli/definitions"

	dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func LoadActionDefinition(act []string) (*def.Action, []string, error) {
	logger.Infof("Reading action def file =>\t%v\n", act)
	action := def.BlankAction()

	act, actionVars := cullCLIVariables(act)
	actionConf, dropped, err := readActionDefinition(act, make(map[string]string), 1)
	if err != nil {
		return action, actionVars, err
	}

	err = marshalActionDefinition(actionConf, action)
	if err != nil {
		return action, actionVars, err
	}

	if len(dropped) != 0 {
		fixSteps(action, dropped)
	}

	return action, actionVars, nil
}

func MockAction(act []string) (*def.Action, []string) {
	action := def.BlankAction()
	action.Name = strings.Join(act, " ")
	logger.Debugf("Mocking action =>\t\t%v\n", act)
	return action, []string{}
}

func MergeStepsAndCLIArgs(action *def.Action, actionVars *[]string, args []string) error {

	return nil
}

func cullCLIVariables(act []string) ([]string, []string) {
	var actionVars []string
	var action []string

	logger.Debugln("Pulling out the named variables passed to the command line.")
	for _, a := range act {
		if strings.Contains(a, ":") {
			r := strings.Replace(a, ":", "=", 1)
			actionVars = append(actionVars, r)
			logger.Debugln("Culling from variable list:\t", r)
		} else {
			action = append(action, a)
		}
	}

	logger.Debugln("Args culled:\t\t\t", actionVars)
	logger.Debugln("Args not culled:\t\t", action)
	logger.Debugln("Success fully parsed the named variables passed to the command line.")
	return action, actionVars
}

func readActionDefinition(actionName []string, dropped map[string]string, varNum int) (*viper.Viper, map[string]string, error) {
	if len(actionName) == 0 {
		return nil, dropped, fmt.Errorf("The marmots could not find the action definition file.\nPlease check your actions with [eris actions ls]")
	}

	logger.Debugln("Reading action definition file:\t", strings.Join(actionName, "_"))
	logger.Debugln("Args to add to the steps:\t", dropped)

	var actionConf = viper.New()

	actionConf.AddConfigPath(dir.ActionsPath)
	actionConf.SetConfigName(strings.Join(actionName, "_"))
	err := actionConf.ReadInConfig()

	// we allow a maximum of 3 additional variables that can be passed as
	//   args to the command. but the read parser needs to be able to find
	//   the file first. so this portion of the function adds a maximum
	//   of three additional variables to be populated before it will
	//   fail.
	if err != nil {
		// we are under the recursion limit, pop off actionName slice and into dropped slice, recurse
		logger.Debugln("Dropping and retrying:\t\t", actionName[len(actionName)-1])
		dropped[fmt.Sprintf("$%d", varNum)] = actionName[len(actionName)-1]
		actionName = actionName[:len(actionName)-1]
		varNum++
		return readActionDefinition(actionName, dropped, varNum)
	} else {
		logger.Debugln("Action definition file successfully read.")
	}

	return actionConf, dropped, nil
}

// marshal from viper to definitions struct
func marshalActionDefinition(actionConf *viper.Viper, action *def.Action) error {
	err := actionConf.Marshal(action)
	if err != nil {
		return fmt.Errorf("Tragic! The marmots could not read that action definition file:\n%v\n", err)
	}
	return nil
}

func fixSteps(action *def.Action, dropReversed map[string]string) {
	logger.Debugln("Replacing $1, $2, $3 in steps with args from command line.")

	if len(dropReversed) == 0 {
		return
	} else {
		logger.Debugln("Variables to replace:\t\t", dropReversed)
	}

	// Because we pop from the end of the args list, the variables
	// in the map $1, $2, etc. are actually the exact opposite of
	// what they should be.
	dropped := make(map[string]string)
	j := 1
	for i := len(dropReversed); i > 0; i-- {
		logger.Debugln("Reversing:\t\t\t", fmt.Sprintf("$%d -> $%d", i, j))
		dropped[fmt.Sprintf("$%d", j)] = dropReversed[fmt.Sprintf("$%d", i)]
		j++
	}

	reg := regexp.MustCompile(`\$\d`)
	for n, step := range action.Steps {
		if reg.MatchString(step) {
			logger.Debugln("Match(es) Found In Step:\t", step)
			for _, m := range reg.FindAllString(step, -1) {
				action.Steps[n] = strings.Replace(step, m, dropped[m], -1)
			}
			logger.Debugln("After replacing the step is:\t", step)
		}
	}

	logger.Debugln("After Fixing Steps, we have:")
	for _, step := range action.Steps {
		logger.Debugln("\t", step)
	}
}

// get the config file's path from the chain name
func configFileNameFromActionName(actionName string) (string, error) {
	actionConf, _, err := readActionDefinition(strings.Split(actionName, "_"), make(map[string]string), 1)
	if err != nil {
		return "", err
	}
	return actionConf.ConfigFileUsed(), nil
}
