package keys

import (
	"testing"

	"github.com/eris-ltd/eris-cli/definitions"
	tests "github.com/eris-ltd/eris-cli/testutils"
)

//vars

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	logLevel = 0
	// logLevel = 1
	//	logLevel = 3

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	tests.IfExit(testsInit())

	exitCode := m.Run()

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		tests.IfExit(tests.TestsTearDown())
	}

	os.Exit(exitCode)
}

func testStartKeys() {
	//more stuff
	do.Name = "keys"
	srv.StartService(do)

}

func TestGenerateKey(t *testing.T) {
	testStartKeys()
	defer testKillKeys() //or some clean function
	//gen a key
	//export it to temp dir
	//parse the filepath,
	//do something with that addr...

}

func TestGetPubKey(t *testing.T) {
	//gen a key
	//

}

func testsInit() error {
	if err := tests.TestsInit("keys"); err != nil {
		return err
	}
	return nil
}
