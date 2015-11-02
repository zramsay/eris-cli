package contracts

import (
	// "fmt"
	// "io/ioutil"
	"os"
	"path"
	// "regexp"
	// "strings"
	"testing"

	def "github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	// "github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/util"

	// . "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

var erisDir string = path.Join(os.TempDir(), "eris")

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	logLevel = 0
	// logLevel = 1
	// logLevel = 2

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	ifExit(testsInit())

	exitCode := m.Run()

	logger.Infoln("Commensing with Tests Tear Down.")
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		ifExit(testsTearDown())
	}

	os.Exit(exitCode)
}

func TestContractsTest(t *testing.T) {

	// TODO: test more app types once we have
	// canonical apps + eth throwaway chains
}

func TestContractsDeploy(t *testing.T) {
	// TODO: finish. not worried about this too much now
	// since test will deploy.
}

func testsInit() error {
	var err error
	// TODO: make a reader/pipe so we can see what is written from tests.
	config.GlobalConfig, err = config.SetGlobalObject(os.Stdout, os.Stderr)
	ifExit(err)

	config.ChangeErisDir(erisDir)

	// init dockerClient
	util.DockerConnect(false, "eris")

	// clone bank...for now.
	// TODO: add better tester

	// this dumps the ipfs service def into the temp dir which
	// has been set as the erisRoot
	do := def.NowDo()
	do.Pull = true
	do.Services = true
	do.Actions = true
	ifExit(ini.Initialize(do))

	logger.Infoln("Test init completed. Starting main test sequence now.")
	return nil
}

func testsTearDown() error {
	return os.RemoveAll(erisDir)
	// return nil
}

func ifExit(err error) {
	if err != nil {
		logger.Errorln(err)
		log.Flush()
		testsTearDown()
		os.Exit(1)
	}
}
