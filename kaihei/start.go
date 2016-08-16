package kaihei

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/util"
	srv "github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/definitions"
)

func StartUpEris(do *definitions.Do) error {

	fmt.Println("starting up your services...")

	// start services
	listOfServices := util.ErisContainersByType(definitions.TypeService, false)

	if len(listOfServices) == 0 {
		return fmt.Errorf("no existing services to start")
	}

	names := make([]string, len(listOfServices))
	for i, serviceName := range listOfServices {
		names[i] = serviceName.ShortName
	}

	fmt.Println(names)

	doStart := definitions.NowDo()
	doStart.ServicesSlice = names
	if err := srv.StartService(doStart); err != nil {
		return err
	}
	
	// start chains
	// do.Name    - name of the chain (required)

	doChain := definitions.NowDo()
	doChain.Name = do.ChainName

	if doChain == nil {
		return nil
	}

	fmt.Println("starting up your chain...")
	if err := chains.StartChain(doChain); err != nil {
		return err
	}

	return nil
}

func ShutUpEris(do *definitions.Do) error {

	fmt.Println("shutting down your services...")

	// start services
	listOfServices := util.ErisContainersByType(definitions.TypeService, false)

	if len(listOfServices) == 0 {
		return fmt.Errorf("no existing services to stop")
	}

	names := make([]string, len(listOfServices))
	for i, serviceName := range listOfServices {
		names[i] = serviceName.ShortName
	}

	fmt.Println(names)

	doStop := definitions.NowDo()
	doStop.Operations.Args = names
	doStop.Timeout = 10
	if err := srv.KillService(doStop); err != nil {
		return err
	}

	// shutdown all chains

	return nil
}
