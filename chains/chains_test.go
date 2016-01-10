package chains

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/logger"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	tests "github.com/eris-ltd/eris-cli/testutils"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

var erisDir string = filepath.Join(os.TempDir(), "eris")
var chainName string = "my_testing_chain_dot_com" // :( [csk]-> :)

func TestMain(m *testing.M) {
	runtime.GOMAXPROCS(1)
	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit("chain"))
	log.Info("Test init completed. Starting main test sequence now")

	layTestChainToml(chainName)

	fmt.Println(m.Run())
}

func TestKnownChain(t *testing.T) {
	do := def.NowDo()
	do.Known = true
	do.Existing = false
	do.Running = false
	do.Operations.Args = []string{"testing"}
	tests.IfExit(util.ListAll(do, "chains"))

	k := strings.Split(do.Result, "\n") // tests output formatting.

	// apparently these have extra space
	if strings.TrimSpace(k[0]) != chainName {
		log.WithField("=>", do.Result).Debug("Result")
		tests.IfExit(fmt.Errorf("Unexpected chain definition file. Got %s, expected %s.", k[0], chainName))
	}
}

func TestChainGraduate(t *testing.T) {
	do := def.NowDo()
	do.Name = chainName
	log.WithField("=>", do.Name).Info("Graduate chain (from tests)")
	if err := GraduateChain(do); err != nil {
		tests.IfExit(err)
	}

	srvDef, err := loaders.LoadServiceDefinition(chainName, false, 1)
	if err != nil {
		tests.IfExit(err)
	}

	image := "quay.io/eris/erisdb:" + version.VERSION
	if srvDef.Service.Image != image {
		tests.IfExit(fmt.Errorf("FAILURE: improper service image on GRADUATE. expected: %s\tgot: %s\n", image, srvDef.Service.Image))
	}

	if srvDef.Service.Command != loaders.ErisChainStart {
		tests.IfExit(fmt.Errorf("FAILURE: improper service command on GRADUATE. expected: %s\tgot: %s\n", loaders.ErisChainStart, srvDef.Service.Command))
	}

	if !srvDef.Service.AutoData {
		tests.IfExit(fmt.Errorf("FAILURE: improper service autodata on GRADUATE. expected: %t\tgot: %t\n", true, srvDef.Service.AutoData))
	}

	if len(srvDef.Dependencies.Services) != 1 {
		tests.IfExit(fmt.Errorf("FAILURE: improper service deps on GRADUATE. expected: [\"keys\"]\tgot: %s\n", srvDef.Dependencies.Services))
	}
}

func TestLoadChainDefinition(t *testing.T) {
	var e error
	log.WithField("=>", chainName).Info("Load chain definition (from tests)")
	chn, e := loaders.LoadChainDefinition(chainName, false, 1)
	if e != nil {
		tests.IfExit(e)
	}

	if chn.Service.Name != chainName {
		tests.IfExit(fmt.Errorf("FAILURE: improper service name on LOAD. expected: %s\tgot: %s", chainName, chn.Service.Name))
	}

	if !chn.Service.AutoData {
		tests.IfExit(fmt.Errorf("FAILURE: data_container not properly read on LOAD."))
	}

	if chn.Operations.DataContainerName == "" {
		tests.IfExit(fmt.Errorf("FAILURE: data_container_name not set."))
	}
}

func TestStartKillChain(t *testing.T) {
	testStartChain(t, chainName)
	testKillChain(t, chainName)
}

func TestRestartChain(t *testing.T) {
	testNewChain(chainName)
	defer testKillChain(t, chainName)

	testKillChain(t, chainName)
	testStartChain(t, chainName)
	testKillChain(t, chainName)
}

