package data

import (
	"fmt"
	"os"
	"path"
	// "strings"
	"testing"

	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
)

var erisDir string = path.Join(os.TempDir(), "eris")
var dataName string = "dataTest1"
var newName string = "dataTest2"

func TestMain(m *testing.M) {
	logger.Level = 0
	// logger.Level = 1
	// logger.Level = 2

	if err := testsInit(); err != nil {
		logger.Errorln(err)
		os.Exit(1)
	}

	exitCode := m.Run()

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		if err := testsTearDown(); err != nil {
			logger.Errorln(err)
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

	if err := ImportDataRaw(dataName, 1); err != nil {
		logger.Errorln(err)
		t.Fail()
	}

	testExist(t, dataName, true)
}

func TestRenameDataRaw(t *testing.T) {
	testExist(t, dataName, true)
	testExist(t, newName, false)

	if err := RenameDataRaw(dataName, newName, 1); err != nil {
		logger.Errorln(err)
		t.FailNow()
	}

	testExist(t, dataName, false)
	testExist(t, newName, true)

	if err := RenameDataRaw(newName, dataName, 1); err != nil {
		logger.Errorln(err)
		t.FailNow()
	}

	testExist(t, dataName, true)
	testExist(t, newName, false)
}

func TestInspectDataRaw(t *testing.T) {
	if err := InspectDataRaw(dataName, "name", 1); err != nil {
		logger.Errorln(err)
		t.FailNow()
	}

	if err := InspectDataRaw(dataName, "config.network_disabled", 1); err != nil {
		logger.Errorln(err)
		t.Fail()
	}
}

func TestExecDataRaw(t *testing.T) {
	args := []string{"mv", "/home/eris/.eris/test", "/home/eris/.eris/tset"}
	if err := ExecDataRaw(dataName, 1, false, args); err != nil {
		logger.Errorln(err)
		t.Fail()
	}
}

func TestExportDataRaw(t *testing.T) {
	if err := ExportDataRaw(dataName, 1); err != nil {
		logger.Errorln(err)
		t.FailNow()
	}

	if _, err := os.Stat(path.Join(common.DataContainersPath, dataName, "tset")); os.IsNotExist(err) {
		logger.Errorf("Tragic! Exported file does not exist: %s\n", err)
		t.Fail()
	}
}

func TestRmDataRaw(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	if err := RmDataRaw(dataName, 1); err != nil {
		logger.Errorln(err)
		t.Fail()
	}
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
		return fmt.Errorf("TRAGIC. Could not initialize the eris dir:\n%s\n", err)
	}

	// init dockerClient
	util.DockerConnect(false)

	return nil
}

func testsTearDown() error {
	// if e := os.RemoveAll(erisDir); e != nil {
	// 	return e
	// }

	return nil
}

func testExist(t *testing.T, name string, toExist bool) {
	var exist bool
	known, _ := ListKnownRaw()
	for _, r := range known {
		if r == name {
			exist = true
		}
	}

	if toExist != exist {
		if toExist {
			logger.Errorf("Could not find an existing instance of %s\n", name)
		} else {
			logger.Errorf("Found an existing instance of %s when I shouldn't have\n", name)
		}
		t.Fail()
	}
}
