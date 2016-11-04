package services

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/testutil"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"
)

const servName = "ipfs"

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	testutil.IfExit(testutil.Init(testutil.Pull{
		Images:   []string{"data", "db", "keys", "ipfs"},
		Services: []string{"do_not_use", "keys", "ipfs"},
	}))

	// Prevent CLI from starting IPFS.
	os.Setenv("ERIS_SKIP_ENSURE", "true")

	exitCode := m.Run()
	testutil.IfExit(testutil.TearDown())
	os.Exit(exitCode)
}

func TestStartKillService(t *testing.T) {
	defer testutil.RemoveAllContainers()

	start(t, servName, false)
	if !util.Running(definitions.TypeService, servName) {
		t.Fatalf("expecting service running")
	}
	if !util.Exists(definitions.TypeData, servName) {
		t.Fatalf("expecting dependent data container exists")
	}

	kill(t, servName, true)
	if util.Running(definitions.TypeService, servName) {
		t.Fatalf("expecting service stopped")
	}
	if util.Exists(definitions.TypeData, servName) {
		t.Fatalf("expecting dependent data container doesn't exist")
	}

}

func TestInspectService1(t *testing.T) {
	defer testutil.RemoveAllContainers()

	start(t, servName, false)

	do := definitions.NowDo()
	do.Name = servName
	do.Operations.Args = []string{"name"}

	if err := InspectService(do); err != nil {
		t.Fatalf("expected service to be inspected, got %v", err)
	}
}

func TestInspectService2(t *testing.T) {
	defer testutil.RemoveAllContainers()

	start(t, servName, false)

	do := definitions.NowDo()
	do.Name = servName
	do.Operations.Args = []string{"config.user"}

	if err := InspectService(do); err != nil {
		t.Fatalf("expected service to be inspected, got %v", err)
	}
}

func TestLogsService(t *testing.T) {
	defer testutil.RemoveAllContainers()

	start(t, servName, false)

	do := definitions.NowDo()
	do.Name = servName
	do.Follow = false
	do.Tail = "5"

	if err := LogsService(do); err != nil {
		t.Fatalf("expected service to return logs, got %v", err)
	}
}

func TestExecService(t *testing.T) {
	defer testutil.RemoveAllContainers()

	start(t, servName, true)

	do := definitions.NowDo()
	do.Name = servName
	do.Operations.Interactive = false
	do.Operations.Args = strings.Fields("ls -la /root/")

	buf, err := ExecService(do)
	if err != nil {
		t.Fatalf("expected to execute service, got %v", err)
	}

	if !strings.Contains(buf.String(), ".") {
		t.Fatalf("expected a file in the exec output, got %v", buf.String())
	}
}

func TestExecServiceBadCommandLine(t *testing.T) {
	defer testutil.RemoveAllContainers()

	start(t, servName, true)

	buf := new(bytes.Buffer)
	config.Global.Writer = buf

	do := definitions.NowDo()
	do.Name = servName
	do.Operations.Interactive = false
	do.Operations.Args = strings.Fields("bad command line")

	if _, err := ExecService(do); err == nil {
		t.Fatal("expected executing service to fail")
	}
}

func TestUpdateService(t *testing.T) {
	defer testutil.RemoveAllContainers()

	start(t, servName, false)

	do := definitions.NowDo()
	do.Name = servName
	do.Pull = false
	do.Timeout = 1
	if err := UpdateService(do); err != nil {
		t.Fatalf("expected service to be updated, got %v", err)
	}

	if !util.Running(definitions.TypeService, servName) {
		t.Fatalf("expecting service running")
	}

}

func TestKillService(t *testing.T) {
	defer testutil.RemoveAllContainers()

	start(t, servName, false)
	if !util.Running(definitions.TypeService, servName) {
		t.Fatalf("expecting service running")
	}
	if !util.Exists(definitions.TypeData, servName) {
		t.Fatalf("expecting dependent data container exists")
	}

	do := definitions.NowDo()
	do.Name = servName
	do.Rm = false
	do.RmD = false
	do.Operations.Args = []string{servName}
	if err := KillService(do); err != nil {
		t.Fatalf("expected service to be stopped, got %v", err)
	}
	if util.Running(definitions.TypeService, servName) {
		t.Fatalf("expecting service doesn't run")
	}
	if !util.Exists(definitions.TypeService, servName) {
		t.Fatalf("expecting service existing")
	}
	if !util.Exists(definitions.TypeData, servName) {
		t.Fatalf("expecting dependent data container exists")
	}
}

func TestRmService(t *testing.T) {
	defer testutil.RemoveAllContainers()

	start(t, servName, false)
	if !util.Running(definitions.TypeService, servName) {
		t.Fatalf("expecting service running")
	}
	if !util.Exists(definitions.TypeData, servName) {
		t.Fatalf("expecting dependent data container exists")
	}

	do := definitions.NowDo()
	do.Name = servName
	do.Operations.Args = []string{servName}
	do.Force = true
	do.File = false
	do.RmD = true
	if err := RmService(do); err != nil {
		t.Fatalf("expected service to be removed, got %v", err)
	}

	if util.Exists(definitions.TypeService, servName) {
		t.Fatalf("expecting service not existing")
	}
	if util.Exists(definitions.TypeData, servName) {
		t.Fatalf("expecting dependent data container not existing")
	}
}

