package actions

import (
	"os"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/tests"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	logger "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
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

	tests.IfExit(tests.TestsInit("actions"))
	exitCode := m.Run()

	log.Info("Tearing tests down")
	tests.IfExit(tests.TestsTearDown())
	os.Exit(exitCode)
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

func testExist(t *testing.T, name string, toExist bool) {
	if existing := tests.TestActionDefinitionFile(name); existing != toExist {
		t.Errorf("expected action definition file %q with status %v, got %v", name, toExist, existing)
	}
}
