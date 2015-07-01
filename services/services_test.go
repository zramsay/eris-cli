package services

import (
  "fmt"
  "os"
  "path"
  "strings"
  "testing"

  "github.com/eris-ltd/eris-cli/util"
  "github.com/eris-ltd/eris-cli/definitions"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
)

var srv *definitions.ServiceDefinition
var erisDir string = path.Join(os.TempDir(), "eris")
var servName string = "ipfs"
var hash string

func TestInit(t *testing.T) {
  // common is initialized on import so
  // we have to manually override these
  // variables to ensure that the tests
  // run correctly.
  common.ErisRoot     = erisDir
  common.ServicesPath = path.Join(common.ErisRoot, "services")

  // this dumps the ipfs service def into the temp dir which
  // has been set as the erisRoot
  util.Initialize(false, false)

  // init dockerClient
  util.DockerConnect(false)

  // set ipfs endpoint
  os.Setenv("ERIS_IPFS_HOST", "http://0.0.0.0")

  // make sure ipfs not running
  for _, r := range ListRunningRaw() {
    if r == "ipfs" {
      fmt.Println("IPFS service is running. Please stop it with eris services stop ipfs.")
      t.Fail()
      TestTearDown(t)
      os.Exit(1)
    }
  }

  // make sure ipfs container does not exist
  for _, r := range ListExistingRaw() {
    if r == "ipfs" {
      fmt.Println("IPFS service exists. Please remove it with eris services rm ipfs.")
      t.Fail()
      TestTearDown(t)
      os.Exit(1)
    }
  }
}

func TestKnownRaw(t *testing.T) {
  k := ListKnownRaw()
  if len(k) != 1 {
    fmt.Printf("More than one service definition found. Something is wrong.\n")
    t.Fail()
    TestTearDown(t)
    os.Exit(1)
  }

  if k[0] != "ipfs" {
    fmt.Printf("Could not find ipfs service definition.\n")
    t.Fail()
    TestTearDown(t)
    os.Exit(1)
  }
}

func TestLoadServiceDefinition(t *testing.T) {
  var e error
  srv, e = LoadServiceDefinition(servName)
  if e != nil {
    fmt.Println(e)
    t.FailNow()
  }

  if srv.Service.Name != servName {
    fmt.Printf("FAILURE: improper service name on LOAD. expected: %s\tgot: %s\n", servName, srv.Service.Name)
    t.FailNow()
  }

  if !srv.Operations.DataContainer {
    fmt.Printf("FAILURE: data_container not properly read on LOAD.\n")
    t.FailNow()
  }

  if srv.Operations.DataContainerName == "" {
    fmt.Printf("FAILURe: data_container_name not set.\n")
    t.FailNow()
  }
}

func TestLoadService(t *testing.T) {
  s, e := LoadService(servName)
  if e != nil {
    fmt.Println(e)
    t.FailNow()
  }

  if s.Name != servName {
    fmt.Printf("FAILURE: improper service name on LOAD_SERVICE. expected: %s\tgot: %s\n", servName, s.Name)
    t.FailNow()
  }
}

func TestStartServiceRaw(t *testing.T) {
  e := StartServiceRaw(servName)
  if e != nil {
    fmt.Println(e)
    t.Fail()
  }

  testRunAndExist(t, servName, true, true)
}

func TestInspectRaw(t *testing.T) {
  e := InspectServiceRaw(servName, "name")
  if e != nil {
    fmt.Println(e)
    t.Fail()
  }

  e = InspectServiceRaw(servName, "config.user")
  if e != nil {
    fmt.Println(e)
    t.Fail()
  }
}

func TestLogsRaw(t *testing.T) {
  e := LogsServiceRaw(servName, false)
  if e != nil {
    fmt.Println(e)
    t.Fail()
  }
}

func TestExecRaw(t *testing.T) {
  if os.Getenv("TEST_IN_CIRCLE") == "true" {
    fmt.Println("Testing in Circle. Where we don't have exec privileges (due to their driver). Skipping test.")
    return
  }
  cmd := strings.Fields("ls -la /root/")
  e := ExecServiceRaw(servName, cmd, false)
  if e != nil {
    fmt.Println(e)
    t.Fail()
  }
}

