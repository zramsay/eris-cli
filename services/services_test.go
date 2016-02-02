package services

import (
	"bytes"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/list"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/tests"
	"github.com/eris-ltd/eris-cli/util"
	ver "github.com/eris-ltd/eris-cli/version"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	logger "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

const servName = "ipfs"

func TestMain(m *testing.M) {
	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit("services"))

	// Prevent CLI from starting IPFS.
	os.Setenv("ERIS_SKIP_ENSURE", "true")

	exitCode := m.Run()
	log.Info("Tearing tests down")
	tests.IfExit(tests.TestsTearDown())
	os.Exit(exitCode)
}

func TestKnownServices(t *testing.T) {
	do := def.NowDo()
	do.Known = true
	do.Existing = false
	do.Running = false
	do.Quiet = true
	if err := list.ListAll(do, "services"); err != nil {
		t.Fatalf("expected list to succeed, got %v", err)
	}

	// Join with ".toml" to avoid an extra for...range.
	definitions := strings.Split(strings.Join(strings.Split(do.Result, "\n"), ".toml ")+".toml", " ")

	sort.Strings(definitions)
	sort.Strings(ver.SERVICE_DEFINITIONS)

	if len(definitions) != len(ver.SERVICE_DEFINITIONS) {
		t.Errorf("expected %v, got %v service definitions", len(ver.SERVICE_DEFINITIONS), len(definitions))
	}

	if reflect.DeepEqual(definitions, ver.SERVICE_DEFINITIONS) != true {
		t.Fatalf("did not find all expected service definitions %v, got %v", ver.SERVICE_DEFINITIONS, definitions)
	}
}

func TestLoadServiceDefinition(t *testing.T) {
	// [pv]: this test belongs to the loaders package.
	srv, err := loaders.LoadServiceDefinition(servName, true, 1)
	if err != nil {
		t.Fatalf("expected definition to load, got %v", err)
	}

	if srv.Name != servName {
		t.Fatalf("improper name on load, expected %v, got %v", servName, srv.Name)
	}

	if srv.Service.Name != servName {
		t.Fatalf("improper service name on load, expected %v, got %v", servName, srv.Service.Name)
	}

	if !srv.Service.AutoData {
		t.Fatal("data_container not properly read on load")
	}

	if srv.Operations.DataContainerName == "" {
		t.Fatal("data_container_name not set")
	}
}

func TestStartKillService(t *testing.T) {
	defer tests.RemoveAllContainers()

	start(t, servName, false)
	if n := util.HowManyContainersRunning(servName, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 running service container, got %v", n)
	}
	if n := util.HowManyContainersExisting(servName, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data container, got %v", n)
	}

	kill(t, servName, true)
	if n := util.HowManyContainersRunning(servName, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 running service container, got %v", n)
	}
	if n := util.HowManyContainersExisting(servName, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data container, got %v", n)
	}

}

func TestInspectService1(t *testing.T) {
	defer tests.RemoveAllContainers()

	start(t, servName, false)

	do := def.NowDo()
	do.Name = servName
	do.Operations.Args = []string{"name"}
	do.Operations.ContainerNumber = 1

	if err := InspectService(do); err != nil {
		t.Fatalf("expected service to be inspected, got %v", err)
	}
}

func TestInspectService2(t *testing.T) {
	defer tests.RemoveAllContainers()

	start(t, servName, false)

	do := def.NowDo()
	do.Name = servName
	do.Operations.Args = []string{"config.user"}
	do.Operations.ContainerNumber = 1

	if err := InspectService(do); err != nil {
		t.Fatalf("expected service to be inspected, got %v", err)
	}
}

func TestLogsService(t *testing.T) {
	defer tests.RemoveAllContainers()

	start(t, servName, false)

	do := def.NowDo()
	do.Name = servName
	do.Follow = false
	do.Tail = "5"

	if err := LogsService(do); err != nil {
		t.Fatalf("expected service to return logs, got %v", err)
	}
}

