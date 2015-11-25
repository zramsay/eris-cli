package testings

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/util"
)

var erisDir = path.Join(os.TempDir(), "eris")
var logger = log.AddLogger("tests")

//hold things...?
type TestingInfo struct {
}

//testType = one of each package, will switch over it for
//make additional tempDirs and vars as needed -> [zr] or not, TBD
func TestsInit(testType string) (err error) {
	// TODO: make a reader/pipe so we can see what is written from tests.
	config.GlobalConfig, err = config.SetGlobalObject(os.Stdout, os.Stderr)
	if err != nil {
		IfExit(fmt.Errorf("TRAGIC. Could not set global config.\n"))
	}

	// common is initialized on import so
	// we have to manually override these
	// variables to ensure that the tests
	// run correctly.
	config.ChangeErisDir(erisDir)

	// init dockerClient (for chains use "eris-test-nyc2-1.8.1"?)
	util.DockerConnect(false, "eris")

	// this dumps the ipfs service def into the temp dir which
	// has been set as the erisRoot
	do := def.NowDo()
	do.Pull = true
	do.Services = true
	do.Actions = true
	do.Yes = true
	if err := ini.Initialize(do); err != nil {
		IfExit(fmt.Errorf("TRAGIC. Could not initialize the eris dir.\n"))
	}

	if testType == "services" {
		checkIPFSnotRunning() //TODO make more general & use for other things?
	}

	logger.Infoln("Test init completed. Starting main test sequence now.")
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
			return true
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
			return true
		}
	}

	return false
}

// Remove a container of some name, type, and number.
func RemoveContainer(name string, t string, n int) error {
	opts := docker.RemoveContainerOptions{
		ID:            util.ContainersName(t, name, n),
		RemoveVolumes: true,
		Force:         true,
	}

	if err := util.DockerClient.RemoveContainer(opts); err != nil {
		return err
	}

	return nil
}

// Remove everything Eris.
func RemoveAllContainers() error {
	return util.Clean(false, false, false, false)
}

// each pacakge will need its own custom stuff if need be
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
		if err := TestsTearDown(); err != nil {
			logger.Errorln(err)
		}
		log.Flush()
		os.Exit(1)
	}
}

//------- helpers --------
func checkIPFSnotRunning() {
	//os.Setenv("ERIS_IPFS_HOST", "http://0.0.0.0") //conflicts with docker-machine
	do := def.NowDo()
	do.Known = false
	do.Existing = false
	do.Running = true
	do.Quiet = true
	do.Operations.Args = []string{"testing"}
	logger.Debugln("Finding the running services.")
	if err := util.ListAll(do, "services"); err != nil {
		IfExit(err)
	}
	res := strings.Split(do.Result, "\n")
	for _, r := range res {
		if r == "ipfs" {
			IfExit(fmt.Errorf("IPFS service is running.\nPlease stop it with.\neris services stop -rx ipfs\n"))
		}
	}
	// make sure ipfs container does not exist
	do = def.NowDo()
	do.Known = false
	do.Existing = true
	do.Running = false
	do.Quiet = true
	do.Operations.Args = []string{"testing"}
	logger.Debugln("Finding the existing services.")
	if err := util.ListAll(do, "services"); err != nil {
		IfExit(err)
	}
	res = strings.Split(do.Result, "\n")
	for _, r := range res {
		if r == "ipfs" {
			IfExit(fmt.Errorf("IPFS service exists.\nPlease remove it with\neris services rm ipfs\n"))
		}
	}
}
