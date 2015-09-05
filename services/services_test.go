package services

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

var srv *def.ServiceDefinition
var erisDir string = path.Join(os.TempDir(), "eris")
var servName string = "ipfs"
var hash string

var DEAD bool // XXX: don't double panic (TODO: Flushing twice blocks)

func fatal(t *testing.T, err error) {
	if !DEAD {
		log.Flush()
		testsTearDown()
		DEAD = true
		panic(err)
	}
}

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	logLevel = 0
	//logLevel = 1
	// logLevel = 3

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	ifExit(testsInit())

	exitCode := m.Run()

	logger.Infoln("Commensing with Tests Tear Down.")
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		ifExit(testsTearDown())
	}

	os.Exit(exitCode)
}

func TestKnownService(t *testing.T) {
	do := def.NowDo()
	ifExit(ListKnown(do))
	k := strings.Split(do.Result, "\n") // tests output formatting.

	if len(k) != 3 {
		ifExit(fmt.Errorf("Did not find exactly 3 service definitions files. Something is wrong.\n"))
	}

	if k[1] != "ipfs" {
		ifExit(fmt.Errorf("Could not find ipfs service definition. Services found =>\t%v\n", k))
	}
}

func TestLoadServiceDefinition(t *testing.T) {
	var e error
	srv, e = loaders.LoadServiceDefinition(servName, true, 1)
	if e != nil {
		logger.Errorln(e)
		fatal(t, e)
	}

	if srv.Name != servName {
		logger.Errorf("FAILURE: improper name on LOAD. expected: %s\tgot: %s\n", servName, srv.Name)
	}

	if srv.Service.Name != servName {
		logger.Errorf("FAILURE: improper service name on LOAD. expected: %s\tgot: %s\n", servName, srv.Service.Name)
		fatal(t, e)
	}

	if !srv.Service.AutoData {
		logger.Errorf("FAILURE: data_container not properly read on LOAD.\n")
		fatal(t, e)
	}

	if srv.Operations.DataContainerName == "" {
		logger.Errorf("FAILURE: data_container_name not set.\n")
		fatal(t, e)
	}
}

func TestStartKillService(t *testing.T) {
	testStartService(t, servName)
	testKillService(t, servName, true)
}

func TestInspectService(t *testing.T) {
	testStartService(t, servName)
	defer testKillService(t, servName, true)

	do := def.NowDo()
	do.Name = servName
	do.Args = []string{"name"}
	do.Operations.ContainerNumber = 1
	logger.Debugf("Inspect service (via tests) =>\t%s:%v:%d\n", servName, do.Args, do.Operations.ContainerNumber)
	e := InspectService(do)
	if e != nil {
		logger.Infof("Error inspecting service =>\t%v\n", e)
		fatal(t, e)
	}

	do = def.NowDo()
	do.Name = servName
	do.Args = []string{"config.user"}
	do.Operations.ContainerNumber = 1
	logger.Debugf("Inspect service (via tests) =>\t%s:%v\n", servName, do.Args)
	e = InspectService(do)
	if e != nil {
		logger.Infof("Error inspecting service =>\t%v\n", e)
		fatal(t, e)
	}
}

func TestLogsService(t *testing.T) {
	testStartService(t, servName)
	defer testKillService(t, servName, true)
	do := def.NowDo()
	do.Name = servName
	do.Follow = false
	do.Tail = "all"
	logger.Debugf("Inspect logs (via tests) =>\t%s:%v\n", servName, do.Tail)
	e := LogsService(do)
	if e != nil {
		logger.Errorln(e)
		fatal(t, e)
	}
}

func TestExecService(t *testing.T) {
	/*if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have exec privileges (due to their driver). Skipping test.")
		return
	}*/

	testStartService(t, servName)
	defer testKillService(t, servName, true)

	do := def.NowDo()
	do.Name = servName
	do.Interactive = false
	do.Args = strings.Fields("ls -la /root/")
	logger.Debugf("Exec-ing serv (via tests) =>\t%s:%v\n", servName, strings.Join(do.Args, " "))
	e := ExecService(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}
}

