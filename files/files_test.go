package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
	"github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"
)

var erisDir string = path.Join(os.TempDir(), "eris")
var file string
var content string = "test content\n"
var hash string

var DEAD bool // XXX: don't double panic (TODO: Flushing twice blocks)
func fatal(t *testing.T, err error) {
	if !DEAD {
		log.Flush()
		testKillIPFS(t)
		testsTearDown()
		DEAD = true
		panic(err)
	}
}

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	//	logLevel = 0
	// logLevel = 1
	logLevel = 3

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		erisDir = os.Getenv("HOME")
	}

	file = path.Join(erisDir, "temp")

	ifExit(testsInit())

	exitCode := m.Run()

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		testKillIPFS(nil)
		ifExit(testsTearDown())
	}

	os.Exit(exitCode)
}

func TestPutFiles(t *testing.T) {
	do := definitions.NowDo()
	do.Name = file
	logger.Infof("Putting File =>\t\t\t%s\n", do.Name)
	if err := PutFiles(do); err != nil {
		fatal(t, err)
	}
	hash = do.Result
	logger.Debugf("My Result =>\t\t\t%s\n", do.Result)
}

func TestGetFiles(t *testing.T) {
	fileName := strings.Replace(file, "temp", "pmet", 1)
	do := definitions.NowDo()
	do.Name = hash
	do.Path = fileName
	if err := GetFiles(do); err != nil {
		fatal(t, err)
	}

	f, err := os.Open(fileName)
	if err != nil {
		fatal(t, err)
	}

	contentPuted, err := ioutil.ReadAll(f)
	if err != nil {
		fatal(t, err)
	}

	if string(contentPuted) != content {
		fatal(t, fmt.Errorf("ERROR: Content Put into IPFS and Pulled out to not match.\nExpected:\t%s\nReceived:\t%s\n", content, string(contentPuted)))
	}
}

func testsInit() error {
	var err error
	// TODO: make a reader/pipe so we can see what is written from tests.
	util.GlobalConfig, err = util.SetGlobalObject(os.Stdout, os.Stderr)
	ifExit(err)

	// common is initialized on import so
	// we have to manually override these
	// variables to ensure that the tests
	// run correctly.
	util.ChangeErisDir(erisDir)

	// init dockerClient
	util.DockerConnect(false, "eris")

	// this dumps the ipfs service def into the temp dir which
	// has been set as the erisRoot
	ifExit(ini.Initialize(true, false))

	// set ipfs endpoint
	os.Setenv("ERIS_IPFS_HOST", "http://0.0.0.0")

	// dump a test file with some stuff
	f, err := os.Create(file)
	ifExit(err)
	f.Write([]byte(content))

	return nil
}

func testsTearDown() error {
	if e := os.RemoveAll(erisDir); e != nil {
		return e
	}

	return nil
}

func testKillIPFS(t *testing.T) {
	serviceName := "ipfs"
	logger.Debugf("Stopping serv (via tests) =>\t%s\n", serviceName)

	do := definitions.NowDo()
	do.Name = serviceName
	do.Args = []string{serviceName}
	do.Rm = true
	do.RmD = true
	if e := services.KillService(do); e != nil {
		t.Fatal(e)
	}
}

func ifExit(err error) {
	if err != nil {
		logger.Errorln(err)
		log.Flush()
		testsTearDown()
		os.Exit(1)
	}
}
