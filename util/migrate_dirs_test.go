package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/eris-ltd/eris-cli/config"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

var erisDir string = path.Join(os.TempDir(), "eris")

//old
var blockchains string = path.Join(erisDir, "blockchains")
var dapps string = path.Join(erisDir, "dapps")
var depDirs = []string{blockchains, dapps}

//new
var chains string = path.Join(erisDir, "chains")
var apps string = path.Join(erisDir, "apps")
var newDirs = []string{chains, apps}

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	logLevel = 0
	// logLevel = 1
	// logLevel = 3

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

func TestMigrationSimple(t *testing.T) {

	testsSetupDirs(depDirs, newDirs, true, false)
	defer testsRemoveDirs(depDirs, newDirs)

	dirsToMigrate := make(map[string]string, len(depDirs))

	for n, d := range depDirs {
		dirsToMigrate[d] = newDirs[n]
	}

	//migrate them
	if err := MigrateDeprecatedDirs(dirsToMigrate, false); err != nil { //false = don't prompt
		ifExit(err) //but some errors are ok ?
	}

	//check the deprecated dirs no longer exist
	for _, depDir := range depDirs {
		if DoesDirExist(depDir) {
			ifExit(fmt.Errorf("something went wrong, deprecated directory (%s) still exists", depDir))
		}

	}
	//check that the new dirs do exist
	for _, newDir := range newDirs {
		if !DoesDirExist(newDir) {
			ifExit(fmt.Errorf("something went wrong, new directory (%s) does not exist", newDir))
		}

	}
}

func TestMigrationMoveFile(t *testing.T) {
	//------------- both dirs exists, deal with accordingly -------------
	// put file in depDir, see if it moved; ensure depDir is removed

	testsSetupDirs(depDirs, newDirs, true, true)
	defer testsRemoveDirs(depDirs, newDirs)

	testFile := "migration_test"
	testDepFile := path.Join(depDirs[0], testFile)

	testDepContent := []byte("some datas")

	if err := ioutil.WriteFile(testDepFile, testDepContent, 0777); err != nil {
		ifExit(err)
	}

	dirsToMigrate := make(map[string]string, len(depDirs))

	for n, d := range depDirs {
		dirsToMigrate[d] = newDirs[n]
	}

	if err := MigrateDeprecatedDirs(dirsToMigrate, false); err != nil { //false = don't prompt
		ifExit(err)
	}

	testNewFile := path.Join(newDirs[0], testFile)
	testNewContent, err := ioutil.ReadFile(testNewFile)
	if err != nil {
		ifExit(err)
	}

	if string(testDepContent) != string(testNewContent) {
		ifExit(fmt.Errorf("something went wrong: depDir (%s) content (%s) is not identical to newDir (%s) content (%s)\n", testDepFile, testDepContent, testNewFile, testNewContent))
	}
}

func testsInit() error {
	var err error
	// TODO: make a reader/pipe so we can see what is written from tests.
	config.GlobalConfig, err = config.SetGlobalObject(os.Stdout, os.Stderr)
	ifExit(err)

	if err := os.Mkdir(erisDir, 0777); err != nil {
		ifExit(err)
	}
	config.ChangeErisDir(erisDir)

	return nil
}

func testsSetupDirs(depDirs, newDirs []string, makeDep, makeNew bool) {
	//should make what needs to be made based on flag logic
	if makeDep {
		//make dirs to be deprecated
		for _, depDir := range depDirs {
			if err := os.Mkdir(depDir, 0777); err != nil {
				ifExit(err)
			}
		}
	}

	if makeNew {
		for _, newDir := range newDirs {
			if err := os.Mkdir(newDir, 0777); err != nil {
				ifExit(err)
			}
		}
	}
}

func testsRemoveDirs(depDirs, newDirs []string) {
	for _, depDir := range depDirs {
		if err := os.RemoveAll(depDir); err != nil {
			ifExit(err)
		}
	}
	for _, newDir := range newDirs {
		if err := os.RemoveAll(newDir); err != nil {
			ifExit(err)
		}
	}
}

func testsTearDown() error {
	if e := os.RemoveAll(erisDir); e != nil {
		return e
	}

	return nil
}

func ifExit(err error) {
	if err != nil {
		logger.Errorln(err)
		log.Flush()
		testsTearDown()
		os.Exit(1)
	}
}
