package clean

import (
	"fmt"
	"os"
	"testing"

	def "github.com/eris-ltd/eris-cli/definitions"
	srv "github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/tests"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	logger "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

var DEAD bool // XXX: don't double panic (TODO: Flushing twice blocks)

func fatal(t *testing.T, err error) {
	if !DEAD {
		tests.TestsTearDown()
		DEAD = true
		panic(err)
	}
}

func TestMain(m *testing.M) {
	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)
	tests.IfExit(testsInit())

	exitCode := m.Run()
	tests.IfExit(tests.TestsTearDown())
	os.Exit(exitCode)
}

func TestClean(t *testing.T) {
	tests.RemoveAllContainers()

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
			Image:           "busybox",
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
	do.Operations.PublishAllPorts = publishAll
	log.WithField("=>", fmt.Sprintf("%s:%d", serviceName, 1)).Debug("Starting service (from tests)")
	e := srv.StartService(do)
	if e != nil {
		log.Infof("Error starting service: %v", e)
		fatal(t, e)
	}

	testExistAndRun(t, serviceName, 1, true, true)
	testNumbersExistAndRun(t, serviceName, 1, 1)
}

func testExistAndRun(t *testing.T, servName string, containerNumber int, toExist, toRun bool) {
	if err := tests.TestExistAndRun(servName, "service", containerNumber, toExist, toRun); err != nil {
		fatal(t, nil)
	}
}

//[zr] TODO move to testings package
func testNumbersExistAndRun(t *testing.T, servName string, containerExist, containerRun int) {
	log.WithFields(log.Fields{
		"=>":        servName,
		"existing#": containerExist,
		"running#":  containerRun,
	}).Info("Checking number of containers for")

	log.WithField("=>", servName).Debug("Checking existing containers for")
	exist := util.HowManyContainersExisting(servName, "service")

	log.WithField("=>", servName).Debug("Checking running containers for")
	run := util.HowManyContainersRunning(servName, "service")

	if exist != containerExist {
		log.WithFields(log.Fields{
			"name":     servName,
			"expected": containerExist,
			"got":      exist,
		}).Error("Wrong number of existing containers")
		fatal(t, nil)
	}

	if run != containerRun {
		log.WithFields(log.Fields{
			"name":     servName,
			"expected": containerExist,
			"got":      run,
		}).Error("Wrong number of running containers")
		fatal(t, nil)
	}

	log.Info("All good")
}

func testsInit() error {
	if err := tests.TestsInit("clean"); err != nil {
		return err
	}
	return nil
}
