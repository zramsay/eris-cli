package actions

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/log"
	"github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/util"
)

var erisDir string = path.Join(os.TempDir(), "eris")
var actionName string = "do not use"
var oldName string = "wanna do some testing"
var newName string = "yeah lets test shit"
var hash string

func TestMain(m *testing.M) {
	var logLevel int

	if os.Getenv("LOG_LEVEL") != "" {
		logLevel, _ = strconv.Atoi(os.Getenv("LOG_LEVEL"))
	} else {
		logLevel = 0
		// logLevel = 1
		// logLevel = 2
	}
	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	ifExit(testsInit())

	exitCode := m.Run()

	logger.Infoln("Commensing with Tests Tear Down.")
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		ifExit(testsTearDown())
	}

	os.Exit(exitCode)
}

func TestListActions(t *testing.T) {
	do := definitions.NowDo()
	ifExit(ListKnown(do))
	k := strings.Split(do.Result, "\n") // tests output formatting.

	if len(k) != 1 {
		ifExit(fmt.Errorf("The wrong number of action definitions have been found. Something is wrong.\n"))
	}

	if k[0] != "do_not_use" {
		ifExit(fmt.Errorf("Could not find \"do not use\" action definition.\n"))
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
	do.Args = strings.Fields(actionName)
	do.Quiet = true
	logger.Infof("Perform Action (from tests) =>\t%v\n", do.Args)
	if err := Do(do); err != nil {
		logger.Errorln(err)
		t.Fail()
	}
}

func TestNewAction(t *testing.T) {
	do := definitions.NowDo()
	do.Args = strings.Fields(oldName)
	logger.Infof("New Action (from tests) =>\t%v\n", do.Args)
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
	do.Args = strings.Fields(oldName)
	do.File = true
	if err := RmAction(do); err != nil {
		logger.Errorln(err)
		t.Fail()
	}
	testExist(t, oldName, false)
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
	util.DockerConnect(false)

	// this dumps the ipfs service def into the temp dir which
	// has been set as the erisRoot
	ifExit(ini.Initialize(false, true, false, false))

	return nil
}

func testsTearDown() error {
	if e := os.RemoveAll(erisDir); e != nil {
		return e
	}

	return nil
}

func testExist(t *testing.T, name string, toExist bool) {
	var exist bool
	name = strings.Replace(name, " ", "_", -1) // dirty
	logger.Infof("\nTesting whether (%s) existing? (%t)\n", name, toExist)
	name = util.DataContainersName(name, 1)

	do := definitions.NowDo()
	do.Quiet = true
	if err := ListKnown(do); err != nil {
		logger.Errorln(err)
		t.Fail()
	}
	res := strings.Split(do.Result, "\n")
	for _, r := range res {
		logger.Debugf("Existing =>\t\t\t%s\n", r)
		if r == util.ContainersShortName(name) {
			exist = true
		}
	}

	if toExist != exist {
		if toExist {
			logger.Infof("Could not find an existing =>\t%s\n", name)
		} else {
			logger.Infof("Found an existing instance of %s when I shouldn't have\n", name)
		}
		t.Fail()
	}

	logger.Infoln("")
}

func ifExit(err error) {
	if err != nil {
		logger.Errorln(err)
		log.Flush()
		testsTearDown()
		os.Exit(1)
	}
}
