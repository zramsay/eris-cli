package perform

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/tests"
	"github.com/eris-ltd/eris-cli/util"
	ver "github.com/eris-ltd/eris-cli/version"

	log "github.com/eris-ltd/eris-logger"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit(tests.ConnectAndPull))

	tests.RemoveAllContainers()

	exitCode := m.Run()
	tests.IfExit(tests.TestsTearDown())
	os.Exit(exitCode)
}

func TestCreateDataSimple(t *testing.T) {
	const (
		name = "testdata"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist")
	}

	ops := loaders.LoadDataDefinition(name)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expecting data container existing")
	}

	// Try to create a duplicate.
	if err := DockerCreateData(ops); err == nil {
		t.Fatalf("expected an error, got nil")
	}
}

func TestRunDataSimple(t *testing.T) {
	const (
		name = "testdata"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist")
	}

	ops := loaders.LoadDataDefinition(name)
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
		name = "testdata"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist")
	}

	ops := loaders.LoadDataDefinition(name)
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
		name = "testdata"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist")
	}

	ops := loaders.LoadDataDefinition(name)
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
		name = "testdata"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist")
	}

	ops := loaders.LoadDataDefinition(name)
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
		name = "testdata"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist")
	}

	ops := loaders.LoadDataDefinition(name)
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
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expecting data container existing")
	}
}

func TestRunServiceNoDataContainer(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Service.AutoData = false
	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
	if util.Exists(def.TypeData, name) {
		t.Fatalf("expecting no dependend data container existing")
	}
}

func TestRunServiceAlreadyRunning(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("1. expecting service container running")
	}
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("1. expecting data container existing")
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected already running service to fail with nil")
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("2. expecting service container running")
	}
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("2. expecting data container existing")
	}
}

func TestRunServiceNonExistentImage(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Service.Image = "non existent"
	if err := DockerRunService(srv.Service, srv.Operations); err == nil {
		t.Fatalf("expected run service to fail")
	}
}