func TestExecService(t *testing.T) {
	defer tests.RemoveAllContainers()

	start(t, servName, true)

	do := def.NowDo()
	do.Name = servName
	do.Operations.Interactive = false
	do.Operations.Args = strings.Fields("ls -la /root/")

	buf, err := ExecService(do)
	if err != nil {
		t.Fatalf("expected to execute service, got %v", err)
	}

	if !strings.Contains(buf.String(), ".bashrc") {
		t.Fatalf("expected a file in the exec output, got %v", buf.String())
	}
}

func TestExecServiceBadCommandLine(t *testing.T) {
	defer tests.RemoveAllContainers()

	start(t, servName, true)

	buf := new(bytes.Buffer)
	config.GlobalConfig.Writer = buf

	do := def.NowDo()
	do.Name = servName
	do.Operations.Interactive = false
	do.Operations.Args = strings.Fields("bad command line")

	if _, err := ExecService(do); err == nil {
		t.Fatal("expected executing service to fail")
	}
}

func TestUpdateService(t *testing.T) {
	defer tests.RemoveAllContainers()

	start(t, servName, false)

	do := def.NowDo()
	do.Name = servName
	do.Pull = false
	do.Timeout = 1
	if err := UpdateService(do); err != nil {
		t.Fatalf("expected service to be updated, got %v", err)
	}

	if n := util.HowManyContainersRunning(servName, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 running service container, got %v", n)
	}
}

func TestKillService(t *testing.T) {
	defer tests.RemoveAllContainers()

	start(t, servName, false)
	if n := util.HowManyContainersRunning(servName, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 running service container, got %v", n)
	}
	if n := util.HowManyContainersExisting(servName, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data container, got %v", n)
	}

	do := def.NowDo()
	do.Name = servName
	do.Rm = false
	do.RmD = false
	do.Operations.Args = []string{servName}
	if err := KillService(do); err != nil {
		t.Fatalf("expected service to be stopped, got %v")
	}
	if n := util.HowManyContainersRunning(servName, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 running service container, got %v", n)
	}
	if n := util.HowManyContainersExisting(servName, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 existing service container, got %v", n)
	}
	if n := util.HowManyContainersExisting(servName, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data container, got %v", n)
	}
}

func TestRmService(t *testing.T) {
	defer tests.RemoveAllContainers()

	start(t, servName, false)
	if n := util.HowManyContainersRunning(servName, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 running service container, got %v", n)
	}
	if n := util.HowManyContainersExisting(servName, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data container, got %v", n)
	}

	do := def.NowDo()
	do.Name = servName
	do.Operations.Args = []string{servName}
	do.Force = true
	do.File = false
	do.RmD = true
	if err := RmService(do); err != nil {
		t.Fatalf("expected service to be removed, got %v", err)
	}

	if n := util.HowManyContainersRunning(servName, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 running service container, got %v", n)
	}
	if n := util.HowManyContainersExisting(servName, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 existing service container, got %v", n)
	}
	if n := util.HowManyContainersExisting(servName, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data container, got %v", n)
	}
}

func TestExportService(t *testing.T) {
	do := def.NowDo()
	do.Name = "ipfs"

	const hash = "QmQ1LZYPNG4wSb9dojRicWCmM4gFLTPKFUhFnMTR3GKuA2"

	// Fake IPFS server.
	os.Setenv("ERIS_IPFS_HOST", "http://127.0.0.1")
	ipfs := tests.NewServer("127.0.0.1:8080")
	ipfs.SetResponse(tests.ServerResponse{
		Code: http.StatusOK,
		Header: map[string][]string{
			"Ipfs-Hash": {hash},
		},
	})
	defer ipfs.Close()

	if err := ExportService(do); err != nil {
		t.Fatalf("expected service to be exported, got %v", err)
	}

	if expected := "/ipfs/"; ipfs.Path() != expected {
		t.Fatalf("called the wrong endpoint; expected %v, got %v", expected, ipfs.Path())
	}

	if expected := "POST"; ipfs.Method() != expected {
		t.Fatalf("used the wrong HTTP method; expected %v, got %v", expected, ipfs.Method())
	}

	if content := tests.FileContents(FindServiceDefinitionFile(do.Name)); content != ipfs.Body() {
		t.Fatalf("sent the bad file; expected %v, got %v", content, ipfs.Body())
	}

	if hash != do.Result {
		t.Fatalf("hash mismatch; expected %v, got %v", hash, do.Result)
	}
}

