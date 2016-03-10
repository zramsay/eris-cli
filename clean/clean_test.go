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

var ToTest = []struct {
	name       string
	all        bool
	containers bool
	scratch    bool
	rmd        bool
	images     bool
	yes        bool
	// uninstall bool => forthcoming
	// volumes bool => forthcoming
}{}

/*{
	{"no-flag", false, false, false, false},
	{"rmd-only", false, true, false, false},
	{"imgs-only", false, false, true, false},
	{"imgs-rmd", false, true, true, false}, //is this  == --all -> prefered behaviour ?
	{"all-only", false, false, false, true},
	{"all-flags", false, true, true, true},
}*/

func TestMain(m *testing.M) {
	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit("clean"))

	exitCode := m.Run()
	tests.IfExit(tests.TestsTearDown())
	os.Exit(exitCode)
}

func TestRemoveAllErisContainers(t *testing.T) {

}

func TestCleanScratchData(t *testing.T) {

}

func TestRemoveErisImages(t *testing.T) {

}

func testCreateNotEris(t *testing.T) {

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
		t.Fatalf("create container error: %v", err)
	}
	return newCont.ID
}

func testCheckSimple(t *testing.T, newContID string) {
	contns, err := util.DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		t.Fatalf("error listing containers: %v", err)
	}

	//if any Eris, fail
	for _, container := range contns {
		if container.Labels["eris:ERIS"] == "true" {
			t.Fatalf("Eris container detected. Clean did not do its job\n")
		}
	}

	var notEris bool
	//make sure "not_eris" still exists
	for _, container := range contns {
		if container.ID == newContID {
			notEris = true
			break
		} else {
			notEris = false
		}
	}

	if !notEris {
		t.Fatalf("Expected running container, did not find %s\n", newContID)
	}

	//teardown non-eris
	rmOpts := docker.RemoveContainerOptions{
		ID:            newContID,
		RemoveVolumes: true,
		Force:         true,
	}

	if err := util.DockerClient.RemoveContainer(rmOpts); err != nil {
		t.Fatalf("error removing container: %v", err)
	}

	contns, err = util.DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		t.Fatalf("error listing containers: %v", err)
	}

	if len(contns) != 0 {
		t.Fatalf("found remaining containers, something went wrong\n")
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
		t.Fatalf("Error starting service: %v", e)
	}

	tests.IfExit(tests.TestExistAndRun(servName, "service", containerNumber, toExist, toRun))
	tests.IfExit(tests.TestNumbersExistAndRun(serviceName, 1, 1))
}
