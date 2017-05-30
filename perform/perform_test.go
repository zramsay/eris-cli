package perform

import (
	"bytes"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/loaders"
	"github.com/monax/monax/log"
	"github.com/monax/monax/testutil"
	"github.com/monax/monax/util"
	"github.com/monax/monax/version"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.WarnLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	testutil.IfExit(testutil.Init(testutil.Pull{
		Images:   []string{"data", "keys", "compilers"},
		Services: []string{"keys", "compilers"},
	}))

	testutil.RemoveAllContainers()

	exitCode := m.Run()
	testutil.IfExit(testutil.TearDown())
	os.Exit(exitCode)
}

func TestCreateDataSimple(t *testing.T) {
	const (
		name = "testdata"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist")
	}

	ops := loaders.LoadDataDefinition(name)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	if !util.Exists(definitions.TypeData, name) {
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

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist")
	}

	ops := loaders.LoadDataDefinition(name)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	ops.Args = strings.Fields("bash -c true")
	if _, err := DockerRunData(ops, nil); err != nil {
		t.Fatalf("expected data successfully run, got %v", err)
	}
}

func TestRunDataBadCommandLine(t *testing.T) {
	const (
		name = "testdata"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeData, name) {
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

	testutil.RemoveAllContainers()
}

func TestExecDataSimple(t *testing.T) {
	const (
		name = "testdata"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeData, name) {
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

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeData, name) {
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

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist")
	}

	ops := loaders.LoadDataDefinition(name)
	if err := DockerCreateData(ops); err != nil {
		t.Fatalf("expected data container created, got %v", err)
	}

	buf := new(bytes.Buffer)
	config.Global.Writer, config.Global.ErrorWriter = buf, buf

	ops.Args = strings.Fields("true")
	if _, err := DockerExecData(ops, nil); err != nil {
		t.Fatalf("expected data successfully run, got %v", err)
	}

	if config.Global.Writer != buf {
		t.Fatalf("expected global writer unchaged after exec")
	}

	if config.Global.ErrorWriter != buf {
		t.Fatalf("expected global error writer unchanged after exec")
	}
}

func TestRunServiceSimple(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting data container existing")
	}
}

func TestRunServiceNoDataContainer(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
	if util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting no dependend data container existing")
	}
}

func TestRunServiceAlreadyRunning(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("1. expecting service container running")
	}
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("1. expecting data container existing")
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected already running service to fail with nil")
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("2. expecting service container running")
	}
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("2. expecting data container existing")
	}
}

