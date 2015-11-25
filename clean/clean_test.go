package clean

import (
	"fmt"
	"os"
	"testing"

	def "github.com/eris-ltd/eris-cli/definitions"
	srv "github.com/eris-ltd/eris-cli/services"
	tests "github.com/eris-ltd/eris-cli/testutils"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

var DEAD bool // XXX: don't double panic (TODO: Flushing twice blocks)

func fatal(t *testing.T, err error) {
	if !DEAD {
		log.Flush()
		tests.TestsTearDown()
		DEAD = true
		panic(err)
	}
}

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	logLevel = 0
	// logLevel = 1
	// logLevel = 3

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	tests.IfExit(testsInit())

	exitCode := m.Run()

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		tests.IfExit(tests.TestsTearDown())
	}

	os.Exit(exitCode)
}

func TestClean(t *testing.T) {
	// since we'll be cleaning all eris containers, check if any exist & flame out
	contns, err := util.DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		fatal(t, err)
	}
	for _, container := range contns {
		if container.Labels["eris:ERIS"] == "true" {
			fatal(t, fmt.Errorf("Eris container detected. Please remove all eris containers prior to initiating tests\n"))
		}
	}

	// each boot 2 contns
	testStartService(t, "ipfs", false)
	testStartService(t, "keys", false)

	opts := docker.CreateContainerOptions{
		Name: "not_eris",
		Config: &docker.Config{
			Image:           "ubuntu",
			AttachStdin:     false,
			AttachStdout:    false,
			AttachStderr:    false,
			Tty:             false,
			OpenStdin:       false,
			NetworkDisabled: true,
			Entrypoint:      []string{},
			Cmd:             []string{},
		},
		HostConfig: &docker.HostConfig{},
	}

	newCont, err := util.DockerClient.CreateContainer(opts)
	if err != nil {
		fatal(t, err)
	}

	//only runs default clean -> need test for others
	if err := util.Clean(false, false, false, false); err != nil {
		fatal(t, err)
	}

	contns, err = util.DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		fatal(t, err)
	}

	//if any Eris, fail
	for _, container := range contns {
		if container.Labels["eris:ERIS"] == "true" {
			fatal(t, fmt.Errorf("Eris container detected. Clean did not do its job\n"))
		}
	}

	var notEris bool
	//make sure "not_eris" still exists
	for _, container := range contns {
		if container.ID == newCont.ID {
			notEris = true
			break
		} else {
			notEris = false
		}
	}

	if !notEris {
		fatal(t, fmt.Errorf("Expected running container, did not find %s\n", newCont.ID))
	}

	//teardown non-eris
	rmOpts := docker.RemoveContainerOptions{
		ID:            newCont.ID,
		RemoveVolumes: true,
		Force:         true,
	}

	if err := util.DockerClient.RemoveContainer(rmOpts); err != nil {
		fatal(t, err)
	}

	contns, err = util.DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		fatal(t, err)
	}

	//check only that not_eris and no eris contns are running
	for _, container := range contns {
		if container.Labels["eris:ERIS"] == "true" || container.ID == newCont.ID {
			fatal(t, fmt.Errorf("found remaining eris containers or test container, something went wrong\n"))
		}
	}
}

//from /services/services_test.go
func testStartService(t *testing.T, serviceName string, publishAll bool) {
	do := def.NowDo()
	do.Operations.Args = []string{serviceName}
	do.Operations.ContainerNumber = 1 //util.AutoMagic(0, "service", true)
	do.Operations.PublishAllPorts = publishAll
	logger.Debugf("Starting service (via tests) =>\t%s:%d\n", serviceName, do.Operations.ContainerNumber)
	e := srv.StartService(do)
	if e != nil {
		logger.Infof("Error starting service =>\t%v\n", e)
		fatal(t, e)
	}

	testExistAndRun(t, serviceName, 1, true, true)
	testNumbersExistAndRun(t, serviceName, 1, 1)
}

func testExistAndRun(t *testing.T, servName string, containerNumber int, toExist, toRun bool) {
	if tests.TestExistAndRun(servName, "services", containerNumber, toExist, toRun) {
		fatal(t, nil)
	}
}

//[zr] TODO move to testings pacakge...wait with [pv] done with moving to testutils
//could also refactor this to use labels && be more generalized...
func testNumbersExistAndRun(t *testing.T, servName string, containerExist, containerRun int) {
	logger.Infof("\nTesting number of (%s) containers. Existing? (%d) and Running? (%d)\n", servName, containerExist, containerRun)

	logger.Debugf("Checking Existing Containers =>\t%s\n", servName)
	exist := util.HowManyContainersExisting(servName, "service")
	logger.Debugf("Checking Running Containers =>\t%s\n", servName)
	run := util.HowManyContainersRunning(servName, "service")

	if exist != containerExist {
		logger.Printf("Wrong number of containers existing for service (%s). Expected (%d). Got (%d).\n", servName, containerExist, exist)
		fatal(t, nil)
	}

	if run != containerRun {
		logger.Printf("Wrong number of containers running for service (%s). Expected (%d). Got (%d).\n", servName, containerRun, run)
		fatal(t, nil)
	}

	logger.Infoln("All good.\n")
}

func testsInit() error {
	if err := tests.TestsInit("clean"); err != nil {
		return err
	}
	return nil
}
