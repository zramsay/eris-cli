package contracts

import (
	"os"
	"testing"

	tests "github.com/eris-ltd/eris-cli/testings"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	logLevel = 0
	// logLevel = 1
	// logLevel = 2

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	tests.IfExit(testsInit())

	exitCode := m.Run()

	logger.Infoln("Commensing with Tests Tear Down.")
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		tests.IfExit(tests.TestsTearDown())
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
	if err := tests.TestsInit("contracts"); err != nil {
		return err
	}
	return nil
}
