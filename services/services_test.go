package services

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/list"
	"github.com/eris-ltd/eris-cli/loaders"
	tests "github.com/eris-ltd/eris-cli/testutils"
	"github.com/eris-ltd/eris-cli/util"
	ver "github.com/eris-ltd/eris-cli/version"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	logger "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

var srv *def.ServiceDefinition

var servName string = "ipfs"

func TestMain(m *testing.M) {
	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	log.SetLevel(log.DebugLevel)

	tests.IfExit(testsInit())

	// Prevent CLI from starting IPFS.
	os.Setenv("ERIS_SKIP_ENSURE", "true")

	exitCode := m.Run()

	log.Info("Tearing tests down")
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		tests.IfExit(tests.TestsTearDown())
	}

	os.Exit(exitCode)
}

func TestKnownServices(t *testing.T) {
	do := def.NowDo()
	do.Known = true
	do.Existing = false
	do.Running = false
	do.Quiet = true
	tests.IfExit(list.ListAll(do, "services"))
	k := strings.Split(do.Result, "\n") // tests output formatting.

	if len(k) != len(ver.SERVICE_DEFINITIONS) {
		tests.IfExit(fmt.Errorf("Did not find correct number of service definitions files, Expected %v, found %v.\n", len(ver.SERVICE_DEFINITIONS), len(k)))
	}

	servDefs := make(map[string]bool)

	for _, srv := range ver.SERVICE_DEFINITIONS {
		servDef := strings.Split(srv, ".")
		servDefs[servDef[0]] = true
	}

	i := 0
	for _, srvFile := range k {
		if servDefs[srvFile] == true {
			i++
		}
	}

	if i != len(ver.SERVICE_DEFINITION) {
		tests.IfExit(fmt.Errorf("Could not find all the expected service definition files.\n"))
	}
}

func TestLoadServiceDefinition(t *testing.T) {
	var e error
	srv, e = loaders.LoadServiceDefinition(servName, true, 1)
	if e != nil {
		log.Error(e)
		tests.IfExit(e)
	}

	if srv.Name != servName {
		log.WithFields(log.Fields{
			"expected": servName,
			"got":      srv.Name,
		}).Error("Improper name on load")
	}

	if srv.Service.Name != servName {
		log.WithFields(log.Fields{
			"expected": servName,
			"got":      srv.Service.Name,
		}).Error("Improper service name on load")

		tests.IfExit(e)
	}

	if !srv.Service.AutoData {
		log.Error("data_container not properly read on load")
		tests.IfExit(e)
	}

	if srv.Operations.DataContainerName == "" {
		log.Error("data_container_name not set")
		tests.IfExit(e)
	}
}

func TestStartKillService(t *testing.T) {
	testStartService(t, servName, false)
	testKillService(t, servName, true)
}

func TestInspectService(t *testing.T) {
	testStartService(t, servName, false)
	defer testKillService(t, servName, true)

	do := def.NowDo()
	do.Name = servName
	do.Operations.Args = []string{"name"}
	do.Operations.ContainerNumber = 1
	log.WithFields(log.Fields{
		"=>":   fmt.Sprintf("%s:%d", servName, do.Operations.ContainerNumber),
		"args": do.Operations.Args,
	}).Debug("Inspect service (from tests)")

	e := InspectService(do)
	if e != nil {
		log.Infof("Error inspecting service: %v", e)
		tests.IfExit(e)
	}

	do = def.NowDo()
	do.Name = servName
	do.Operations.Args = []string{"config.user"}
	do.Operations.ContainerNumber = 1
	log.WithFields(log.Fields{
		"=>":   servName,
		"args": do.Operations.Args,
	}).Debug("Inspect service (from tests)")
	e = InspectService(do)
	if e != nil {
		log.Infof("Error inspecting service: %v", e)
		tests.IfExit(e)
	}
}