// TODO: this isn't actually testing much!
func TestExecChain(t *testing.T) {
	testStartChain(t, chainName)
	defer testKillChain(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Operations.Args = strings.Fields("ls -la /home/eris/.eris")
	do.Operations.Interactive = false
	log.WithField("=>", do.Name).Info("Executing chain (from tests)")
	e := ExecChain(do)
	if e != nil {
		log.Error(e)
		t.Fail()
	}
}

func TestPlopChain(t *testing.T) {
	cfgFilePath := filepath.Join(common.ChainsPath, "default", "config.toml")

	do := def.NowDo()
	do.ConfigFile = cfgFilePath
	do.Name = chainName
	do.Operations.ContainerNumber = 1
	do.Operations.PublishAllPorts = true
	log.WithField("=>", chainName).Info("Creating chain (from tests)")
	tests.IfExit(NewChain(do))
	defer testKillChain(t, chainName)

	do = def.NowDo()
	do.Type = "config"
	do.ChainID = chainName

	newWriter := new(bytes.Buffer)
	config.GlobalConfig.Writer = newWriter
	e := PlopChain(do)
	if e != nil {
		log.Error(e)
		t.Fail()
	}

	cfgFile, err := ioutil.ReadFile(cfgFilePath)
	tests.IfExit(err)
	cfgFilePlop := newWriter.Bytes()

	// remove [13] that shows up everywhere ...
	cfgFile = bytes.Replace(cfgFile, []byte{13}, []byte{}, -1)
	cfgFilePlop = bytes.Replace(cfgFilePlop, []byte{13}, []byte{}, -1)

	if !bytes.Equal(cfgFile, cfgFilePlop) {
		log.Errorf("Error: Got: %s. Expected: %s", cfgFilePlop, cfgFile)
		log.WithFields(log.Fields{
			"got":      cfgFilePlop,
			"expected": cfgFile,
		}).Error("Error comparing config files")
		t.Fail()
	}
}

// eris chains new --dir _ -g _
// the default chain_id is my_tests, so should be overwritten
func TestChainsNewDirGen(t *testing.T) {
	chainID := "testChainsNewDirGen"
	myDir := filepath.Join(common.DataContainersPath, chainID)
	if err := os.MkdirAll(myDir, 0700); err != nil {
		tests.IfExit(err)
	}
	contents := "this is a file in the directory\n"
	if err := ioutil.WriteFile(filepath.Join(myDir, "file.file"), []byte(contents), 0664); err != nil {
		tests.IfExit(err)
	}

	do := def.NowDo()
	do.GenesisFile = filepath.Join(common.ChainsPath, "default", "genesis.json")
	do.Name = chainID
	do.Path = myDir
	do.Operations.ContainerNumber = 1
	do.Operations.PublishAllPorts = true
	log.WithField("=>", do.Name).Info("Creating chain (from tests)")
	tests.IfExit(NewChain(do))

	// remove the data container
	defer removeChainContainer(t, chainID, do.Operations.ContainerNumber)

	// verify the contents of file.file - swap config writer with bytes.Buffer
	// TODO: functions for facilitating this
	oldWriter := config.GlobalConfig.Writer
	newWriter := new(bytes.Buffer)
	config.GlobalConfig.Writer = newWriter
	ops := loaders.LoadDataDefinition(do.Name, do.Operations.ContainerNumber)
	util.Merge(ops, do.Operations)
	ops.Args = []string{"cat", fmt.Sprintf("/home/eris/.eris/file.file")}
	b, err := perform.DockerRunData(ops, nil)
	if err != nil {
		tests.IfExit(err)
	}

	config.GlobalConfig.Writer = oldWriter
	result := trimResult(string(b))
	contents = trimResult(contents)
	if result != contents {
		tests.IfExit(fmt.Errorf("file not faithfully copied. Got: %s \n Expected: %s", result, contents))
	}

	// verify the chain_id got swapped in the genesis.json
	// TODO: functions for facilitating this
	oldWriter = config.GlobalConfig.Writer
	newWriter = new(bytes.Buffer)
	config.GlobalConfig.Writer = newWriter
	ops = loaders.LoadDataDefinition(do.Name, do.Operations.ContainerNumber)
	util.Merge(ops, do.Operations)
	ops.Args = []string{"cat", fmt.Sprintf("/home/eris/.eris/chains/%s/genesis.json", chainID)} //, "|", "jq", ".chain_id"}
	b, err = perform.DockerRunData(ops, nil)
	if err != nil {
		tests.IfExit(err)
	}

	config.GlobalConfig.Writer = oldWriter
	result = string(b)

	s := struct {
		ChainID string `json:"chain_id"`
	}{}
	if err := json.Unmarshal([]byte(result), &s); err != nil {
		tests.IfExit(err)
	}

	if s.ChainID != chainID {
		tests.IfExit(fmt.Errorf("ChainID mismatch: got %s, expected %s", s.ChainID, chainID))
	}
}

// eris chains new -c _ -csv _
func TestChainsNewConfigAndCSV(t *testing.T) {
	chainID := "testChainsNewConfigAndCSV"
	do := def.NowDo()
	do.Name = chainID
	do.ConfigFile = filepath.Join(common.ChainsPath, "default", "config.toml")
	do.CSV = filepath.Join(common.ChainsPath, "default", "genesis.csv")
	do.Operations.ContainerNumber = 1
	do.Operations.PublishAllPorts = true
	log.WithField("=>", do.Name).Info("Creating chain (from tests)")
	tests.IfExit(NewChain(do))
	_, err := ioutil.ReadFile(do.ConfigFile)
	if err != nil {
		tests.IfExit(err)
	}

	// remove the data container
	defer removeChainContainer(t, chainID, do.Operations.ContainerNumber)

	// verify the contents of config.toml
	ops := loaders.LoadDataDefinition(do.Name, do.Operations.ContainerNumber)
	util.Merge(ops, do.Operations)
	ops.Args = []string{"cat", fmt.Sprintf("/home/eris/.eris/chains/%s/config.toml", chainID)}
	result := trimResult(string(runContainer(t, ops)))

	configDefault := filepath.Join(erisDir, "chains", "default", "config.toml")
	read, err := ioutil.ReadFile(configDefault)
	if err != nil {
		tests.IfExit(err)
	}
	contents := trimResult(string(read))

	if result != contents {
		tests.IfExit(fmt.Errorf("config not properly copied. Got: %s \n Expected: %s", result, contents))
	}

	// verify the contents of genesis.json (should have the validator from the csv)
	ops = loaders.LoadDataDefinition(do.Name, do.Operations.ContainerNumber)
	util.Merge(ops, do.Operations)
	ops.Args = []string{"cat", fmt.Sprintf("/home/eris/.eris/chains/%s/genesis.json", chainID)}
	result = string(runContainer(t, ops))
	var found bool
	for _, s := range strings.Split(result, "\n") {
		if strings.Contains(s, ini.DefaultPubKeys[0]) {
			found = true
			break
		}
	}
	if !found {
		tests.IfExit(fmt.Errorf("Did not find pubkey %s in genesis.json: %s", ini.DefaultPubKeys[0], result))
	}
}

// eris chains new --options
func TestChainsNewConfigOpts(t *testing.T) {
	// XXX: need to use a different chainID or remove the local tmp/eris/data/chainID dir with each test!
	chainID := "testChainsNewConfigOpts"
	do := def.NowDo()

	do.Name = chainID
	do.ConfigOpts = []string{"moniker=satoshi", "p2p=1.1.1.1:42", "fast-sync=true"}
	do.Operations.ContainerNumber = 1
	do.Operations.PublishAllPorts = true
	log.WithField("=>", do.Name).Info("Creating chain (from tests)")
	tests.IfExit(NewChain(do))

	// remove the data container
	defer removeChainContainer(t, chainID, do.Operations.ContainerNumber)

	// verify the contents of config.toml
	ops := loaders.LoadDataDefinition(do.Name, do.Operations.ContainerNumber)
	util.Merge(ops, do.Operations)
	ops.Args = []string{"cat", fmt.Sprintf("/home/eris/.eris/chains/%s/config.toml", chainID)}
	result := string(runContainer(t, ops))

	spl := strings.Split(result, "\n")
	var found bool
	for _, s := range spl {
		if ensureTomlValue(t, s, "moniker", "satoshi") {
			found = true
		}
		if ensureTomlValue(t, s, "node_laddr", "1.1.1.1:42") {
			found = true
		}
		if ensureTomlValue(t, s, "fast_sync", "true") {
			found = true
		}
	}
	if !found {
		tests.IfExit(fmt.Errorf("failed to find fields: %s", result))
	}
}

func TestLogsChain(t *testing.T) {
	testStartChain(t, chainName)
	defer testKillChain(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Follow = false
	do.Tail = "all"
	log.WithFields(log.Fields{
		"=>":   do.Name,
		"tail": do.Tail,
	}).Info("Getting chain logs (from tests)")
	e := LogsChain(do)
	if e != nil {
		tests.IfExit(e)
	}
}

func TestUpdateChain(t *testing.T) {
	testStartChain(t, chainName)
	defer testKillChain(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Pull = false
	do.Operations.PublishAllPorts = true
	log.WithField("=>", do.Name).Info("Updating chain (from tests)")
	if e := UpdateChain(do); e != nil {
		tests.IfExit(e)
	}

	testExistAndRun(t, chainName, true, true)
}

func TestInspectChain(t *testing.T) {
	testStartChain(t, chainName)
	defer testKillChain(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Operations.Args = []string{"name"}
	do.Operations.ContainerNumber = 1
	log.WithFields(log.Fields{
		"=>":   chainName,
		"args": do.Operations.Args,
	}).Debug("Inspecting chain (from tests)")
	if e := InspectChain(do); e != nil {
		tests.IfExit(fmt.Errorf("Error inspecting chain =>\t%v\n", e))
	}
}

func TestRenameChain(t *testing.T) {
	aChain := "hichain"
	rename1 := "niahctset"
	rename2 := chainName
	testNewChain(aChain)
	defer testKillChain(t, rename2)

	do := def.NowDo()
	do.Name = aChain
	do.NewName = rename1
	log.WithFields(log.Fields{
		"from": do.Name,
		"to":   do.NewName,
	}).Info("Renaming chain (from tests)")

	if e := RenameChain(do); e != nil {
		tests.IfExit(e)
	}

	testExistAndRun(t, rename1, true, true)

	do = def.NowDo()
	do.Name = rename1
	do.NewName = rename2
	log.WithFields(log.Fields{
		"from": do.Name,
		"to":   do.NewName,
	}).Info("Renaming chain (from tests)")
	if e := RenameChain(do); e != nil {
		tests.IfExit(e)
	}

	testExistAndRun(t, chainName, true, true)
}

func TestRmChain(t *testing.T) {
	testStartChain(t, chainName)

	do := def.NowDo()
	do.Operations.Args, do.Rm, do.RmD = []string{"keys"}, true, true
	log.WithField("=>", do.Name).Info("Removing keys (from tests)")
	if e := services.KillService(do); e != nil {
		tests.IfExit(e)
	}

	do = def.NowDo()
	do.Name, do.Rm, do.RmD = chainName, false, false
	log.WithField("=>", do.Name).Info("Stopping chain (from tests)")
	if e := KillChain(do); e != nil {
		tests.IfExit(e)
	}
	testExistAndRun(t, chainName, true, false)

	do = def.NowDo()
	do.Name = chainName
	do.RmD = true
	log.WithField("=>", do.Name).Info("Removing chain (from tests)")
	if e := RmChain(do); e != nil {
		tests.IfExit(e)
	}

	testExistAndRun(t, chainName, false, false)
}

func TestServiceLinkNoChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "quay.io/eris/ipfs"
data_container = true
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	do := def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.Operations.ContainerNumber = 1
	if err := services.StartService(do); err == nil {
		t.Fatalf("expect start service to fail, got nil")
	}
}

func TestServiceLinkBadChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "quay.io/eris/ipfs"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	do := def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.Operations.ContainerNumber = 1
	do.ChainName = "non-existent-chain"
	if err := services.StartService(do); err == nil {
		t.Fatalf("expect start service to fail, got nil")
	}
}

func TestServiceLinkBadChainWithoutChainInDefinition(t *testing.T) {
	defer tests.RemoveAllContainers()

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
[service]
name = "fake"
image = "quay.io/eris/ipfs"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	do := def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.Operations.ContainerNumber = 1
	do.ChainName = "non-existent-chain"

	// [pv]: is this a bug? the service which doesn't have a
	// "chain" in its definition file doesn't care about linking at all.
	if err := services.StartService(do); err != nil {
		t.Fatalf("expect service to start, got %v", err)
	}

	if n := util.HowManyContainersRunning("fake", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data containers, got %v", n)
	}
}

func TestServiceLink(t *testing.T) {
	defer tests.RemoveAllContainers()

	do := def.NowDo()
	do.Name = chainName
	do.Operations.ContainerNumber = 1
	if err := NewChain(do); err != nil {
		t.Fatalf("could not start a new chain, got %v", err)
	}

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "quay.io/eris/ipfs"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if n := util.HowManyContainersExisting(chainName, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 test chain containers, got %v", n)
	}

	if n := util.HowManyContainersExisting("fake", def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data containers, got %v", n)
	}

	do = def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.Operations.ContainerNumber = 1
	do.ChainName = chainName
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if n := util.HowManyContainersRunning("fake", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 fake service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 fake data containers, got %v", n)
	}

	links := tests.Links("fake", def.TypeService, 1)
	if len(links) != 1 || !strings.Contains(links[0], chainName) {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

func TestServiceLinkWithDataContainer(t *testing.T) {
	defer tests.RemoveAllContainers()

	do := def.NowDo()
	do.Name = chainName
	do.Operations.ContainerNumber = 1
	if err := NewChain(do); err != nil {
		t.Fatalf("could not start a new chain, got %v", err)
	}

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "quay.io/eris/ipfs"
data_container = true
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if n := util.HowManyContainersExisting(chainName, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 test chain containers, got %v", n)
	}

	if n := util.HowManyContainersExisting("fake", def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data containers, got %v", n)
	}

	do = def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.Operations.ContainerNumber = 1
	do.ChainName = chainName
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if n := util.HowManyContainersRunning("fake", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data containers, got %v", n)
	}

	links := tests.Links("fake", def.TypeService, 1)
	if len(links) != 1 || !strings.Contains(links[0], chainName) {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

func TestServiceLinkLiteral(t *testing.T) {
	defer tests.RemoveAllContainers()

	do := def.NowDo()
	do.Name = chainName
	do.Operations.ContainerNumber = 1
	if err := NewChain(do); err != nil {
		t.Fatalf("could not start a new chain, got %v", err)
	}

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
chain = "`+chainName+`:fake"

[service]
name = "fake"
image = "quay.io/eris/ipfs"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if n := util.HowManyContainersExisting(chainName, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 test chain containers, got %v", n)
	}

	if n := util.HowManyContainersExisting("fake", def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data containers, got %v", n)
	}

	do = def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.Operations.ContainerNumber = 1
	do.ChainName = chainName
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if n := util.HowManyContainersRunning("fake", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 fake service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 fake data containers, got %v", n)
	}

	links := tests.Links("fake", def.TypeService, 1)
	if len(links) != 1 || !strings.Contains(links[0], chainName) {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

func TestServiceLinkBadLiteral(t *testing.T) {
	defer tests.RemoveAllContainers()

	do := def.NowDo()
	do.Name = chainName
	do.Operations.ContainerNumber = 1
	if err := NewChain(do); err != nil {
		t.Fatalf("could not start a new chain, got %v", err)
	}

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
chain = "blah-blah:blah"

[service]
name = "fake"
image = "quay.io/eris/ipfs"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if n := util.HowManyContainersExisting(chainName, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 test chain containers, got %v", n)
	}

	do = def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.Operations.ContainerNumber = 1
	do.ChainName = chainName
	// [pv]: probably a bug. Bad literal chain link in a definition
	// file doesn't affect the service start. Links is not nil.
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	links := tests.Links("fake", def.TypeService, 1)
	if len(links) != 1 || !strings.Contains(links[0], chainName) {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

func TestServiceLinkKeys(t *testing.T) {
	defer tests.RemoveAllContainers()

	do := def.NowDo()
	do.Name = chainName
	do.Operations.ContainerNumber = 1
	if err := NewChain(do); err != nil {
		t.Fatalf("could not start a new chain, got %v", err)
	}

	if n := util.HowManyContainersExisting(chainName, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 test chain containers, got %v", n)
	}

	do = def.NowDo()
	do.Operations.Args = []string{"keys"}
	do.Operations.ContainerNumber = 1
	do.ChainName = chainName
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if n := util.HowManyContainersExisting("keys", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 test chain containers, got %v", n)
	}

	links := tests.Links("keys", def.TypeService, 1)
	if len(links) != 0 {
		t.Fatalf("expected service links be empty, got %v", links)
	}
}

func TestServiceLinkChainedService(t *testing.T) {
	defer tests.RemoveAllContainers()

	do := def.NowDo()
	do.Name = chainName
	do.Operations.ContainerNumber = 1
	if err := NewChain(do); err != nil {
		t.Fatalf("could not start a new chain, got %v", err)
	}

	if err := tests.FakeServiceDefinition(erisDir, "fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "quay.io/eris/ipfs"

[dependencies]
services = [ "sham" ]
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if err := tests.FakeServiceDefinition(erisDir, "sham", `
chain = "$chain:sham"

[service]
name = "sham"
image = "quay.io/eris/keys"
data_container = true
`); err != nil {
		t.Fatalf("can't create a sham service definition: %v", err)
	}

	if n := util.HowManyContainersExisting(chainName, def.TypeChain); n != 1 {
		t.Fatalf("expecting 1 test chain containers, got %v", n)
	}

	do = def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.Operations.ContainerNumber = 1
	do.ChainName = chainName
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if n := util.HowManyContainersRunning("fake", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 fake service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("fake", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 fake data containers, got %v", n)
	}
	if n := util.HowManyContainersRunning("sham", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 sham service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting("sham", def.TypeData); n != 1 {
		t.Fatalf("expecting 1 sham data containers, got %v", n)
	}

	// [pv]: second service doesn't reference the chain.
	links := tests.Links("fake", def.TypeService, 1)
	if len(links) != 2 || (!strings.Contains(links[1], chainName) && !strings.Contains(links[0], chainName)) {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

//------------------------------------------------------------------
// testing utils

func testStartChain(t *testing.T, chain string) {
	do := def.NowDo()
	do.Name = chain
	do.Operations.ContainerNumber = 1
	do.Operations.PublishAllPorts = true
	log.WithField("=>", do.Name).Info("Starting chain (from tests)")
	if e := StartChain(do); e != nil {
		log.Error(e)
		tests.IfExit(nil)
	}
	testExistAndRun(t, chain, true, true)
}

func testKillChain(t *testing.T, chain string) {
	// log.SetLoggers(2, os.Stdout, os.Stderr)
	testExistAndRun(t, chain, true, true)

	do := def.NowDo()
	do.Operations.Args, do.Rm, do.RmD = []string{"keys"}, true, true
	log.WithField("=>", do.Name).Info("Stopping service (from tests)")
	if e := services.KillService(do); e != nil {
		tests.IfExit(e)
	}

	do = def.NowDo()
	do.Name, do.Rm, do.RmD = chain, true, true
	log.WithField("=>", do.Name).Info("Stopping chain (from tests)")
	if e := KillChain(do); e != nil {
		tests.IfExit(e)
	}
	testExistAndRun(t, chain, false, false)
}

func testExistAndRun(t *testing.T, chainName string, toExist, toRun bool) {
	if tests.TestExistAndRun(chainName, "chains", 1, toExist, toRun) {
		tests.IfExit(nil)
	}
}

func testNewChain(chain string) {
	do := def.NowDo()
	do.GenesisFile = filepath.Join(common.ChainsPath, "default", "genesis.json")
	do.Name = chain
	do.Operations.ContainerNumber = 1
	do.Operations.PublishAllPorts = true
	log.WithField("=>", chain).Info("Creating chain")
	tests.IfExit(NewChain(do))
}

func removeChainContainer(t *testing.T, chainID string, cNum int) {
	do := def.NowDo()
	do.Name = chainID
	do.Rm, do.Force, do.RmD = true, true, true
	do.Operations.ContainerNumber = cNum
	if err := KillChain(do); err != nil {
		tests.IfExit(err)
	}
}

func runContainer(t *testing.T, ops *def.Operation) []byte {
	oldWriter := config.GlobalConfig.Writer
	newWriter := new(bytes.Buffer)
	config.GlobalConfig.Writer = newWriter

	b, err := perform.DockerRunData(ops, nil)
	if err != nil {
		tests.IfExit(err)
	}
	log.WithFields(log.Fields{
		"=>":   ops.DataContainerName,
		"args": ops.Args,
	}).Debug("Container ran (from tests)")
	config.GlobalConfig.Writer = oldWriter
	return b
}

func ensureTomlValue(t *testing.T, s, field, value string) bool {
	if strings.Contains(s, field) {
		if !strings.Contains(s, value) {
			tests.IfExit(fmt.Errorf("Expected %s to be %s. Got: %s", field, value, s))
		}
		return true
	}
	return false
}

func trimResult(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\n")
	spl := strings.Split(s, "\n")
	for i, t := range spl {
		t = strings.TrimSpace(t)
		spl[i] = t
	}
	return strings.Trim(strings.Join(spl, "\n"), "\n")
}

func layTestChainToml(name string) {
	chain := loaders.MockChainDefinition(name, name, false, 1)

	// write the chain definition file ...
	fileName := filepath.Join(erisDir, "chains", name)
	if _, err := os.Stat(fileName); err != nil {
		if err = WriteChainDefinitionFile(chain, fileName); err != nil {
			panic(fmt.Errorf("error writing chain definition to file: %v", err))
		}
	}
}
