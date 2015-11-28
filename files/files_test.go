package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/services"
	tests "github.com/eris-ltd/eris-cli/testutils"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
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
		tests.TestsTearDown()
		DEAD = true
		panic(err)
	}
}

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	logLevel = 0
	// logLevel = 1
	// logLevel = 3

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		erisDir = os.Getenv("HOME")
	}

	file = path.Join(erisDir, "temp")

	tests.IfExit(testsInit())
	exitCode := m.Run()

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		testKillIPFS(nil)
		tests.IfExit(tests.TestsTearDown())
	}

	os.Exit(exitCode)
}

func TestPutFiles(t *testing.T) {
	do := definitions.NowDo()
	do.Name = file
	logger.Infof("Putting File =>\t\t\t%s\n", do.Name)

	// because IPFS is testy, we retry the put up to
	// 10 times.
	passed := false
	for i := 0; i < 9; i++ {
		if err := testPutFiles(do); err != nil {
			time.Sleep(3 * time.Second)
			continue
		} else {
			passed = true
			break
		}
	}
	if !passed {
		// final time will throw
		if err := testPutFiles(do); err != nil {
			fatal(t, err)
		}
	}

	hash = do.Result
	logger.Debugf("My Result =>\t\t\t%s\n", do.Result)
}

func TestGetFiles(t *testing.T) {
	fileName := strings.Replace(file, "temp", "pmet", 1)
	do := definitions.NowDo()
	do.Name = hash
	do.Path = fileName
	// because IPFS is testy, we retry the put up to
	// 10 times.
	passed := false
	for i := 0; i < 9; i++ {
		if err := testGetFiles(do); err != nil {
			time.Sleep(3 * time.Second)
			continue
		} else {
			passed = true
			break
		}
	}
	if !passed {
		// final time will throw
		if err := testGetFiles(do); err != nil {
			fatal(t, err)
		}
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
	if err := tests.TestsInit("files"); err != nil {
		return err
	}

	f, err := os.Create(file)
	tests.IfExit(err)
	f.Write([]byte(content))

	do1 := definitions.NowDo()
	do1.Operations.Args = []string{"ipfs"}
	err = services.StartService(do1)
	tests.IfExit(err)
	time.Sleep(5 * time.Second) // boot time

	return nil
}

func testKillIPFS(t *testing.T) {
	serviceName := "ipfs"
	logger.Debugf("Stopping serv (via tests) =>\t%s\n", serviceName)

	do := definitions.NowDo()
	do.Name = serviceName
	do.Operations.Args = []string{serviceName}
	do.Rm = true
	do.RmD = true
	if e := services.KillService(do); e != nil {
		t.Fatal(e)
	}
}

func testPutFiles(do *definitions.Do) error {
	if err := PutFiles(do); err != nil {
		return err
	}
	return nil
}

func testGetFiles(do *definitions.Do) error {
	if err := GetFiles(do); err != nil {
		return err
	}
	return nil
}
