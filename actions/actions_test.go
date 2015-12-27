package actions

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/logger"
	tests "github.com/eris-ltd/eris-cli/testutils"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

var actionName string = "do not use"
var oldName string = "wanna do some testing"
var newName string = "yeah lets test shit"
var hash string

func TestMain(m *testing.M) {
	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(testsInit())

	exitCode := m.Run()

	log.Info("Tearing tests down")
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
		log.Errorf("Error: action did not load properly: %v", e)
		t.FailNow()
	}

	actionName = strings.Replace(actionName, "_", " ", -1)
	if act.Name != actionName {
		log.Errorf("Error: improper action name on LOAD. expected: %s got: %s", actionName, act.Name)
		t.Fail()
	}
}

func TestDoAction(t *testing.T) {
	do := definitions.NowDo()
	do.Operations.Args = strings.Fields(actionName)
	do.Quiet = true
	log.WithField("args", do.Operations.Args).Info("Performing action (from tests)")
	if err := Do(do); err != nil {
		log.Error(err)
		t.Fail()
	}
}

func TestNewAction(t *testing.T) {
	do := definitions.NowDo()
	do.Operations.Args = strings.Fields(oldName)
	log.WithField("args", do.Operations.Args).Info("New action (from tests)")
	if err := NewAction(do); err != nil {
		log.Error(err)
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
	log.WithFields(log.Fields{
		"from": do.Name,
		"to":   do.NewName,
	}).Info("Renaming action (from tests)")
	if err := RenameAction(do); err != nil {
		log.Error(err)
		t.Fail()
	}
	testExist(t, newName, true)
	testExist(t, oldName, false)

	do = definitions.NowDo()
	do.Name = newName
	do.NewName = oldName
	log.WithFields(log.Fields{
		"from": do.Name,
		"to":   do.NewName,
	}).Info("Renaming action (from tests)")
	if err := RenameAction(do); err != nil {
		log.Error(err)
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
		log.Error(err)
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