func TestExecServiceSimple(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Interactive = true
	srv.Operations.Args = strings.Fields("uptime")
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceBufferNotOverwritten(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	srv, err := loaders.LoadServiceDefinition(name)
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

func TestExecServiceAlwaysRestart(t *testing.T) {
	const (
		name = "restart-keys"
	)

	defer tests.RemoveAllContainers()

	if err := tests.FakeServiceDefinition(name, `
name = "`+name+`"

[service]
name = "`+name+`"
image = "`+path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_KEYS)+`"
data_container = true
exec_host = "ERIS_KEYS_HOST"
restart = "always"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Interactive = false
	srv.Operations.Args = strings.Fields("uname")
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceMaxAttemptsRestart(t *testing.T) {
	const (
		name = "restart-keys"
	)

	defer tests.RemoveAllContainers()

	if err := tests.FakeServiceDefinition(name, `
name = "`+name+`"

[service]
name = "`+name+`"
image = "`+path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_KEYS)+`"
data_container = true
exec_host = "ERIS_KEYS_HOST"
restart = "max:99"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Interactive = false
	srv.Operations.Args = strings.Fields("uname")
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceNeverRestart(t *testing.T) {
	const (
		name = "restart-keys"
	)

	defer tests.RemoveAllContainers()

	if err := tests.FakeServiceDefinition(name, `
name = "`+name+`"

[service]
name = "`+name+`"
image = "`+path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_KEYS)+`"
data_container = true
exec_host = "ERIS_KEYS_HOST"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Interactive = false
	srv.Operations.Args = strings.Fields("uname")
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceVolume(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Args = strings.Fields("uptime")
	srv.Operations.Volume = filepath.Join(config.GlobalConfig.ErisDir)
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceMount(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Args = strings.Fields("uptime")
	srv.Service.Volumes = []string{
		config.GlobalConfig.ErisDir + ":" + "/tmp",
		config.GlobalConfig.ErisDir + ":" + "/custom",
	}
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceBadMount1(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Args = strings.Fields("uptime")
	srv.Service.Volumes = []string{""}
	if _, err := DockerExecService(srv.Service, srv.Operations); err == nil {
		t.Fatalf("expected service container creation to fail")
	}
}

func TestExecServiceBadMount2(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Args = strings.Fields("uptime")
	srv.Service.Volumes = []string{config.GlobalConfig.ErisDir + ":"}
	if _, err := DockerExecService(srv.Service, srv.Operations); err == nil {
		t.Fatalf("expected service container creation to fail")
	}
}

func TestExecServiceLogOutput(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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
		name = "keys"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}
	srv, err := loaders.LoadServiceDefinition(name)
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
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceTwiceWithoutData(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if util.Exists(def.TypeData, name) {
		t.Fatalf("expecting dependend data container doesn't exist")
	}
}

func TestExecServiceBadCommandLine(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Interactive = false
	srv.Operations.Args = strings.Fields("/bad/command/line")
	if _, err := DockerExecService(srv.Service, srv.Operations); err == nil {
		t.Fatalf("expected failure, got %v", err)
	}

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceNonInteractive(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Interactive = false
	srv.Operations.Args = strings.Fields("uptime")
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceAfterRunService(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceAfterRunServiceWithPublishedPorts2(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestContainerExistsSimple(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if exists := ContainerExists(srv.Operations.SrvContainerName); exists != true {
		t.Fatalf("expecting service container existing, got false")
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if exists := ContainerExists(srv.Operations.SrvContainerName); exists != true {
		t.Fatalf("expecting data container existing, got false")
	}
}

func TestContainerExistsBadName(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.SrvContainerName = "some-random-name"
	if exists := ContainerExists(srv.Operations.SrvContainerName); exists != false {
		t.Fatalf("expecting service container not existing, got true")
	}
}

func TestContainerExistsAfterRemove(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if exists := ContainerExists(srv.Operations.SrvContainerName); exists == false {
		t.Fatalf("expecting service container exists, got false")
	}

	tests.RemoveContainer(name, def.TypeService)

	if exists := ContainerExists(srv.Operations.SrvContainerName); exists == true {
		t.Fatalf("expecting service container not existing after remove, got true")
	}
}

func TestContainerRunningSimple(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if running := ContainerRunning(srv.Operations.SrvContainerName); running == false {
		t.Fatalf("expecting service container running, got false")
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if running := ContainerRunning(srv.Operations.SrvContainerName); running == true {
		t.Fatalf("expecting data container not running, got true")
	}
}

func TestContainerRunningBadName(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if running := ContainerRunning(srv.Operations.SrvContainerName); running == false {
		t.Fatalf("expecting service container running, got false")
	}

	srv.Operations.SrvContainerName = "random-bad-name"
	if running := ContainerRunning(srv.Operations.SrvContainerName); running == true {
		t.Fatalf("expecting data container not running, got true")
	}
}

func TestContainerRunningAfterRemove(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if running := ContainerRunning(srv.Operations.SrvContainerName); running == false {
		t.Fatalf("expecting service container exists, got false")
	}

	tests.RemoveContainer(name, def.TypeService)

	if running := ContainerRunning(srv.Operations.SrvContainerName); running == true {
		t.Fatalf("expecting service container not existing after remove, got true")
	}
}

func TestRemoveWithoutData(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if err := DockerStop(srv.Service, srv.Operations, 5); err != nil {
		t.Fatalf("expected service container stopped, got %v", err)
	}

	if !util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container existing (before removal)")
	}

	if err := DockerRemove(srv.Service, srv.Operations, false, true, false); err != nil {
		t.Fatalf("expected service container removed, got %v", err)
	}

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist (after removal)")
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expected data container existing (before removal)")
	}

	if err := DockerRemove(srv.Service, srv.Operations, false, true, false); err != nil {
		t.Fatalf("expected service container removed, got %v", err)
	}

	if util.Exists(def.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist (after removal)")
	}
}

func TestRemoveWithData(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if err := DockerStop(srv.Service, srv.Operations, 5); err != nil {
		t.Fatalf("expected service container stopped, got %v", err)
	}

	if !util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container exist (before removal)")
	}

	if !util.Exists(def.TypeData, name) {
		t.Fatalf("expecting data container exist (before removal)")
	}

	if err := DockerRemove(srv.Service, srv.Operations, true, true, false); err != nil {
		t.Fatalf("expected service container removed, got %v", err)
	}

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist (after removal)")
	}

	if util.Exists(def.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist (after removal)")
	}
}

func TestRemoveNonExistent(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.SrvContainerName = "non existent"
	if err := DockerRemove(srv.Service, srv.Operations, true, true, false); err != nil {
		t.Fatalf("expected container removal will fail with nil")
	}
}

func TestRemoveServiceWithoutStopping(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerStop(srv.Service, srv.Operations, 5); err != nil {
		t.Fatalf("expected service container to stop, got %v", err)
	}

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container don't run (after stop)")
	}

	if util.Running(def.TypeData, name) {
		t.Fatalf("expecting data container don't run (after stop)")
	}
}

func TestStopDataContainer(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if err := DockerStop(srv.Service, srv.Operations, 5); err != nil {
		t.Fatalf("expected container to stop, got %v", err)
	}
}

func TestRebuildSimple(t *testing.T) {
	const (
		name    = "ipfs"
		timeout = 5
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerRebuild(srv.Service, srv.Operations, false, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
}

func TestRebuildBadName(t *testing.T) {
	const (
		name    = "ipfs"
		timeout = 5
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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
		timeout = 5
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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
		timeout = 0
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerRebuild(srv.Service, srv.Operations, false, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
}

func TestRebuildNotRunning(t *testing.T) {
	const (
		name    = "ipfs"
		timeout = 5
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if err := DockerStop(srv.Service, srv.Operations, timeout); err != nil {
		t.Fatalf("expected service container stopped, got %v", err)
	}

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't run")
	}

	if err := DockerRebuild(srv.Service, srv.Operations, false, 5); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't run")
	}
}

func TestRebuildPullDisallow(t *testing.T) {
	const (
		name    = "keys"
		timeout = 5
	)

	defer tests.RemoveAllContainers()

	tests.RemoveImage(name)

	os.Setenv("ERIS_PULL_APPROVE", "true")

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerRebuild(srv.Service, srv.Operations, false, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
}

func TestRebuildPull(t *testing.T) {
	const (
		name    = "keys"
		timeout = 5
	)

	defer tests.RemoveAllContainers()

	tests.RemoveImage(name)

	os.Setenv("ERIS_PULL_APPROVE", "true")

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerRebuild(srv.Service, srv.Operations, true, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
}

func TestRebuildPullRepeat(t *testing.T) {
	const (
		name    = "keys"
		timeout = 5
	)

	defer tests.RemoveAllContainers()

	os.Setenv("ERIS_PULL_APPROVE", "true")

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerRebuild(srv.Service, srv.Operations, true, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
}

func TestPullSimple(t *testing.T) {
	const (
		name = "keys"
	)

	defer tests.RemoveAllContainers()

	os.Setenv("ERIS_PULL_APPROVE", "true")

	tests.RemoveImage(name)

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerPull(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected image pulled, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
}

func TestPullRepeat(t *testing.T) {
	const (
		name = "keys"
	)

	defer tests.RemoveAllContainers()

	os.Setenv("ERIS_PULL_APPROVE", "true")

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerPull(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected image pulled, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
}

func TestPullBadName(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	srv, err := loaders.LoadServiceDefinition(name)
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
		name = "ipfs"
		tail = "100"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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

func TestLogsNilConfig(t *testing.T) {
	const (
		name = "ipfs"
		tail = "1"
	)

	defer tests.RemoveAllContainers()

	savedConfig := config.GlobalConfig
	config.GlobalConfig = nil
	defer func() { config.GlobalConfig = savedConfig }()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if err := DockerStop(srv.Service, srv.Operations, 5); err != nil {
		t.Fatalf("expected service container to stop, got %v", err)
	}

	if err := DockerLogs(srv.Service, srv.Operations, false, tail); err != nil {
		t.Fatalf("expected logs pulled, got %v", err)
	}
}

func TestLogsFollow(t *testing.T) {
	const (
		name = "ipfs"
		tail = "1"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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
		name = "ipfs"
		tail = "100"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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
		name = "ipfs"
		tail = "0"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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
		name = "ipfs"
		tail = "1"
	)

	defer tests.RemoveAllContainers()

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.SrvContainerName = "bad name"
	if err := DockerLogs(srv.Service, srv.Operations, false, tail); err == nil {
		t.Fatalf("expected logs to fail")
	}
}

func TestLogsBadServiceName(t *testing.T) {
	const (
		name = "ipfs"
		tail = "1"
	)
	defer tests.RemoveAllContainers()

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.SrvContainerName = "bad-name"
	if err := DockerLogs(srv.Service, srv.Operations, false, tail); err == nil {
		t.Fatalf("expected logs to fail")
	}
}

func TestInspectSimple(t *testing.T) {
	const (
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
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
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	srv, err := loaders.LoadServiceDefinition(name)
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
		name = "ipfs"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.SrvContainerName = "bad name"
	if err := DockerInspect(srv.Service, srv.Operations, "all"); err == nil {
		t.Fatalf("expected inspect to fail")
	}
}

func TestRenameSimple(t *testing.T) {
	const (
		name    = "testdata"
		newName = "newname"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	ops := loaders.LoadDataDefinition(name)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	if err := DockerRename(ops, newName); err != nil {
		t.Fatalf("expected container renamed, got %v", err)
	}

	if util.Exists(def.TypeData, name) {
		t.Fatalf("expecting old data container doesn't exist")
	}

	if !util.Exists(def.TypeData, newName) {
		t.Fatalf("expecting renamed data container exists")
	}
}

func TestRenameService(t *testing.T) {
	const (
		name    = "ipfs"
		newName = "newname"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerRename(srv.Operations, newName); err != nil {
		t.Fatalf("expected container renamed, got %v", err)
	}

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting old service container not running")
	}

	if !util.Running(def.TypeService, newName) {
		t.Fatalf("expecting new service container running")
	}
}

func TestRenameEmptyName(t *testing.T) {
	const (
		name    = "ipfs"
		newName = ""
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerRename(srv.Operations, newName); err == nil {
		t.Fatalf("expected empty name rename to fail")
	}
}

func TestRenameServiceStopped(t *testing.T) {
	const (
		name    = "ipfs"
		newName = "newname"
	)

	defer tests.RemoveAllContainers()

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}
	if err := DockerStop(srv.Service, srv.Operations, 5); err != nil {
		t.Fatalf("expected service container be stopped, got %v", err)
	}

	if util.Running(def.TypeService, name) {
		t.Fatalf("expecting service container doesn't run")
	}
	if !util.Exists(def.TypeService, name) {
		t.Fatalf("expecting service container exist")
	}

	if err := DockerRename(srv.Operations, newName); err != nil {
		t.Fatalf("expected container renamed, got %v", err)
	}

	if util.Exists(def.TypeService, name) {
		t.Fatalf("expecting old service container doesn't exist")
	}

	if util.Running(def.TypeService, newName) {
		t.Fatalf("expecting new service container doesn't run")
	}
	if !util.Exists(def.TypeService, newName) {
		t.Fatalf("expecting new service container exist")
	}
}

func TestRenameBadName(t *testing.T) {
	const (
		name    = "ipfs"
		newName = "newname"
	)

	defer tests.RemoveAllContainers()

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.SrvContainerName = "bad name"
	if err := DockerRename(srv.Operations, newName); err == nil {
		t.Fatalf("expected rename to fail, got nil")
	}
}

func TestBuildSimple(t *testing.T) {
	const (
		image = "test-image-1"
	)

	dockerfile := `FROM ` + path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_KEYS)

	if err := DockerBuild(image, dockerfile); err != nil {
		t.Fatalf("expected image to be built, got %v", err)
	}

	if err := DockerRemoveImage(image, true); err != nil {
		t.Fatalf("expected image to be remove, got %v", err)
	}
}

func TestBuildBad(t *testing.T) {
	const (
		image = "test-image-2"
	)

	defer DockerRemoveImage(image, true)

	dockerfile := `@^@%^@#%^&&#@%`

	if err := DockerBuild(image, dockerfile); err == nil {
		t.Fatalf("expected image build to fail")
	}
}

func TestBuildImage(t *testing.T) {
	const (
		image = "test-image-3"
	)

	defer DockerRemoveImage(image, true)

	dockerfile := `FROM ###^@%^@#%^&&#@%`

	if err := DockerBuild(image, dockerfile); err == nil {
		t.Fatalf("expected image build to fail")
	}
}

func TestBuildEmptyImage(t *testing.T) {
	const (
		image = "test-image-4"
	)

	defer DockerRemoveImage(image, true)

	if err := DockerBuild(image, ``); err == nil {
		t.Fatalf("expected image build to fail")
	}
}

func TestRemoveImageBadName(t *testing.T) {
	if err := DockerRemoveImage("bad name", true); err == nil {
		t.Fatalf("expected remove image to fail")
	}
}
