package actions

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/definitions"
	tests "github.com/eris-ltd/eris-cli/testutils"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

var actionName string = "do not use"
var oldName string = "wanna do some testing"
var newName string = "yeah lets test shit"
var hash string

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	logLevel = 0
	// logLevel = 1
	// logLevel = 2
	// logLevel = 3

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	tests.IfExit(testsInit())

	exitCode := m.Run()

	logger.Infoln("Commensing with Tests Tear Down.")
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		tests.IfExit(tests.TestsTearDown())
	}

	os.Exit(exitCode)
}

func TestListActions(t *testing.T) {
	do := definitions.NowDo()
	do.Known = true
	do.Running = false
	do.Existing = false
	do.Operations.Args = []string{"testing"}
	tests.IfExit(util.ListAll(do, "actions"))
	k := strings.Split(do.Result, "\n") // tests output formatting.

	if len(k) != 1 {
		tests.IfExit(fmt.Errorf("The wrong number of action definitions have been found. Something is wrong.\n"))
	}

	if k[0] != "do_not_use" {
		tests.IfExit(fmt.Errorf("Could not find \"do not use\" action definition.\n"))
	}
}

func TestLoadActionDefinition(t *testing.T) {
	var e error
	actionName = strings.Replace(actionName, " ", "_", -1)
	act, _, e := LoadActionDefinition(actionName)
	if e != nil {
		logger.Errorf("Action did not load properly =>\t%v\n", e)
		t.FailNow()
	}

	actionName = strings.Replace(actionName, "_", " ", -1)
	if act.Name != actionName {
		logger.Errorf("FAILURE: improper action name on LOAD. expected: %s\tgot: %s\n", actionName, act.Name)
		t.Fail()
	}
}

func TestDoAction(t *testing.T) {
	do := definitions.NowDo()
	do.Operations.Args = strings.Fields(actionName)
	do.Quiet = true
	logger.Infof("Perform Action (from tests) =>\t%v\n", do.Operations.Args)
	if err := Do(do); err != nil {
		logger.Errorln(err)
		t.Fail()
	}
}

func TestNewAction(t *testing.T) {
	do := definitions.NowDo()
	do.Operations.Args = strings.Fields(oldName)
	logger.Infof("New Action (from tests) =>\t%v\n", do.Operations.Args)
	if err := NewAction(do); err != nil {
		logger.Errorln(err)
		t.Fail()
	}
	testExist(t, oldName, true)
}

func TestRenameAction(t *testing.T) {
	testExist(t, newName, false)
	testExist(t, oldName, true)

	do := definitions.NowDo()
	do.Name = oldName
	do.NewName = newName
	logger.Infof("Renaming Action (from tests) =>\t%s:%s\n", do.Name, do.NewName)
	if err := RenameAction(do); err != nil {
		logger.Errorln(err)
		t.Fail()
	}
	testExist(t, newName, true)
	testExist(t, oldName, false)

	do = definitions.NowDo()
	do.Name = newName
	do.NewName = oldName
	logger.Infof("Renaming Action (from tests) =>\t%s:%s\n", do.Name, do.NewName)
	if err := RenameAction(do); err != nil {
		logger.Errorln(err)
		t.Fail()
	}
	testExist(t, newName, false)
	testExist(t, oldName, true)
}

func TestRemoveAction(t *testing.T) {
	do := definitions.NowDo()
	do.Operations.Args = strings.Fields(oldName)
	do.File = true
	if err := RmAction(do); err != nil {
		logger.Errorln(err)
		t.Fail()
	}
	testExist(t, oldName, false)
}

func testsInit() error {
	if err := tests.TestsInit("actions"); err != nil {
		return err
	}
	return nil
}

func testExist(t *testing.T, name string, toExist bool) {
	if tests.TestExistAndRun(name, "actions", 1, toExist, false) {
		t.Fail()
	}
}