func TestImportService(t *testing.T) {
	do := def.NowDo()
	do.Name = "eth"
	do.Hash = "QmQ1LZYPNG4wSb9dojRicWCmM4gFLTPKFUhFnMTR3GKuA2"

	content := `name = "ipfs"

[service]
name = "ipfs"
image = "` + path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_IPFS) + `"`

	// Fake IPFS server.
	os.Setenv("ERIS_IPFS_HOST", "http://127.0.0.1")
	ipfs := tests.NewServer("127.0.0.1:8080")
	ipfs.SetResponse(tests.ServerResponse{
		Code: http.StatusOK,
		Body: content,
	})
	defer ipfs.Close()

	if err := ImportService(do); err != nil {
		t.Fatalf("expected service to be imported, got %v", err)
	}

	if expected := "/ipfs/" + do.Hash; ipfs.Path() != expected {
		t.Fatalf("called the wrong endpoint; expected %v, got %v", expected, ipfs.Path())
	}

	if expected := "GET"; ipfs.Method() != expected {
		t.Fatalf("used the wrong HTTP method; expected %v, got %v\n", expected, ipfs.Method())
	}

	if imported := tests.FileContents(FindServiceDefinitionFile(do.Name)); imported != content {
		t.Fatalf("returned unexpected content; expected: %v, got %v", content, imported)
	}
}

func TestNewService(t *testing.T) {
	defer tests.RemoveAllContainers()

	do := def.NowDo()
	servName := "keys"
	do.Name = servName
	do.Operations.Args = []string{path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_KEYS)}
	if err := NewService(do); err != nil {
		t.Fatalf("expected a new service to be created, got %v", err)
	}

	do = def.NowDo()
	do.Operations.Args = []string{servName}
	if err := StartService(do); err != nil {
		t.Fatalf("expected service to be started, got %v", err)
	}

	if n := util.HowManyContainersRunning(servName, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 running service container, got %v", n)
	}
	if n := util.HowManyContainersExisting(servName, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data container, got %v", n)
	}

	kill(t, servName, true)
	if n := util.HowManyContainersExisting(servName, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 service containers, got %v", n)
	}
	if n := util.HowManyContainersExisting(servName, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data container, got %v", n)
	}

}

