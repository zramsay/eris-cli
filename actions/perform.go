package actions

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/services"

	log "github.com/eris-ltd/eris-logger"
)

func Do(do *definitions.Do) error {
	log.WithFields(log.Fields{
		"chain":    do.ChainName,
		"services": do.ServicesSlice,
	}).Info("Performing action")

	var err error
	var actionVars []string
	do.Action, actionVars, err = LoadActionDefinition(strings.Join(do.Operations.Args, "_"))
	if err != nil {
		return err
	}

	resolveServices(do)
	resolveChain(do)
	fixChain(do.Action, do.ChainName)

	if err := StartServicesAndChains(do); err != nil {
		return err
	}

	if err := PerformCommand(do.Action, actionVars, do.Quiet); err != nil {
		return err
	}

	return nil
}

// [zr] actions probably shouldn't _start_ chains, only check that they are running
func StartServicesAndChains(do *definitions.Do) error {
	// start the services and chains
	doSrvs := definitions.NowDo()
	if do.Action.Dependencies == nil || len(do.Action.Dependencies.Services) == 0 {
		log.Debug("No services to start")
	} else {
		doSrvs.Operations.Args = do.Action.Dependencies.Services
		log.WithField("args", doSrvs.Operations.Args).Debug("Starting services")
		if err := services.StartService(doSrvs); err != nil {
			return err
		}
	}

	doChns := definitions.NowDo()
	doChns.Name = do.Action.Chain
	if doChns.Name == "" {
		log.Debug("No chain to start")
	} else {
		log.WithField("=>", doChns.Name).Debug("Starting chain")
		if err := chains.StartChain(do); err != nil {
			return err
		}
	}

	return nil
}

func PerformCommand(action *definitions.Action, actionVars []string, quiet bool) error {
	log.WithField("action", action.Name).Info("Performing action")

	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	log.WithField("directory", dir).Debug()

	// pull actionVars (first given from command line) and
	// combine with the environment variables (given in the
	// action definition files) and finally combine with
	// the hosts os.Environ() to provide the full set of
	// variables to be consumed during the steps phase.
	for k, v := range action.Environment {
		actionVars = append(actionVars, fmt.Sprintf("%s=%s", k, v))
	}

	for _, v := range actionVars {
		log.WithField("variable", v).Debug()
	}

	actionVars = append(os.Environ(), actionVars...)

	for n, step := range action.Steps {
		cmd := exec.Command("sh", "-c", step)
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/c", step)
		}
		cmd.Env = actionVars
		cmd.Dir = dir

		log.WithField("=>", strings.Join(cmd.Args, " ")).Debugf("Performing step %d", n+1)

		prev, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("error running command (%v): %s", err, prev)
		}

		if !quiet {
			log.Warn(strings.TrimSpace(string(prev)))
		}

		if n != 0 {
			actionVars = actionVars[:len(actionVars)-1]
		}
		actionVars = append(actionVars, ("prev=" + strings.TrimSpace(string(prev))))
	}

	log.Info("Action performed")
	return nil
}

func resolveChain(do *definitions.Do) {
	if do.ChainName == "" { // do.ChainName populated via CLI flag
		do.Action.Chain = do.ChainName
	}

	if do.Action.Chain == "$chain" { // requires chains via the CLI
		do.Action.Chain = do.ChainName
	}
}

func resolveServices(do *definitions.Do) {
	if do.Action.Dependencies != nil {
		do.Action.Dependencies.Services = append(do.Action.Dependencies.Services, do.ServicesSlice...)
	}
	log.WithField("args", do.Operations.Args).Debug("Services to start")
}
