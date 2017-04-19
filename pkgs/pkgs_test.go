package pkgs

import (
	"fmt"
	"io/ioutil"
	"os"
	//"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monax/cli/chains"
	"github.com/monax/cli/config"
	"github.com/monax/cli/data"
	"github.com/monax/cli/definitions"
	//"github.com/monax/cli/loaders"
	"github.com/monax/cli/log"
	"github.com/monax/cli/services"
	"github.com/monax/cli/testutil"
	//"github.com/monax/cli/util"
	//"github.com/monax/cli/version"
)

var goodPkg string = filepath.Join(config.AppsPath, "good", "package.json")
var badPkg string = filepath.Join(config.AppsPath, "bad", "package.json")
var emptyPkg string = filepath.Join(config.AppsPath, "empty", "package.json")

var chainName = "pkg-test-chain"

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	testutil.IfExit(testutil.Init(testutil.Pull{
		Images:   []string{"data", "db", "keys"},
		Services: []string{"keys"},
	}))

	exitCode := m.Run()
	killKeys()
	log.Info("Tearing tests down")
	testutil.IfExit(testutil.TearDown())
	os.Exit(exitCode)
}

func TestJobManagerBasicRunning(t *testing.T) {
	const (
		filename = "./epm.yaml"
		jobsfile = `
jobs:

- name: setStorageBase
  set:
    val: 5

- name: setAccount
  account:
    address: 1234567890
`
	)
	err := ioutil.WriteFile(filename, []byte(jobsfile), 0644)
	defer os.Remove(filename)
	if err != nil {
		t.Fatalf("cannot write config file %v", err)
	}
	do := definitions.NowDo()
	do.YAMLPath = filename
	output, err := loaders.LoadJobs(do)
	if err != nil {
		t.Fatalf("could not load jobs: %v", err)
	}
	if err = output.RunJobs(); err != nil {
		t.Fatalf("running jobs resulted in err: %v", err)
	}
}

func TestJobManagerLegacy(t *testing.T) {
	const (
		filename = "./epm.yaml"
		jobsfile = `
jobs:

- name: setStorageBase
  job:
    set:
      val: 5
      
- name: something
  job:
    account:
      address: 123457890
`
	)
	err := ioutil.WriteFile(filename, []byte(jobsfile), 0644)
	defer os.Remove(filename)
	if err != nil {
		t.Fatalf("cannot write config file %v", err)
	}
	do := definitions.NowDo()
	do.YAMLPath = filename
	output, err := loaders.LoadJobs(do)
	if err != nil {
		t.Fatalf("could not load jobs: %v", err)
	}
	if err = output.RunJobs(); err != nil {
		t.Fatalf("running jobs resulted in err: %v", err)
	}
}

func killKeys() {
	do := definitions.NowDo()
	do.Operations.Args = []string{"keys"}
	do.Rm = true
	do.RmD = true
	services.KillService(do)
}
