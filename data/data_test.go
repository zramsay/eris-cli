package data

import (
	"os"
	"path"
	"testing"

	"github.com/eris-ltd/eris-cli/definitions"
	tests "github.com/eris-ltd/eris-cli/testutils"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

var dataName string = "dataTest1"
var newName string = "dataTest2"

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	logLevel = 0
	// logLevel = 1
	//logLevel = 3

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	tests.IfExit(testsInit())

	exitCode := m.Run()

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		tests.IfExit(tests.TestsTearDown())
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

func TestExecData(t *testing.T) {
	do := definitions.NowDo()
	do.Name = dataName
	do.Operations.Args = []string{"mv", "/home/eris/.eris/test", "/home/eris/.eris/tset"}
	do.Operations.Interactive = false
	do.Operations.ContainerNumber = 1

	logger.Infof("Exec-ing Data (from tests) =>\t%s:%v\n", do.Name, do.Operations.Args)
	if err := ExecData(do); err != nil {
		logger.Errorln(err)
		t.Fail()
	}
}

func TestExportData(t *testing.T) {
	do := definitions.NowDo()
	do.Name = dataName
	do.ErisPath = common.DataContainersPath
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
	do.Operations.Args = []string{"name"}
	do.Operations.ContainerNumber = 1
	logger.Infof("Inspecting Data (from tests) =>\t%s:%v\n", do.Name, do.Operations.Args)
	if err := InspectData(do); err != nil {
		logger.Errorln(err)
		t.FailNow()
	}

	do = definitions.NowDo()
	do.Name = dataName
	do.Operations.Args = []string{"config.network_disabled"}
	do.Operations.ContainerNumber = 1
	logger.Infof("Inspecting Data (from tests) =>\t%s:%v\n", do.Name, do.Operations.Args)
	if err := InspectData(do); err != nil {
		logger.Errorln(err)
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
	if err := tests.TestsInit("data"); err != nil {
		return err
	}
	return nil
}

func testExist(t *testing.T, name string, toExist bool) {
	if tests.TestExistAndRun(name, "data", 1, toExist, false) {
		t.Fail()
	}
}
