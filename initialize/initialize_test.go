package initialize

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/util"
	ver "github.com/eris-ltd/eris-cli/version"

	"github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
)

var erisDir = filepath.Join(os.TempDir(), "eris")
var servDir = filepath.Join(erisDir, "services")
var chnDir = filepath.Join(erisDir, "chains")
var chnDefDir = filepath.Join(chnDir, "default")

// TODO [zr] remove toadserver references/tests

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	ifExit(testsInit())

	exitCode := m.Run()
	ifExit(testsTearDown())
	os.Exit(exitCode)
}

func TestInitErisRootDir(t *testing.T) {
	_, err := checkThenInitErisRoot(false)
	if err != nil {
		ifExit(err)
	}

	for _, dir := range common.MajorDirs {
		if !util.DoesDirExist(dir) {
			ifExit(fmt.Errorf("Could not find the %s subdirectory", dir))
		}
	}
}

func TestMigration(t *testing.T) {
	//already has its own test
}

func TestPullImages(t *testing.T) {
	//already tested by virtue of being needed for tool level tests
}

func TestDropServiceDefaults(t *testing.T) {
	if err := testDrops(servDir, "services"); err != nil {
		ifExit(fmt.Errorf("error dropping services: %v\n", err))
	}
}

func testDrops(dir, kind string) error {
	var dirGit = filepath.Join(dir, "git")

	if err := os.MkdirAll(dirGit, 0777); err != nil {
		ifExit(err)
	}

	switch kind {
	case "services":
		if err := dropServiceDefaults(dirGit, ver.SERVICE_DEFINITIONS); err != nil {
			ifExit(err)
		}
	}

	readDirs(dirGit)

	return nil
}

func readDirs(dirGit string) {
	_, err := ioutil.ReadDir(dirGit)
	if err != nil {
		ifExit(err)
	}
}

func testsInit() error {
	common.ChangeErisRoot(erisDir)

	var err error
	config.Global, err = config.New(os.Stdout, os.Stderr)
	if err != nil {
		ifExit(fmt.Errorf("TRAGIC. Could not set global config.\n"))
	}

	util.DockerConnect(false, "eris")

	log.Info("Test init completed. Starting main test sequence now")
	return nil

}

func testsCompareFiles(path1, path2 string) error {
	//skip dirs
	if util.DoesDirExist(path1) || util.DoesDirExist(path2) {
		return nil
	}
	file1, err := ioutil.ReadFile(path1)
	if err != nil {
		return err
	}

	file2, err := ioutil.ReadFile(path2)
	if err != nil {
		return err
	}

	if !bytes.Equal(file1, file2) {
		return fmt.Errorf("Error: Got %s\nExpected %s", string(file1), string(file1))
	}
	return nil
}

func testsTearDown() error {
	return os.RemoveAll(erisDir)
}

//copied from testutils
func ifExit(err error) {
	if err != nil {
		log.Error(err)
		if err := testsTearDown(); err != nil {
			log.Error(err)
		}
		os.Exit(1)
	}
}
