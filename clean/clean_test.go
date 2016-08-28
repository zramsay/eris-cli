package clean

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	srv "github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/tests"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"

	docker "github.com/fsouza/go-dockerclient"
)

const customImage = "sample"

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit(tests.ConnectAndPull, "keys", "ipfs"))

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

func TestCleanLatentChainDatas(t *testing.T) {
	defer util.RemoveAllErisContainers()

	// create a couple chains
	chain0 := "clean-me-0"
	chain1 := "clean-me-1"
	testStartChain(chain0, t)
	testStartChain(chain1, t)

	// remove them (w/o --dir)
	doClean := definitions.NowDo()
	// default settings except for ChnDirs
	doClean.Yes = true
	doClean.Containers = true
	doClean.Scratch = true
	doClean.ChnDirs = false
	doClean.RmD = false
	doClean.Images = false

	if err := Clean(doClean); err != nil {
		t.Fatalf("error cleaning chains: %v", err)
	}

	// check that their dirs & .toml exist
	testCheckChainDirsExist([]string{chain0, chain1}, true, t)

	// run clean with --chn-dir
	doClean.ChnDirs = true
	if err := Clean(doClean); err != nil {
		t.Fatalf("error cleaning chains: %v", err)
	}
	// check that they're gone
	testCheckChainDirsExist([]string{chain0, chain1}, false, t)
}

func testCheckChainDirsExist(chains []string, yes bool, t *testing.T) {
	if yes { // fail if dirs/files don't exist
		for _, chn := range chains {
			if !util.DoesDirExist(filepath.Join(common.ChainsPath, chn)) {
				t.Fatalf("chain directory does not exist when it should")
			}
			_, err := loaders.LoadChainConfigFile(chn) // list.Known only Prints to stdout
			if err != nil {
				t.Fatalf("can't load chain defintion file")
			}
		}
	} else { // !yes, fail if dirs/files do exist
		for _, chn := range chains {
			if util.DoesDirExist(filepath.Join(common.ChainsPath, chn)) {
				t.Fatalf("chain directory exists when it shouldn't")
			}
			_, err := loaders.LoadChainConfigFile(chn)
			if err == nil {
				t.Fatalf("no error loading chain def that shouldn't exist")
			}
		}
	}
}

func testCreateNotEris(name string, t *testing.T) string {
	if err := perform.DockerBuild(customImage, "FROM "+path.Join(config.Global.DefaultRegistry, config.Global.ImageKeys)); err != nil {
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
	doMake := definitions.NowDo()
	doMake.Name = chainName
	doMake.ChainType = "simplechain"
	if err := chains.MakeChain(doMake); err != nil {
		t.Fatalf("expected a chain to be made, got %v", err)
	}

	do := definitions.NowDo()
	do.Name = chainName
	do.Operations.PublishAllPorts = true
	do.Path = filepath.Join(common.ChainsPath, chainName)
	do.ConfigFile = filepath.Join(common.ChainsPath, "default", "config.toml") // TODO remove
	if err := chains.StartChain(do); err != nil {
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
