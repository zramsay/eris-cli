package data

import (
	"fmt"

	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
)

func MockService(name string, containerNumber int) (*def.Service, *def.ServiceOperation) {
	srv := &def.Service{}
	ops := &def.ServiceOperation{}
	ops.SrvContainerName = nameToContainerName(name, containerNumber)
	return srv, ops
}

func nameToContainerName(name string, containerNumber int) string {
	return fmt.Sprintf("eris_data_%s_%d", name, containerNumber)
}

func checkServiceGiven(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("No Data Container Given. Please rerun command with a known data container.")
	}
	return nil
}

func parseKnown(name string, num int) bool {
	name = util.NameAndNumber(name, num)
	known, _ := ListKnownRaw()
	if len(known) != 0 {
		for _, srv := range known {
			if srv == name {
				return true
			}
		}
	}
	return false
}