func TestUpdateService(t *testing.T) {
	testStartService(t, servName)
	defer testKillService(t, servName, true)
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	do := def.NowDo()
	do.Name = servName
	do.SkipPull = true
	do.Timeout = 1
	logger.Debugf("Update serv (via tests) =>\t%s\n", servName)
	e := UpdateService(do)
	if e != nil {
		logger.Errorln(e)
		fatal(t, e)
	}

	testExistAndRun(t, servName, 1, true, true)
	testNumbersExistAndRun(t, servName, 1, 1)
}

func TestKillRmService(t *testing.T) {
	testStartService(t, servName)
	do := def.NowDo()
	do.Name = servName
	do.Rm = false
	do.RmD = false
	do.Args = []string{servName}
	logger.Debugf("Stopping serv (via tests) =>\t%s\n", servName)
	if e := KillService(do); e != nil {
		logger.Errorln(e)
		fatal(t, e)
	}

	testExistAndRun(t, servName, 1, true, false)
	testNumbersExistAndRun(t, servName, 1, 0)

	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	do = def.NowDo()
	do.Name = servName
	do.Args = []string{servName}
	do.File = false
	do.RmD = true
	logger.Debugf("Removing serv (via tests) =>\t%s\n", servName)
	if e := RmService(do); e != nil {
		logger.Errorln(e)
		fatal(t, e)
	}

	testExistAndRun(t, servName, 1, false, false)
	testNumbersExistAndRun(t, servName, 0, 0)
}

func TestImportService(t *testing.T) {
	//	testStartService(t, "ipfs")
	//	defer testKillService(t, "ipfs", true)
	//XXX above functions paniced; this worked

	do := def.NowDo()
	do.Name = "ipfs"
	do.Operations.ContainerNumber = 1
	err := EnsureRunning(do)
	if err != nil {
		logger.Errorln(err)
		fatal(t, err)
	}

	servName := "eth"
	do.Name = servName
	do.Hash = "QmQ1LZYPNG4wSb9dojRicWCmM4gFLTPKFUhFnMTR3GKuA2"
	logger.Debugf("Import-ing serv (via tests) =>\t%s:%v\n", do.Name, do.Hash)

	e := ImportService(do)
	if e != nil {
		logger.Errorln(e)
		fatal(t, e)
	}

	testExistAndRun(t, "ipfs", 1, true, true)
}

func TestExportService(t *testing.T) {
	do := def.NowDo()
	do.Name = "ipfs"
	err := ExportService(do) //ExportService has EnsureRunning builtin
	if err != nil {
		logger.Errorln(err)
		fatal(t, err)
	}

	testExistAndRun(t, "ipfs", 1, true, true)
}

func TestNewService(t *testing.T) {
	do := def.NowDo()
	servName := "not-keys"
	do.Name = servName
	do.Args = []string{"eris/keys"}
	logger.Debugf("New-ing serv (via tests) =>\t%s:%v\n", do.Name, do.Args)
	e := NewService(do)
	if e != nil {
		logger.Errorln(e)
		fatal(t, e)
	}

	do = def.NowDo()
	do.Args = []string{servName}
	// do.Operations.ContainerNumber = util.AutoMagic(0, "service")
	//do.Operations.ContainerNumber = 1
	logger.Debugf("Starting serv (via tests) =>\t%v:%d\n", do.Args, do.Operations.ContainerNumber)
	e = StartService(do)
	if e != nil {
		logger.Errorln(e)
		fatal(t, e)
	}
	defer testKillService(t, servName, true)

	testExistAndRun(t, servName, 1, true, true)
	testNumbersExistAndRun(t, servName, 1, 1)
}