func TestLogsService(t *testing.T) {
	testStartService(t, servName, false)
	defer testKillService(t, servName, true)
	do := def.NowDo()
	do.Name = servName
	do.Follow = false
	do.Tail = "5"
	log.WithFields(log.Fields{
		"=>":   servName,
		"tail": do.Tail,
	}).Debug("Inspect logs (from tests)")
	e := LogsService(do)
	if e != nil {
		log.Error(e)
		tests.IfExit(e)
	}
}

func TestExecService(t *testing.T) {
	testStartService(t, servName, true)
	defer testKillService(t, servName, true)

	do := def.NowDo()
	do.Name = servName
	do.Operations.Interactive = false
	do.Operations.Args = strings.Fields("ls -la /root/")
	log.WithFields(log.Fields{
		"=>":   servName,
		"args": do.Operations.Args,
	}).Debug("Executing service (from tests)")
	e := ExecService(do)
	if e != nil {
		log.Error(e)
		t.Fail()
	}
}

func TestUpdateService(t *testing.T) {
	testStartService(t, servName, false)
	defer testKillService(t, servName, true)
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		log.Warn("Testing in Circle where we don't have rm privileges. Skipping test")
		return
	}

	do := def.NowDo()
	do.Name = servName
	do.Pull = false
	do.Timeout = 1
	log.WithField("=>", servName).Debug("Update service (from tests)")
	e := UpdateService(do)
	if e != nil {
		log.Error(e)
		tests.IfExit(e)
	}

	testExistAndRun(t, servName, 1, true, true)
	testNumbersExistAndRun(t, servName, 1, 1)
}

func TestKillRmService(t *testing.T) {
	testStartService(t, servName, false)
	do := def.NowDo()
	do.Name = servName
	do.Rm = false
	do.RmD = false
	do.Operations.Args = []string{servName}
	log.WithField("=>", servName).Debug("Stopping service (from tests)")
	if e := KillService(do); e != nil {
		log.Error(e)
		tests.IfExit(e)
	}

	testExistAndRun(t, servName, 1, true, false)
	testNumbersExistAndRun(t, servName, 1, 0)

	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		log.Warn("Testing in Circle where we don't have rm privileges. Skipping test")
		return
	}

	do = def.NowDo()
	do.Name = servName
	do.Operations.Args = []string{servName}
	do.File = false
	do.RmD = true
	log.WithField("=>", servName).Debug("Removing service (from tests)")
	if e := RmService(do); e != nil {
		log.Error(e)
		tests.IfExit(e)
	}

	testExistAndRun(t, servName, 1, false, false)
	testNumbersExistAndRun(t, servName, 0, 0)
}

func TestExportService(t *testing.T) {
	do := def.NowDo()
	do.Name = "ipfs"

	hash := "QmQ1LZYPNG4wSb9dojRicWCmM4gFLTPKFUhFnMTR3GKuA2"

	// Fake IPFS server.
	os.Setenv("ERIS_IPFS_HOST", "http://127.0.0.1")
	ipfs := tests.NewServer("127.0.0.1:8080")
	ipfs.SetResponse(tests.ServerResponse{
		Code: http.StatusOK,
		Header: map[string][]string{
			"Ipfs-Hash": {hash},
		},
	})
	defer ipfs.Close()

	if err := ExportService(do); err != nil {
		tests.IfExit(err)
	}

	if expected := "/ipfs/"; ipfs.Path() != expected {
		tests.IfExit(fmt.Errorf("Called the wrong endpoint; expected %v, got %v\n", expected, ipfs.Path()))
	}

	if expected := "POST"; ipfs.Method() != expected {
		tests.IfExit(fmt.Errorf("Used the wrong HTTP method; expected %v, got %v\n", expected, ipfs.Method()))
	}

	if content := tests.FileContents(FindServiceDefinitionFile(do.Name)); content != ipfs.Body() {
		tests.IfExit(fmt.Errorf("Sent the bad file; expected %q, got %q\n", content, ipfs.Body()))
	}

	if hash != do.Result {
		tests.IfExit(fmt.Errorf("Hash mismatch; expected %q, got %q\n", hash, do.Result))
	}
}

