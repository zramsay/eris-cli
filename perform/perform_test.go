package perform

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/tests"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	logger "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

func TestMain(m *testing.M) {
	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit("perform"))

	tests.RemoveAllContainers()

	exitCode := m.Run()
	tests.IfExit(tests.TestsTearDown())
	os.Exit(exitCode)
}

func TestCreateDataSimple(t *testing.T) {
	const (
		name   = "testdata"
		number = 199
	)

	defer tests.RemoveAllContainers()

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
}

func TestRunDataSimple(t *testing.T) {
	const (
		name   = "testdata"
		number = 199
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	ops := loaders.LoadDataDefinition(name, number)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	ops.Args = strings.Fields("uptime")
	if _, err := DockerRunData(ops, nil); err != nil {
		t.Fatalf("expected data successfully run, got %v", err)
	}
}

func TestRunDataBadCommandLine(t *testing.T) {
	const (
		name   = "testdata"
		number = 199
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	ops := loaders.LoadDataDefinition(name, number)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	ops.Args = strings.Fields("/bad/command/line")
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

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	ops := loaders.LoadDataDefinition(name, number)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	ops.Args = strings.Fields("uptime")
	buf, err := DockerExecData(ops, nil)
	if err != nil {
		t.Fatalf("expected data successfully run, got %v", err)
	}
	if !strings.Contains(buf.String(), "up") {
		t.Fatalf("expected to find text in the output, got %s", buf.String())
	}
}

func TestExecDataBadCommandLine(t *testing.T) {
	const (
		name   = "testdata"
		number = 199
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	ops := loaders.LoadDataDefinition(name, number)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	ops.Args = strings.Fields("/bad/command/line")
	if _, err := DockerExecData(ops, nil); err == nil {
		t.Fatalf("expected command line error, got nil")
	}
}

func TestExecDataBufferNotOverwritten(t *testing.T) {
	const (
		name   = "testdata"
		number = 199
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	ops := loaders.LoadDataDefinition(name, number)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer, config.GlobalConfig.ErrorWriter = buf, buf

	ops.Args = strings.Fields("true")
	if _, err := DockerExecData(ops, nil); err != nil {
		t.Fatalf("expected data successfully run, got %v", err)
	}

	if config.GlobalConfig.Writer != buf {
		t.Fatalf("expected global writer unchaged after exec")
	}

	if config.GlobalConfig.ErrorWriter != buf {
		t.Fatalf("expected global error writer unchanged after exec")
	}
}

func TestRunServiceSimple(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
}

func TestRunServiceNoDataContainer(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
}

func TestRunServiceTwoServices(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
}

func TestRunServiceTwoServicesPublishedPorts(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
}

func TestExecServiceSimple(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Interactive = true
	srv.Operations.Args = strings.Fields("uptime")
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 dependent data container, got %v", n)
	}
}

func TestExecServiceBufferNotOverwritten(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Args = strings.Fields("true")
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer, config.GlobalConfig.ErrorWriter = buf, buf

	if config.GlobalConfig.Writer != buf {
		t.Fatalf("expected global writer unchaged after exec")
	}

	if config.GlobalConfig.ErrorWriter != buf {
		t.Fatalf("expected global error writer unchanged after exec")
	}
}

func TestExecServiceLogOutput(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Args = strings.Fields("echo test")
	buf, err := DockerExecService(srv.Service, srv.Operations)
	if err != nil {
		t.Fatalf("expected service run, got %v", err)
	}

	if strings.TrimSpace(buf.String()) != "test" {
		t.Fatalf("expecting a certain log output, got %q", buf.String())
	}
}

func TestExecServiceLogOutputLongRunning(t *testing.T) {
	const (
		name   = "keys"
		number = 99
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Args = strings.Fields("du -sh /usr")
	buf, err := DockerExecService(srv.Service, srv.Operations)
	if err != nil {
		t.Fatalf("expected service container run, got %v", err)
	}

	if !strings.Contains(buf.String(), "/usr") {
		t.Fatalf("expecting a certain log output, got %q", buf.String())
	}
}

func TestExecServiceLogOutputInteractive(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Args = strings.Fields("echo test")
	srv.Operations.Interactive = true
	buf, err := DockerExecService(srv.Service, srv.Operations)
	if err != nil {
		t.Fatalf("expected service container run, got %v", err)
	}

	if strings.TrimSpace(buf.String()) != "test" {
		t.Fatalf("expecting a certain log output, got %q", buf.String())
	}
}

func TestExecServiceTwice(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Interactive = true
	srv.Operations.Args = strings.Fields("uptime")

	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("1. expected service container created, got %v", err)
	}

	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("2. expected service container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 dependent data container, got %v", n)
	}
}

func TestExecServiceTwiceWithoutData(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Service.AutoData = false
	srv.Operations.Interactive = true
	srv.Operations.Args = strings.Fields("uptime")
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("1. expected service container created, got %v", err)
	}

	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("2. expected service container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 dependent data containers, got %v", n)
	}
}