func TestRenameService(t *testing.T) {
	do := def.NowDo()
	do.Name = "keys"
	do.Args = []string{"eris/keys"}
	logger.Debugf("New-ing serv (via tests) =>\t%s:%v\n", do.Name, do.Args)
	if e := NewService(do); e != nil {
		logger.Errorln(e)
		fatal(t, nil)
	}

	testStartService(t, "keys")
	defer testKillService(t, "keys", true)

	// log.SetLoggers(2, os.Stdout, os.Stderr)
	do = def.NowDo()
	do.Name = "keys"
	do.NewName = "syek"
	// do.Operations.ContainerNumber = util.AutoMagic(0, "service")
	//do.Operations.ContainerNumber = 1
	logger.Debugf("Renaming serv (via tests) =>\t%s:%v\n", do.Name, do.NewName)
	if e := RenameService(do); e != nil {
		logger.Errorf("Error (tests fail) =>\t\t%v\n", e)
		fatal(t, nil)
	}

	testExistAndRun(t, "syek", 1, true, true)
	testExistAndRun(t, "keys", 1, false, false)

	testNumbersExistAndRun(t, "syek", 1, 1)
	testNumbersExistAndRun(t, "keys", 0, 0)

	do = def.NowDo()
	do.Name = "syek"
	do.NewName = "keys"
	// do.Operations.ContainerNumber = util.AutoMagic(0, "service")
	//do.Operations.ContainerNumber = 1
	logger.Debugf("Renaming serv (via tests) =>\t%s:%v\n", do.Name, do.NewName)
	if e := RenameService(do); e != nil {
		logger.Errorf("Error (tests fail) =>\t\t%v\n", e)
		fatal(t, e)
	}

	testExistAndRun(t, "keys", 1, true, true)
	testExistAndRun(t, "syek", 1, false, false)

	testNumbersExistAndRun(t, "syek", 0, 0)
	testNumbersExistAndRun(t, "keys", 1, 1)
	// log.SetLoggers(0, os.Stdout, os.Stderr)
}

func TestCatService(t *testing.T) {
	do := def.NowDo()
	do.Name = "do_not_use"
	if err := CatService(do); err != nil {
		fatal(t, err)
	}

	if do.Result != ini.DefaultIpfs2() {
		fatal(t, fmt.Errorf("Cat Service on keys does not match DefaultKeys. Got %s \n Expected %s", do.Result, ini.DefaultIpfs2()))
	}
}

func TestStartKillServiceWithDependencies(t *testing.T) {
	do := def.NowDo()
	do.Args = []string{"do_not_use"}
	//do.Operations.ContainerNumber = 1
	// do.Operations.ContainerNumber = util.AutoMagic(0, "service")
	logger.Debugf("Starting service with deps =>\t%s:%s\n", servName, "keys")
	if e := StartService(do); e != nil {
		logger.Infof("Error starting service =>\t%v\n", e)
		fatal(t, e)
	}

	defer func() {
		testKillService(t, "do_not_use", true)

		testExistAndRun(t, servName, 1, false, false)
		testNumbersExistAndRun(t, servName, 0, 0)

		// XXX: option for kill to kill dependencies too
		testKillService(t, "keys", true)
		//testExistAndRun(t, "keys", 1, false, false)
		//testNumbersExistAndRun(t, "keys", 1, 0)
	}()

	testExistAndRun(t, servName, 1, true, true)
	testExistAndRun(t, "keys", 1, true, true)

	testNumbersExistAndRun(t, "keys", 1, 1)
	testNumbersExistAndRun(t, servName, 1, 1)
}

//----------------------------------------------------------------------
// test utils!

func testStartService(t *testing.T, serviceName string) {
	do := def.NowDo()
	do.Args = []string{serviceName}
	do.Operations.ContainerNumber = 1 //util.AutoMagic(0, "service", true)
	logger.Debugf("Starting service (via tests) =>\t%s:%d\n", serviceName, do.Operations.ContainerNumber)
	e := StartService(do)
	if e != nil {
		logger.Infof("Error starting service =>\t%v\n", e)
		fatal(t, e)
	}

	testExistAndRun(t, serviceName, 1, true, true)
	testNumbersExistAndRun(t, serviceName, 1, 1)
}

func testKillService(t *testing.T, serviceName string, wipe bool) {
	logger.Debugf("Stopping serv (via tests) =>\t%s\n", servName)

	do := def.NowDo()
	do.Name = serviceName
	do.Args = []string{serviceName}
	if wipe {
		do.Rm = true
		do.RmD = true
	}
	e := KillService(do)
	if e != nil {
		logger.Errorln(e)
		fatal(t, e)
	}
	testExistAndRun(t, serviceName, 1, !wipe, false)
	testNumbersExistAndRun(t, serviceName, 0, 0)
}

