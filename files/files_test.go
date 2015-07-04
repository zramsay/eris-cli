package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/util"
)

var erisDir string = path.Join(os.TempDir(), "eris")
var file string
var content string = "test content\n"
var hash string

func TestMain(m *testing.M) {
	logger.Level = 0
	// logger.Level = 1
	// logger.Level = 2

	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		erisDir = os.Getenv("HOME")
	}

	file = path.Join(erisDir, "temp")

	if err := testsInit(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	exitCode := m.Run()

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		if err := testsTearDown(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	os.Exit(exitCode)
}

func TestPutFilesRaw(t *testing.T) {
	var err error
	hash, err = PutFilesRaw(file)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	logger.Debugln(hash)
}

func TestGetFilesRaw(t *testing.T) {
	fileName := strings.Replace(file, "temp", "pmet", 1)
	if err := GetFilesRaw(hash, fileName); err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	contentPuted, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	if string(contentPuted) != content {
		fmt.Printf("ERROR: Content Put into IPFS and Pulled out to not match.\nExpected:\t%s\nReceived:\t%s\n", content, string(contentPuted))
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

	// set ipfs endpoint
	os.Setenv("ERIS_IPFS_HOST", "http://0.0.0.0")

	// dump a test file with some stuff
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	f.Write([]byte(content))

	return nil
}

func testsTearDown() error {
	// if e := os.RemoveAll(erisDir); e != nil {
	// 	return e
	// }

	return nil
}
