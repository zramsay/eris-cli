package pkgs

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"
)

func RunPackage(do *definitions.Do) error {
	var err error

	// Populates chainID from the chain
	// TODO link properly & get chainID not from chainName
	// XXX temp hack :-1:
	//do.ChainID = do.ChainName
	do.ChainName = fmt.Sprintf("tcp://%s:%s", do.ChainName, "46657") // TODO flexible port
	if err = util.GetChainID(do); err != nil {
		return err
	}

	// Load the package if it doesn't exist
	if do.Package == nil {
		do.Package, err = loaders.LoadPackage(do.YAMLPath)
		if err != nil {
			return err
		}
	}

	if do.LocalCompiler {
		if err := bootCompiler(); err != nil {
			return err
		}
		getLocalCompilerData(do)
	}

	return perform.RunJobs(do)
}

func bootCompiler() error {

	// add the compilers to the local services if the flag is pushed
	// [csk] note - when we move to default local compilers we'll remove
	// the compilers service completely and this will need to get
	// reworked to utilize DockerRun with a populated service def.
	doComp := definitions.NowDo()
	doComp.Name = "compilers"
	return services.StartService(doComp)
}

// getLocalCompilerData populates the IP:port combo for the compilers.
func getLocalCompilerData(do *definitions.Do) {
	// [csk]: note this is brittle we should only expose one port in the
	// docker file by default for the compilers service we can expose more
	// forcibly

	do.Compiler = "http://compilers:9099"
}
