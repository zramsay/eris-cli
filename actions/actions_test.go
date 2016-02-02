package actions

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/list"
	"github.com/eris-ltd/eris-cli/tests"
	ver "github.com/eris-ltd/eris-cli/version"

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

func TestListKnownActions(t *testing.T) {
	do := definitions.NowDo()
	do.Quiet = true
	tests.IfExit(list.ListActions(do))
	k := strings.Split(do.Result, "\n") // tests output formatting.

	ver.ACTION_DEFINITIONS = append(ver.ACTION_DEFINITIONS, "do_not_use.toml")

	if len(k) != len(ver.ACTION_DEFINITIONS) {
		tests.IfExit(fmt.Errorf("Did not find correct number of action definitions files, Expected %v, found %v.\n", len(ver.ACTION_DEFINITIONS), len(k)))
	}

	actDefs := make(map[string]bool)

	for _, act := range ver.ACTION_DEFINITIONS {
		actDef := strings.Split(act, ".")
		actDefs[actDef[0]] = true
	}

	i := 0
	for _, actFile := range k {
		if actDefs[actFile] == true {
			i++
		}
	}

	if i != len(ver.ACTION_DEFINITIONS) {
		tests.IfExit(fmt.Errorf("Could not find all the expected action definition files.\n"))
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

func testExist(t *testing.T, name string, toExist bool) {
	if existing := tests.TestActionDefinitionFile(name); existing != toExist {
		t.Errorf("expected action definition file %q with status %v, got %v", name, toExist, existing)
	}
}
