package actions

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/eris-ltd/eris-cli/chains"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/services"
)

func DoRaw(action *def.Action, actionVars []string, quiet bool) error {
	if err := StartServicesAndChains(action); err != nil {
		return err
	}

	if err := PerformCommand(action, actionVars, quiet); err != nil {
		return err
	}
	return nil
}

func StartServicesAndChains(action *def.Action) error {
	// start the services and chains
	wg, ch := new(sync.WaitGroup), make(chan error, 1)

	runningServices := services.ListRunningRaw(true)
	services.StartGroup(ch, wg, action.ServiceDeps, runningServices, "service", &def.Operation{ContainerNumber: 1}, services.StartServiceRaw) // TODO:CNUM

	runningChains := chains.ListRunningRaw(true)
	services.StartGroup(ch, wg, []string{action.Chain}, runningChains, "chain", &def.Operation{ContainerNumber: 1}, chains.StartChainRaw) // TODO:CNUM

	go func() {
		wg.Wait()
		ch <- nil
	}()

	return <-ch
}

func PerformCommand(action *def.Action, actionVars []string, quiet bool) error {
	logger.Infof("Performing Action:\t%s.\n", action.Name)

	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	logger.Debugf("Directory for action:\t%s\n", dir)

	// pull actionVars (first given from command line) and
	// combine with the environment variables (given in the
	// action definition files) and finally combine with
	// the hosts os.Environ() to provide the full set of
	// variables to be consumed during the steps phase.
	for k, v := range action.Environment {
		actionVars = append(actionVars, fmt.Sprintf("%s=%s", k, v))
	}

	for _, v := range actionVars {
		logger.Debugf("Variable for action:\t%s\n", v)
	}

	actionVars = append(os.Environ(), actionVars...)

	for n, step := range action.Steps {
		cmd := exec.Command("sh", "-c", step)
		cmd.Env = actionVars
		cmd.Dir = dir

		logger.Debugf("Performing Step %d:\t%s\n", n+1, strings.Join(cmd.Args, " "))

		prev, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("error running command (%v): %s", err, prev)
		}

		if !quiet {
			logger.Println(strings.TrimSpace(string(prev)))
		}

		if n != 0 {
			actionVars = actionVars[:len(actionVars)-1]
		}
		actionVars = append(actionVars, ("prev=" + strings.TrimSpace(string(prev))))
	}

	logger.Infoln("Action performed")
	return nil
}
