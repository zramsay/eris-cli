package actions

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/services"

	dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func Do(cmd *cobra.Command, args []string) {
	verbose := cmd.Flags().Lookup("verbose").Changed
	action, actionVars := loadActionConfigAndVars(args)
	startServicesAndChains(cmd, action, verbose)
	PerformCommand(action, actionVars, verbose)
}

func loadActionConfigAndVars(act []string) (*def.Action, []string) {
	var action def.Action
	var actionConf = viper.New()
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

	actionConf.AddConfigPath(dir.ActionsPath)
	actionConf.SetConfigName(strings.Join(act, "_"))
	actionConf.ReadInConfig()

	err := actionConf.Marshal(&action)
	if err != nil {
		// TODO: error handling
		fmt.Println(err)
		os.Exit(1)
	}

	return &action, actionVars
}

func startServicesAndChains(cmd *cobra.Command, action *def.Action, verbose bool) {
	// start the services and chains
	var wg sync.WaitGroup
	var skip bool

	running := services.ListRunningRaw()
	for _, srv := range action.Services {
		skip = false

		for _, run := range running {
			if srv == run {
				if verbose {
					fmt.Println("Service already started, Skipping: ", srv)
				}
				skip = true
			}
		}
		if skip {
			continue
		}

		wg.Add(1)
		go func(service string) {
			defer wg.Done()
			services.Start(cmd, []string{service})
		}(srv)
	}

	running = chains.ListRunningRaw()
	for _, chn := range action.Chains {
		skip = false

		for _, run := range running {
			if chn == run {
				if verbose {
					fmt.Println("Chain already started, Skipping: ", chn)
				}
				skip = true
			}
		}
		if skip {
			continue
		}

		wg.Add(1)
		go func(chain string) {
			defer wg.Done()
			chains.Start(cmd, []string{chain})
		}(chn)
	}

	wg.Wait()
}

func PerformCommand(action *def.Action, actionVars []string, verbose bool) {
	dir, err := os.Getwd()
	if err != nil {
		// TODO: error handling
		fmt.Println(err)
		os.Exit(1)
	}
	actionVars = append(os.Environ(), actionVars...)

	fmt.Println("Performing Action: " + action.Name)

	for n, step := range action.Steps {
		cmd := exec.Command("sh", "-c", step)
		cmd.Env = actionVars
		cmd.Dir = dir

		prev, err := cmd.Output()
		if err != nil {
			// TODO: error handling
			fmt.Println(err)
			os.Exit(1)
		}

		if verbose {
			fmt.Println(strings.TrimSpace(string(prev)))
		}

		if n != 0 {
			actionVars = actionVars[:len(actionVars)-1]
		}
		actionVars = append(actionVars, ("prev=" + strings.TrimSpace(string(prev))))
	}

	fmt.Println("Action performed")
}
