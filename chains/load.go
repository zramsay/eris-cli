package chains

import (
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

func IsChainExisting(chain *definitions.Chain) bool {
	log.WithField("=>", chain.Name).Debug("Checking chain existing")
	cName := util.FindChainContainer(chain.Name, true)
	if cName == nil {
		return false
	}
	chain.Operations.SrvContainerID = cName.ContainerID
	return true
}

func IsChainRunning(chain *definitions.Chain) bool {
	log.WithField("=>", chain.Name).Debug("Checking chain running")
	cName := util.FindChainContainer(chain.Name, false)
	if cName == nil {
		return false
	}
	chain.Operations.SrvContainerID = cName.ContainerID
	return true
}
