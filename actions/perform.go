package actions

import (
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/services"

	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Do(cmd *cobra.Command, args []string) {
	action, actionVars, err := LoadActionDefinition(args)
	if err != nil {
		logger.Errorln(err)
		return
	}
	err = DoRaw(action, actionVars)
	if err != nil {
		logger.Errorln(err)
		return
	}
}

func DoRaw(action *def.Action, actionVars []string) error {
	err := StartServicesAndChains(action)
	if err != nil {
		return err
	}

	err = PerformCommand(action, actionVars)
	if err != nil {
		return err
	}
	return nil
}

func StartServicesAndChains(action *def.Action) error {
	// start the services and chains
	var wg sync.WaitGroup
	var skip bool

	running := services.ListRunningRaw()
	for _, srv := range action.Services {
		skip = false

		for _, run := range running {
			if srv == run {
				logger.Infoln("Service already started, Skipping: ", srv)
				skip = true
			}
		}
		if skip {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := services.StartServiceRaw(srv)
			if err != nil {
				logger.Println("Service already started, Skipping: ", srv)
			}
		}()
	}

	running = chains.ListRunningRaw()
	for _, chn := range action.Chains {
		skip = false

		for _, run := range running {
			if chn == run {
				logger.Infoln("Chain already started, Skipping: ", chn)
				skip = true
			}
		}
		if skip {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := chains.StartChainRaw(chn)
			if err != nil {
				logger.Println("Chain already started, Skipping: ", chn)
			}
		}()
	}

	wg.Wait()
	return nil
}

func PerformCommand(action *def.Action, actionVars []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	actionVars = append(os.Environ(), actionVars...)

	logger.Println("Performing Action: ", action.Name)

	for n, step := range action.Steps {
		cmd := exec.Command("sh", "-c", step)
		cmd.Env = actionVars
		cmd.Dir = dir

		prev, err := cmd.Output()
		if err != nil {
			return err
		}

		logger.Infoln(strings.TrimSpace(string(prev)))

		if n != 0 {
			actionVars = actionVars[:len(actionVars)-1]
		}
		actionVars = append(actionVars, ("prev=" + strings.TrimSpace(string(prev))))
	}

	logger.Println("Action performed")
	return nil
}
