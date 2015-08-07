package services

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"testing"

	def "github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/log"
)

var srv *def.ServiceDefinition
var erisDir string = path.Join(os.TempDir(), "eris")
var servName string = "ipfs"
var hash string

func TestMain(m *testing.M) {
	var logLevel int

	if os.Getenv("LOG_LEVEL") != "" {
		logLevel, _ = strconv.Atoi(os.Getenv("LOG_LEVEL"))
	} else {
		logLevel = 0
		// logLevel = 1
		// logLevel = 2
	}
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

	if len(k) != 2 {
		ifExit(fmt.Errorf("More than two service definitions found. Something is wrong.\n"))
	}

	if k[0] != "ipfs" {
		ifExit(fmt.Errorf("Could not find ipfs service definition. Services found =>\t%v\n", k))
	}
}

func TestLoadServiceDefinition(t *testing.T) {
	var e error
	srv, e = loaders.LoadServiceDefinition(servName, true, 1)
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

func TestStartService(t *testing.T) {
	do := def.NowDo()
	do.Args = []string{servName}
	do.Operations.ContainerNumber = util.AutoMagic(0, "service", true)
	logger.Debugf("Starting service (via tests) =>\t%s:%d\n", servName, do.Operations.ContainerNumber)
	e := StartService(do)
	if e != nil {
		logger.Infof("Error starting service =>\t%v\n", e)
		t.Fail()
	}

	testExistAndRun(t, servName, 1, true, true)
	testNumbersExistAndRun(t, servName, 1, 1)
}

func TestInspectService(t *testing.T) {
	do := def.NowDo()
	do.Name = servName
	do.Args = []string{"name"}
	do.Operations.ContainerNumber = 1
	logger.Debugf("Inspect service (via tests) =>\t%s:%v:%d\n", servName, do.Args, do.Operations.ContainerNumber)
	e := InspectService(do)
	if e != nil {
		logger.Infof("Error inspecting service =>\t%v\n", e)
		t.FailNow()
	}

	do = def.NowDo()
	do.Name = servName
	do.Args = []string{"config.user"}
	do.Operations.ContainerNumber = 1
	logger.Debugf("Inspect service (via tests) =>\t%s:%v\n", servName, do.Args)
	e = InspectService(do)
	if e != nil {
		logger.Infof("Error inspecting service =>\t%v\n", e)
		t.Fail()
	}
}

func TestLogsService(t *testing.T) {
	do := def.NowDo()
	do.Name = servName
	do.Follow = false
	do.Tail = "all"
	logger.Debugf("Inspect logs (via tests) =>\t%s:%v\n", servName, do.Tail)
	e := LogsService(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}
}

func TestUpdateService(t *testing.T) {
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
		t.Fail()
	}

	testExistAndRun(t, servName, 1, true, true)
	testNumbersExistAndRun(t, servName, 1, 1)
}

