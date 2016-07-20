package chains

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/tests"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
)

var (
	erisDir      = filepath.Join(os.TempDir(), "eris")
	chainName    = "test-chain"
	chainNameFix = "test-chain-fix"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit(tests.ConnectAndPull))

	exitCode := m.Run()

	tests.IfExit(tests.TestsTearDown())

	os.Exit(exitCode)
}

func TestStartChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)

	if !util.Running(def.TypeChain, chainName) {
		t.Fatalf("expecting chain running")
	}
	if !util.Exists(def.TypeData, chainName) {
		t.Fatalf("expecting dependent data container exists")
	}

	kill(t, chainName)
	if util.Running(def.TypeChain, chainName) {
		t.Fatalf("expecting chain doesn't run")
	}
	if util.Exists(def.TypeData, chainName) {
		t.Fatalf("expecting data container doesn't exist")
	}
}

func TestRestartChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)

	if !util.Running(def.TypeChain, chainName) {
		t.Fatalf("expecting chain running")
	}
	if !util.Exists(def.TypeData, chainName) {
		t.Fatalf("expecting data container exists")
	}

	//kill(t, chainName)
	do := def.NowDo()
	do.Name = chainName
	do.Rm = true
	do.RmD = true
	do.Volumes = true
	do.Force = true

	if err := KillChain(do); err != nil {
		t.Fatalf("expected chain to stop, got %v", err)
	}

	if util.Running(def.TypeChain, chainName) {
		t.Fatalf("expecting chain doesn't run")
	}
	if util.Exists(def.TypeData, chainName) {
		t.Fatalf("expecting data container doesn't exist")
	}

	start(t, chainName)
	if !util.Running(def.TypeChain, chainName) {
		t.Fatalf("expecting chain running")
	}
	if !util.Exists(def.TypeData, chainName) {
		t.Fatalf("expecting data container exists")
	}

	kill(t, chainName)
	if util.Running(def.TypeChain, chainName) {
		t.Fatalf("expecting chain doesn't run")
	}
	if util.Exists(def.TypeData, chainName) {
		t.Fatalf("expecting data container doesn't exist")
	}
}

func TestExecChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Operations.Args = []string{"ls", common.ErisContainerRoot}
	buf, err := ExecChain(do)
	if err != nil {
		t.Fatalf("expected chain to execute, got %v", err)
	}

	if dir := "chains"; !strings.Contains(buf.String(), dir) {
		t.Fatalf("expected to find %q dir in eris root", dir)
	}
}

func TestExecChainBadCommandLine(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Operations.Args = strings.Fields("bad command line")
	if _, err := ExecChain(do); err == nil {
		t.Fatalf("expected chain to fail")
	}
}

func TestCatChainLocalConfig(t *testing.T) {
	defer tests.RemoveAllContainers()

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer = buf

	create(t, chainName)
	defer kill(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Type = "toml"
	if err := CatChain(do); err != nil {
		t.Fatalf("expected getting a local config to succeed, got %v", err)
	}

	if buf.String() != tests.FileContents(filepath.Join(erisDir, "chains", chainName, chainName+".toml")) {
		t.Fatalf("expected the local config file to match, got %v", buf.String())
	}
}

func TestCatChainContainerConfig(t *testing.T) {
	defer tests.RemoveAllContainers()

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer = buf

	const chain = "test-cat-cont-config"

	create(t, chain)
	defer kill(t, chain)

	do := def.NowDo()
	do.Name = chain
	do.Type = "config"
	if err := CatChain(do); err != nil {
		t.Fatalf("expected getting a local config to succeed, got %v", err)
	}

	if !strings.Contains(buf.String(), "defaulttester.com") {
		t.Fatalf("expected the config file to contain an expected string, got %v", buf.String())
	}
}

func TestCatChainContainerGenesis(t *testing.T) {
	defer tests.RemoveAllContainers()

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer = buf

	create(t, chainName)
	defer kill(t, chainName)

	do := def.NowDo()
	do.Type = "genesis"
	do.Name = chainName
	if err := CatChain(do); err != nil {
		t.Fatalf("expected getting a local config to succeed, got %v", err)
	}

	if !strings.Contains(buf.String(), "accounts") || !strings.Contains(buf.String(), "validators") {
		t.Fatalf("expected the genesis file to contain expected strings, got %v", buf.String())
	}
}

func TestChainsNewDirGenesis(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-dir-gen"
	create(t, chain)
	defer kill(t, chain)

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/chains/%s/genesis.json", chain)}
	if out := exec(t, chain, args); !strings.Contains(out, chain) {
		t.Fatalf("expected chain_id to be equal to chain name in genesis file, got %v", out)
	}
}

func TestChainsNewConfig(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-config-new"
	create(t, chain)
	defer kill(t, chain)

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/chains/%s/config.toml", chain)}
	if out := exec(t, chain, args); !strings.Contains(out, "defaulttester.com") {
		t.Fatalf("expected the config file to contain an expected string, got %v", out)
	}
}

// chains new should import the priv_validator.json (available in mint form)
// into eris-keys (available in eris form) so it can be used by the rest
// of the platform
func TestChainsNewKeysImported(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-config-keys"
	create(t, chain)
	defer kill(t, chain)

	if !util.Running(def.TypeChain, chain) {
		t.Fatalf("expecting chain running")
	}

	keysOut, err := services.ExecHandler("keys", []string{"ls", "/home/eris/.eris/keys/data"})
	if err != nil {
		t.Fatalf("expecting to list keys, got %v", err)
	}

	keysOutString := strings.Split(util.TrimString(keysOut.String()), "\n")[0]

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/keys/data/%s/%s", keysOutString, keysOutString)}

	if out := exec(t, chain, args); !strings.Contains(out, keysOutString) {
		t.Fatalf("expected to find keys in container, got %v", out)
	}
}

func TestLogsChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Follow = false
	do.Tail = "all"
	if err := LogsChain(do); err != nil {
		t.Fatalf("failed to fetch container logs")
	}
}

func TestUpdateChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Pull = false
	do.Operations.PublishAllPorts = true
	if err := UpdateChain(do); err != nil {
		t.Fatalf("expected chain to update, got %v", err)
	}

	if !util.Running(def.TypeChain, chainName) {
		t.Fatalf("expecting chain running")
	}
}

func TestInspectChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Operations.Args = []string{"name"}
	if err := InspectChain(do); err != nil {
		t.Fatalf("expected chain to be inspected, got %v", err)
	}
}

func TestRmChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	do := def.NowDo()
	do.Operations.Args, do.Rm, do.RmD = []string{"keys"}, true, true
	if err := services.KillService(do); err != nil {
		t.Fatalf("expected service to be stopped, got %v", err)
	}

	do = def.NowDo()
	do.Name, do.Rm, do.RmD = chainName, false, false
	if err := KillChain(do); err != nil {
		t.Fatalf("expected chain to be stopped, got %v", err)
	}
	if !util.Exists(def.TypeChain, chainName) {
		t.Fatalf("expecting chain running")
	}

	do = def.NowDo()
	do.Name = chainName
	do.RmD = true
	if err := RemoveChain(do); err != nil {
		t.Fatalf("expected chain to be removed, got %v", err)
	}
	if util.Exists(def.TypeChain, chainName) {
		t.Fatalf("expecting chain to be removed")
	}
}

func TestServiceLinkNoChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	if err := tests.FakeServiceDefinition("fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "`+path.Join(config.GlobalConfig.Config.ERIS_REG_DEF, config.GlobalConfig.Config.ERIS_IMG_IPFS)+`"
data_container = true
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	do := def.NowDo()
	do.Operations.Args = []string{"fake"}
	if err := services.StartService(do); err == nil {
		t.Fatalf("expect start service to fail, got nil")
	}
}