func TestExecServiceBadCommandLine(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Interactive = false
	srv.Operations.Args = strings.Fields("/bad/command/line")
	if _, err := DockerExecService(srv.Service, srv.Operations); err == nil {
		t.Fatalf("expected failure, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 dependent data container, got %v", n)
	}
}

func TestExecServiceNonInteractive(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Interactive = false
	srv.Operations.Args = strings.Fields("uptime")
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 dependent data container, got %v", n)
	}
}

func TestExecServiceAfterRunService(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
	srv.Operations.Args = strings.Fields("uptime")
	if _, err := DockerExecService(srv.Service, srv.Operations); err == nil {
		t.Fatalf("expected failure due to unpublished ports, got %v", err)
	}
}

func TestExecServiceAfterRunServiceWithPublishedPorts1(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
	srv.Operations.Args = strings.Fields("uptime")
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected exec container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 dependent data container, got %v", n)
	}
}

func TestExecServiceAfterRunServiceWithPublishedPorts2(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
	srv.Operations.Args = strings.Fields("uptime")
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected exec container created, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 dependent data container, got %v", n)
	}
}

func TestContainerExistsSimple(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
}

func TestContainerExistsBadName(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
}

func TestContainerExistsAfterRemove(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
}

func TestContainerRunningSimple(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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

	if _, running := ContainerRunning(srv.Operations); running == false {
		t.Fatalf("expecting service container running, got false")
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if _, running := ContainerRunning(srv.Operations); running == true {
		t.Fatalf("expecting data container not running, got true")
	}
}

func TestContainerRunningBadName(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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

	if _, running := ContainerRunning(srv.Operations); running == false {
		t.Fatalf("expecting service container running, got false")
	}

	srv.Operations.SrvContainerName = "random-bad-name"
	if _, running := ContainerRunning(srv.Operations); running == true {
		t.Fatalf("expecting data container not running, got true")
	}
}

func TestContainerRunningAfterRemove(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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

	if _, running := ContainerRunning(srv.Operations); running == false {
		t.Fatalf("expecting service container exists, got false")
	}

	tests.RemoveContainer(name, def.TypeService, number)

	if _, running := ContainerRunning(srv.Operations); running == true {
		t.Fatalf("expecting service container not existing after remove, got true")
	}
}

func TestDataContainerExistsSimple(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
}

func TestDataContainerExistsBadName(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
}

func TestDataContainerExistsAfterRemove(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
}

func TestRemoveWithoutData(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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

	if err := DockerRemove(srv.Service, srv.Operations, false, true, false); err != nil {
		t.Fatal("expected service container removed, got %v", err)
	}

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running (after removal), got %v", n)
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if n := util.HowManyContainersExisting(name, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data container existing (before removal), got %v", n)
	}

	if err := DockerRemove(srv.Service, srv.Operations, false, true, false); err != nil {
		t.Fatal("expected service container removed, got %v", err)
	}

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data container running (after removal), got %v", n)
	}
}

func TestRemoveWithData(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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

	if err := DockerRemove(srv.Service, srv.Operations, true, true, false); err != nil {
		t.Fatal("expected service container removed, got %v", err)
	}

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running (after removal), got %v", n)
	}

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data container running (after removal), got %v", n)
	}
}

func TestRemoveServiceWithoutStopping(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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

	if err := DockerRemove(srv.Service, srv.Operations, true, true, false); err == nil {
		t.Fatal("expected service remove to fail, got nil")
	}
}

func TestStopSimple(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
		t.Fatalf("expecting 0 service containers existing (after stop), got %v", n)
	}
}

func TestStopDataContainer(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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
}

func TestRebuildSimple(t *testing.T) {
	const (
		name    = "ipfs"
		number  = 99
		timeout = 5
	)

	defer tests.RemoveAllContainers()

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

	if err := DockerRebuild(srv.Service, srv.Operations, false, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running, got %v", n)
	}
}

