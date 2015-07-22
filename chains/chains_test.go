package chains

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/log"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
)

var erisDir string = path.Join(os.TempDir(), "eris")
var chainName string = "testchain"
var hash string

func TestMain(m *testing.M) {
	var logLevel int
	var err error

	if os.Getenv("LOG_LEVEL") != "" {
		logLevel, _ = strconv.Atoi(os.Getenv("LOG_LEVEL"))
	} else {
		logLevel = 0
		// logLevel = 1
		// logLevel = 2
	}
	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	testsInit()
	logger.Infoln("Test init completed. Starting main test sequence now.")

	exitCode := m.Run()

	logger.Infoln("Commensing with Tests Tear Down.")
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		err = testsTearDown()
		if err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}
	}

	os.Exit(exitCode)
}

func TestKnownChain(t *testing.T) {
	do := def.NowDo()
	ifExit(ListKnown(do))

	k := strings.Split(do.Result, "\n") // tests output formatting.

	if k[0] != "" {
		logger.Debugf("Result =>\t\t%s\n", do.Result)
		ifExit(fmt.Errorf("Found a chain definition file. Something is wrong."))
	}
}

func TestNewChain(t *testing.T) {
	do := def.NowDo()
	do.GenesisFile = path.Join(common.BlockchainsPath, "genesis", "default.json")
	do.Name = chainName
	do.Operations.ContainerNumber = 1
	logger.Infof("Creating chain (from tests) =>\t%s\n", do.Name)
	e := NewChain(do) // configFile and dir are not needed for the tests.
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}
}

func TestLoadChainDefinition(t *testing.T) {
	var e error
	logger.Infof("Load chain def (from tests) =>\t%s\n", chainName)
	chn, e := loaders.LoadChainDefinition(chainName, false, 1)
	if e != nil {
		logger.Errorln(e)
		t.FailNow()
	}

	if chn.Service.Name != chainName {
		logger.Errorln("FAILURE: improper service name on LOAD. expected: %s\tgot: %s", chainName, chn.Service.Name)
		t.FailNow()
	}

	if !chn.Service.AutoData {
		logger.Errorln("FAILURE: data_container not properly read on LOAD.")
		t.FailNow()
	}

	if chn.Operations.DataContainerName == "" {
		logger.Errorln("FAILURE: data_container_name not set.")
		t.Fail()
	}
}

func TestStartChain(t *testing.T) {
	do := def.NowDo()
	do.Name = chainName
	do.Operations.ContainerNumber = 1
	logger.Infof("Starting chain (from tests) =>\t%s\n", do.Name)
	e := StartChain(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testExistAndRun(t, chainName, true, true)
}

func TestLogsChain(t *testing.T) {
	do := def.NowDo()
	do.Name = chainName
	do.Follow = false
	do.Tail = "all"
	logger.Infof("Get chain logs (from tests) =>\t%s:%s\n", do.Name, do.Tail)
	e := LogsChain(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}
}

func TestUpdateChain(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	do := def.NowDo()
	do.Name = chainName
	do.SkipPull = true
	logger.Infof("Updating chain (from tests) =>\t%s\n", do.Name)
	e := UpdateChain(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testExistAndRun(t, chainName, true, true)
}

func TestInspectChain(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	do := def.NowDo()
	do.Name = chainName
	do.Args = []string{"name"}
	do.Operations.ContainerNumber = 1
	logger.Debugf("Inspect chain (via tests) =>\t%s:%v\n", chainName, do.Args)
	e := InspectChain(do)
	if e != nil {
		logger.Infof("Error inspecting chain =>\t%v\n", e)
		t.FailNow()
	}

}

func TestRenameChain(t *testing.T) {
	do := def.NowDo()
	do.Name = chainName
	do.NewName = "niahctset"
	logger.Infof("Renaming chain (from tests) =>\t%s:%s\n", do.Name, do.NewName)
	e := RenameChain(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testExistAndRun(t, "niahctset", true, true)

	do = def.NowDo()
	do.Name = "niahctset"
	do.NewName = chainName
	logger.Infof("Renaming chain (from tests) =>\t%s:%s\n", do.Name, do.NewName)
	e = RenameChain(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testExistAndRun(t, chainName, true, true)
}

func TestKillChain(t *testing.T) {
	// log.SetLoggers(2, os.Stdout, os.Stderr)
	testExistAndRun(t, chainName, true, true)

	do := def.NowDo()
	do.Args = []string{"keys"}
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		do.Rm = true
		do.RmD = true
	}
	logger.Infof("Removing keys (from tests) =>\n%s\n", do.Name)
	e := services.KillService(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	do = def.NowDo()
	do.Name = chainName
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		do.Rm = true
		do.RmD = true
	}
	logger.Infof("Stopping chain (from tests) =>\t%s\n", do.Name)
	e = KillChain(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		testExistAndRun(t, chainName, false, false)
	} else {
		testExistAndRun(t, chainName, true, false)
	}
	// log.SetLoggers(0, os.Stdout, os.Stderr)
}

// TODO: finish this....
// func TestServiceWithChainDependencies(t *testing.T) {
// 	do := definitions.NowDo()
// 	do.Name = "keys"
// 	do.Args = []string{"eris/keys"}
// 	err := services.NewService(do)
// 	if err != nil {
// 		logger.Errorln(err)
// 		t.FailNow()
// 	}

// 	services.TestCatService(t)

// }

func TestRmChain(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	do := def.NowDo()
	do.Name = chainName
	do.RmD = true
	logger.Infof("Removing chain (from tests) =>\n%s\n", do.Name)
	e := RmChain(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testExistAndRun(t, chainName, false, false)
}

func testExistAndRun(t *testing.T, chainName string, toExist, toRun bool) {
	var exist, run bool
	logger.Infof("\nTesting whether (%s) is running? (%t) and existing? (%t)\n", chainName, toRun, toExist)
	chainName = util.ChainContainersName(chainName, 1) // not worried about containerNumbers, deal with multiple containers in services tests

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
		if r == util.ContainersShortName(chainName) {
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
		if r == util.ContainersShortName(chainName) {
			run = true
		}
	}

	if toExist != exist {
		if toExist {
			logger.Infof("Could not find an existing =>\t%s\n", chainName)
		} else {
			logger.Infof("Found an existing instance of %s when I shouldn't have\n", chainName)
		}
		t.Fail()
	}

	if toRun != run {
		if toRun {
			logger.Infof("Could not find a running =>\t%s\n", chainName)
		} else {
			logger.Infof("Found a running instance of %s when I shouldn't have\n", chainName)
		}
		t.Fail()
	}

	logger.Debugln("")
}

func testsInit() error {
	var err error
	// TODO: make a reader/pipe so we can see what is written from tests.
	util.GlobalConfig, err = util.SetGlobalObject(os.Stdout, os.Stderr)
	if err != nil {
		ifExit(fmt.Errorf("TRAGIC. Could not set global config.\n"))
	}

	// common is initialized on import so
	// we have to manually override these
	// variables to ensure that the tests
	// run correctly.
	util.ChangeErisDir(erisDir)

	// this dumps the ipfs service def into the temp dir which
	// has been set as the erisRoot
	if err := util.Initialize(false, false); err != nil {
		ifExit(fmt.Errorf("TRAGIC. Could not initialize the eris dir.\n"))
	}

	// init dockerClient
	util.DockerConnect(false)
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
