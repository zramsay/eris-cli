package initialize

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/ipfs"
	logger "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

var erisDir = filepath.Join(os.TempDir(), "eris")
var servDir = filepath.Join(erisDir, "services")
var actDir = filepath.Join(erisDir, "actions")
var chnDir = filepath.Join(erisDir, "chains")
var chnDefDir = filepath.Join(chnDir, "default")
var toadUp bool

func TestMain(m *testing.M) {

	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	ifExit(testsInit())

	toadUp = toadServerUp()

	exitCode := m.Run()
	log.Info("Commensing with Tests Tear Down.")
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
	var dirToad = filepath.Join(dir, "toad")
	var dirGit = filepath.Join(dir, "git")

	if err := os.MkdirAll(dirToad, 0777); err != nil {
		ifExit(err)
	}

	if err := os.MkdirAll(dirGit, 0777); err != nil {
		ifExit(err)
	}
	switch kind {
	case "services":
		//pull from toadserver
		if toadUp {
			if err := dropServiceDefaults(dirToad, "toadserver"); err != nil {
				ifExit(err)
			}
		}
		//pull from rawgit
		if err := dropServiceDefaults(dirGit, "rawgit"); err != nil {
			ifExit(err)
		}
	case "actions":
		if toadUp {
			if err := dropActionDefaults(dirToad, "toadserver"); err != nil {
				ifExit(err)
			}
		}
		if err := dropActionDefaults(dirGit, "rawgit"); err != nil {
			ifExit(err)
		}

	case "chains":
		if toadUp {
			if err := dropChainDefaults(dirToad, "toadserver"); err != nil {
				ifExit(err)
			}
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

	if toadUp {
		for _, toad := range toads {
			for _, git := range gits {
				if toad.Name() == git.Name() {
					tsFile := filepath.Join(dirToad, toad.Name())
					gitFile := filepath.Join(dirGit, git.Name())
					//read and compare files
					if err := testsCompareFiles(tsFile, gitFile); err != nil {
						ifExit(fmt.Errorf("error comparing files: %v\n", err))
					}
				}
			}
		}
	}
	return nil
}

func toadServerUp() bool {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	url := fmt.Sprintf("%s:11113/getfile/%s", ipfs.SexyUrl(), "keys.toml")
	_, err := client.Get(url)
	if err != nil {
		return false
	}
	return true
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