// export is not currently working
// consistently getting the following error
// no matter what I do:
//
// Post http://0.0.0.0:8080/ipfs/: read tcp 127.0.0.1:8080: connection reset by peer
//
// I have no idea why it keeps rerouting 0.0.0.0 -> 127.0.0.1
// works fine outside of test environment
// func TestExportRaw(t *testing.T) {
//   e := ExportServiceRaw(servName)
//   if e != nil {
//     fmt.Println(e)
//     t.Fail()
//   }

//   // need to grab the hash
//   hash, e = exportFile(servName)
//   if e != nil {
//     fmt.Println(e)
//     t.Fail()
//   }
// }

// import is also not currently working
//
// Get http://0.0.0.0:8080/ipfs/Qma8GzJ7dHezN8GfrNzuq9JD199WgbQC7Qz29wwMX7JHf3: net/http: transport closed before response was received
//
// suspect this is related to the above testing error
// func TestImportRaw(t *testing.T) {
//   e := ImportServiceRaw("sfpi", "ipfs:Qma8GzJ7dHezN8GfrNzuq9JD199WgbQC7Qz29wwMX7JHf3")
//   if e != nil {
//     fmt.Println(e)
//     t.Fail()
//   }
// }

func TestUpdateRaw(t *testing.T) {
  if os.Getenv("TEST_IN_CIRCLE") == "true" {
    fmt.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
    return
  }

  e := UpdateServiceRaw(servName, true)
  if e != nil {
    fmt.Println(e)
    t.Fail()
  }

  testRunAndExist(t, servName, true, true)
}

func TestKillRaw(t *testing.T) {
  e := KillServiceRaw(true, false, servName)
  if e != nil {
    fmt.Println(e)
    t.Fail()
  }

  testRunAndExist(t, servName, true, false)
}

func TestRmRaw(t *testing.T) {
  if os.Getenv("TEST_IN_CIRCLE") == "true" {
    fmt.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
    return
  }

  s := []string{servName}
  e := RmServiceRaw(s, false)
  if e != nil {
    fmt.Println(e)
    t.Fail()
  }

  testRunAndExist(t, servName, false, false)
}

func TestNewRaw(t *testing.T) {
  e := NewServiceRaw("keys", "eris/keys")
  if e != nil {
    fmt.Println(e)
    t.FailNow()
  }

  e = StartServiceRaw("keys")
  if e != nil {
    fmt.Println(e)
    t.Fail()
  }

  testRunAndExist(t, "keys", true, true)
}

func TestRenameRaw(t *testing.T) {
  e := RenameServiceRaw("keys", "syek")
  if e != nil {
    fmt.Println(e)
    t.Fail()
  }

  testRunAndExist(t, "syek", true, true)

  e = RenameServiceRaw("syek", "keys")
  if e != nil {
    fmt.Println(e)
    t.Fail()
  }

  testRunAndExist(t, "keys", true, true)
}

// tests remove+kill
func TestKillRawPostNew(t *testing.T) {
  if os.Getenv("TEST_IN_CIRCLE") == "true" {
    fmt.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
    return
  }

  e := KillServiceRaw(true, true, "keys")
  if e != nil {
    fmt.Println(e)
    t.Fail()
  }

  testRunAndExist(t, "keys", false, false)
}

func TestTearDown(t *testing.T) {
  e := os.RemoveAll(erisDir)
  if e != nil {
    fmt.Println(e)
    t.FailNow()
    os.Exit(1)
  }
}

func testRunAndExist(t *testing.T, servName string, toExist, toRun bool) {
  var exist, run bool
  for _, r := range ListExistingRaw() {
    if r == servName {
      exist = true
    }
  }
  for _, r := range ListRunningRaw() {
    if r == servName {
      run = true
    }
  }

  if toRun != run {
    if toRun {
      fmt.Println("Could not find a running instance of ipfs")
      t.Fail()
    } else {
      fmt.Println("Found a running instance of ipfs when I shouldn't have")
      t.Fail()
    }
  }

  if toExist != exist {
    if toExist {
      fmt.Println("Could not find an existing instance of ipfs")
      t.Fail()
    } else {
      fmt.Println("Found an existing instance of ipfs when I shouldn't have")
      t.Fail()
    }
  }
}