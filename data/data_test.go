package data

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
	"github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

var erisDir string = path.Join(os.TempDir(), "eris")
var dataName string = "dataTest1"
var newName string = "dataTest2"

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	logLevel = 0
	// logLevel = 1
	// logLevel = 2

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	if err := testsInit(); err != nil {
		logger.Errorln(err)
		os.Exit(1)
	}

	exitCode := m.Run()

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		if err := testsTearDown(); err != nil {
			logger.Errorln(err)
			log.Flush()
			os.Exit(1)
		}
	}

	os.Exit(exitCode)
}

func TestImportDataRawNoPriorExist(t *testing.T) {
	newDataDir := path.Join(common.DataContainersPath, dataName)
	if err := os.MkdirAll(newDataDir, 0777); err != nil {
		logger.Errorln(err)
		t.FailNow()
		os.Exit(1)
	}

	f, err := os.Create(path.Join(newDataDir, "test"))
	if err != nil {
		logger.Errorln(err)
		t.FailNow()
		os.Exit(1)
	}
	defer f.Close()

	do := definitions.NowDo()
	do.Name = dataName
	do.Operations.ContainerNumber = 1
	logger.Infof("Importing Data (from tests) =>\t%s\n", do.Name)
	if err := ImportData(do); err != nil {
		logger.Errorln(err)
		t.Fail()
	}

	testExist(t, dataName, true)
}

func TestRenameData(t *testing.T) {
	testExist(t, dataName, true)
	testExist(t, newName, false)

	do := definitions.NowDo()
	do.Name = dataName
	do.NewName = newName
	do.Operations.ContainerNumber = 1
	logger.Infof("Renaming Data (from tests) =>\t%s:%s\n", do.Name, do.NewName)
	if err := RenameData(do); err != nil {
		logger.Errorln(err)
		t.FailNow()
	}

	testExist(t, dataName, false)
	testExist(t, newName, true)

	do = definitions.NowDo()
	do.Name = newName
	do.NewName = dataName
	do.Operations.ContainerNumber = 1
	logger.Infof("Renaming Data (from tests) =>\t%s:%s\n", do.Name, do.NewName)
	if err := RenameData(do); err != nil {
		logger.Errorln(err)
		t.FailNow()
	}

	testExist(t, dataName, true)
	testExist(t, newName, false)
}

func TestInspectData(t *testing.T) {
	do := definitions.NowDo()
	do.Name = dataName
	do.Args = []string{"name"}
	do.Operations.ContainerNumber = 1
	logger.Infof("Inspecting Data (from tests) =>\t%s:%v\n", do.Name, do.Args)
	if err := InspectData(do); err != nil {
		logger.Errorln(err)
		t.FailNow()
	}

	do = definitions.NowDo()
	do.Name = dataName
	do.Args = []string{"config.network_disabled"}
	do.Operations.ContainerNumber = 1
	logger.Infof("Inspecting Data (from tests) =>\t%s:%v\n", do.Name, do.Args)
	if err := InspectData(do); err != nil {
		logger.Errorln(err)
		t.Fail()
	}
}

func TestExecData(t *testing.T) {
	do := definitions.NowDo()
	do.Name = dataName
	do.Args = []string{"mv", "/home/eris/.eris/test", "/home/eris/.eris/tset"}
	do.Interactive = false
	do.Operations.ContainerNumber = 1
	logger.Infof("Exec-ing Data (from tests) =>\t%s:%v\n", do.Name, do.Args)
	if err := ExecData(do); err != nil {
		logger.Errorln(err)
		t.Fail()
	}
}

func TestExportData(t *testing.T) {
	do := definitions.NowDo()
	do.Name = dataName
	do.Operations.ContainerNumber = 1
	if err := ExportData(do); err != nil {
		logger.Errorln(err)
		t.FailNow()
	}

	if _, err := os.Stat(path.Join(common.DataContainersPath, dataName, "tset")); os.IsNotExist(err) {
		logger.Errorf("Tragic! Exported file does not exist: %s\n", err)
		t.Fail()
	}
}

func TestRmData(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	do := definitions.NowDo()
	do.Name = dataName
	do.Operations.ContainerNumber = 1
	if err := RmData(do); err != nil {
		logger.Errorln(err)
		t.Fail()
	}

	do = definitions.NowDo()
	do.Name = newName
	do.Operations.ContainerNumber = 1
	RmData(do) // don't reap this error, it is just to check its Rm'ed
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
	ifExit(ini.Initialize(false, false, false))

	return nil
}

func testsTearDown() error {
	if e := os.RemoveAll(erisDir); e != nil {
		return e
	}

	return nil
}

func testExist(t *testing.T, name string, toExist bool) {
	var exist bool
	logger.Infof("\nTesting whether (%s) existing? (%t)\n", name, toExist)
	name = util.DataContainersName(name, 1)

	do := definitions.NowDo()
	do.Quiet = true
	if err := ListKnown(do); err != nil {
		logger.Errorln(err)
		t.FailNow()
	}
	res := strings.Split(do.Result, "\n")
	for _, r := range res {
		logger.Debugf("Existing =>\t\t\t%s\n", r)
		if r == util.ContainersShortName(name) {
			exist = true
		}
	}

	if toExist != exist {
		if toExist {
			logger.Infof("Could not find an existing =>\t%s\n", name)
		} else {
			logger.Infof("Found an existing instance of %s when I shouldn't have\n", name)
		}
		t.Fail()
	}
}

func ifExit(err error) {
	if err != nil {
		logger.Errorln(err)
		log.Flush()
		testsTearDown()
		os.Exit(1)
	}
}