func TestKillService(t *testing.T) {
	do := def.NowDo()
	do.Name = servName
	do.Rm = false
	do.RmD = false
	do.Args = []string{servName}
	logger.Debugf("Stopping serv (via tests) =>\t%s\n", servName)
	e := KillService(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testExistAndRun(t, servName, 1, true, false)
	testNumbersExistAndRun(t, servName, 1, 0)
}

func TestRmService(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	do := def.NowDo()
	do.Name = servName
	do.Args = []string{servName}
	do.File = false
	do.RmD = true
	logger.Debugf("Removing serv (via tests) =>\t%s\n", servName)
	e := RmService(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testExistAndRun(t, servName, 1, false, false)
	testNumbersExistAndRun(t, servName, 0, 0)
}

func TestNewService(t *testing.T) {
	do := def.NowDo()
	do.Name = "keys"
	do.Args = []string{"eris/keys"}
	logger.Debugf("New-ing serv (via tests) =>\t%s:%v\n", do.Name, do.Args)
	e := NewService(do)
	if e != nil {
		logger.Errorln(e)
		t.FailNow()
	}

	do = def.NowDo()
	do.Args = []string{"keys"}
	// do.Operations.ContainerNumber = util.AutoMagic(0, "service")
	//do.Operations.ContainerNumber = 1
	logger.Debugf("Stating serv (via tests) =>\t%v:%d\n", do.Args, do.Operations.ContainerNumber)
	e = StartService(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testExistAndRun(t, "keys", 1, true, true)
	testNumbersExistAndRun(t, "keys", 1, 1)
}

func TestRenameService(t *testing.T) {
	// log.SetLoggers(2, os.Stdout, os.Stderr)
	do := def.NowDo()
	do.Name = "keys"
	do.NewName = "syek"
	// do.Operations.ContainerNumber = util.AutoMagic(0, "service")
	//do.Operations.ContainerNumber = 1
	logger.Debugf("Renaming serv (via tests) =>\t%s:%v\n", do.Name, do.NewName)
	e := RenameService(do)
	if e != nil {
		logger.Errorf("Error (tests fail) =>\t\t%v\n", e)
		t.Fail()
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
	e = RenameService(do)
	if e != nil {
		logger.Errorf("Error (tests fail) =>\t\t%v\n", e)
		t.Fail()
	}

	testExistAndRun(t, "keys", 1, true, true)
	testExistAndRun(t, "syek", 1, false, false)

	testNumbersExistAndRun(t, "syek", 0, 0)
	testNumbersExistAndRun(t, "keys", 1, 1)
	// log.SetLoggers(0, os.Stdout, os.Stderr)
}

func TestKillServicePostNew(t *testing.T) {
	do := def.NowDo()
	do.Args = []string{"keys"}
	// do.Operations.ContainerNumber = util.AutoMagic(0, "service")
	//do.Operations.ContainerNumber = 1
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		do.Rm = true
		do.RmD = true
	}
	logger.Debugf("Killing service post new =>\t%s\n", do.Args)
	e := KillService(do)
	if e != nil {
		logger.Errorln(e)
		t.FailNow()
	}

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		testExistAndRun(t, "keys", 1, false, false)
		testExistAndRun(t, servName, 1, false, false)

		testNumbersExistAndRun(t, "keys", 0, 0)
		testNumbersExistAndRun(t, servName, 0, 0)
	} else {
		testExistAndRun(t, "keys", 1, true, false)
		testExistAndRun(t, servName, 1, true, false)

		testNumbersExistAndRun(t, "keys", 1, 0)
		testNumbersExistAndRun(t, servName, 1, 0)
	}
}

func TestCatService(t *testing.T) {
	filePath := path.Join(ServicesPath, "keys.toml")
	addition := `services = ["` + servName + `"]` + "\n\n"
	additReg := `services = \["` + servName + `"\]`
	reg := regexp.MustCompile(additReg)

	orig, err := ioutil.ReadFile(filePath)
	if err != nil {
		logger.Errorln(err)
		t.FailNow()
	}
	orig = append([]byte(addition), orig...)
	ioutil.WriteFile(filePath, orig, 0777)

	var sec []byte
	sec, err = ioutil.ReadFile(filePath)
	secd := string(sec)
	if !reg.MatchString(secd) {
		logger.Errorf("FAIL: Additional Service Not Found Pre Cat. I got: %s\n", secd)
		t.FailNow()
	}

	do := def.NowDo()
	do.Name = "keys"
	err = CatService(do)
	if err != nil {
		logger.Errorln(err)
		t.FailNow()
	}

	secd = do.Result
	if !reg.MatchString(secd) {
		logger.Errorf("FAIL: Additional Service Not Found Post Cat. I got: %s\n", secd)
		t.FailNow()
	}
}

func TestStartServiceWithDependencies(t *testing.T) {
	do := def.NowDo()
	do.Args = []string{"keys"}
	//do.Operations.ContainerNumber = 1
	// do.Operations.ContainerNumber = util.AutoMagic(0, "service")
	logger.Debugf("Starting service with deps =>\t%s:%s\n", "keys", servName)
	e := StartService(do)
	if e != nil {
		logger.Infof("Error starting service =>\t%v\n", e)
		t.Fail()
	}

	testExistAndRun(t, servName, 1, true, true)
	testExistAndRun(t, "keys", 1, true, true)

	testNumbersExistAndRun(t, "keys", 1, 1)
	testNumbersExistAndRun(t, servName, 1, 1)
}

// tests remove+kill
func TestKillServiceWithDependencies(t *testing.T) {
	do := def.NowDo()
	do.Args = []string{"keys"}
	do.All = true
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		do.Rm = true
		do.RmD = true
	}
	logger.Debugf("Kill service with deps =>\t%v\n", do.Args)
	e := KillService(do)
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		testExistAndRun(t, servName, 1, false, false)
		testExistAndRun(t, "keys", 1, false, false)

		testNumbersExistAndRun(t, "keys", 0, 0)
		testNumbersExistAndRun(t, servName, 0, 0)
	} else {
		testExistAndRun(t, servName, 1, true, false)
		testExistAndRun(t, "keys", 1, true, false)

		testNumbersExistAndRun(t, "keys", 1, 0)
		testNumbersExistAndRun(t, servName, 1, 0)
	}
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
		t.FailNow()
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
		t.FailNow()
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
		t.Fail()
	}

	if toRun != run {
		if toRun {
			logger.Printf("Could not find a running =>\t%s\n", servName)
		} else {
			logger.Printf("Found a running instance of %s when I shouldn't have\n", servName)
		}
		t.Fail()
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
		t.Fail()
	}

	if run != containerRun {
		logger.Printf("Wrong number of containers running for service (%s). Expected (%d). Got (%d).\n", servName, containerRun, run)
		t.Fail()
	}

	logger.Infoln("All good.\n")
}

func testsInit() error {
	var err error
	// TODO: make a reader/pipe so we can see what is written from tests.
	util.GlobalConfig, err = util.SetGlobalObject(os.Stdout, os.Stderr)
	ifExit(err)

	// common is initialized on import so
	// we have to manually override these
	// variables to ensure that the tests
	// run correctly.
	util.ChangeErisDir(erisDir)

	// init dockerClient
	util.DockerConnect(false)

	// this dumps the ipfs service def into the temp dir which
	// has been set as the erisRoot
	ifExit(ini.Initialize(false, false, false, false))

	// set ipfs endpoint
	os.Setenv("ERIS_IPFS_HOST", "http://0.0.0.0")

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
}

func ifExit(err error) {
	if err != nil {
		logger.Errorln(err)
		log.Flush()
		testsTearDown()
		os.Exit(1)
	}
}
