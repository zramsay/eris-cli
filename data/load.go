package data

import (
  "fmt"
  "os"
  "strconv"

  def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"
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

func checkServiceGiven(args []string) {
  if len(args) == 0 {
    // TODO: betterly error handling
    fmt.Println("No Data Container Given. Please rerun command with a known data container.")
    os.Exit(1)
  }
}

func parseKnown(name string) bool {
  known := ListKnownRaw()
  if len(known) != 0 {
    for _, srv := range known {
      if srv == name {
        return true
      }
    }
  }
  return false
}
