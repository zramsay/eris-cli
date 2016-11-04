package data

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/testutil"
)

var dataName string = "dataTest1"
var newName string = "dataTest2"

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	testutil.IfExit(testutil.Init(testutil.Pull{
		Images: []string{"data"},
	}))

	exitCode := m.Run()
	testutil.IfExit(testutil.TearDown())
	os.Exit(exitCode)
}

// TODO add some export/import test robustness
func TestImportDataRawNoPriorExist(t *testing.T) {
	testCreateDataByImport(t, dataName)
	defer testKillDataCont(t, dataName)
}

func TestExportData(t *testing.T) {
	testCreateDataByImport(t, dataName)
	defer testKillDataCont(t, dataName)

	do := definitions.NowDo()
	do.Name = dataName
	do.Source = config.ErisContainerRoot
	do.Destination = filepath.Join(config.DataContainersPath, do.Name)
	if err := ExportData(do); err != nil {
		log.Error(err)
		t.FailNow()
	}

	if _, err := os.Stat(filepath.Join(config.DataContainersPath, dataName, "test")); os.IsNotExist(err) {
		log.Errorf("Exported file does not exist: %s", err)
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
	newDataDir := filepath.Join(config.DataContainersPath, name)
	if err := os.MkdirAll(newDataDir, 0777); err != nil {
		t.Fatalf("err mkdir: %v", err)
	}

	f, err := os.Create(filepath.Join(newDataDir, "test"))
	if err != nil {
		t.Fatalf("err creating file: %v", err)
	}
	defer f.Close()

	do := definitions.NowDo()
	do.Name = name
	do.Source = filepath.Join(config.DataContainersPath, do.Name)
	do.Destination = config.ErisContainerRoot
	if err := ImportData(do); err != nil {
		t.Fatalf("error importing data: %v", err)
	}

	testExist(t, name, true)
}

func testKillDataCont(t *testing.T, name string) {
	testCreateDataByImport(t, name)
	testExist(t, name, true)

	do := definitions.NowDo()
	do.Name = name
	if err := RmData(do); err != nil {
		t.Fatalf("error rm data: %v", err)
	}

	testExist(t, name, false)
}

func testExist(t *testing.T, name string, toExist bool) {
	if err := testutil.ExistAndRun(name, "data", toExist, false); err != nil {
		t.Fail()
	}
}
