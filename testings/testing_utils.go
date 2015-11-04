package testings

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

//var DEAD bool // XXX: don't double panic (TODO: Flushing twice blocks)
var erisDir = path.Join(os.TempDir(), "eris")

var logger = log.AddLogger("tests")

//hold things...?
type TestingInfo struct {
}

//testType = one of each package, will switch over it for
//make additional tempDirs and vars as needed -> [zr] or not, TBD
func TestsInit(testType string) error {
	var err error
	// TODO: make a reader/pipe so we can see what is written from tests.
	config.GlobalConfig, err = config.SetGlobalObject(os.Stdout, os.Stderr)
	if err != nil {
		fmt.Errorf("TRAGIC. Could not set global config.\n")
		//ifExit(fmt.Errorf("TRAGIC. Could not set global config.\n"))
	}

	// common is initialized on import so
	// we have to manually override these
	// variables to ensure that the tests
	// run correctly.
	config.ChangeErisDir(erisDir)

	// init dockerClient
	if testType == "chain" {
		util.DockerConnect(false, "eris-test-nyc2-1.8.1") //hmm -> for local tests
	} else {
		util.DockerConnect(false, "eris")
	}

	// this dumps the ipfs service def into the temp dir which
	// has been set as the erisRoot
	do := def.NowDo()
	do.Pull = true
	do.Services = true
	do.Actions = true
	if err := ini.Initialize(do); err != nil {
		fmt.Errorf("TRAGIC. Could not initialize the eris dir.\n")
		//ifExit(fmt.Errorf("TRAGIC. Could not initialize the eris dir.\n"))
	}

	return nil
}

//return to handle failings in each pkg
//typ = type of test for dealing with do.() details
func TestExistAndRun(name, typ string, contNum int, toExist, toRun bool) bool {
	var exist, run bool
	if typ == "actions" {
		name = strings.Replace(name, " ", "_", -1) // dirty
	}

	//logger.Infof("\nTesting whether (%s) existing? (%t)\n", name, toExist)
	logger.Infof("\nTesting whether (%s) is running? (%t) and existing? (%t)\n", name, toRun, toExist)
	if typ == "chains" {
		name = util.ChainContainersName(name, 1) // not worried about containerNumbers, deal with multiple containers in services tests
	} else if typ == "services" {
		name = util.ServiceContainersName(name, contNum)

	} else {
		name = util.DataContainersName(name, 1)
	}
	do := def.NowDo()
	do.Quiet = true
	do.Operations.Args = []string{"testing"}

	if typ == "data" || typ == "chains" || typ == "services" {
		do.Existing = true
	} else if typ == "actions" {
		do.Known = true
	}
	if err := util.ListAll(do, typ); err != nil {
		logger.Errorln(err)
		return true
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
		return true
	}
	//func should always be testing for toExist, only sometimes tested for runining
	if typ == "chains" || typ == "services" {
		do.Running = true
		do.Existing = false //unset
		if err := util.ListAll(do, typ); err != nil {
			return true //	fatal(t, err)
		}
		logger.Debugln("RUNNING RESULT:", do.Result)
		res = strings.Split(do.Result, "\n")
		for _, r := range res {
			logger.Debugf("Running =>\t\t\t%s\n", r)
			if r == util.ContainersShortName(name) {
				run = true
			}
		}

		if toRun != run {
			if toRun {
				logger.Infof("Could not find a running =>\t%s\n", name)
			} else {
				logger.Infof("Found a running instance of %s when I shouldn't have\n", name)
			}
			return true //	fatal(t, nil)
		}
	}

	return false
}

// each pacakge will need its own custom shit if need be
// do it through a custom pre-process ifExit in each package that
// calls tests.IfExit()
func TestsTearDown() error {
	if err := os.RemoveAll(erisDir); err != nil {
		return err
	}
	return nil
}

func IfExit(err error) {
	if err != nil {
		logger.Errorln(err)
		log.Flush()
		if err := TestsTearDown(); err != nil {
			logger.Errorln(err)
		}
		os.Exit(1)
	}
}
