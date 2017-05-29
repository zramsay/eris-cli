package services

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/testutil"
	"github.com/monax/monax/util"
)

const servName = "keys"

func TestMain(m *testing.M) {
	log.SetLevel(log.WarnLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	testutil.IfExit(testutil.Init(testutil.Pull{
		Images:   []string{"data", "db", "keys"},
		Services: []string{"keys"},
	}))

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
