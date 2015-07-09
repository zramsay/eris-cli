package actions

import (
	"fmt"
	"regexp"
	"strings"

	def "github.com/eris-ltd/eris-cli/definitions"

	dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func LoadActionDefinition(actionName string) (*def.Action, []string, error) {
	logger.Infof("Reading action def file =>\t%v\n", actionName)
	act := strings.Split(actionName, "_")
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

func MockAction(act string) (*def.Action, []string) {
	action := def.BlankAction()
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
			logger.Debugln("Culling from variable list =>\t%s\n", r)
		} else {
			action = append(action, a)
		}
	}

	logger.Debugf("Args culled =>\t\t\t%s\n", actionVars)
	logger.Debugf("Args not culled =>\t\t%s\n", action)
	logger.Infof("Successfully parsed the named variables passed to the command line.\n")
	return action, actionVars
}

func readActionDefinition(actionName []string, dropped map[string]string, varNum int) (*viper.Viper, map[string]string, error) {
	if len(actionName) == 0 {
		logger.Debugf("Fail. actionName, drop, varN =>\t%v:%v:%v\n", actionName, dropped, varNum)
		return nil, dropped, fmt.Errorf("The marmots could not find the action definition file.\nPlease check your actions with [eris actions ls]")
	}

	logger.Debugf("Read action definition file =>\t%s\n", strings.Join(actionName, "_"))
	logger.Debugf("Args to add to the steps =>\t%s\n", dropped)

	var actionConf = viper.New()

	actionConf.AddConfigPath(dir.ActionsPath)
	actionConf.SetConfigName(strings.Join(actionName, "_"))
	err := actionConf.ReadInConfig()

	if err != nil {
		logger.Debugf("Dropping and retrying =>\t%s\n", actionName[len(actionName)-1])
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
