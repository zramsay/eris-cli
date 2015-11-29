package perform

import (
	"os"
	"testing"

	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	tests "github.com/eris-ltd/eris-cli/testutils"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	logLevel = 0
	// logLevel = 1
	// logLevel = 3

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	tests.IfExit(tests.TestsInit("perform"))

	tests.RemoveAllContainers()

	exitCode := m.Run()

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		tests.IfExit(tests.TestsTearDown())
	}

	os.Exit(exitCode)
}

func TestCreateDataSimple(t *testing.T) {
	const (
		name   = "testdata"
		number = 199
	)

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	ops := loaders.LoadDataDefinition(name, number)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 containers, got %v", n)
	}

	// Try to create a duplicate.
	if err := DockerCreateData(ops); err == nil {
		t.Fatalf("expected an error, got nil")
	}

	tests.RemoveAllContainers()
}

func TestRunDataSimple(t *testing.T) {
	const (
		name   = "testdata"
		number = 199
	)

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	ops := loaders.LoadDataDefinition(name, number)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	ops.Args = []string{"uptime"}
	if _, err := DockerRunData(ops, nil); err != nil {
		t.Fatalf("expected data successfully run, got %v", err)
	}

	tests.RemoveAllContainers()
}

func TestRunDataBadCommandLine(t *testing.T) {
	const (
		name   = "testdata"
		number = 199
	)

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	ops := loaders.LoadDataDefinition(name, number)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	ops.Args = []string{"/bad/command/line"}
	if _, err := DockerRunData(ops, nil); err == nil {
		t.Fatalf("expected command line error, got nil")
	}

	tests.RemoveAllContainers()
}

func TestExecDataSimple(t *testing.T) {
	const (
		name   = "testdata"
		number = 199
	)

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	ops := loaders.LoadDataDefinition(name, number)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	ops.Args = []string{"uptime"}
	if err := DockerExecData(ops, nil); err != nil {
		t.Fatalf("expected data successfully run, got %v", err)
	}

	tests.RemoveAllContainers()
}

func TestExecDataBadCommandLine(t *testing.T) {
	const (
		name   = "testdata"
		number = 199
	)

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	ops := loaders.LoadDataDefinition(name, number)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	ops.Args = []string{"/bad/command/line"}
	if err := DockerExecData(ops, nil); err == nil {
		t.Fatalf("expected command line error, got nil")
	}

	tests.RemoveAllContainers()
}

func TestRunServiceSimple(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 dependent data container, got %v", n)
	}

	tests.RemoveAllContainers()
}

func TestRunServiceNoDataContainer(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Service.AutoData = false
	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting no dependent data containers, got %v", n)
	}

	tests.RemoveAllContainers()
}

func TestRunServiceTwoServices(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv1, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("1. could not load service definition %v", err)
	}

	if err := DockerRunService(srv1.Service, srv1.Operations); err != nil {
		t.Fatalf("1. expected service container created, got %v", err)
	}

	srv2, err := loaders.LoadServiceDefinition(name, true, number+1)
	if err != nil {
		t.Fatalf("2. could not load service definition %v", err)
	}

	if err := DockerRunService(srv2.Service, srv2.Operations); err == nil {
		t.Fatalf("2. expected service failure due to occupied ports, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 2 {
		t.Fatalf("expecting 2 dependent data container, got %v", n)
	}

	tests.RemoveAllContainers()
}

func TestRunServiceTwoServicesPublishedPorts(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv1, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("1. could not load service definition %v", err)
	}

	if err := DockerRunService(srv1.Service, srv1.Operations); err != nil {
		t.Fatalf("1. expected service container created, got %v", err)
	}

	srv2, err := loaders.LoadServiceDefinition(name, true, number+1)
	if err != nil {
		t.Fatalf("2. could not load service definition %v", err)
	}

	srv2.Operations.PublishAllPorts = true
	if err := DockerRunService(srv2.Service, srv2.Operations); err != nil {
		t.Fatalf("2. expected service container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 2 {
		t.Fatalf("expecting 2 services running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 2 {
		t.Fatalf("expecting 2 dependent data container, got %v", n)
	}

	tests.RemoveAllContainers()
}

func TestExecServiceSimple(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Interactive = true
	srv.Operations.Args = []string{"uptime"}
	if err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 dependent data container, got %v", n)
	}

	tests.RemoveAllContainers()
}

func TestExecServiceTwice(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Interactive = true
	srv.Operations.Args = []string{"uptime"}
	if err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("1. expected service container created, got %v", err)
	}

	if err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("2. expected service container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 dependent data container, got %v", n)
	}

	tests.RemoveAllContainers()
}

func TestExecServiceTwiceWithoutData(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Service.AutoData = false
	srv.Operations.Interactive = true
	srv.Operations.Args = []string{"uptime"}
	if err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("1. expected service container created, got %v", err)
	}

	if err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("2. expected service container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 dependent data containers, got %v", n)
	}

	tests.RemoveAllContainers()
}

func TestExecServiceBadCommandLine(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Interactive = false
	srv.Operations.Args = []string{"/bad/command/line"}
	if err := DockerExecService(srv.Service, srv.Operations); err == nil {
		t.Fatalf("expected failure, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 dependent data container, got %v", n)
	}

	tests.RemoveAllContainers()
}

func TestExecServiceNonInteractive(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Interactive = false
	srv.Operations.Args = []string{"uptime"}
	if err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 dependent data container, got %v", n)
	}

	tests.RemoveAllContainers()
}

func TestExecServiceAfterRunService(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	srv.Operations.Interactive = true
	srv.Operations.Args = []string{"uptime"}
	if err := DockerExecService(srv.Service, srv.Operations); err == nil {
		t.Fatalf("expected failure due to unpublished ports, got %v", err)
	}

	tests.RemoveAllContainers()
}