func TestImportService(t *testing.T) {
	do := def.NowDo()
	do.Name = "eth"
	do.Hash = "QmQ1LZYPNG4wSb9dojRicWCmM4gFLTPKFUhFnMTR3GKuA2"
	log.WithFields(log.Fields{
		"=>":   do.Name,
		"hash": do.Hash,
	}).Debug("Importing service (from tests)")

	content := `name = "ipfs"

[service]
name = "ipfs"
image = "quay.io/eris/ipfs"`

	// Fake IPFS server.
	os.Setenv("ERIS_IPFS_HOST", "http://127.0.0.1")
	ipfs := tests.NewServer("127.0.0.1:8080")
	ipfs.SetResponse(tests.ServerResponse{
		Code: http.StatusOK,
		Body: content,
	})
	defer ipfs.Close()

	if err := ImportService(do); err != nil {
		tests.IfExit(err)
	}

	if expected := "/ipfs/" + do.Hash; ipfs.Path() != expected {
		tests.IfExit(fmt.Errorf("Called the wrong endpoint; expected %v, got %v\n", expected, ipfs.Path()))
	}

	if expected := "GET"; ipfs.Method() != expected {
		tests.IfExit(fmt.Errorf("Used the wrong HTTP method; expected %v, got %v\n", expected, ipfs.Method()))
	}

	if imported := tests.FileContents(FindServiceDefinitionFile(do.Name)); imported != content {
		tests.IfExit(fmt.Errorf("Returned unexpected content; expected: %q, got %q", content, imported))
	}
}

func TestNewService(t *testing.T) {
	do := def.NowDo()
	servName := "keys"
	do.Name = servName
	do.Operations.Args = []string{"quay.io/eris/keys"}

	log.WithFields(log.Fields{
		"=>":   do.Name,
		"args": do.Operations.Args,
	}).Debug("Creating a new service (from tests)")
	e := NewService(do)
	if e != nil {
		log.Error(e)
		tests.IfExit(e)
	}

	do = def.NowDo()
	do.Operations.Args = []string{servName}
	log.WithFields(log.Fields{
		"container number": do.Operations.ContainerNumber,
		"args":             do.Operations.Args,
	}).Debug("Starting service (from tests)")
	e = StartService(do)
	if e != nil {
		log.Error(e)
		tests.IfExit(e)
	}

	testExistAndRun(t, servName, 1, true, true)
	testNumbersExistAndRun(t, servName, 1, 1)
	testKillService(t, servName, true)
	testExistAndRun(t, servName, 1, false, false)
}

func TestRenameService(t *testing.T) {
	testStartService(t, "keys", false)
	testExistAndRun(t, "keys", 1, true, true)
	testNumbersExistAndRun(t, "keys", 1, 1)

	do := def.NowDo()
	do.Name = "keys"
	do.NewName = "syek"
	log.WithFields(log.Fields{
		"from": do.Name,
		"to":   do.NewName,
	}).Debug("Renaming service (from tests)")
	if e := RenameService(do); e != nil {
		tests.IfExit(fmt.Errorf("Error (tests fail) =>\t\t%v\n", e))
	}

	testExistAndRun(t, "syek", 1, true, true)
	testExistAndRun(t, "keys", 1, false, false)
	testNumbersExistAndRun(t, "syek", 1, 1)
	testNumbersExistAndRun(t, "keys", 0, 0)

	do = def.NowDo()
	do.Name = "syek"
	do.NewName = "keys"
	log.WithFields(log.Fields{
		"from": do.Name,
		"to":   do.NewName,
	}).Debug("Renaming service (from tests)")
	if e := RenameService(do); e != nil {
		log.Errorf("Error (tests fail): %v", e)
		tests.IfExit(e)
	}

	testExistAndRun(t, "keys", 1, true, true)
	testExistAndRun(t, "syek", 1, false, false)
	testNumbersExistAndRun(t, "keys", 1, 1)
	testNumbersExistAndRun(t, "syek", 0, 0)

	testKillService(t, "keys", true)
	testExistAndRun(t, "keys", 1, false, false)
}

