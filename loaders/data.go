package loaders

import (
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/Sirupsen/logrus"
)

// LoadDataDefinition returns an Operation structure for a blank data container
// specified by a name dataName
func LoadDataDefinition(dataName string) *definitions.Operation {

	log.WithField("=>", dataName).Debug("Loading data definition")

	ops := definitions.BlankOperation()
	ops.ContainerType = definitions.TypeData
	ops.SrvContainerName = util.DataContainerName(dataName)
	ops.DataContainerName = util.DataContainerName(dataName)
	ops.Labels = util.Labels(dataName, ops)

	return ops
}