func TestRebuildBadName(t *testing.T) {
	const (
		name    = "ipfs"
		number  = 99
		timeout = 5
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	// XXX: DockerRebuild bug.
	srv.Operations.SrvContainerName = "bad name"
	if err := DockerRebuild(srv.Service, srv.Operations, false, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}
}

func TestRebuildNotCreated(t *testing.T) {
	const (
		name    = "ipfs"
		number  = 99
		timeout = 5
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	// XXX: DockerRebuild bug.
	if err := DockerRebuild(srv.Service, srv.Operations, false, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}
}

func TestRebuildTimeout0(t *testing.T) {
	const (
		name    = "ipfs"
		number  = 99
		timeout = 0
	)

	defer tests.RemoveAllContainers()

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

	if err := DockerRebuild(srv.Service, srv.Operations, false, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running, got %v", n)
	}
}

func TestRebuildNotRunning(t *testing.T) {
	const (
		name    = "ipfs"
		number  = 99
		timeout = 5
	)

	defer tests.RemoveAllContainers()

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

	if err := DockerStop(srv.Service, srv.Operations, timeout); err != nil {
		t.Fatal("expected service container stopped, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}

	if err := DockerRebuild(srv.Service, srv.Operations, false, 5); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
}

func TestRebuildPullDisallow(t *testing.T) {
	const (
		name    = "keys"
		number  = 99
		timeout = 5
	)

	defer tests.RemoveAllContainers()

	tests.RemoveImage(name)

	os.Setenv("ERIS_PULL_APPROVE", "true")

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

	if err := DockerRebuild(srv.Service, srv.Operations, false, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running, got %v", n)
	}
}

func TestRebuildPull(t *testing.T) {
	const (
		name    = "keys"
		number  = 99
		timeout = 5
	)

	defer tests.RemoveAllContainers()

	tests.RemoveImage(name)

	os.Setenv("ERIS_PULL_APPROVE", "true")

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

	if err := DockerRebuild(srv.Service, srv.Operations, true, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running, got %v", n)
	}
}

func TestRebuildPullRepeat(t *testing.T) {
	const (
		name    = "keys"
		number  = 99
		timeout = 5
	)

	defer tests.RemoveAllContainers()

	os.Setenv("ERIS_PULL_APPROVE", "true")

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

	if err := DockerRebuild(srv.Service, srv.Operations, true, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running, got %v", n)
	}
}

func TestPullSimple(t *testing.T) {
	const (
		name   = "keys"
		number = 99
	)

	defer tests.RemoveAllContainers()

	os.Setenv("ERIS_PULL_APPROVE", "true")

	tests.RemoveImage(name)

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

	if err := DockerPull(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected image pulled, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running, got %v", n)
	}
}

func TestPullRepeat(t *testing.T) {
	const (
		name   = "keys"
		number = 99
	)

	defer tests.RemoveAllContainers()

	os.Setenv("ERIS_PULL_APPROVE", "true")

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

	if err := DockerPull(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected image pulled, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running, got %v", n)
	}
}

func TestPullBadName(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.SrvContainerName = "bad name"
	// XXX: DockerPull bug.
	// if err := DockerPull(srv.Service, srv.Operations); err != nil {
	// 	t.Fatalf("expected container pulled, got %v", err)
	// }
}

func TestLogsSimple(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
		tail   = "100"
	)

	defer tests.RemoveAllContainers()

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
		t.Fatalf("expected service container to stop, got %v", err)
	}

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer = buf

	if err := DockerLogs(srv.Service, srv.Operations, false, tail); err != nil {
		t.Fatalf("expected logs pulled, got %v", err)
	}

	if !strings.Contains(buf.String(), "Starting IPFS") {
		t.Fatalf("expected certain log entries, got %q", buf.String())
	}
}

func TestLogsFollow(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
		tail   = "1"
	)

	defer tests.RemoveAllContainers()

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
		t.Fatalf("expected service container to stop, got %v", err)
	}

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer = buf

	if err := DockerLogs(srv.Service, srv.Operations, true, tail); err != nil {
		t.Fatalf("expected logs pulled, got %v", err)
	}
}

func TestLogsTail(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
		tail   = "100"
	)

	defer tests.RemoveAllContainers()

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
		t.Fatalf("expected service container to stop, got %v", err)
	}

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer = buf

	if err := DockerLogs(srv.Service, srv.Operations, false, tail); err != nil {
		t.Fatalf("expected logs pulled, got %v", err)
	}

	if !strings.Contains(buf.String(), "Starting IPFS") {
		t.Fatalf("expected certain log entries, got %q", buf.String())
	}
}

func TestLogsTail0(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
		tail   = "0"
	)

	defer tests.RemoveAllContainers()

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
		t.Fatalf("expected service container to stop, got %v", err)
	}

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer = buf

	if err := DockerLogs(srv.Service, srv.Operations, false, tail); err != nil {
		t.Fatalf("expected logs pulled, got %v", err)
	}

	if strings.TrimSpace(buf.String()) != "" {
		t.Fatalf("expected certain log entries, got %q", buf.String())
	}
}

