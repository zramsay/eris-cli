package clean

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	srv "github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/tests"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	logger "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

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
	testFailIfAnyContainers(t)

	// start a bunch of eris containers
	testStartService("ipfs", t)
	testStartService("tor", t)
	testStartService("keys", t)

	testStartChain("dirty-chain0", t)
	testStartChain("dirty-chain1", t)

	testCreateDataContainer("filthy-data0", t)
	testCreateDataContainer("filthy-data1", t)

	// start some not eris containers (using busybox)
	notEris0 := testCreateNotEris("not_eris0", t)
	notEris1 := testCreateNotEris("not_eris1", t)

	// run the command
	tests.IfExit(util.RemoveAllErisContainers())

	// check that both not_eris still exist
	// and no Eris containers exist
	testCheckSimple(notEris0, t)
	testCheckSimple(notEris1, t)

	// remove what's left
	testRemoveNotEris(notEris0, t)
	testRemoveNotEris(notEris1, t)

	// fail is anything remains
	testFailIfAnyContainers(t)

}

func testFailIfAnyContainers(t *testing.T) {
	contns, err := util.DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		t.Fatalf("error listing containers: %v", err)
	}

	if len(contns) != 0 {
		t.Fatalf("found (%v) remaining containers, something went wrong\n", len(contns))
	}
}

// it works...any working test
// will take too long IMO [zr]
func TestRemoveErisImages(t *testing.T) {
}

func testCreateNotEris(name string, t *testing.T) string {
	opts := docker.CreateContainerOptions{
		Name: name,
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

func testRemoveNotEris(contID string, t *testing.T) {
	rmOpts := docker.RemoveContainerOptions{
		ID:            contID,
		RemoveVolumes: true,
		Force:         true,
	}
	if err := util.DockerClient.RemoveContainer(rmOpts); err != nil {
		t.Fatalf("error removing container: %v", err)
	}
}

func testCheckSimple(newContID string, t *testing.T) {
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
}

//from /services/services_test.go
func testStartService(serviceName string, t *testing.T) {
	do := definitions.NowDo()
	do.Operations.Args = []string{serviceName}
	do.Operations.PublishAllPorts = true
	log.WithField("=>", fmt.Sprintf("%s:%d", serviceName, 1)).Debug("Starting service (from tests)")
	if err := srv.StartService(do); err != nil {
		t.Fatalf("Error starting service: %v", err)
	}

	tests.IfExit(tests.TestExistAndRun(serviceName, "service", true, true))
	tests.IfExit(tests.TestNumbersExistAndRun(serviceName, 1, 1))
}

func testStartChain(chainName string, t *testing.T) {
	do := definitions.NowDo()
	do.Name = chainName
	do.Operations.PublishAllPorts = true
	if err := chains.NewChain(do); err != nil {
		t.Fatalf("starting chain %v failed: %v", chainName, err)
	}
}

func testCreateDataContainer(dataName string, t *testing.T) {
	newDataDir := filepath.Join(common.DataContainersPath, dataName)
	if err := os.MkdirAll(newDataDir, 0777); err != nil {
		t.Fatalf("err mkdir: %v\n", err)
	}

	f, err := os.Create(filepath.Join(newDataDir, "test"))
	if err != nil {
		t.Fatalf("err creating file: %v\n", err)
	}
	defer f.Close()

	do := definitions.NowDo()
	do.Name = dataName
	do.Source = filepath.Join(common.DataContainersPath, do.Name)
	do.Destination = common.ErisContainerRoot
	log.WithField("=>", do.Name).Info("Importing data (from tests)")
	if err := data.ImportData(do); err != nil {
		t.Fatalf("error importing data: %v\n", err)
	}

	if err := tests.TestExistAndRun(dataName, "data", true, false); err != nil {
		t.Fatalf("error creating data cont: %v\n", err)
	}

}
