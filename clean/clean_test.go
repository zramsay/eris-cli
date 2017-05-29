package clean

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/monax/monax/chains"
	"github.com/monax/monax/config"
	"github.com/monax/monax/data"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/loaders"
	"github.com/monax/monax/log"
	"github.com/monax/monax/perform"
	"github.com/monax/monax/services"
	"github.com/monax/monax/testutil"
	"github.com/monax/monax/util"
	"github.com/monax/monax/version"

	docker "github.com/fsouza/go-dockerclient"
)

const customImage = "sample"

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	testutil.IfExit(testutil.Init(testutil.Pull{
		Images:   []string{"keys", "data", "db"},
		Services: []string{"keys"},
	}))

	exitCode := m.Run()
	testutil.IfExit(testutil.TearDown())
	os.Exit(exitCode)
}

func TestRemoveAllMonaxContainers(t *testing.T) {
	defer util.RemoveAllMonaxContainers()

	// Start a bunch of monax containers.
	testStartService("keys", t)

	testStartChain("dirty-chain0", t)
	testStartChain("dirty-chain1", t)

	testCreateDataContainer("filthy-data0", t)
	testCreateDataContainer("filthy-data1", t)

	// Start some non-Monax containers.
	notMonax0 := testCreateNotMonax("not_monax0", t)
	notMonax1 := testCreateNotMonax("not_monax1", t)

	testutil.IfExit(util.RemoveAllMonaxContainers())

	// Check that both not_monax still exist and no Monax containers exist.
	testCheckSimple(notMonax0, t)
	testCheckSimple(notMonax1, t)

	// Remove what's left.
	testRemoveNotMonax(notMonax0, t)
	testRemoveNotMonax(notMonax1, t)

	// Remove custom built image.
	testRemoveNotMonaxImage(t)
}

func TestCleanLatentChainDatas(t *testing.T) {
	defer util.RemoveAllMonaxContainers()

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
			if !util.DoesDirExist(filepath.Join(config.ChainsPath, chn)) {
				t.Fatalf("chain directory does not exist when it should")
			}
			_, err := loaders.LoadChainDefinition(chn) // list.Known only Prints to stdout
			if err != nil {
				t.Fatalf("can't load chain defintion file")
			}
		}
	} else { // !yes, fail if dirs/files do exist
		for _, chn := range chains {
			if util.DoesDirExist(filepath.Join(config.ChainsPath, chn)) {
				t.Fatalf("chain directory exists when it shouldn't")
			}
			_, err := loaders.LoadChainDefinition(chn, filepath.Join(config.ChainsPath, chn, "config.toml"))
			if err == nil {
				t.Fatalf("no error loading chain def that shouldn't exist")
			}
		}
	}
}

func testCreateNotMonax(name string, t *testing.T) string {
	if err := perform.DockerBuild(customImage, "FROM "+path.Join(version.DefaultRegistry, version.ImageKeys)); err != nil {
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

func testRemoveNotMonax(contID string, t *testing.T) {
	rmOpts := docker.RemoveContainerOptions{
		ID:            contID,
		RemoveVolumes: true,
		Force:         true,
	}
	if err := util.DockerClient.RemoveContainer(rmOpts); err != nil {
		t.Fatalf("error removing container: %v", err)
	}
}

func testRemoveNotMonaxImage(t *testing.T) {
	if err := perform.DockerRemoveImage(customImage, true); err != nil {
		t.Fatalf("expected custom image to be removed, got %v", err)
	}
}

func testCheckSimple(newContID string, t *testing.T) {
	contns, err := util.DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		t.Fatalf("error listing containers: %v", err)
	}

	// If any Monax containers exist, fail.
	util.MonaxContainers(func(name string, details *util.Details) bool {
		t.Fatalf("expected no Monax containers running")
		return true
	}, false)

	var notMonax bool
	for _, container := range contns {
		if container.ID == newContID {
			notMonax = true
			break
		} else {
			notMonax = false
		}
	}

	if !notMonax {
		t.Fatalf("expected running container, did not find %s", newContID)
	}
}

func testStartService(serviceName string, t *testing.T) {
	do := definitions.NowDo()
	do.Operations.Args = []string{serviceName}
	do.Operations.PublishAllPorts = true
	if err := services.StartService(do); err != nil {
		t.Fatalf("error starting service: %v", err)
	}

	testutil.IfExit(testutil.ExistAndRun(serviceName, "service", true, true))
	testutil.IfExit(testutil.NumbersExistAndRun(serviceName, true, true))
}

func testStartChain(chainName string, t *testing.T) {
	doMake := definitions.NowDo()
	doMake.Name = chainName
	doMake.ChainType = "simplechain"
	doMake.ChainImageName = path.Join(version.DefaultRegistry, version.ImageDB)
	if err := chains.MakeChain(doMake); err != nil {
		t.Fatalf("expected a chain to be made, got %v", err)
	}

	do := definitions.NowDo()
	do.Name = chainName
	do.Operations.PublishAllPorts = true
	do.Path = filepath.Join(config.ChainsPath, chainName, fmt.Sprintf("%s_full_000", chainName)) // --init-dir
	if err := chains.StartChain(do); err != nil {
		t.Fatalf("starting chain %v failed: %v", chainName, err)
	}
}

func testCreateDataContainer(dataName string, t *testing.T) {
	newDataDir := filepath.Join(config.DataContainersPath, dataName)
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
	do.Source = filepath.Join(config.DataContainersPath, do.Name)
	do.Destination = config.MonaxContainerRoot
	if err := data.ImportData(do); err != nil {
		t.Fatalf("error importing data: %v", err)
	}

	if err := testutil.ExistAndRun(dataName, "data", true, false); err != nil {
		t.Fatalf("error creating data cont: %v", err)
	}

}