func TestExecServiceAfterRunServiceWithPublishedPorts1(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.PublishAllPorts = true
	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	srv.Operations.Interactive = true
	srv.Operations.Args = []string{"uptime"}
	if err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected exec container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 dependent data container, got %v", n)
	}

	tests.RemoveAllContainers()
}

func TestExecServiceAfterRunServiceWithPublishedPorts2(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	srv.Operations.PublishAllPorts = true
	srv.Operations.Interactive = true
	srv.Operations.Args = []string{"uptime"}
	if err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected exec container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 dependent data container, got %v", n)
	}

	tests.RemoveAllContainers()
}

func TestContainerExistsSimple(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if _, exists := ContainerExists(srv.Operations); exists != true {
		t.Fatalf("expecting service container existing, got false")
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if _, exists := ContainerExists(srv.Operations); exists != true {
		t.Fatalf("expecting data container existing, got false")
	}

	tests.RemoveAllContainers()
}

func TestContainerExistsBadName(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.SrvContainerName = "some-random-name"
	if _, exists := ContainerExists(srv.Operations); exists != false {
		t.Fatalf("expecting service container not existing, got true")
	}

	tests.RemoveAllContainers()
}

func TestContainerExistsAfterRemove(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if _, exists := ContainerExists(srv.Operations); exists == false {
		t.Fatalf("expecting service container exists, got false")
	}

	tests.RemoveContainer(name, def.TypeService, number)

	if _, exists := ContainerExists(srv.Operations); exists == true {
		t.Fatalf("expecting service container not existing after remove, got true")
	}

	tests.RemoveAllContainers()
}

func TestContainerRunningSimple(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if _, exists := ContainerRunning(srv.Operations); exists == false {
		t.Fatalf("expecting service container running, got false")
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if _, exists := ContainerRunning(srv.Operations); exists == true {
		t.Fatalf("expecting data container not running, got true")
	}

	tests.RemoveAllContainers()
}

func TestContainerRunningBadName(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if _, exists := ContainerRunning(srv.Operations); exists == false {
		t.Fatalf("expecting service container running, got false")
	}

	srv.Operations.SrvContainerName = "random-bad-name"
	if _, exists := ContainerRunning(srv.Operations); exists == true {
		t.Fatalf("expecting data container not running, got true")
	}

	tests.RemoveAllContainers()
}

func TestContainerRunningAfterRemove(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if _, exists := ContainerRunning(srv.Operations); exists == false {
		t.Fatalf("expecting service container exists, got false")
	}

	tests.RemoveContainer(name, def.TypeService, number)

	if _, exists := ContainerRunning(srv.Operations); exists == true {
		t.Fatalf("expecting service container not existing after remove, got true")
	}

	tests.RemoveAllContainers()
}

func TestDataContainerExistsSimple(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if _, exists := DataContainerExists(srv.Operations); exists != true {
		t.Fatalf("expecting data container existing, got false")
	}

	tests.RemoveAllContainers()
}

func TestDataContainerExistsBadName(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.SrvContainerName = "some-random-name"
	if _, exists := DataContainerExists(srv.Operations); exists != false {
		t.Fatalf("expecting service container not existing, got true")
	}

	tests.RemoveAllContainers()
}

func TestDataContainerExistsAfterRemove(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if _, exists := DataContainerExists(srv.Operations); exists == false {
		t.Fatalf("expecting service container exists, got false")
	}

	tests.RemoveContainer(name, def.TypeData, number)

	if _, exists := DataContainerExists(srv.Operations); exists == true {
		t.Fatalf("expecting service container not existing after remove, got true")
	}

	tests.RemoveAllContainers()
}

func TestRemoveWithoutData(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if err := DockerStop(srv.Service, srv.Operations, 5); err != nil {
		t.Fatal("expected service container stopped, got %v", err)
	}

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running (before removal), got %v", n)
	}

	if err := DockerRemove(srv.Service, srv.Operations, false, true); err != nil {
		t.Fatal("expected service container removed, got %v", err)
	}

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running (after removal), got %v", n)
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data container existing (before removal), got %v", n)
	}

	if err := DockerRemove(srv.Service, srv.Operations, false, true); err != nil {
		t.Fatal("expected service container removed, got %v", err)
	}

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data container running (after removal), got %v", n)
	}

	tests.RemoveAllContainers()
}

func TestRemoveWithData(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if err := DockerStop(srv.Service, srv.Operations, 5); err != nil {
		t.Fatal("expected service container stopped, got %v", err)
	}

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container existing (before removal), got %v", n)
	}

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data container existing (before removal), got %v", n)
	}

	if err := DockerRemove(srv.Service, srv.Operations, true, true); err != nil {
		t.Fatal("expected service container removed, got %v", err)
	}

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running (after removal), got %v", n)
	}

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data container running (after removal), got %v", n)
	}

	tests.RemoveAllContainers()
}

func TestRemoveServiceWithoutStopping(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if err := DockerRemove(srv.Service, srv.Operations, true, true); err == nil {
		t.Fatal("expected service remove to fail, got nil")
	}

	tests.RemoveAllContainers()
}

func TestStopSimple(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service containers running, got %v", n)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service containers existing, got %v", n)
	}

	if err := DockerStop(srv.Service, srv.Operations, 5); err != nil {
		t.Fatalf("expected service container to stop, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service containers running (after stop), got %v", n)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 1 service containers existing (after stop), got %v", n)
	}

	tests.RemoveAllContainers()
}

func TestStopDataContainer(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if err := DockerStop(srv.Service, srv.Operations, 5); err == nil {
		t.Fatalf("expected stop to fail, got %v", err)
	}

	tests.RemoveAllContainers()
}
