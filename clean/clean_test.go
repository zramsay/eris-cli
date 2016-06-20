package clean

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/perform"
	srv "github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/tests"
	"github.com/eris-ltd/eris-cli/util"
	ver "github.com/eris-ltd/eris-cli/version"

	"github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"

	docker "github.com/fsouza/go-dockerclient"
)

const customImage = "sample"

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit(tests.ConnectAndPull))

	exitCode := m.Run()
	tests.IfExit(tests.TestsTearDown())
	os.Exit(exitCode)
}

func TestRemoveAllErisContainers(t *testing.T) {
	defer util.RemoveAllErisContainers()

	// Start a bunch of eris containers.
	testStartService("ipfs", t)
	testStartService("keys", t)

	testStartChain("dirty-chain0", t)
	testStartChain("dirty-chain1", t)

	testCreateDataContainer("filthy-data0", t)
	testCreateDataContainer("filthy-data1", t)

	// Start some non-Eris containers.
	notEris0 := testCreateNotEris("not_eris0", t)
	notEris1 := testCreateNotEris("not_eris1", t)

	tests.IfExit(util.RemoveAllErisContainers())

	// Check that both not_eris still exist and no Eris containers exist.
	testCheckSimple(notEris0, t)
	testCheckSimple(notEris1, t)

	// Remove what's left.
	testRemoveNotEris(notEris0, t)
	testRemoveNotEris(notEris1, t)

	// Remove custom built image.
	testRemoveNotErisImage(t)
}

func testCreateNotEris(name string, t *testing.T) string {
	if err := perform.DockerBuild(customImage, "FROM "+path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_KEYS)); err != nil {
		t.Fatalf("expected to build a custom image, got %v", err)
	}

	opts := docker.CreateContainerOptions{
		Name: name,
		Config: &docker.Config{
			Image:           customImage,
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

func testRemoveNotErisImage(t *testing.T) {
	if err := perform.DockerRemoveImage(customImage, true); err != nil {
		t.Fatalf("expected custom image to be removed, got %v", err)
	}
}

func testCheckSimple(newContID string, t *testing.T) {
	contns, err := util.DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		t.Fatalf("error listing containers: %v", err)
	}

	// If any Eris containers exist, fail.
	util.ErisContainers(func(name string, details *util.Details) bool {
		t.Fatalf("expected no Eris containers running")
		return true
	}, false)

	var notEris bool
	for _, container := range contns {
		if container.ID == newContID {
			notEris = true
			break
		} else {
			notEris = false
		}
	}

	if !notEris {
		t.Fatalf("expected running container, did not find %s", newContID)
	}
}

func testStartService(serviceName string, t *testing.T) {
	do := definitions.NowDo()
	do.Operations.Args = []string{serviceName}
	do.Operations.PublishAllPorts = true
	if err := srv.StartService(do); err != nil {
		t.Fatalf("error starting service: %v", err)
	}

	tests.IfExit(tests.TestExistAndRun(serviceName, "service", true, true))
	tests.IfExit(tests.TestNumbersExistAndRun(serviceName, true, true))
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
		t.Fatalf("error mkdir: %v\n", err)
	}

	f, err := os.Create(filepath.Join(newDataDir, "test"))
	if err != nil {
		t.Fatalf("error creating file: %v", err)
	}
	defer f.Close()

	do := definitions.NowDo()
	do.Name = dataName
	do.Source = filepath.Join(common.DataContainersPath, do.Name)
	do.Destination = common.ErisContainerRoot
	if err := data.ImportData(do); err != nil {
		t.Fatalf("error importing data: %v", err)
	}

	if err := tests.TestExistAndRun(dataName, "data", true, false); err != nil {
		t.Fatalf("error creating data cont: %v", err)
	}

}
