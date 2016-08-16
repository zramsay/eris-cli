package kaihei

import (
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/tests"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit(tests.ConnectAndPull))
	//mockChainDefinitionFile(chainName)

	exitCode := m.Run()

	tests.IfExit(tests.TestsTearDown())

	os.Exit(exitCode)
}

func TestStartAndStopEris(t *testing.T) {
	defer tests.RemoveAllContainers()

	var (
		chainZero = "chain-zero"
		chainOne  = "chain-one"
	)

	createChain(t, chainZero)
	createChain(t, chainOne)

	servicesToStart := []string{"keys, ipfs"} // ideally logs but then the tests would need to pull them in

	startServices(t, servicesToStart)

	// check these four things are running...

	doShut := definitions.NowDo()
	if err := ShutUpEris(doShut); err != nil {
		t.Fatalf("expected to shut down eris, got %v", err)
	}

	// check that everything is shut down

	doStart := definitions.NowDo()
	if err := StartUpEris(doStart); err != nil {
		t.Fatalf("expected to start up eris, got %v", err)
	}

	// check these four things are running...

}

func startServices(t *testing.T, services []string) {
	doStart := definitions.NowDo()
	doStart.Operations.Args = services
	if err := services.StartServices(doStart); err != nil {
		t.Fatalf("expected services to start, got %v", err)
	}
}

func createChain(t *testing.T, chain string) {
	do := def.NowDo()
	do.ConfigFile = filepath.Join(common.ChainsPath, "default", "config.toml")
	do.Name = chain
	do.Operations.PublishAllPorts = true
	if err := NewChain(do); err != nil {
		t.Fatalf("expected a new chain to be created, got %v", err)
	}
}