func TestCatService(t *testing.T) {
	do := def.NowDo()
	do.Name = "ipfs"
	if err := CatService(do); err != nil {
		tests.IfExit(err)
	}
	//if init worked properly...?
	read, err := ioutil.ReadFile(filepath.Join(config.GlobalConfig.ErisDir, "services", "ipfs.toml"))
	if err != nil {
		tests.IfExit(err)
	}
	if do.Result != string(read) {
		tests.IfExit(fmt.Errorf("Cat Service on ipfs does not match Default. Got %s \n Expected %s", do.Result, string(read)))
	}
}

func TestStartKillServiceWithDependencies(t *testing.T) {
	do := def.NowDo()
	do.Operations.Args = []string{"do_not_use"}
	log.WithFields(log.Fields{
		"service":    servName,
		"dependency": "keys",
	}).Debug("Starting service with dependency (from tests)")
	if e := StartService(do); e != nil {
		log.Infof("Error starting service: %v", e)
		tests.IfExit(e)
	}

	defer func() {
		testKillService(t, "do_not_use", true)

		testExistAndRun(t, servName, 1, false, false)
		testNumbersExistAndRun(t, servName, 0, 0)

		testKillService(t, "keys", true)
	}()

	testExistAndRun(t, servName, 1, true, true)
	testExistAndRun(t, "keys", 1, true, true)

	testNumbersExistAndRun(t, "keys", 1, 1)
	testNumbersExistAndRun(t, servName, 1, 1)
}

//----------------------------------------------------------------------
// test utils!

func testStartService(t *testing.T, serviceName string, publishAll bool) {
	do := def.NowDo()
	do.Operations.Args = []string{serviceName}
	do.Operations.ContainerNumber = 1 //util.AutoMagic(0, "service", true)
	do.Operations.PublishAllPorts = publishAll
	log.WithField("=>", fmt.Sprintf("%s:%d", serviceName, do.Operations.ContainerNumber)).Debug("Starting service (from tests)")
	e := StartService(do)
	if e != nil {
		log.Infof("Error starting service: %v", e)
		tests.IfExit(e)
	}

	testExistAndRun(t, serviceName, 1, true, true)
	testNumbersExistAndRun(t, serviceName, 1, 1)
}

func testKillService(t *testing.T, serviceName string, wipe bool) {
	log.WithField("=>", servName).Debug("Stopping service (from tests)")

	do := def.NowDo()
	do.Name = serviceName
	do.Operations.Args = []string{serviceName}
	if wipe {
		do.Rm = true
		do.RmD = true
	}
	e := KillService(do)
	if e != nil {
		log.Error(e)
		tests.IfExit(e)
	}
	testExistAndRun(t, serviceName, 1, !wipe, false)
	testNumbersExistAndRun(t, serviceName, 0, 0)
}

func testExistAndRun(t *testing.T, servName string, containerNumber int, toExist, toRun bool) {
	if err := tests.TestExistAndRun(servName, "services", containerNumber, toExist, toRun); err != nil {
		tests.IfExit(nil)
	}
}

func testNumbersExistAndRun(t *testing.T, servName string, containerExist, containerRun int) {
	log.WithFields(log.Fields{
		"=>":        servName,
		"existing#": containerExist,
		"running#":  containerRun,
	}).Info("Checking number of containers for")

	log.WithField("=>", servName).Debug("Checking existing containers for")
	exist := util.HowManyContainersExisting(servName, "service")

	log.WithField("=>", servName).Debug("Checking running containers for")
	run := util.HowManyContainersRunning(servName, "service")

	if exist != containerExist {
		tests.IfExit(fmt.Errorf("Wrong number of containers existing for service (%s). Expected (%d). Got (%d).\n", servName, containerExist, exist))
	}

	if run != containerRun {
		tests.IfExit(fmt.Errorf("Wrong number of containers running for service (%s). Expected (%d). Got (%d).\n", servName, containerRun, run))
	}

	log.Info("All good")
}

func testsInit() error {
	if err := tests.TestsInit("services"); err != nil {
		return err
	}
	return nil
}
