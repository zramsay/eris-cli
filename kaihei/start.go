package kaihei

import (
	"fmt"

//	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/definitions"
)

func StartUpEris(do *definitions.Do) error {

	// start services
	listofServices := util.ErisContainersByType(definitions.TypeService, true)
	names := make([]string, len(listofServices))
	for i, serviceName := range listofServices {
		names[i] = serviceName.ShortName
	}

	fmt.Println(names)

	doStart := definitions.NowDo()
	doStart.ServicesSlice = names
	if err := services.StartService(doStart); err != nil {
		return err
	}	
	// start chains
	// 	IfExit(ArgCheck(1, "ge", cmd, args))
	// do.Name = args[0]
	// do.Run = true
	// IfExit(chns.StartChain(do))
	return nil
}

func ShutUpEris(do *definitions.Do) error {
	fmt.Println("todo: stop:")
	return nil
}
