package loaders

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

// LoadDataDefinition returns an Operation structure for a blank data container
// specified by a name dataName and a cNum number.
func LoadDataDefinition(dataName string, cNum int) *definitions.Operation {
	if cNum == 0 {
		cNum = 1
	}

	log.WithField("=>", fmt.Sprintf("%s:%d", dataName, cNum)).Debug("Loading data definition")

	ops := definitions.BlankOperation()
	ops.ContainerNumber = cNum
	ops.ContainerType = definitions.TypeData
	ops.SrvContainerName = util.DataContainersName(dataName, cNum)
	ops.DataContainerName = util.DataContainersName(dataName, cNum)
	ops.Labels = util.Labels(dataName, ops)

	return ops
}