func TestLogsBadName(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
		tail   = "1"
	)

	defer tests.RemoveAllContainers()

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	// XXX: DockerLogs bug.
	srv.Operations.SrvContainerName = "bad name"
	if err := DockerLogs(srv.Service, srv.Operations, false, tail); err != nil {
		t.Fatalf("expected logs pulled, got %v", err)
	}
}

func TestInspectSimple(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer = buf

	if err := DockerInspect(srv.Service, srv.Operations, "all"); err != nil {
		t.Fatalf("expected inspect to succeed, got %v", err)
	}

	if !strings.Contains(buf.String(), "IPAddress") {
		t.Fatalf("expect to get IPAddress with inspect, got %q", buf.String())
	}
}

func TestInspectLine(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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

	// XXX: DockerInspect "line" doesn't redirect its output.
	if err := DockerInspect(srv.Service, srv.Operations, "line"); err != nil {
		t.Fatalf("expected inspect to succeed, got %v", err)
	}
}

func TestInspectField(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

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

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer = buf

	if err := DockerInspect(srv.Service, srv.Operations, "Config.WorkingDir"); err != nil {
		t.Fatalf("expected inspect to succeed, got %v", err)
	}

	if !strings.Contains(buf.String(), "/home/eris") {
		t.Fatalf("expect a certain value, got %q", buf.String())
	}
}

func TestInspectStoppedContainer(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if err := DockerStop(srv.Service, srv.Operations, 5); err != nil {
		t.Fatalf("expected service container be stopped, got %v", err)
	}

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer = buf

	if err := DockerInspect(srv.Service, srv.Operations, "Config.WorkingDir"); err != nil {
		t.Fatalf("expected inspect to succeed, got %v", err)
	}

	if !strings.Contains(buf.String(), "/home/eris") {
		t.Fatalf("expect a certain value, got %q", buf.String())
	}
}

func TestInspectBadName(t *testing.T) {
	const (
		name   = "ipfs"
		number = 99
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	// XXX: DockerInspect bug.
	srv.Operations.SrvContainerName = "bad name"
	if err := DockerInspect(srv.Service, srv.Operations, "all"); err != nil {
		t.Fatalf("expected inspect to succeed, got %v", err)
	}
}

func TestRenameSimple(t *testing.T) {
	const (
		name    = "testdata"
		newName = "newname"
		number  = 199
	)

	defer tests.RemoveAllContainers()

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}

	ops := loaders.LoadDataDefinition(name, number)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	if err := DockerRename(ops, newName); err != nil {
		t.Fatalf("expected container renamed, got %v", err)
	}

	if n := util.HowManyContainersExisting(name, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 containers, got %v", n)
	}
	if n := util.HowManyContainersExisting(newName, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 containers, got %v", n)
	}
}

func TestRenameService(t *testing.T) {
	const (
		name    = "ipfs"
		newName = "newname"
		number  = 99
	)

	defer tests.RemoveAllContainers()

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
	if n := util.HowManyContainersExisting(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container existing, got %v", n)
	}

	if err := DockerRename(srv.Operations, newName); err != nil {
		t.Fatalf("expected container renamed, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container existing, got %v", n)
	}

	if n := util.HowManyContainersRunning(newName, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(newName, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container existing, got %v", n)
	}
}

func TestRenameEmptyName(t *testing.T) {
	const (
		name    = "ipfs"
		newName = ""
		number  = 99
	)

	defer tests.RemoveAllContainers()

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
	if n := util.HowManyContainersExisting(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container existing, got %v", n)
	}

	// XXX: DockerRename bug.
	if err := DockerRename(srv.Operations, newName); err != nil {
		t.Fatalf("expected container renamed, got %v", err)
	}
}

func TestRenameServiceStopped(t *testing.T) {
	const (
		name    = "ipfs"
		newName = "newname"
		number  = 99
	)

	defer tests.RemoveAllContainers()

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
		t.Fatalf("expected service container be stopped, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container existing, got %v", n)
	}

	if err := DockerRename(srv.Operations, newName); err != nil {
		t.Fatalf("expected container renamed, got %v", err)
	}

	if n := util.HowManyContainersRunning(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(name, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container existing, got %v", n)
	}

	if n := util.HowManyContainersRunning(newName, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service container running, got %v", n)
	}
	if n := util.HowManyContainersExisting(newName, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 service container existing, got %v", n)
	}
}

func TestRenameBadName(t *testing.T) {
	const (
		name    = "ipfs"
		newName = "newname"
		number  = 99
	)

	defer tests.RemoveAllContainers()

	srv, err := loaders.LoadServiceDefinition(name, true, number)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.SrvContainerName = "bad name"
	if err := DockerRename(srv.Operations, newName); err == nil {
		t.Fatalf("expected rename to fail, got nil")
	}
}