func TestRenameService(t *testing.T) {
	defer tests.RemoveAllContainers()

	start(t, "keys", false)
	if n := util.HowManyContainersRunning("keys", def.TypeService); n != 1 {
		t.Fatalf("#1. expecting 1 running service container, got %v", n)
	}
	if n := util.HowManyContainersExisting("keys", def.TypeData); n != 1 {
		t.Fatalf("#1. expecting 1 data container, got %v", n)
	}

	do := def.NowDo()
	do.Name = "keys"
	do.NewName = "syek"
	if err := RenameService(do); err != nil {
		t.Fatalf("expected service to be renamed, got %v", err)
	}

	if n := util.HowManyContainersRunning("keys", def.TypeService); n != 0 {
		t.Fatalf("#2. expecting 0 running service container, got %v", n)
	}
	if n := util.HowManyContainersExisting("keys", def.TypeData); n != 0 {
		t.Fatalf("#2. expecting 0 data container, got %v", n)
	}
	if n := util.HowManyContainersRunning("syek", def.TypeService); n != 1 {
		t.Fatalf("#3. expecting 1 running service container, got %v", n)
	}
	if n := util.HowManyContainersExisting("syek", def.TypeData); n != 1 {
		t.Fatalf("#3. expecting 1 data container, got %v", n)
	}

	do = def.NowDo()
	do.Name = "syek"
	do.NewName = "keys"
	if err := RenameService(do); err != nil {
		t.Fatalf("expected service to be renamed back, got %v", err)
	}

	if n := util.HowManyContainersRunning("syek", def.TypeService); n != 0 {
		t.Fatalf("#4. expecting 0 running service container, got %v", n)
	}
	if n := util.HowManyContainersExisting("syek", def.TypeData); n != 0 {
		t.Fatalf("#4. expecting 0 data container, got %v", n)
	}
	if n := util.HowManyContainersRunning("keys", def.TypeService); n != 1 {
		t.Fatalf("#5. expecting 1 running service container, got %v", n)
	}
	if n := util.HowManyContainersExisting("keys", def.TypeData); n != 1 {
		t.Fatalf("#5. expecting 1 data container, got %v", n)
	}
}

func TestCatService(t *testing.T) {
	do := def.NowDo()
	do.Name = servName
	if err := CatService(do); err != nil {
		t.Fatalf("expected cat to succeed, got %v", err)
	}

	if out := tests.FileContents(filepath.Join(config.GlobalConfig.ErisDir, "services", "ipfs.toml")); out != do.Result {
		t.Fatalf("expected local config to be returned %v, got %v", out, do.Result)
	}
}

func TestStartKillServiceWithDependencies(t *testing.T) {
	defer tests.RemoveAllContainers()

	do := def.NowDo()
	do.Operations.Args = []string{"do_not_use"}
	if err := StartService(do); err != nil {
		t.Fatalf("expected service to start, got %v", err)
	}

	if n := util.HowManyContainersRunning(servName, def.TypeService); n != 1 {
		t.Fatalf("expecting 1 running service container, got %v", n)
	}
	if n := util.HowManyContainersExisting(servName, def.TypeData); n != 1 {
		t.Fatalf("expecting 1 data container, got %v", n)
	}
	if n := util.HowManyContainersRunning("keys", def.TypeService); n != 1 {
		t.Fatalf("expecting 1 running dependent service container, got %v", n)
	}
	if n := util.HowManyContainersExisting("keys", def.TypeData); n != 1 {
		t.Fatalf("expecting 1 dependent data container, got %v", n)
	}

	kill(t, "do_not_use", true)

	if n := util.HowManyContainersRunning(servName, def.TypeService); n != 0 {
		t.Fatalf("expecting 0 running service container, got %v", n)
	}
	if n := util.HowManyContainersExisting(servName, def.TypeData); n != 0 {
		t.Fatalf("expecting 0 data container, got %v", n)
	}
	if n := util.HowManyContainersRunning("keys", def.TypeService); n != 0 {
		t.Fatalf("expecting 0 running dependent service container, got %v", n)
	}
	if n := util.HowManyContainersExisting("keys", def.TypeData); n != 0 {
		t.Fatalf("expecting 0 dependent data container, got %v", n)
	}
}

func start(t *testing.T, serviceName string, publishAll bool) {
	do := def.NowDo()
	do.Operations.Args = []string{serviceName}
	do.Operations.ContainerNumber = 1
	do.Operations.PublishAllPorts = publishAll
	if err := StartService(do); err != nil {
		t.Fatalf("expected service to start, got %v", err)
	}
}

func kill(t *testing.T, serviceName string, wipe bool) {
	do := def.NowDo()
	do.Name = serviceName
	do.Operations.Args = []string{serviceName}
	if wipe {
		do.Rm = true
		do.RmD = true
	}
	if err := KillService(do); err != nil {
		t.Fatalf("expected service to be stopped, got %v", err)
	}
}
