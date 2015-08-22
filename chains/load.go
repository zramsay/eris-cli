package chains

import (
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
)

func IsChainExisting(chain *definitions.Chain) bool {
	logger.Debugf("Does Chain Exist? =>\t\t%s:%d\n", chain.Name, chain.Operations.ContainerNumber)
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
