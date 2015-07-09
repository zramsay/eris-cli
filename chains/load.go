package chains

import (
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
)
func IsChainExisting(chain *definitions.Chain) bool {
	cName := util.FindChainContainer(chain.Name, chain.Operations.ContainerNumber, true)
	if cName == nil {
		return false
	}
	chain.Operations.SrvContainerID = cName.ContainerID
	return true
}

func IsChainRunning(chain *definitions.Chain) bool {
	cName := util.FindChainContainer(chain.Name, chain.Operations.ContainerNumber, false)
	if cName == nil {
		return false
	}
	chain.Operations.SrvContainerID = cName.ContainerID
	return true

}

// check if given chain is known
func isKnownChain(name string) bool {
	known := util.GetGlobalLevelConfigFilesByType("chains", false)
	if len(known) != 0 {
		for _, srv := range known {
			if srv == name {
				return true
			}
		}
	}
	return false
}

func findChainDefinitionFile(name string) string {
	return util.GetFileByNameAndType("chains", name)
}