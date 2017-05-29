package chains

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/services"
	"github.com/monax/monax/testutil"
	"github.com/monax/monax/util"
	"github.com/monax/monax/version"
)

var chainName = "test-chain"

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	testutil.IfExit(testutil.Init(testutil.Pull{
		Images: []string{"data", "db", "keys"},
	}))

	exitCode := m.Run()

	testutil.IfExit(testutil.TearDown())

	os.Exit(exitCode)
}

func TestStartChain(t *testing.T) {
	defer testutil.RemoveAllContainers()

	create(t, chainName)

	if !util.Running(definitions.TypeChain, chainName) {
		t.Fatalf("expecting chain running")
	}
	if !util.Exists(definitions.TypeData, chainName) {
		t.Fatalf("expecting dependent data container exists")
	}

	kill(t, chainName)
	if util.Running(definitions.TypeChain, chainName) {
		t.Fatalf("expecting chain doesn't run")
	}
	if util.Exists(definitions.TypeData, chainName) {
		t.Fatalf("expecting data container doesn't exist")
	}
}

func TestStartStopStartChain(t *testing.T) {
	defer testutil.RemoveAllContainers()

	// make a chain
	create(t, chainName)
	if !util.Running(definitions.TypeChain, chainName) {
		t.Fatalf("expecting chain running")
	}
	if !util.Exists(definitions.TypeData, chainName) {
		t.Fatalf("expecting data container exists")
	}

	// stop it
	stop(t, chainName)
	if util.Running(definitions.TypeChain, chainName) {
		t.Fatalf("expecting chain doesn't run")
	}
	if !util.Exists(definitions.TypeData, chainName) {
		t.Fatalf("expecting data container exists")
	}

	// start it back up again
	start(t, chainName)
	if !util.Running(definitions.TypeChain, chainName) {
		t.Fatalf("expecting chain running")
	}
	if !util.Exists(definitions.TypeData, chainName) {
		t.Fatalf("expecting data container exists")
	}

	// kill it
	kill(t, chainName)
	if util.Running(definitions.TypeChain, chainName) {
		t.Fatalf("expecting chain doesn't run")
	}
	if util.Exists(definitions.TypeData, chainName) {
		t.Fatalf("expecting data container doesn't exist")
	}
}

func TestExecChain(t *testing.T) {
	defer testutil.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	do := definitions.NowDo()
	do.Name = chainName
	do.Operations.Args = []string{"ls", config.MonaxContainerRoot}
	buf, err := ExecChain(do)
	if err != nil {
		t.Fatalf("expected chain to execute, got %v", err)
	}

	if dir := "chains"; !strings.Contains(buf.String(), dir) {
		t.Fatalf("expected to find %q dir in monax root", dir)
	}
}

func TestExecChainBadCommandLine(t *testing.T) {
	defer testutil.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	do := definitions.NowDo()
	do.Name = chainName
	do.Operations.Args = strings.Fields("bad command line")
	if _, err := ExecChain(do); err == nil {
		t.Fatalf("expected chain to fail")
	}
}

func TestChainsNewDirGenesis(t *testing.T) {
	defer testutil.RemoveAllContainers()

	const chain = "test-dir-gen"
	create(t, chain)
	defer kill(t, chain)

	args := []string{"cat", fmt.Sprintf("/home/monax/.monax/chains/%s/genesis.json", chain)}
	if out := exec(t, chain, args); !strings.Contains(out, chain) {
		t.Fatalf("expected chain_id to be equal to chain name in genesis file, got %v", out)
	}
}

func TestChainsNewConfig(t *testing.T) {
	defer testutil.RemoveAllContainers()

	const chain = "test-config-new"
	create(t, chain)
	defer kill(t, chain)

	args := []string{"cat", fmt.Sprintf("/home/monax/.monax/chains/%s/config.toml", chain)}
	if out := exec(t, chain, args); !strings.Contains(out, "moniker") {
		t.Fatalf("expected the config file to contain an expected string, got %v", out)
	}
}