func testExistAndRun(t *testing.T, servName string, containerNumber int, toExist, toRun bool) {
	var exist, run bool
	logger.Infof("\nTesting whether (%s) is running? (%t) and existing? (%t)\n", servName, toRun, toExist)
	servName = util.ServiceContainersName(servName, containerNumber)

	do := def.NowDo()
	do.Quiet = true
	do.Args = []string{"testing"}
	if err := ListExisting(do); err != nil {
		logger.Errorln(err)
		fatal(t, err)
	}
	res := strings.Split(do.Result, "\n")
	for _, r := range res {
		logger.Debugf("Existing =>\t\t\t%s\n", r)
		if r == util.ContainersShortName(servName) {
			exist = true
		}
	}

	do = def.NowDo()
	do.Quiet = true
	do.Args = []string{"testing"}
	if err := ListRunning(do); err != nil {
		logger.Errorln(err)
		fatal(t, err)
	}
	res = strings.Split(do.Result, "\n")
	for _, r := range res {
		logger.Debugf("Running =>\t\t\t%s\n", r)
		if r == util.ContainersShortName(servName) {
			run = true
		}
	}

	if toExist != exist {
		if toExist {
			logger.Printf("Could not find an existing =>\t%s\n", servName)
		} else {
			logger.Printf("Found an existing instance of %s when I shouldn't have\n", servName)
		}
		fatal(t, nil)
	}

	if toRun != run {
		if toRun {
			logger.Printf("Could not find a running =>\t%s\n", servName)
		} else {
			logger.Printf("Found a running instance of %s when I shouldn't have\n", servName)
		}
		fatal(t, nil)
	}

	logger.Infoln("All good.\n")
}

func testNumbersExistAndRun(t *testing.T, servName string, containerExist, containerRun int) {
	logger.Infof("\nTesting number of (%s) containers. Existing? (%d) and Running? (%d)\n", servName, containerExist, containerRun)

	logger.Debugf("Checking Existing Containers =>\t%s\n", servName)
	exist := util.HowManyContainersExisting(servName, "service")
	logger.Debugf("Checking Running Containers =>\t%s\n", servName)
	run := util.HowManyContainersRunning(servName, "service")

	if exist != containerExist {
		logger.Printf("Wrong number of containers existing for service (%s). Expected (%d). Got (%d).\n", servName, containerExist, exist)
		fatal(t, nil)
	}

	if run != containerRun {
		logger.Printf("Wrong number of containers running for service (%s). Expected (%d). Got (%d).\n", servName, containerRun, run)
		fatal(t, nil)
	}

	logger.Infoln("All good.\n")
}

func testsInit() error {
	var err error
	// TODO: make a reader/pipe so we can see what is written from tests.
	config.GlobalConfig, err = config.SetGlobalObject(os.Stdout, os.Stderr)
	ifExit(err)

	// common is initialized on import so
	// we have to manually override these
	// variables to ensure that the tests
	// run correctly.
	config.ChangeErisDir(erisDir)

	// init dockerClient
	util.DockerConnect(false, "eris")

	// this dumps the ipfs service def into the temp dir which
	// has been set as the erisRoot
	ifExit(ini.Initialize(true))

	// set ipfs endpoint
	//os.Setenv("ERIS_IPFS_HOST", "http://0.0.0.0") //conflicts with docker-machine

	// make sure ipfs not running
	do := def.NowDo()
	do.Quiet = true
	logger.Debugln("Finding the running services.")
	if err := ListRunning(do); err != nil {
		ifExit(err)
	}
	res := strings.Split(do.Result, "\n")
	for _, r := range res {
		if r == "ipfs" {
			ifExit(fmt.Errorf("IPFS service is running.\nPlease stop it with.\neris services stop -rx ipfs\n"))
		}
	}
	// make sure ipfs container does not exist
	do = def.NowDo()
	do.Quiet = true
	if err := ListExisting(do); err != nil {
		ifExit(err)
	}
	res = strings.Split(do.Result, "\n")
	for _, r := range res {
		if r == "ipfs" {
			ifExit(fmt.Errorf("IPFS service exists.\nPlease remove it with\neris services rm ipfs\n"))
		}
	}

	logger.Infoln("Test init completed. Starting main test sequence now.")
	return nil
}

func testsTearDown() error {
	return os.RemoveAll(erisDir)
	// return nil
}

func ifExit(err error) {
	if err != nil {
		logger.Errorln(err)
		log.Flush()
		testsTearDown()
		os.Exit(1)
	}
}
