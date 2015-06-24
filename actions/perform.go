package actions

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Do(config *util.ErisCli, cmd *cobra.Command, args []string) {
	action, actionVars, err := LoadActionDefinition(args)
	if err != nil {
		fmt.Fprintln(config.ErrorWriter, err)
		return
	}
	err = DoRaw(action, actionVars, cmd.Flags().Lookup("verbose").Changed, config.Writer, config.ErrorWriter)
	if err != nil {
		fmt.Fprintln(config.ErrorWriter, err)
		return
	}
}

func DoRaw(action *def.Action, actionVars []string, verbose bool, w, ew io.Writer) error {
	err := StartServicesAndChains(action, verbose, w, ew)
	if err != nil {
		return err
	}

	err = PerformCommand(action, actionVars, verbose, w)
	if err != nil {
		return err
	}
	return nil
}

func StartServicesAndChains(action *def.Action, verbose bool, w, ew io.Writer) error {
	// start the services and chains
	var wg sync.WaitGroup
	var skip bool

	running := services.ListRunningRaw()
	for _, srv := range action.Services {
		skip = false

		for _, run := range running {
			if srv == run {
				if verbose {
					fmt.Fprintln(w, "Service already started, Skipping: ", srv)
				}
				skip = true
			}
		}
		if skip {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := services.StartServiceRaw(srv, verbose, w)
			if err != nil {
				fmt.Fprintln(w, "Service already started, Skipping: ", srv)
			}
		}()
	}

	running = chains.ListRunningRaw()
	for _, chn := range action.Chains {
		skip = false

		for _, run := range running {
			if chn == run {
				if verbose {
					fmt.Fprintln(w, "Chain already started, Skipping: ", chn)
				}
				skip = true
			}
		}
		if skip {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := chains.StartChainRaw(chn, verbose, w)
			if err != nil {
				fmt.Fprintln(w, "Chain already started, Skipping: ", chn)
			}
		}()
	}

	wg.Wait()
	return nil
}

func PerformCommand(action *def.Action, actionVars []string, verbose bool, w io.Writer) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	actionVars = append(os.Environ(), actionVars...)

	fmt.Fprintln(w, "Performing Action: ", action.Name)

	for n, step := range action.Steps {
		cmd := exec.Command("sh", "-c", step)
		cmd.Env = actionVars
		cmd.Dir = dir

		prev, err := cmd.Output()
		if err != nil {
			return err
		}

		if verbose {
			fmt.Fprintln(w, strings.TrimSpace(string(prev)))
		}

		if n != 0 {
			actionVars = actionVars[:len(actionVars)-1]
		}
		actionVars = append(actionVars, ("prev=" + strings.TrimSpace(string(prev))))
	}

	fmt.Fprintln(w, "Action performed")
	return nil
}
