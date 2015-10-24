package loaders

import (
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
)

// LoadDataDefinitions returns returns a container operations structure for
// a blank data container specified by a name dataName and a cNum number.
func LoadDataDefinition(dataName string, cNum int) *definitions.Operation {
	if cNum == 0 {
		cNum = 1
	}

	logger.Debugf("Loading Data Definition =>\t%s:%d\n", dataName, cNum)

	ops := definitions.BlankOperation()
	ops.ContainerNumber = cNum
	ops.ContainerType = definitions.TypeData
	ops.SrvContainerName = util.DataContainersName(dataName, cNum)
	ops.DataContainerName = util.DataContainersName(dataName, cNum)
	ops.Labels = util.Labels(dataName, ops)

	return ops
}
