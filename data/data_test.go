package data

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/tests"

	"github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
)

var dataName string = "dataTest1"
var newName string = "dataTest2"

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit(tests.ConnectAndPull))

	exitCode := m.Run()
	tests.IfExit(tests.TestsTearDown())
	os.Exit(exitCode)
}

//TODO add some export/import test robustness
func TestImportDataRawNoPriorExist(t *testing.T) {
	testCreateDataByImport(t, dataName)
	defer testKillDataCont(t, dataName)
}

func TestExportData(t *testing.T) {
	testCreateDataByImport(t, dataName)
	defer testKillDataCont(t, dataName)

	do := definitions.NowDo()
	do.Name = dataName
	do.Source = common.ErisContainerRoot
	do.Destination = filepath.Join(common.DataContainersPath, do.Name)
	if err := ExportData(do); err != nil {
		log.Error(err)
		t.FailNow()
	}

	if _, err := os.Stat(filepath.Join(common.DataContainersPath, dataName, "test")); os.IsNotExist(err) {
		log.Errorf("Tragic! Exported file does not exist: %s", err)
		t.Fail()
	}

}

func TestExecData(t *testing.T) {
	testCreateDataByImport(t, dataName)
	defer testKillDataCont(t, dataName)

	do := definitions.NowDo()
	do.Name = dataName
	do.Operations.Args = []string{"mv", "/home/eris/.eris/test", "/home/eris/.eris/tset"}
	do.Operations.Interactive = false

	log.WithFields(log.Fields{
		"data container": do.Name,
		"args":           do.Operations.Args,
	}).Info("Executing data (from tests)")
	if _, err := ExecData(do); err != nil {
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
		t.Fatalf("err mkdir: %v\n", err)
	}

	f, err := os.Create(filepath.Join(newDataDir, "test"))
	if err != nil {
		t.Fatalf("err creating file: %v\n", err)
	}
	defer f.Close()

	do := definitions.NowDo()
	do.Name = name
	do.Source = filepath.Join(common.DataContainersPath, do.Name)
	do.Destination = common.ErisContainerRoot
	log.WithField("=>", do.Name).Info("Importing data (from tests)")
	if err := ImportData(do); err != nil {
		t.Fatalf("error importing data: %v\n", err)
	}

	testExist(t, name, true)
}

func testKillDataCont(t *testing.T, name string) {
	testCreateDataByImport(t, name)
	testExist(t, name, true)

	do := definitions.NowDo()
	do.Name = name
	if err := RmData(do); err != nil {
		t.Fatalf("error rm data: %v\n", err)
	}

	testExist(t, name, false)
}

func testExist(t *testing.T, name string, toExist bool) {
	if err := tests.TestExistAndRun(name, "data", toExist, false); err != nil {
		t.Fail()
	}
}
