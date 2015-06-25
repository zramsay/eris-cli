package data

import (
	"fmt"
	"strconv"

	def "github.com/eris-ltd/eris-cli/definitions"
)

func mockService(name string) (*def.Service, *def.ServiceOperation) {
	srv := &def.Service{}
	ops := &def.ServiceOperation{}
	ops.SrvContainerName = nameToContainerName(name)
	return srv, ops
}

func nameToContainerName(name string) string {
	containerNumber := 1 // tmp
	return "eris_data_" + name + "_" + strconv.Itoa(containerNumber)
}

func checkServiceGiven(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("No Data Container Given. Please rerun command with a known data container.")
	}
	return nil
}

func parseKnown(name string) bool {
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
