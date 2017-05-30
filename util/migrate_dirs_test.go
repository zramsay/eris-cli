package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/monax/monax/config"
	"github.com/monax/monax/log"
)

var monaxDir string = filepath.Join(os.TempDir(), "monax")

//old
var blockchains string = filepath.Join(monaxDir, "blockchains")
var dapps string = filepath.Join(monaxDir, "dapps")
var depDirs = []string{blockchains, dapps}

//new
var chains string = filepath.Join(monaxDir, "chains")
var apps string = filepath.Join(monaxDir, "apps")
var newDirs = []string{chains, apps}

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	if err := testsInit(); err != nil {
		log.Fatal(err)
	}

	exitCode := m.Run()
	if err := testsTearDown(); err != nil {
		log.Fatal(err)
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
	testDepFile := filepath.Join(depDirs[0], testFile)

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

	testNewFile := filepath.Join(newDirs[0], testFile)
	testNewContent, err := ioutil.ReadFile(testNewFile)
	if err != nil {
		ifExit(err)
	}

	if string(testDepContent) != string(testNewContent) {
		ifExit(fmt.Errorf("something went wrong: depDir (%s) content (%s) is not identical to newDir (%s) content (%s)\n", testDepFile, testDepContent, testNewFile, testNewContent))
	}
}

func testsInit() error {
	config.ChangeMonaxRoot(monaxDir)

	// TODO: make a reader/pipe so we can see what is written from tests.
	var err error
	config.Global, err = config.New(os.Stdout, os.Stderr)
	if err != nil {
		ifExit(err)
	}

	if err := os.Mkdir(monaxDir, 0777); err != nil {
		if runtime.GOOS != "windows" {
			// windows returns an error here
			ifExit(err)
		}
	}

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
	return os.RemoveAll(monaxDir)
}

func ifExit(err error) {
	if err != nil {
		log.Error(err)
		testsTearDown()
		os.Exit(1)
	}
}
