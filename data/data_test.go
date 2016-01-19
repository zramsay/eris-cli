package data

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eris-ltd/eris-cli/definitions"
	tests "github.com/eris-ltd/eris-cli/testutils"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	logger "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

var dataName string = "dataTest1"
var newName string = "dataTest2"

func TestMain(m *testing.M) {
	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit("data"))

	exitCode := m.Run()

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		tests.IfExit(tests.TestsTearDown())
	}

	os.Exit(exitCode)
}

//TODO add some export/import test robustness
func TestImportDataRawNoPriorExist(t *testing.T) {
	testCreateDataByImport(t, dataName) //was basically copied from this test
	defer testKillDataCont(t, dataName)
}

func TestExportData(t *testing.T) {
	testCreateDataByImport(t, dataName)
	defer testKillDataCont(t, dataName)

	do := definitions.NowDo()
	do.Name = dataName
	do.Source = common.ErisContainerRoot
	do.Destination = filepath.Join(common.DataContainersPath, do.Name)
	do.Operations.ContainerNumber = 1
	if err := ExportData(do); err != nil {
		log.Error(err)
		t.FailNow()
	}

	if _, err := os.Stat(filepath.Join(common.DataContainersPath, dataName, "test")); os.IsNotExist(err) {
		log.Errorf("Tragic! Exported file does not exist: %s", err)
		t.Fail()
	}

}

//TODO !!
func TestListDatas(t *testing.T) {

}

func TestExecData(t *testing.T) {
	testCreateDataByImport(t, dataName)
	defer testKillDataCont(t, dataName)

	do := definitions.NowDo()
	do.Name = dataName
	do.Operations.Args = []string{"mv", "/home/eris/.eris/test", "/home/eris/.eris/tset"}
	do.Operations.Interactive = false
	do.Operations.ContainerNumber = 1

	log.WithFields(log.Fields{
		"data container": do.Name,
		"args":           do.Operations.Args,
	}).Info("Executing data (from tests)")
	if err := ExecData(do); err != nil {
		log.Error(err)
		t.Fail()
	}

	//TODO check that the file was actually moved! (TestExport _used_ todo that)
}

func TestRenameData(t *testing.T) {
	testCreateDataByImport(t, dataName)
	defer testKillDataCont(t, dataName)

	testExist(t, dataName, true)
	testExist(t, newName, false)

	do := definitions.NowDo()
	do.Name = dataName
	do.NewName = newName
	do.Operations.ContainerNumber = 1
	log.WithFields(log.Fields{
		"from": do.Name,
		"to":   do.NewName,
	}).Info("Renaming data (from tests)")
	if err := RenameData(do); err != nil {
		log.Error(err)
		t.FailNow()
	}

	testExist(t, dataName, false)
	testExist(t, newName, true)

	do = definitions.NowDo()
	do.Name = newName
	do.NewName = dataName
	do.Operations.ContainerNumber = 1
	log.WithFields(log.Fields{
		"from": do.Name,
		"to":   do.NewName,
	}).Info("Renaming data (from tests)")
	if err := RenameData(do); err != nil {
		log.Error(err)
		t.FailNow()
	}

	testExist(t, dataName, true)
	testExist(t, newName, false)
}

func TestInspectData(t *testing.T) {
	testCreateDataByImport(t, dataName)
	defer testKillDataCont(t, dataName)

	do := definitions.NowDo()
	do.Name = dataName
	do.Operations.Args = []string{"name"}
	do.Operations.ContainerNumber = 1
	log.WithFields(log.Fields{
		"data container": do.Name,
		"args":           do.Operations.Args,
	}).Info("Inspecting data (from tests)")
	if err := InspectData(do); err != nil {
		log.Error(err)
		t.FailNow()
	}

	do = definitions.NowDo()
	do.Name = dataName
	do.Operations.Args = []string{"config.network_disabled"}
	do.Operations.ContainerNumber = 1
	log.WithFields(log.Fields{
		"data container": do.Name,
		"args":           do.Operations.Args,
	}).Info("Inspecting data (from tests)")
	if err := InspectData(do); err != nil {
		log.Error(err)
		t.Fail()
	}
}

//now is testKillDataCont()
func TestRmData(t *testing.T) {
	testCreateDataByImport(t, dataName)
	testExist(t, dataName, true)

	testKillDataCont(t, dataName)
	testExist(t, dataName, false)
}

//creates a new data container w/ dir to be used by a test
//maybe give create opts? => paths, files, file contents, etc
func testCreateDataByImport(t *testing.T, name string) {
	newDataDir := filepath.Join(common.DataContainersPath, name)
	if err := os.MkdirAll(newDataDir, 0777); err != nil {
		log.Error(err)
		t.FailNow()
		os.Exit(1)
	}

	f, err := os.Create(filepath.Join(newDataDir, "test"))
	if err != nil {
		log.Error(err)
		t.FailNow()
		os.Exit(1)
	}
	defer f.Close()

	do := definitions.NowDo()
	do.Name = name
	do.Source = filepath.Join(common.DataContainersPath, do.Name)
	do.Destination = common.ErisContainerRoot
	do.Operations.ContainerNumber = 1
	log.WithField("=>", do.Name).Info("Importing data (from tests)")
	if err := ImportData(do); err != nil {
		log.Error(err)
		t.Fail()
	}

	testExist(t, name, true)
}

func testKillDataCont(t *testing.T, name string) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		log.Warn("Testing in Circle. Where we don't have rm privileges. Skipping test")
		return
	}

	testCreateDataByImport(t, name)
	testExist(t, name, true)

	do := definitions.NowDo()
	do.Name = name
	do.Operations.ContainerNumber = 1
	if err := RmData(do); err != nil {
		log.Error(err)
		t.Fail()
	}

	testExist(t, name, false)
}

func testExist(t *testing.T, name string, toExist bool) {
	if err := tests.TestExistAndRun(name, "data", 1, toExist, false); err != nil {
		t.Fail()
	}
}