// chains start (--init-dir) should import the priv_validator.json (available in mint form)
// into monax-keys (available in monax form) so it can be used by the rest
// of the platform
func TestChainsNewKeysImported(t *testing.T) {
	defer testutil.RemoveAllContainers()

	const chain = "test-config-keys"
	create(t, chain)
	defer kill(t, chain)

	if !util.Running(definitions.TypeChain, chain) {
		t.Fatalf("expecting chain running")
	}

	keysOut, err := services.ExecHandler("keys", []string{"ls", "/home/monax/.monax/keys/data"})
	if err != nil {
		t.Fatalf("expecting to list keys, got %v", err)
	}

	keysOutString0 := strings.Fields(strings.TrimSpace(keysOut.String()))[0]

	args := []string{"cat", fmt.Sprintf("/home/monax/.monax/keys/data/%s/%s", keysOutString0, keysOutString0)}

	keysOut1, err := services.ExecHandler("keys", args)
	if err != nil {
		t.Fatalf("expecting to cat keys, got %v", err)
	}

	keysOutString1 := strings.Fields(strings.TrimSpace(keysOut1.String()))[0]

	if !strings.Contains(keysOutString1, keysOutString0) { // keysOutString0 is the substring (addr only)
		t.Fatalf("keys do not match, key0: %v, key1: %v", keysOutString0, keysOutString1)
	}
}

func TestLogsChain(t *testing.T) {
	defer testutil.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	do := definitions.NowDo()
	do.Name = chainName
	do.Follow = false
	do.Tail = "all"
	if err := LogsChain(do); err != nil {
		t.Fatalf("failed to fetch container logs")
	}
}

func TestInspectChain(t *testing.T) {
	defer testutil.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	do := definitions.NowDo()
	do.Name = chainName
	do.Operations.Args = []string{"name"}
	if err := InspectChain(do); err != nil {
		t.Fatalf("expected chain to be inspected, got %v", err)
	}
}

func TestRmChain(t *testing.T) {
	defer testutil.RemoveAllContainers()

	create(t, chainName)

	do := definitions.NowDo()
	do.Operations.Args, do.Rm, do.RmD = []string{"keys"}, true, true
	if err := services.KillService(do); err != nil {
		t.Fatalf("expected service to be stopped, got %v", err)
	}

	kill(t, chainName) // implements RemoveChain
	if util.Exists(definitions.TypeChain, chainName) {
		t.Fatalf("expecting chain not running")
	}
}

func TestServiceLinkKeys(t *testing.T) {
	defer testutil.RemoveAllContainers()

	const chain = "chain-test-keys"
	create(t, chain)
	defer kill(t, chain)

	if !util.Exists(definitions.TypeChain, chain) {
		t.Fatalf("expecting test chain running")
	}

	do := definitions.NowDo()
	do.Operations.Args = []string{"keys"}
	do.ChainName = chain
	if err := services.StartService(do); err != nil {
		t.Fatalf("expecting service to start, got %v", err)
	}

	if !util.Running(definitions.TypeService, "keys") {
		t.Fatalf("expecting keys service running")
	}

	links := testutil.Links("keys", definitions.TypeService)
	if len(links) != 0 {
		t.Fatalf("expected service links be empty, got %v", links)
	}
}

func create(t *testing.T, chain string) {
	doMake := definitions.NowDo()
	doMake.Name = chain
	doMake.ChainType = "simplechain"
	// added because this is now set by the flag
	doMake.ChainImageName = path.Join(version.DefaultRegistry, version.ImageDB)
	if err := MakeChain(doMake); err != nil {
		t.Fatalf("expected a chain to be made, got %v", err)
	}

	do := definitions.NowDo()
	do.Name = chain
	do.Operations.PublishAllPorts = true
	do.Path = filepath.Join(config.ChainsPath, chain, fmt.Sprintf("%s_full_000", chain)) // --init-dir
	if err := StartChain(do); err != nil {
		t.Fatalf("expected a new chain to be created, got %v", err)
	}
}

func start(t *testing.T, chain string) {
	do := definitions.NowDo()
	do.Name = chain
	do.Operations.PublishAllPorts = true
	if err := StartChain(do); err != nil {
		t.Fatalf("starting chain %v failed: %v", chain, err)
	}
}

func stop(t *testing.T, chain string) {
	do := definitions.NowDo()
	do.Name = chain
	do.Force = true
	if err := StopChain(do); err != nil {
		t.Fatalf("stopping chain %v failed: %v", chain, err)
	}
}

func kill(t *testing.T, chain string) {
	do := definitions.NowDo()
	do.Operations.Args, do.Rm, do.RmD = []string{"keys"}, true, true
	if err := services.KillService(do); err != nil {
		t.Fatalf("killing keys service failed: %v", err)
	}

	do = definitions.NowDo()
	do.Name, do.RmHF, do.RmD, do.Force = chain, true, true, true
	if err := RemoveChain(do); err != nil {
		t.Fatalf("killing chain failed: %v", err)
	}
}

func exec(t *testing.T, chain string, args []string) string {
	do := definitions.NowDo()
	do.Name = chain
	do.Operations.Args = args
	buf, err := ExecChain(do)
	if err != nil {
		log.Error(buf)
		t.Fatalf("expected chain to execute, got %v", err)
	}

	return buf.String()
}
