package initialize

import (
	"fmt"
	"io/ioutil"
	"os"
	//"os/exec"
	"path"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	//def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

//XXX can't dry because testutils imports this package

var erisDir = path.Join(os.TempDir(), "eris")
var servDir = path.Join(erisDir, "services")
var actDir = path.Join(erisDir, "actions")
var chnDir = path.Join(erisDir, "chains")
var chnDefDir = path.Join(chnDir, "default")

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	logLevel = 0
	// logLevel = 1
	// logLevel = 3

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	ifExit(testsInit())

	exitCode := m.Run()

	logger.Infoln("Commensing with Tests Tear Down.")
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		ifExit(testsTearDown())
	}

	os.Exit(exitCode)
}

func TestInitErisRootDir(t *testing.T) {
	//will this overwrite my current eris ?
	//thats the point of the next test/flag
	_, err := checkThenInitErisRoot()
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

//TestDropService/Action/ChainDefaults are basically just tests
//that the toadserver is up and running & that the files there
//match the definition files in each eris-service/actions/chains
func TestDropServiceDefaults(t *testing.T) {
	if err := testDrops(servDir, "services"); err != nil {
		ifExit(fmt.Errorf("error dropping services: %v\n", err))
	}
}

func TestDropActionDefaults(t *testing.T) {
	if err := testDrops(actDir, "actions"); err != nil {
		ifExit(fmt.Errorf("error dropping actions: %v\n", err))
	}
}

func TestDropChainDefaults(t *testing.T) {
	if err := testDrops(chnDir, "chains"); err != nil {
		ifExit(fmt.Errorf("errors dropping chains: %v\n", err))
	}
}

func testDrops(dir, kind string) error {
	var dirToad = path.Join(dir, "toad")
	var dirGit = path.Join(dir, "git")

	if err := os.MkdirAll(dirToad, 0777); err != nil {
		ifExit(err)
	}

	if err := os.MkdirAll(dirGit, 0777); err != nil {
		ifExit(err)
	}
	switch kind {
	case "services":
		//pull from toadserver
		if err := dropServiceDefaults(dirToad, "toadserver"); err != nil {
			ifExit(err)
		}
		//pull from rawgit
		if err := dropServiceDefaults(dirGit, "rawgit"); err != nil {
			ifExit(err)
		}
	case "actions":
		if err := dropActionDefaults(dirToad, "toadserver"); err != nil {
			ifExit(err)
		}
		if err := dropActionDefaults(dirGit, "rawgit"); err != nil {
			ifExit(err)
		}

	case "chains":
		if err := dropChainDefaults(dirToad, "toadserver"); err != nil {
			ifExit(err)
		}
		if err := dropChainDefaults(dirGit, "rawgit"); err != nil {
			ifExit(err)
		}
	}
	//read dirs
	toads, err := ioutil.ReadDir(dirToad)
	if err != nil {
		ifExit(err)
	}
	gits, err := ioutil.ReadDir(dirGit)
	if err != nil {
		ifExit(err)
	}

	for _, toad := range toads {
		for _, git := range gits {
			if toad.Name() == git.Name() {
				tsFile := path.Join(dirToad, toad.Name())
				gitFile := path.Join(dirGit, git.Name())
				//read and compare files
				if err := testsCompareFiles(tsFile, gitFile); err != nil {
					ifExit(fmt.Errorf("error comparing files: %v\n", err))
				}
			}
		}
	}
	return nil
}

func testsInit() error {
	var err error
	config.GlobalConfig, err = config.SetGlobalObject(os.Stdout, os.Stderr)
	if err != nil {
		ifExit(fmt.Errorf("TRAGIC. Could not set global config.\n"))
	}

	// common is initialized on import so
	// we have to manually override these
	// variables to ensure that the tests
	// run correctly.
	config.ChangeErisDir(erisDir)

	util.DockerConnect(false, "eris")

	logger.Infoln("Test init completed. Starting main test sequence now.")
	return nil

}

func testsCompareFiles(path1, path2 string) error {
	file1, err := ioutil.ReadFile(path1)
	if err != nil {
		return err
	}

	file2, err := ioutil.ReadFile(path2)
	if err != nil {
		return err
	}

	if string(file1) != string(file2) {
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
		logger.Errorln(err)
		if err := testsTearDown(); err != nil {
			logger.Errorln(err)
		}
		log.Flush()
		os.Exit(1)
	}
}