func TestRunServiceNonExistentImage(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

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
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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

	if util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceBufferNotOverwritten(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Args = strings.Fields("true")
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	buf := new(bytes.Buffer)
	config.Global.Writer, config.Global.ErrorWriter = buf, buf

	if config.Global.Writer != buf {
		t.Fatalf("expected global writer unchaged after exec")
	}

	if config.Global.ErrorWriter != buf {
		t.Fatalf("expected global error writer unchanged after exec")
	}
}

func TestExecServiceAlwaysRestart(t *testing.T) {
	const (
		name = "restart-keys"
	)

	defer testutil.RemoveAllContainers()

	if err := testutil.FakeServiceDefinition(name, `
name = "`+name+`"

[service]
name = "`+name+`"
image = "`+path.Join(version.DefaultRegistry, version.ImageKeys)+`"
data_container = true
exec_host = "MONAX_KEYS_HOST"
restart = "always"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if util.Exists(definitions.TypeService, name) {
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

	if util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceMaxAttemptsRestart(t *testing.T) {
	const (
		name = "restart-keys"
	)

	defer testutil.RemoveAllContainers()

	if err := testutil.FakeServiceDefinition(name, `
name = "`+name+`"

[service]
name = "`+name+`"
image = "`+path.Join(version.DefaultRegistry, version.ImageKeys)+`"
data_container = true
exec_host = "MONAX_KEYS_HOST"
restart = "max:99"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if util.Exists(definitions.TypeService, name) {
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

	if util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceNeverRestart(t *testing.T) {
	const (
		name = "restart-keys"
	)

	defer testutil.RemoveAllContainers()

	if err := testutil.FakeServiceDefinition(name, `
name = "`+name+`"

[service]
name = "`+name+`"
image = "`+path.Join(version.DefaultRegistry, version.ImageKeys)+`"
data_container = true
exec_host = "MONAX_KEYS_HOST"
`); err != nil {
		t.Fatalf("can't create a fake service definition: %v", err)
	}

	if util.Exists(definitions.TypeService, name) {
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

	if util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceVolume(t *testing.T) {
	const (
		name = "compilers"
	)

	// Don't work on Windows without MSYS or Cygwin.
	// https://github.com/docker/docker/issues/12751
	if runtime.GOOS == "windows" {
		return
	}

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Args = strings.Fields("uptime")
	srv.Operations.Volume = config.MonaxRoot
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceMount(t *testing.T) {
	const (
		name = "compilers"
	)

	// Don't work on Windows without MSYS or Cygwin.
	// https://github.com/docker/docker/issues/12751
	if runtime.GOOS == "windows" {
		return
	}

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Args = strings.Fields("uptime")
	srv.Service.Volumes = []string{
		config.MonaxRoot + ":" + "/tmp",
		config.MonaxRoot + ":" + "/custom",
	}
	if _, err := DockerExecService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceBadMount1(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.Args = strings.Fields("uptime")
	srv.Service.Volumes = []string{config.MonaxRoot + ":"}
	if _, err := DockerExecService(srv.Service, srv.Operations); err == nil {
		t.Fatalf("expected service container creation to fail")
	}
}

func TestExecServiceLogOutput(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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

	if util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceTwiceWithoutData(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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

	if util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting dependend data container doesn't exist")
	}
}

func TestExecServiceBadCommandLine(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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

	if util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceNonInteractive(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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

	if util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container not running")
	}
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

// TODO fix!
// this test (last 3 lines) does *not* fail when it should
// see following two tests
/*
func TestExecServiceAfterRunService(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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
}*/

func TestExecServiceAfterRunServiceWithPublishedPorts1(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestExecServiceAfterRunServiceWithPublishedPorts2(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting dependend data container existing")
	}
}

func TestContainerExistsSimple(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if exists := ContainerExists(srv.Operations.SrvContainerName); !exists {
		t.Fatalf("expecting service container existing, got false")
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if exists := ContainerExists(srv.Operations.SrvContainerName); !exists {
		t.Fatalf("expecting data container existing, got false")
	}
}

func TestContainerExistsBadName(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	srv.Operations.SrvContainerName = "some-random-name"
	if exists := ContainerExists(srv.Operations.SrvContainerName); exists {
		t.Fatalf("expecting service container not existing, got true")
	}
}

func TestContainerExistsAfterRemove(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if exists := ContainerExists(srv.Operations.SrvContainerName); !exists {
		t.Fatalf("expecting service container exists, got false")
	}

	testutil.RemoveContainer(name, definitions.TypeService)

	if exists := ContainerExists(srv.Operations.SrvContainerName); exists {
		t.Fatalf("expecting service container not existing after remove, got true")
	}
}

func TestContainerRunningSimple(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if running := ContainerRunning(srv.Operations.SrvContainerName); !running {
		t.Fatalf("expecting service container running, got false")
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if running := ContainerRunning(srv.Operations.SrvContainerName); running {
		t.Fatalf("expecting data container not running, got true")
	}
}

func TestContainerRunningBadName(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if running := ContainerRunning(srv.Operations.SrvContainerName); !running {
		t.Fatalf("expecting service container running, got false")
	}

	srv.Operations.SrvContainerName = "random-bad-name"
	if running := ContainerRunning(srv.Operations.SrvContainerName); running {
		t.Fatalf("expecting data container not running, got true")
	}
}

func TestContainerRunningAfterRemove(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if running := ContainerRunning(srv.Operations.SrvContainerName); !running {
		t.Fatalf("expecting service container exists, got false")
	}

	testutil.RemoveContainer(name, definitions.TypeService)

	if running := ContainerRunning(srv.Operations.SrvContainerName); running {
		t.Fatalf("expecting service container not existing after remove, got true")
	}
}

func TestRemoveWithoutData(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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

	if !util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container existing (before removal)")
	}

	if err := DockerRemove(srv.Service, srv.Operations, false, true, false); err != nil {
		t.Fatalf("expected service container removed, got %v", err)
	}

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist (after removal)")
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName
	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("expected data container existing (before removal)")
	}

	if err := DockerRemove(srv.Service, srv.Operations, false, true, false); err != nil {
		t.Fatalf("expected service container removed, got %v", err)
	}

	if util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist (after removal)")
	}
}

func TestRemoveWithData(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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

	if !util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container exist (before removal)")
	}

	if !util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting data container exist (before removal)")
	}

	if err := DockerRemove(srv.Service, srv.Operations, true, true, false); err != nil {
		t.Fatalf("expected service container removed, got %v", err)
	}

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist (after removal)")
	}

	if util.Exists(definitions.TypeData, name) {
		t.Fatalf("expecting data container doesn't exist (after removal)")
	}
}

func TestRemoveNonExistent(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

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
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerStop(srv.Service, srv.Operations, 5); err != nil {
		t.Fatalf("expected service container to stop, got %v", err)
	}

	if util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container don't run (after stop)")
	}

	if util.Running(definitions.TypeData, name) {
		t.Fatalf("expecting data container don't run (after stop)")
	}
}

func TestStopDataContainer(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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
		name    = "compilers"
		timeout = 5
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerRebuild(srv.Service, srv.Operations, false, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
}

func TestRebuildBadName(t *testing.T) {
	const (
		name    = "compilers"
		timeout = 5
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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
		name    = "compilers"
		timeout = 5
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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
		name    = "compilers"
		timeout = 0
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerRebuild(srv.Service, srv.Operations, false, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
}

func TestRebuildNotRunning(t *testing.T) {
	const (
		name    = "compilers"
		timeout = 5
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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

	if util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't run")
	}

	if err := DockerRebuild(srv.Service, srv.Operations, false, 5); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't run")
	}
}

func TestRebuildPullDisallow(t *testing.T) {
	const (
		name    = "keys"
		timeout = 5
	)

	defer testutil.RemoveAllContainers()

	testutil.RemoveImage(name)

	os.Setenv("MONAX_PULL_APPROVE", "true")

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerRebuild(srv.Service, srv.Operations, false, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
}

func TestRebuildPull(t *testing.T) {
	const (
		name    = "keys"
		timeout = 5
	)

	defer testutil.RemoveAllContainers()

	testutil.RemoveImage(name)

	os.Setenv("MONAX_PULL_APPROVE", "true")

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerRebuild(srv.Service, srv.Operations, true, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
}

func TestRebuildPullRepeat(t *testing.T) {
	const (
		name    = "keys"
		timeout = 5
	)

	defer testutil.RemoveAllContainers()

	os.Setenv("MONAX_PULL_APPROVE", "true")

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerRebuild(srv.Service, srv.Operations, true, timeout); err != nil {
		t.Fatalf("expected container rebuilt, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
}

func TestPullSimple(t *testing.T) {
	const (
		name = "keys"
	)

	defer testutil.RemoveAllContainers()

	os.Setenv("MONAX_PULL_APPROVE", "true")

	testutil.RemoveImage(name)

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerPull(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected image pulled, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
}

func TestPullRepeat(t *testing.T) {
	const (
		name = "keys"
	)

	defer testutil.RemoveAllContainers()

	os.Setenv("MONAX_PULL_APPROVE", "true")

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}

	if err := DockerPull(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected image pulled, got %v", err)
	}

	if !util.Running(definitions.TypeService, name) {
		t.Fatalf("expecting service container running")
	}
}

func TestPullBadName(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

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

// TODO: [ben] issue-1262: perform/TestLogsSimple fails
// https://github.com/monax/monax/issues/1262
func testLogsSimple(t *testing.T) {
	const (
		name = "compilers"
		tail = "100"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
		t.Fatalf("expecting service container doesn't exist")
	}

	srv, err := loaders.LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("could not load service definition %v", err)
	}

	if err := DockerRunService(srv.Service, srv.Operations); err != nil {
		t.Fatalf("expected service container created, got %v", err)
	}

	// this test fails if using keys rather than IPFS
	// except if using the ErrorWriter instead
	if err := DockerStop(srv.Service, srv.Operations, 5); err != nil {
		t.Fatalf("expected service container to stop, got %v", err)
	}

	buf := new(bytes.Buffer)
	bufErr := new(bytes.Buffer)

	config.Global.Writer = buf
	config.Global.ErrorWriter = bufErr

	if err := DockerLogs(srv.Service, srv.Operations, false, tail); err != nil {
		t.Fatalf("expected logs pulled, got %v", err)
	}

	if !strings.Contains(bufErr.String(), "Starting monax-keys") {
		t.Fatalf("expected certain log entries, got %q", bufErr.String())
	}
}

func TestLogsNilConfig(t *testing.T) {
	const (
		name = "compilers"
		tail = "1"
	)

	defer testutil.RemoveAllContainers()

	savedConfig := config.Global
	config.Global = nil
	defer func() { config.Global = savedConfig }()

	if util.Exists(definitions.TypeService, name) {
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
		name = "compilers"
		tail = "1"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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
	config.Global.Writer = buf

	if err := DockerLogs(srv.Service, srv.Operations, true, tail); err != nil {
		t.Fatalf("expected logs pulled, got %v", err)
	}
}

/*func TestLogsTail(t *testing.T) {
	const (
		name = "keys"
		tail = "100"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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
	config.Global.Writer = buf

	if err := DockerLogs(srv.Service, srv.Operations, false, tail); err != nil {
		t.Fatalf("expected logs pulled, got %v", err)
	}

	if !strings.Contains(buf.String(), "Starting monax-keys") {
		t.Fatalf("expected certain log entries, got %q", buf.String())
	}
}*/

func TestLogsTail0(t *testing.T) {
	const (
		name = "compilers"
		tail = "0"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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
	config.Global.Writer = buf

	if err := DockerLogs(srv.Service, srv.Operations, false, tail); err != nil {
		t.Fatalf("expected logs pulled, got %v", err)
	}

	if strings.TrimSpace(buf.String()) != "" {
		t.Fatalf("expected certain log entries, got %q", buf.String())
	}
}

func TestLogsBadName(t *testing.T) {
	const (
		name = "compilers"
		tail = "1"
	)

	defer testutil.RemoveAllContainers()

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
		name = "compilers"
		tail = "1"
	)
	defer testutil.RemoveAllContainers()

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
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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
	config.Global.Writer = buf

	if err := DockerInspect(srv.Service, srv.Operations, "all"); err != nil {
		t.Fatalf("expected inspect to succeed, got %v", err)
	}

	if !strings.Contains(buf.String(), "IPAddress") {
		t.Fatalf("expect to get IPAddress with inspect, got %q", buf.String())
	}
}

func TestInspectLine(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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
	config.Global.Writer = buf

	if err := DockerInspect(srv.Service, srv.Operations, "Config.WorkingDir"); err != nil {
		t.Fatalf("expected inspect to succeed, got %v", err)
	}

	if !strings.Contains(buf.String(), "/home/monax") {
		t.Fatalf("expect a certain value, got %q", buf.String())
	}
}

func TestInspectStoppedContainer(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

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
	config.Global.Writer = buf

	if err := DockerInspect(srv.Service, srv.Operations, "Config.WorkingDir"); err != nil {
		t.Fatalf("expected inspect to succeed, got %v", err)
	}

	if !strings.Contains(buf.String(), "/home/monax") {
		t.Fatalf("expect a certain value, got %q", buf.String())
	}
}

func TestInspectBadName(t *testing.T) {
	const (
		name = "compilers"
	)

	defer testutil.RemoveAllContainers()

	if util.Exists(definitions.TypeService, name) {
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

func TestBuildSimple(t *testing.T) {
	const (
		image = "test-image-1"
	)

	dockerfile := `FROM ` + path.Join(version.DefaultRegistry, version.ImageKeys)

	if err := DockerBuild(image, dockerfile); err != nil {
		t.Fatalf("expected image to be built, got %v", err)
	}

	if err := DockerRemoveImage(image, true); err != nil {
		t.Fatalf("expected image to be removed, got %v", err)
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
