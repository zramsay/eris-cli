package chains

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

func IsChainExisting(chain *definitions.Chain) bool {
	log.WithField("=>", fmt.Sprintf("%s:%d", chain.Name, chain.Operations.ContainerNumber)).Debug("Checking chain existing")
	cName := util.FindChainContainer(chain.Name, chain.Operations.ContainerNumber, true)
	if cName == nil {
		return false
	}
	chain.Operations.SrvContainerID = cName.ContainerID
	return true
}

func IsChainRunning(chain *definitions.Chain) bool {
	log.WithField("=>", fmt.Sprintf("%s:%d", chain.Name, chain.Operations.ContainerNumber)).Debug("Checking chain running")
	cName := util.FindChainContainer(chain.Name, chain.Operations.ContainerNumber, false)
	if cName == nil {
		return false
	}
	chain.Operations.SrvContainerID = cName.ContainerID
	return true
}
