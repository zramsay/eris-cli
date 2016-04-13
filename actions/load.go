package actions

import (
	"fmt"
	"regexp"
	"strings"

	def "github.com/eris-ltd/eris-cli/definitions"

	log "github.com/Sirupsen/logrus"
	dir "github.com/eris-ltd/common/go/common"

	"github.com/spf13/viper"
)

func LoadActionDefinition(actionName string) (*def.Action, []string, error) {
	log.WithField("file", actionName).Info("Reading action definition file")
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
	log.WithField("file", act).Debug("Mocking action")
	return action, []string{}
}

func cullCLIVariables(act []string) ([]string, []string) {
	var actionVars []string
	var action []string

	log.Debug("Pulling out named variables")
	for _, a := range act {
		if strings.Contains(a, ":") {
			r := strings.Replace(a, ":", "=", 1)
			actionVars = append(actionVars, r)
			log.WithField("=>", r).Debug("Culling from variable list")
		} else {
			action = append(action, a)
		}
	}

	log.Info("Successfully parsed named variables")
	log.WithFields(log.Fields{
		"culled":     actionVars,
		"not culled": action,
	}).Debug()
	return action, actionVars
}

func readActionDefinition(actionName []string, dropped map[string]string, varNum int) (*viper.Viper, map[string]string, error) {
	if len(actionName) == 0 {
		log.WithFields(log.Fields{
			"action": actionName,
			"drop":   dropped,
			"var#":   varNum,
		}).Debug("Failed to load action definition file")
		return nil, dropped, fmt.Errorf("The marmots could not find the action definition file.\nPlease check your actions with [eris actions ls]")
	}

	log.WithField("file", strings.Join(actionName, "_")).Debug("Preparing to read action definition file")
	log.WithField("drop", dropped).Debug()

	var actionConf = viper.New()

	actionConf.AddConfigPath(dir.ActionsPath)
	actionConf.SetConfigName(strings.Join(actionName, "_"))
	err := actionConf.ReadInConfig()

	if err != nil {
		log.WithField("action", actionName[len(actionName)-1]).Debug("Dropping and retrying")
		dropped[fmt.Sprintf("$%d", varNum)] = actionName[len(actionName)-1]
		actionName = actionName[:len(actionName)-1]
		varNum++
		return readActionDefinition(actionName, dropped, varNum)
	} else {
		log.Debug("Successfully read action definition file")
	}

	return actionConf, dropped, nil
}

// marshal from viper to definitions struct
func marshalActionDefinition(actionConf *viper.Viper, action *def.Action) error {
	err := actionConf.Unmarshal(action)
	if err != nil {
		return fmt.Errorf("Tragic! The marmots could not read that action definition file:\n%v\n", err)
	}
	return nil
}

func fixSteps(action *def.Action, dropReversed map[string]string) {
	log.Debug("Replacing $1, $2, $3 in steps with args from command line")

	if len(dropReversed) == 0 {
		return
	} else {
		log.WithField("replace", dropReversed).Debug()
	}

	// Because we pop from the end of the args list, the variables
	// in the map $1, $2, etc. are actually the exact opposite of
	// what they should be.
	dropped := make(map[string]string)
	j := 1
	for i := len(dropReversed); i > 0; i-- {
		log.WithField("=>", fmt.Sprintf("$%d:$%d", i, j)).Debug("Reversing")
		dropped[fmt.Sprintf("$%d", j)] = dropReversed[fmt.Sprintf("$%d", i)]
		j++
	}

	reg := regexp.MustCompile(`\$\d`)
	for n, step := range action.Steps {
		if reg.MatchString(step) {
			log.WithField("matched", step).Debug()
			for _, m := range reg.FindAllString(step, -1) {
				action.Steps[n] = strings.Replace(step, m, dropped[m], -1)
			}
			log.WithField("replaced", step).Debug()
		}
	}

	log.Debug("Checking fixed steps")
	for _, step := range action.Steps {
		log.WithField("step", step).Debug()
	}
}

func fixChain(action *def.Action, chainName string) {
	// Steps can include a $chain variable which will be populated *solely*
	// the --chain flag which has been passed via the command line
	if chainName == "" {
		return
	}

	reg := regexp.MustCompile(`\$chain`)
	for n, step := range action.Steps {
		if reg.MatchString(step) {
			log.WithField("matched", step).Debug()
			for _, m := range reg.FindAllString(step, -1) {
				action.Steps[n] = strings.Replace(step, m, chainName, -1)
			}
			log.WithField("replaced", step).Debug()
		}
	}

	log.Debug("Checking fixed chain")
	for _, step := range action.Steps {
		log.WithField("step", step).Debug()
	}
}
