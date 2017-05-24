package pkgs

import (
	"os"
	"testing"

	"github.com/monax/monax/log"
	"github.com/monax/monax/testutil"
)

// TODO write well-defined user stories for these tests

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	testutil.IfExit(testutil.Init(testutil.Pull{
		Images:   []string{"data", "db", "keys", "compilers"},
		Services: []string{"keys", "compilers"},
	}))

	exitCode := m.Run()
	log.Info("Tearing tests down")
	testutil.IfExit(testutil.TearDown())
	os.Exit(exitCode)
}

func TestBadPathsGiven(t *testing.T) {
}

func TestImportEPMYamlInMainDir(t *testing.T) {
}

func TestImportEPMYamlNotInContractDir(t *testing.T) {
}

func TestImportMainDirRel(t *testing.T) {
}

func TestImportContractDirRel(t *testing.T) {
}

func TestImportContractDirAbs(t *testing.T) {
}

func TestImportContractDirAsFile(t *testing.T) {
}

func TestImportABIDirRel(t *testing.T) {
}

func TestImportABIDirAbs(t *testing.T) {
}

func TestImportABIDirAsFile(t *testing.T) {
}

func TestExportEPMOutputsInMainDir(t *testing.T) {
}

func TestExportEPMOutputsNotInMainDir(t *testing.T) {
}