func TestMakeService(t *testing.T) {
	defer testutil.RemoveAllContainers()

	do := definitions.NowDo()
	servName := "keys"
	do.Name = servName
	do.Operations.Args = []string{path.Join(version.DefaultRegistry, version.ImageKeys)}
	if err := MakeService(do); err != nil {
		t.Fatalf("expected a new service to be created, got %v", err)
	}

	do = definitions.NowDo()
	do.Operations.Args = []string{servName}
	if err := StartService(do); err != nil {
		t.Fatalf("expected service to be started, got %v", err)
	}

	if !util.Running(definitions.TypeService, servName) {
		t.Fatalf("expecting service running")
	}
	if !util.Exists(definitions.TypeData, servName) {
		t.Fatalf("expecting dependent data container exists")
	}

	kill(t, servName, true)
	if util.Exists(definitions.TypeService, servName) {
		t.Fatalf("expecting service not existing")
	}
	if util.Exists(definitions.TypeData, servName) {
		t.Fatalf("expecting dependent data container not existing")
	}

}

func TestRenameService(t *testing.T) {
	defer testutil.RemoveAllContainers()

	start(t, "keys", false)
	if !util.Running(definitions.TypeService, "keys") {
		t.Fatalf("expecting keys service running")
	}
	if !util.Exists(definitions.TypeData, "keys") {
		t.Fatalf("expecting keys data container exists")
	}

	do := definitions.NowDo()
	do.Name = "keys"
	do.NewName = "syek"
	if err := RenameService(do); err != nil {
		t.Fatalf("expected service to be renamed, got %v", err)
	}

	if util.Running(definitions.TypeService, "keys") {
		t.Fatalf("expecting keys service not running")
	}
	if util.Exists(definitions.TypeData, "keys") {
		t.Fatalf("expecting keys data container doesn't exist")
	}
	if !util.Running(definitions.TypeService, "syek") {
		t.Fatalf("expecting syek service running")
	}
	if !util.Exists(definitions.TypeData, "syek") {
		t.Fatalf("expecting keys data container exists")
	}

	do = definitions.NowDo()
	do.Name = "syek"
	do.NewName = "keys"
	if err := RenameService(do); err != nil {
		t.Fatalf("expected service to be renamed back, got %v", err)
	}

	if util.Running(definitions.TypeService, "syek") {
		t.Fatalf("expecting syek service not running")
	}
	if util.Exists(definitions.TypeData, "syek") {
		t.Fatalf("expecting syek data container doesn't exist")
	}
	if !util.Running(definitions.TypeService, "keys") {
		t.Fatalf("expecting keys service running")
	}
	if !util.Exists(definitions.TypeData, "keys") {
		t.Fatalf("expecting keys data container exists")
	}
}

func TestCatService(t *testing.T) {
	do := definitions.NowDo()
	do.Name = servName
	buf := new(bytes.Buffer)
	config.Global.Writer = buf
	out, err := CatService(do)
	if err != nil {
		t.Fatalf("expected cat to succeed, got %v", err)
	}

	if cmp := testutil.FileContents(filepath.Join(config.ErisRoot, "services", "ipfs.toml")); out != cmp {
		t.Fatalf("expected local config to be returned %v, got %v", cmp, out)
	}
}

func TestStartKillServiceWithDependencies(t *testing.T) {
	defer testutil.RemoveAllContainers()

	do := definitions.NowDo()
	do.Operations.Args = []string{"do_not_use"} // [csk] we should make a fake service def instead of using this file

	if err := StartService(do); err != nil {
		t.Fatalf("expected service to start, got %v", err)
	}

	if !util.IsService("do_not_use", true) {
		t.Fatalf("expecting service container running, got false")
	}

	if !util.IsData("do_not_use") {
		t.Fatalf("expecting data container existing got false")
	}

	if !util.Running(definitions.TypeService, "keys") {
		t.Fatalf("expecting keys service running")
	}
	if !util.Exists(definitions.TypeData, "keys") {
		t.Fatalf("expecting keys data container exists")
	}

	kill(t, "do_not_use", true)

	if util.Running(definitions.TypeService, servName) {
		t.Fatalf("expecting test service not running")
	}
	if util.Exists(definitions.TypeData, servName) {
		t.Fatalf("expecting test data container doesn't exist")
	}
	if util.Running(definitions.TypeService, "keys") {
		t.Fatalf("expecting keys service not running")
	}
	if util.Exists(definitions.TypeData, "keys") {
		t.Fatalf("expecting keys data container doesn't exist")
	}

}

func start(t *testing.T, serviceName string, publishAll bool) {
	do := definitions.NowDo()
	do.Operations.Args = []string{serviceName}
	do.Operations.PublishAllPorts = publishAll
	if err := StartService(do); err != nil {
		t.Fatalf("expected service to start, got %v", err)
	}
}

func kill(t *testing.T, serviceName string, wipe bool) {
	do := definitions.NowDo()
	do.Name = serviceName
	do.Operations.Args = []string{serviceName}
	if wipe {
		do.Force = true
		do.Rm = true
		do.RmD = true
	}
	if err := KillService(do); err != nil {
		t.Fatalf("expected service to be stopped, got %v", err)
	}
}