func TestServiceLinkBadChain(t *testing.T) {
	defer tests.RemoveAllContainers()

	if err := tests.FakeServiceDefinition("fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "`+path.Join(config.GlobalConfig.Config.ERIS_REG_DEF, config.GlobalConfig.Config.ERIS_IMG_IPFS)+`"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	do := def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.ChainName = "non-existent-chain"
	if err := services.StartService(do); err == nil {
		t.Fatalf("expect start service to fail, got nil")
	}
}

func TestServiceLinkBadChainWithoutChainInDefinition(t *testing.T) {
	defer tests.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	if err := tests.FakeServiceDefinition("fake", `
[service]
name = "fake"
image = "`+path.Join(config.GlobalConfig.Config.ERIS_REG_DEF, config.GlobalConfig.Config.ERIS_IMG_IPFS)+`"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	do := def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.ChainName = "non-existent-chain"

	// [pv]: is this a bug? the service which doesn't have a
	// "chain" in its definition file doesn't care about linking at all.
	if err := services.StartService(do); err != nil {
		t.Fatalf("expect service to start, got %v", err)
	}

	if !util.Running(def.TypeService, "fake") {
		t.Fatalf("expecting fake service running")
	}
	if util.Exists(def.TypeData, "fake") {
		t.Fatalf("expecting fake data container doesn't exist")
	}
}

func TestServiceLink(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-chain-link"
	create(t, chain)
	defer kill(t, chain)

	if err := tests.FakeServiceDefinition("fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "`+path.Join(config.GlobalConfig.Config.ERIS_REG_DEF, config.GlobalConfig.Config.ERIS_IMG_KEYS)+`"
data_container = false
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if !util.Exists(def.TypeChain, chain) {
		t.Fatalf("expecting fake chain container")
	}
	if util.Running(def.TypeService, "fake") {
		t.Fatalf("expecting fake service running")
	}
	if util.Exists(def.TypeData, "fake") {
		t.Fatalf("expecting fake data container doesn't exist")
	}

	do := def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.ChainName = chain
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if !util.Running(def.TypeService, "fake") {
		t.Fatalf("expecting fake service not running")
	}
	if util.Exists(def.TypeData, "fake") {
		t.Fatalf("expecting fake data container doesn't exist")
	}

	links := tests.Links("fake", def.TypeService)
	if len(links) != 1 || !strings.Contains(links[0], "/fake") {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

func TestServiceLinkWithDataContainer(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-chain-data-container"

	create(t, chain)
	defer kill(t, chain)

	if err := tests.FakeServiceDefinition("fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "`+path.Join(config.GlobalConfig.Config.ERIS_REG_DEF, config.GlobalConfig.Config.ERIS_IMG_IPFS)+`"
data_container = true
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if !util.Exists(def.TypeChain, chain) {
		t.Fatalf("expecting test chain container")
	}
	if util.Running(def.TypeService, "fake") {
		t.Fatalf("expecting fake service not running")
	}
	if util.Exists(def.TypeData, "fake") {
		t.Fatalf("expecting fake data container doesn't exist")
	}

	do := def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.ChainName = chain
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if !util.Running(def.TypeService, "fake") {
		t.Fatalf("expecting fake service running")
	}
	if !util.Exists(def.TypeData, "fake") {
		t.Fatalf("expecting fake data container exists")
	}

	links := tests.Links("fake", def.TypeService)
	if len(links) != 1 || !strings.Contains(links[0], "/fake") {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

func TestServiceLinkLiteral(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-chain-literal"

	create(t, chain)
	defer kill(t, chain)

	if err := tests.FakeServiceDefinition("fake", `
chain = "`+chain+`:fake"

[service]
name = "fake"
image = "`+path.Join(config.GlobalConfig.Config.ERIS_REG_DEF, config.GlobalConfig.Config.ERIS_IMG_KEYS)+`"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if !util.Exists(def.TypeChain, chain) {
		t.Fatalf("expecting fake chain container")
	}
	if util.Running(def.TypeService, "fake") {
		t.Fatalf("expecting fake service not running")
	}
	if util.Exists(def.TypeData, "fake") {
		t.Fatalf("expecting fake data container doesn't exist")
	}

	do := def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.ChainName = chain
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if !util.Running(def.TypeService, "fake") {
		t.Fatalf("expecting fake service running")
	}
	if util.Exists(def.TypeData, "fake") {
		t.Fatalf("expecting fake data container exists")
	}

	links := tests.Links("fake", def.TypeService)
	if len(links) != 1 || !strings.Contains(links[0], "/fake") {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

func TestServiceLinkBadLiteral(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-chain-bad-literal"

	create(t, chain)
	defer kill(t, chain)

	if err := tests.FakeServiceDefinition("fake", `
chain = "blah-blah:blah"

[service]
name = "fake"
image = "`+path.Join(config.GlobalConfig.Config.ERIS_REG_DEF, config.GlobalConfig.Config.ERIS_IMG_IPFS)+`"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if !util.Running(def.TypeChain, chain) {
		t.Fatalf("expecting test chain container")
	}

	do := def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.ChainName = chain
	// [pv]: probably a bug. Bad literal chain link in a definition
	// file doesn't affect the service start. Links is not nil.
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	links := tests.Links("fake", def.TypeService)
	if len(links) != 1 || !strings.Contains(links[0], "/blah") {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

func TestServiceLinkChainedService(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "test-chained-service"

	if err := tests.FakeServiceDefinition("fake", `
chain = "$chain:fake"

[service]
name = "fake"
image = "`+path.Join(config.GlobalConfig.Config.ERIS_REG_DEF, config.GlobalConfig.Config.ERIS_IMG_KEYS)+`"

[dependencies]
services = [ "sham" ]
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if err := tests.FakeServiceDefinition("sham", `
chain = "$chain:sham"

[service]
name = "sham"
image = "`+path.Join(config.GlobalConfig.Config.ERIS_REG_DEF, config.GlobalConfig.Config.ERIS_IMG_KEYS)+`"
data_container = true
`); err != nil {
		t.Fatalf("can't create a sham service definition: %v", err)
	}

	if util.Running(def.TypeChain, chain) {
		t.Fatalf("expecting test chain container doesn't run")
	}

	create(t, chain) // [zr] why was the NewChain here?
	defer kill(t, chain)

	if !util.Exists(def.TypeChain, chain) {
		t.Fatalf("expecting test chain container exists")
	}

	do := def.NowDo()
	do.Operations.Args = []string{"fake"}
	do.ChainName = chain
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if !util.Running(def.TypeService, "fake") {
		t.Fatalf("expecting fake service running")
	}
	if util.Exists(def.TypeData, "fake") {
		t.Fatalf("expecting fake data container doesn't exist")
	}
	if !util.Running(def.TypeService, "sham") {
		t.Fatalf("expecting sham service running")
	}
	if !util.Exists(def.TypeData, "sham") {
		t.Fatalf("expecting sham data container exist")
	}

	// [pv]: second service doesn't reference the chain.
	links := tests.Links("fake", def.TypeService)

	if len(links) != 2 || !strings.Contains(strings.Join(links, " "), "/fake") || !strings.Contains(strings.Join(links, " "), "/sham") {
		t.Fatalf("expected service be linked to a test chain, got %v", links)
	}
}

func TestServiceLinkKeys(t *testing.T) {
	defer tests.RemoveAllContainers()

	const chain = "chain-test-keys"
	create(t, chain)
	defer kill(t, chain)

	if !util.Exists(def.TypeChain, chain) {
		t.Fatalf("expecting test chain running")
	}

	do := def.NowDo()
	do.Operations.Args = []string{"keys"}
	do.ChainName = chain
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if !util.Running(def.TypeService, "keys") {
		t.Fatalf("expecting keys service running")
	}

	links := tests.Links("keys", def.TypeService)
	if len(links) != 0 {
		t.Fatalf("expected service links be empty, got %v", links)
	}
}

func create(t *testing.T, chain string) {
	doMake := def.NowDo()
	doMake.Name = chain
	doMake.ChainType = "simplechain"
	if err := MakeChain(doMake); err != nil {
		t.Fatalf("expected a chain to be made, got %v", err)
	}

	do := def.NowDo()
	do.ConfigFile = filepath.Join(common.ChainsPath, "default", "config.toml")
	do.Name = chain
	do.Path = filepath.Join(common.ChainsPath, chain)
	do.Operations.PublishAllPorts = true
	if err := StartChain(do); err != nil {
		t.Fatalf("expected a new chain to be created, got %v", err)
	}
}

// [zr] only diff btwn this & the above command is do.ConfigFile = ...
func start(t *testing.T, chain string) {
	do := def.NowDo()
	do.Name = chain
	do.Operations.PublishAllPorts = true
	if err := StartChain(do); err != nil {
		t.Fatalf("starting chain %v failed: %v", chain, err)
	}
}

func kill(t *testing.T, chain string) {
	do := def.NowDo()
	do.Operations.Args, do.Rm, do.RmD = []string{"keys"}, true, true
	if err := services.KillService(do); err != nil {
		t.Fatalf("killing keys service failed: %v", err)
	}

	do = def.NowDo()
	do.Name, do.RmHF, do.RmD, do.Force = chain, true, true, true
	if err := RemoveChain(do); err != nil {
		t.Fatalf("killing chain failed: %v", err)
	}
}

func exec(t *testing.T, chain string, args []string) string {
	do := def.NowDo()
	do.Name = chain
	do.Operations.Args = args
	buf, err := ExecChain(do)
	if err != nil {
		log.Error(buf)
		t.Fatalf("expected chain to execute, got %v", err)
	}

	return buf.String()
}

func mockChainDefinitionFile(name string) error {
	definition := loaders.MockChainDefinition(name, name)

	return WriteChainDefinitionFile(definition, filepath.Join(erisDir, "chains", name, name))
}
