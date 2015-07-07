package services

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/data"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
)

var srv *def.ServiceDefinition
var erisDir string = path.Join(os.TempDir(), "eris")
var servName string = "ipfs"
var hash string

func TestMain(m *testing.M) {
	logger.Level = 0
	// logger.Level = 1
	// logger.Level = 2

	if err := testsInit(); err != nil {
		fmt.Printf("Could not initialize the tests\n%v", err)
		os.Exit(1)
	}

	exitCode := m.Run()

	var e1, e2, e3 error
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		e1 = data.RmDataRaw("keys", 1)
		if e1 != nil {
			fmt.Println(e1)
		}
		e2 = data.RmDataRaw("ipfs", 1)
		if e2 != nil {
			fmt.Println(e2)
		}
	}

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		e3 = testsTearDown()
		if e3 != nil {
			fmt.Println(e3)
		}
	}

	if e1 != nil || e2 != nil || e3 != nil {
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func TestKnownServiceRaw(t *testing.T) {
	k := ListKnownRaw()

	if len(k) != 2 {
		logger.Errorf("More than two service definitions found. Something is wrong.\n")
		t.Fail()
		testsTearDown()
		os.Exit(1)
	}

	if k[1] != "ipfs" {
		logger.Errorf("Could not find ipfs service definition.\n")
		t.Fail()
		testsTearDown()
		os.Exit(1)
	}
}

func TestLoadServiceDefinition(t *testing.T) {
	var e error
	srv, e = LoadServiceDefinition(servName, 1)
	if e != nil {
		logger.Errorln(e)
		t.FailNow()
	}

	if srv.Name != servName {
		logger.Errorf("FAILURE: improper name on LOAD. expected: %s\tgot: %s\n", servName, srv.Name)
	}

	if srv.Service.Name != servName {
		logger.Errorf("FAILURE: improper service name on LOAD. expected: %s\tgot: %s\n", servName, srv.Service.Name)
		t.FailNow()
	}

	if !srv.Service.AutoData {
		logger.Errorf("FAILURE: data_container not properly read on LOAD.\n")
		t.FailNow()
	}

	if srv.Operations.DataContainerName == "" {
		logger.Errorf("FAILURE: data_container_name not set.\n")
		t.Fail()
	}
}

func TestStartServiceRaw(t *testing.T) {
	e := StartServiceRaw(servName, 1, def.BlankOperation())
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testRunAndExist(t, servName, 1, true, true)
}

func TestInspectServiceRaw(t *testing.T) {
	e := InspectServiceRaw(servName, "name", 1)
	if e != nil {
		logger.Errorln(e)
		t.FailNow()
	}

	e = InspectServiceRaw(servName, "config.user", 1)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}
}

func TestLogsServiceRaw(t *testing.T) {
	e := LogsServiceRaw(servName, false, "all", 1)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}
}

func TestExecServiceRaw(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have exec privileges (due to their driver). Skipping test.")
		return
	}
	cmd := strings.Fields("ls -la /root/")
	e := ExecServiceRaw(servName, cmd, false, 1)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}
}

func TestUpdateServiceRaw(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	e := UpdateServiceRaw(servName, true, 1)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testRunAndExist(t, servName, 1, true, true)
}

func TestKillServiceRaw(t *testing.T) {
	e := KillServiceRaw(true, false, false, 1, servName)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testRunAndExist(t, servName, 1, true, false)
}

func TestRmServiceRaw(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	s := []string{servName}
	e := RmServiceRaw(s, 1, false, false)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testRunAndExist(t, servName, 1, false, false)
}

func TestNewServiceRaw(t *testing.T) {
	e := NewServiceRaw("keys", "eris/keys")
	if e != nil {
		logger.Errorln(e)
		t.FailNow()
	}

	e = StartServiceRaw("keys", 1, new(def.Operation))
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testRunAndExist(t, "keys", 1, true, true)
}

func TestRenameServiceRaw(t *testing.T) {
	e := RenameServiceRaw("keys", "syek", 1)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testRunAndExist(t, "syek", 1, true, true)

	e = RenameServiceRaw("syek", "keys", 1)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testRunAndExist(t, "keys", 1, true, true)
}

// tests remove+kill
func TestKillServiceRawPostNew(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	e := KillServiceRaw(true, true, false, 1, "keys")
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}

	testRunAndExist(t, "keys", 1, false, false)
}

func testsInit() error {
	// common is initialized on import so
	// we have to manually override these
	// variables to ensure that the tests
	// run correctly.
	util.ChangeErisDir(erisDir)

	// this dumps the ipfs service def into the temp dir which
	// has been set as the erisRoot
	if err := util.Initialize(false, false); err != nil {
		return fmt.Errorf("TRAGIC. Could not initialize the eris dir: %s.\n", err)
	}

	// init dockerClient
	util.DockerConnect(false)

	// set ipfs endpoint
	os.Setenv("ERIS_IPFS_HOST", "http://0.0.0.0")

	// make sure ipfs not running
	for _, r := range ListRunningRaw(false) {
		if r == "ipfs" {
			return fmt.Errorf("IPFS service is running.\nPlease stop it with.\neris services stop -rx ipfs\n")
		}
	}

	// make sure ipfs container does not exist
	for _, r := range ListExistingRaw(false) {
		if r == "ipfs" {
			return fmt.Errorf("IPFS service exists.\nPlease remove it with\neris services rm ipfs\n")
		}
	}

	return nil
}

func testsTearDown() error {
	return os.RemoveAll(erisDir)
}

func testRunAndExist(t *testing.T, servName string, containerNumber int, toExist, toRun bool) {
	var exist, run bool
	servName = util.ServiceContainersName(servName, containerNumber)
	for _, r := range ListExistingRaw(false) {
		logger.Debugf("Existing =>\t%s\n", r)
		if r == util.ContainersShortName(servName) {
			exist = true
		}
	}
	for _, r := range ListRunningRaw(false) {
		logger.Debugf("Running => \t%s\n", r)
		if r == util.ContainersShortName(servName) {
			run = true
		}
	}

	if toRun != run {
		if toRun {
			logger.Errorf("Could not find a running instance of\t%s\n", servName)
		} else {
			logger.Errorf("Found a running instance of %s when I shouldn't have\n", servName)
		}
		t.Fail()
	}

	if toExist != exist {
		if toExist {
			logger.Errorf("Could not find an existing instance of\t%s\n", servName)
		} else {
			logger.Errorf("Found an existing instance of %s when I shouldn't have\n", servName)
		}
		t.Fail()
	}
}
