package pkgs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/testutil"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"
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
		Images:   []string{"data", "db", "pm", "cm", "keys", "quay.io/eris/compilers"},
		Services: []string{"keys", "ipfs", "compilers"},
	}))

	exitCode := m.Run()
	killKeys()
	log.Info("Tearing tests down")
	testutil.IfExit(testutil.TearDown())
	os.Exit(exitCode)
}

func TestServicesBooted(t *testing.T) {
	defer testutil.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	// write a good pkg.json
	if err := writeGoodPkgJson(); err != nil {
		t.Fatalf("unexpected error writing package.json: %v", err)
	}

	defer func() {
		if err := os.RemoveAll(filepath.Dir(goodPkg)); err != nil {
			t.Fatalf("error removing good package.json directory: %v", err)
		}
	}()

	// load a good pkg.json
	pkg, err := loaders.LoadPackage(goodPkg, "")
	if err != nil {
		t.Fatalf("unexpected error loading package: %v", err)
	}

	do := definitions.NowDo()
	do.Name = pkg.Name
	pkg.ChainName = chainName

	if err := BootServicesAndChain(do, pkg); err != nil {
		CleanUp(do, pkg)
		t.Fatalf("error booting chains and services: %v", err)
	}
	defer CleanUp(do, pkg)

	// check dependencies on
	for _, servName := range pkg.Dependencies.Services {
		if !util.Running(definitions.TypeService, servName) {
			t.Fatalf("expected service to run")
		}
		if !util.Exists(definitions.TypeData, servName) {
			t.Fatalf("expected data container to exist")
		}
	}

	// turn off dependencies
	doOff := definitions.NowDo()
	doOff.Operations.Args = pkg.Dependencies.Services
	doOff.RmD = true
	doOff.Rm = true
	if err := services.KillService(doOff); err != nil {
		t.Fatalf("error turning off services: %v", err)
	}

	// check dependencies off
	for _, servName := range pkg.Dependencies.Services {
		if util.Running(definitions.TypeService, servName) {
			t.Fatalf("expected service to stop")
		}
		if util.Exists(definitions.TypeData, servName) {
			t.Fatalf("expected data container not existing")
		}
	}
}

func TestCompilersBootedOnLocalCompilersFlag(t *testing.T) {
	defer testutil.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)
	// write a good pkg.json

	if err := writeGoodPkgJson(); err != nil {
		t.Fatalf("unexpected error writing package.json: %v", err)
	}

	// load a good pkg.json
	pkg, err := loaders.LoadPackage(goodPkg, "")
	if err != nil {
		t.Fatalf("unexpected error loading package: %v", err)
	}

	do := definitions.NowDo()
	do.Name = pkg.Name
	do.LocalCompiler = true
	pkg.ChainName = chainName

	defer func() {
		CleanUp(do, pkg)
		if err := os.RemoveAll(filepath.Dir(goodPkg)); err != nil {
			t.Fatalf("error removing good package.json directory: %v", err)
		}
	}()

	if err := BootServicesAndChain(do, pkg); err != nil {
		t.Fatalf("error booting chains and services: %v", err)
	}

	// check compilers on
	if !util.Running(definitions.TypeService, "compilers") {
		t.Fatalf("expected compiler to start")
	}
	if !util.Exists(definitions.TypeData, "compilers") {
		t.Fatalf("expected compilers data container to exist")
	}

	CleanUp(do, pkg)

	// check compilers off
	if util.Running(definitions.TypeService, "compilers") {
		t.Fatalf("expected compilers to stop")
	}
	if util.Exists(definitions.TypeData, "compilers") {
		t.Fatalf("expected compilers data container to not exist")
	}
}

func _TestKnownChainBoots(t *testing.T) {
	defer testutil.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	name := "good"
	pkg := loaders.DefaultPackage(name, chainName)
	doBoot := definitions.NowDo()
	defer CleanUp(doBoot, pkg)

	if err := BootServicesAndChain(doBoot, pkg); err != nil {
		CleanUp(doBoot, pkg)
		t.Fatalf("error booting chains and services: %v", err)
	}

	if !util.Running(definitions.TypeChain, chainName) {
		t.Fatalf("expected chain to run")
	}
	if !util.Exists(definitions.TypeData, chainName) {
		t.Fatalf("expected data container to exist")
	}
}

func TestLinkingToServicesAndChains(t *testing.T) {
	defer testutil.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	// write a good pkg.json
	if err := writeGoodPkgJson(); err != nil {
		t.Fatalf("unexpected error writing package.json: %v", err)
	}

	// load a good pkg.json
	pkg, err := loaders.LoadPackage(goodPkg, "")
	if err != nil {
		t.Fatalf("unexpected error loading package: %v", err)
	}

	do := definitions.NowDo()
	do.Name = pkg.Name
	do.LocalCompiler = true
	pkg.ChainName = chainName

	if err := BootServicesAndChain(do, pkg); err != nil {
		CleanUp(do, pkg)
		t.Fatalf("error booting chains and services: %v", err)
	}

	defer func() {
		CleanUp(do, pkg)

		doOff := definitions.NowDo()
		doOff.Operations.Args = pkg.Dependencies.Services
		doOff.RmD = true
		doOff.Rm = true
		if err := services.KillService(doOff); err != nil {
			t.Fatalf("error turning off services: %v", err)
		}

		if err := os.RemoveAll(filepath.Dir(goodPkg)); err != nil {
			t.Fatalf("error removing good package.json directory: %v", err)
		}
	}()

	if err := DefinePkgActionService(do, pkg); err != nil {
		t.Fatalf("unexpected error formulating the pkg service: %v", err)
	}

	if do.Service.Name != pkg.Name+"_tmp_"+do.Name {
		t.Fatalf("wrong service name, expected %s got %s", pkg.Name+"_tmp_"+do.Name, do.Service.Name)
	}

	if do.Service.Image != path.Join(version.DefaultRegistry, version.ImagePM) {
		t.Fatalf("wrong service image, expected %s got %s", path.Join(version.DefaultRegistry, version.ImagePM), do.Service.Image)
	}

	if !do.Service.AutoData {
		t.Fatalf("unexpectedly data containers are not turned on")
	}

	if do.Service.WorkDir != path.Join(config.ErisContainerRoot, "apps", filepath.Base(do.Path)) {
		t.Fatalf("wrong working directory, expected %s, got %s", path.Join(config.ErisContainerRoot, "apps", filepath.Base(do.Path)), do.Service.WorkDir)
	}

	if do.Service.User != "eris" {
		t.Fatalf("wrong user for the containers, expected %s, got %s", "eris", do.Service.User)
	}

	if err := checkLinks(do); err != nil {
		t.Fatalf("expected links to check out, got %v", err)
	}
}

func TestBadPathsGiven(t *testing.T) {
	defer testutil.RemoveAllContainers()

	create(t, chainName)
	defer kill(t, chainName)

	name := "homiedontplay"
	pkg := loaders.DefaultPackage(name, chainName)
	pkg.ChainName = chainName
	do := definitions.NowDo()

	defer func() {
		CleanUp(do, pkg)
	}()

	if err := DefinePkgActionService(do, pkg); err != nil {
		t.Fatalf("unexpected error formulating the pkg service: %v", err)
	}

	do.Path = "/qwerty"
	if err := getDataContainerSorted(do, true); err == nil {
		t.Fatalf("expected error not received")
	}
}

func TestImportEPMYamlInMainDir(t *testing.T) {
	dirName := "testerSteven"
	chainName := "simpletestingChain"
	name := "homiedontplay"
	contents := "marmots"
	dir := filepath.Join(config.AppsPath, dirName)

	pkg := loaders.DefaultPackage(name, chainName)
	pkg.ChainName = "temp"
	do := definitions.NowDo()

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}
	}()

	if err := DefinePkgActionService(do, pkg); err != nil {
		t.Fatalf("unexpected error formulating the pkg service: %v", err)
	}

	if err := writeTestFile(filepath.Join(dir, "epm.yaml"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}

	do.Path = dir
	do.PackagePath = filepath.Join(dir, "contracts")
	do.ABIPath = filepath.Join(dir, "abi")
	do.EPMConfigFile = filepath.Join(dir, "epm.yaml")
	if err := getDataContainerSorted(do, true); err != nil {
		t.Fatalf("unexpected error received on data import: %v", err)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/apps/%s/epm.yaml", dirName)}
	if out := exec(t, name, args); !strings.Contains(out, contents) {
		t.Fatalf("unexpected error in getting epm.yaml, expected %s, got %v", contents, out)
	}
}

func TestImportEPMYamlNotInContractDir(t *testing.T) {
	dirName := "testerSteven"
	dirName2 := "testerRichard"
	chainName := "simpletestingChain"
	name := "homiedontplay"
	contents := "marmots"
	dir := filepath.Join(config.AppsPath, dirName)
	dir2 := filepath.Join(config.AppsPath, dirName2)

	pkg := loaders.DefaultPackage(name, chainName)
	pkg.ChainName = "temp"
	do := definitions.NowDo()

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}

		if err := os.RemoveAll(dir2); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}
	}()

	if err := DefinePkgActionService(do, pkg); err != nil {
		t.Fatalf("unexpected error formulating the pkg service: %v", err)
	}

	file := filepath.Join(dir2, "epm.yaml")
	if err := os.MkdirAll(dir, 0775); err != nil {
		t.Fatalf("unexpected error making a test directory: %v", err)
	}
	if err := os.MkdirAll(dir2, 0775); err != nil {
		t.Fatalf("unexpected error making a test directory: %v", err)
	}
	f, err := os.Create(file)
	if err != nil {
		t.Fatalf("unexpected error creating a file in test directory: %v", err)
	}
	_, err = f.Write([]byte(contents))
	if err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}
	f.Close()

	do.Path = dir
	do.PackagePath = filepath.Join(dir, "contracts")
	do.ABIPath = filepath.Join(dir, "abi")
	do.EPMConfigFile = filepath.Join(dir2, "epm.yaml")
	if err := getDataContainerSorted(do, true); err != nil {
		t.Fatalf("unexpected error received on data import: %v", err)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/apps/%s/epm.yaml", dirName)}
	if out := exec(t, name, args); !strings.Contains(out, contents) {
		t.Fatalf("unexpected error in getting epm.yaml, expected %s, got %v", contents, out)
	}

	if err := os.RemoveAll(filepath.Join(dir2, "epm.yaml")); err != nil {
		t.Fatalf("unexpected error removing epm.yaml for a test: %v", err)
	}

	CleanUp(do, pkg)

	if out2, _ := ioutil.ReadFile(filepath.Join(dir2, "epm.yaml")); !strings.Contains(string(out2), contents) {
		t.Fatalf("unexpected error in getting epm.yaml, expected %s, got %s", contents, out2)
	}
}

func TestImportMainDirRel(t *testing.T) {
	pwd, _ := os.Getwd()
	os.Chdir(config.AppsPath)

	dirName := "testerSteven"
	chainName := "simpletestingChain"
	name := "homiedontplay"
	contents := "marmots"
	dir := filepath.Join(".", dirName)

	pkg := loaders.DefaultPackage(name, chainName)
	pkg.ChainName = "temp"
	do := definitions.NowDo()

	if err := DefinePkgActionService(do, pkg); err != nil {
		t.Fatalf("unexpected error formulating the pkg service: %v", err)
	}

	if err := writeTestFile(filepath.Join(dir, "epm.yaml"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}

	defer func() {
		CleanUp(do, pkg)

		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("unexpected error removing directory: %v", err)
		}

		os.Chdir(pwd)
	}()

	do.Path = filepath.Join(".", dirName)
	do.PackagePath = filepath.Join(dir, "contracts")
	do.ABIPath = filepath.Join(dir, "abi")
	do.EPMConfigFile = filepath.Join(dir, "epm.yaml")
	if err := getDataContainerSorted(do, true); err != nil {
		t.Fatalf("unexpected error received on data import: %v", err)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/apps/%s/epm.yaml", dirName)}
	if out := exec(t, name, args); !strings.Contains(out, contents) {
		t.Fatalf("unexpected error in getting epm.yaml, expected %s, got %v", contents, out)
	}
}

func TestImportMainDirAsFile(t *testing.T) {
	dirName := "testerSteven"
	chainName := "simpletestingChain"
	name := "homiedontplay"
	contents := "marmots"
	dir := filepath.Join(config.AppsPath, dirName)

	pkg := loaders.DefaultPackage(name, chainName)
	pkg.ChainName = "temp"
	do := definitions.NowDo()

	defer func() {
		CleanUp(do, pkg)

		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}
	}()

	if err := DefinePkgActionService(do, pkg); err != nil {
		t.Fatalf("unexpected error formulating the pkg service: %v", err)
	}

	if err := writeTestFile(filepath.Join(dir, "epm.yaml"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}

	do.Path = filepath.Join(dir, "epm.yaml")
	do.PackagePath = filepath.Join(dir, "contracts")
	do.ABIPath = filepath.Join(dir, "abi")
	do.EPMConfigFile = filepath.Join(dir, "epm.yaml")
	if err := getDataContainerSorted(do, true); err != nil {
		t.Fatalf("unexpected error received on data import: %v", err)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/apps/%s/epm.yaml", dirName)}
	if out := exec(t, name, args); !strings.Contains(out, contents) {
		t.Fatalf("unexpected error in getting epm.yaml, expected %s, got %v", contents, out)
	}
}

func TestImportContractDirRel(t *testing.T) {
	pwd, _ := os.Getwd()
	os.Chdir(config.AppsPath)

	dirName := "testerSteven"
	dirName2 := "testerRichard"
	chainName := "simpletestingChain"
	name := "homiedontplay"
	contents := "marmots"
	dir := filepath.Join(config.AppsPath, dirName)
	dir2 := filepath.Join(config.AppsPath, dirName2)

	pkg := loaders.DefaultPackage(name, chainName)
	pkg.ChainName = "temp"
	do := definitions.NowDo()

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}

		if err := os.RemoveAll(dir2); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}

		os.Chdir(pwd)
	}()

	if err := DefinePkgActionService(do, pkg); err != nil {
		t.Fatalf("unexpected error formulating the pkg service: %v", err)
	}

	if err := writeTestFile(filepath.Join(dir, "epm.yaml"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}
	if err := writeTestFile(filepath.Join(dir2, "fakeContract"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}

	do.Path = filepath.Join(dir)
	do.PackagePath = filepath.Join(".", filepath.Base(dir2))
	do.ABIPath = filepath.Join(dir, "abi")
	do.EPMConfigFile = filepath.Join(dir, "epm.yaml")
	if err := getDataContainerSorted(do, true); err != nil {
		t.Fatalf("unexpected error received on data import: %v", err)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/apps/%s/contracts/fakeContract", dirName)}
	if out := exec(t, name, args); !strings.Contains(out, contents) {
		t.Fatalf("unexpected error in getting fakeContract, expected %s, got %v", contents, out)
	}

	if err := os.RemoveAll(dir2); err != nil {
		t.Fatalf("error removing directory: %v", err)
	}

	CleanUp(do, pkg)

	if out2, _ := ioutil.ReadFile(filepath.Join(dir2, "fakeContract")); !strings.Contains(string(out2), contents) {
		t.Fatalf("unexpected error in getting fakeContract, expected %s, got %s", contents, out2)
	}
}

func TestImportContractDirAbs(t *testing.T) {
	dirName := "testerSteven"
	dirName2 := "testerRichard"
	chainName := "simpletestingChain"
	name := "homiedontplay"
	contents := "marmots"
	dir := filepath.Join(config.AppsPath, dirName)
	dir2 := filepath.Join(config.AppsPath, dirName2)

	pkg := loaders.DefaultPackage(name, chainName)
	pkg.ChainName = "temp"
	do := definitions.NowDo()

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}

		if err := os.RemoveAll(dir2); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}
	}()

	if err := DefinePkgActionService(do, pkg); err != nil {
		t.Fatalf("unexpected error formulating the pkg service: %v", err)
	}

	if err := writeTestFile(filepath.Join(dir, "epm.yaml"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}
	if err := writeTestFile(filepath.Join(dir2, "fakeContract"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}

	do.Path = filepath.Join(dir)
	do.PackagePath = dir2
	do.ABIPath = filepath.Join(dir, "abi")
	do.EPMConfigFile = filepath.Join(dir, "epm.yaml")
	if err := getDataContainerSorted(do, true); err != nil {
		t.Fatalf("unexpected error received on data import: %v", err)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/apps/%s/contracts/fakeContract", dirName)}
	if out := exec(t, name, args); !strings.Contains(out, contents) {
		t.Fatalf("unexpected error in getting fakeContract, expected %s, got %v", contents, out)
	}

	if err := os.RemoveAll(dir2); err != nil {
		t.Fatalf("error removing directory: %v", err)
	}

	CleanUp(do, pkg)

	if out2, _ := ioutil.ReadFile(filepath.Join(dir2, "fakeContract")); !strings.Contains(string(out2), contents) {
		t.Fatalf("unexpected error in getting fakeContract, expected %s, got %s", contents, out2)
	}
}

func TestImportContractDirAsFile(t *testing.T) {
	dirName := "testerSteven"
	dirName2 := "testerRichard"
	chainName := "simpletestingChain"
	name := "homiedontplay"
	contents := "marmots"
	dir := filepath.Join(config.AppsPath, dirName)
	dir2 := filepath.Join(config.AppsPath, dirName2)

	pkg := loaders.DefaultPackage(name, chainName)
	pkg.ChainName = "temp"
	do := definitions.NowDo()

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}

		if err := os.RemoveAll(dir2); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}
	}()

	if err := DefinePkgActionService(do, pkg); err != nil {
		t.Fatalf("unexpected error formulating the pkg service: %v", err)
	}

	if err := writeTestFile(filepath.Join(dir, "epm.yaml"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}
	if err := writeTestFile(filepath.Join(dir2, "fakeContract"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}

	do.Path = filepath.Join(dir)
	do.PackagePath = filepath.Join(dir2, "fakeContract")
	do.ABIPath = filepath.Join(dir, "abi")
	do.EPMConfigFile = filepath.Join(dir, "epm.yaml")
	if err := getDataContainerSorted(do, true); err != nil {
		t.Fatalf("unexpected error received on data import: %v", err)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/apps/%s/contracts/fakeContract", dirName)}
	if out := exec(t, name, args); !strings.Contains(out, contents) {
		t.Fatalf("unexpected error in getting fakeContract, expected %s, got %v", contents, out)
	}

	CleanUp(do, pkg)

	if out2, _ := ioutil.ReadFile(filepath.Join(dir2, "fakeContract")); !strings.Contains(string(out2), contents) {
		t.Fatalf("unexpected error in getting fakeContract, expected %s, got %s", contents, out2)
	}
}

func TestImportABIDirRel(t *testing.T) {
	pwd, _ := os.Getwd()
	os.Chdir(config.AppsPath)

	dirName := "testerSteven"
	dirName2 := "testerRichard"
	chainName := "simpletestingChain"
	name := "homiedontplay"
	contents := "marmots"
	dir := filepath.Join(config.AppsPath, dirName)
	dir2 := filepath.Join(config.AppsPath, dirName2)

	pkg := loaders.DefaultPackage(name, chainName)
	pkg.ChainName = "temp"
	do := definitions.NowDo()

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}

		if err := os.RemoveAll(dir2); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}

		os.Chdir(pwd)
	}()

	if err := DefinePkgActionService(do, pkg); err != nil {
		t.Fatalf("unexpected error formulating the pkg service: %v", err)
	}

	if err := writeTestFile(filepath.Join(dir, "epm.yaml"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}
	if err := writeTestFile(filepath.Join(dir2, "fakeContract"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}

	do.Path = filepath.Join(dir)
	do.PackagePath = filepath.Join(dir, "contracts")
	do.ABIPath = filepath.Join(".", filepath.Base(dir2))
	do.EPMConfigFile = filepath.Join(dir, "epm.yaml")
	if err := getDataContainerSorted(do, true); err != nil {
		t.Fatalf("unexpected error received on data import: %v", err)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/apps/%s/abi/fakeContract", dirName)}
	if out := exec(t, name, args); !strings.Contains(out, contents) {
		t.Fatalf("unexpected error in getting fakeContract, expected %s, got %v", contents, out)
	}

	if err := os.RemoveAll(dir2); err != nil {
		t.Fatalf("error removing directory: %v", err)
	}

	CleanUp(do, pkg)

	if out2, _ := ioutil.ReadFile(filepath.Join(dir2, "fakeContract")); !strings.Contains(string(out2), contents) {
		t.Fatalf("unexpected error in getting fakeContract, expected %s, got %s", contents, out2)
	}
}

func TestImportABIDirAbs(t *testing.T) {
	dirName := "testerSteven"
	dirName2 := "testerRichard"
	chainName := "simpletestingChain"
	name := "homiedontplay"
	contents := "marmots"
	dir := filepath.Join(config.AppsPath, dirName)
	dir2 := filepath.Join(config.AppsPath, dirName2)

	pkg := loaders.DefaultPackage(name, chainName)
	pkg.ChainName = "temp"
	do := definitions.NowDo()

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}

		if err := os.RemoveAll(dir2); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}
	}()

	if err := DefinePkgActionService(do, pkg); err != nil {
		t.Fatalf("unexpected error formulating the pkg service: %v", err)
	}

	if err := writeTestFile(filepath.Join(dir, "epm.yaml"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}
	if err := writeTestFile(filepath.Join(dir2, "fakeContract"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}

	do.Path = filepath.Join(dir)
	do.PackagePath = filepath.Join(dir, "contracts")
	do.ABIPath = dir2
	do.EPMConfigFile = filepath.Join(dir, "epm.yaml")
	if err := getDataContainerSorted(do, true); err != nil {
		t.Fatalf("unexpected error received on data import: %v", err)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/apps/%s/abi/fakeContract", dirName)}
	if out := exec(t, name, args); !strings.Contains(out, contents) {
		t.Fatalf("unexpected error in getting fakeContract, expected %s, got %v", contents, out)
	}

	if err := os.RemoveAll(dir2); err != nil {
		t.Fatalf("error removing directory: %v", err)
	}

	CleanUp(do, pkg)

	if out2, _ := ioutil.ReadFile(filepath.Join(dir2, "fakeContract")); !strings.Contains(string(out2), contents) {
		t.Fatalf("unexpected error in getting fakeContract, expected %s, got %s", contents, out2)
	}
}

func TestImportABIDirAsFile(t *testing.T) {
	dirName := "testerSteven"
	dirName2 := "testerRichard"
	chainName := "simpletestingChain"
	name := "homiedontplay"
	contents := "marmots"
	dir := filepath.Join(config.AppsPath, dirName)
	dir2 := filepath.Join(config.AppsPath, dirName2)

	pkg := loaders.DefaultPackage(name, chainName)
	pkg.ChainName = "temp"
	do := definitions.NowDo()

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}

		if err := os.RemoveAll(dir2); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}
	}()

	if err := DefinePkgActionService(do, pkg); err != nil {
		t.Fatalf("unexpected error formulating the pkg service: %v", err)
	}

	if err := writeTestFile(filepath.Join(dir, "epm.yaml"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}
	if err := writeTestFile(filepath.Join(dir2, "fakeContract"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}

	do.Path = filepath.Join(dir)
	do.PackagePath = filepath.Join(dir, "contracts")
	do.ABIPath = filepath.Join(dir2, "fakeContract")
	do.EPMConfigFile = filepath.Join(dir, "epm.yaml")
	if err := getDataContainerSorted(do, true); err != nil {
		t.Fatalf("unexpected error received on data import: %v", err)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/apps/%s/abi/fakeContract", dirName)}
	if out := exec(t, name, args); !strings.Contains(out, contents) {
		t.Fatalf("unexpected error in getting fakeContract, expected %s, got %v", contents, out)
	}

	CleanUp(do, pkg)

	if out2, _ := ioutil.ReadFile(filepath.Join(dir2, "fakeContract")); !strings.Contains(string(out2), contents) {
		t.Fatalf("unexpected error in getting fakeContract, expected %s, got %s", contents, out2)
	}
}

func TestExportEPMOutputsInMainDir(t *testing.T) {
	dirName := "testerSteven"
	chainName := "simpletestingChain"
	name := "homiedontplay"
	contents := "marmots"
	dir := filepath.Join(config.AppsPath, dirName)

	pkg := loaders.DefaultPackage(name, chainName)
	pkg.ChainName = "temp"
	do := definitions.NowDo()

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}
	}()

	if err := DefinePkgActionService(do, pkg); err != nil {
		t.Fatalf("unexpected error formulating the pkg service: %v", err)
	}

	if err := writeTestFile(filepath.Join(dir, "epm.yaml"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}

	if err := writeTestFile(filepath.Join(dir, "epm.csv"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}

	do.Path = dir
	do.PackagePath = filepath.Join(dir, "contracts")
	do.ABIPath = filepath.Join(dir, "abi")
	do.EPMConfigFile = filepath.Join(dir, "epm.yaml")
	if err := getDataContainerSorted(do, true); err != nil {
		t.Fatalf("unexpected error received on data import: %v", err)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/apps/%s/epm.csv", dirName)}
	if out := exec(t, name, args); !strings.Contains(out, contents) {
		t.Fatalf("unexpected error in getting epm.csv, expected %s, got %v", contents, out)
	}

	if err := os.RemoveAll(filepath.Join(dir, "epm.csv")); err != nil {
		t.Fatalf("error removing file: %v", err)
	}

	CleanUp(do, pkg)

	if out2, _ := ioutil.ReadFile(filepath.Join(dir, "epm.csv")); !strings.Contains(string(out2), contents) {
		t.Fatalf("unexpected error in getting epm.csv, expected %s, got %s", contents, out2)
	}
}

func TestExportEPMOutputsNotInMainDir(t *testing.T) {
	dirName := "testerSteven"
	dirName2 := "testerRichard"
	chainName := "simpletestingChain"
	name := "homiedontplay"
	contents := "marmots"
	dir := filepath.Join(config.AppsPath, dirName)
	dir2 := filepath.Join(config.AppsPath, dirName2)

	pkg := loaders.DefaultPackage(name, chainName)
	pkg.ChainName = "temp"
	do := definitions.NowDo()

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}

		if err := os.RemoveAll(dir2); err != nil {
			t.Fatalf("error removing directory: %v", err)
		}
	}()

	if err := DefinePkgActionService(do, pkg); err != nil {
		t.Fatalf("unexpected error formulating the pkg service: %v", err)
	}

	if err := writeTestFile(filepath.Join(dir, "epm.csv"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}

	if err := writeTestFile(filepath.Join(dir2, "epm.yaml"), contents); err != nil {
		t.Fatalf("unexpected error writing to test file: %v", err)
	}

	do.Path = dir
	do.PackagePath = filepath.Join(dir, "contracts")
	do.ABIPath = filepath.Join(dir, "abi")
	do.EPMConfigFile = filepath.Join(dir2, "epm.yaml")
	if err := getDataContainerSorted(do, true); err != nil {
		t.Fatalf("unexpected error received on data import: %v", err)
	}

	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/apps/%s/epm.csv", dirName)}
	if out := exec(t, name, args); !strings.Contains(out, contents) {
		t.Fatalf("unexpected error in getting epm.csv, expected %s, got %v", contents, out)
	}

	if err := os.RemoveAll(filepath.Join(dir2, "epm.csv")); err != nil {
		t.Fatalf("error removing file: %v", err)
	}

	CleanUp(do, pkg)

	if out2, _ := ioutil.ReadFile(filepath.Join(dir2, "epm.csv")); !strings.Contains(string(out2), contents) {
		t.Fatalf("unexpected error in getting epm.csv, expected %s, got %s", contents, out2)
	}
}

func startKeys() error {
	doKeys := definitions.NowDo()
	doKeys.Operations.Args = []string{"keys"}
	doKeys.Rm = true
	doKeys.RmD = true
	if err := services.StartService(doKeys); err != nil {
		return err
	}
	return nil
}

func killKeys() {
	do := definitions.NowDo()
	do.Operations.Args = []string{"keys"}
	do.Rm = true
	do.RmD = true
	services.KillService(do)
}

func writeGoodPkgJson() error {
	if _, err := os.Stat(goodPkg); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(goodPkg), 0755); err != nil {
			return err
		}
	}
	return ioutil.WriteFile(goodPkg, []byte(goodPkgContents()), 0644)
}

func goodPkgContents() string {
	return `{
  "name": "idis_app",
  "version": "0.0.1",
  "dependencies": {
    "eris-contracts": "^0.13.1",
    "prompt": "*"
  },
  "eris": {
		"package_id": "XXXXX",
		"chain_name": "simplechain",
		"chain_id": "YYYYYY",
		"chain_types": ["mint"],
		"environment": {
			"ASDF": "1234"
		},
		"dependencies": {
			"services": ["keys", "ipfs"]
		}
  }
}
`
}

func writeTestFile(filename, contents string) error {
	file := filepath.Join(filename)
	if err := os.MkdirAll(filepath.Dir(filename), 0775); err != nil {
		return err
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(contents))
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func writeBadPkgJson() error {
	if _, err := os.Stat(badPkg); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(badPkg), 0755); err != nil {
			return err
		}
	}
	return ioutil.WriteFile(badPkg, []byte(badPkgContents()), 0644)
}

func badPkgContents() string {
	return `{
  "name": "idis_app",
  "version": "0.0.1",
  "dependencies": {
    "eris-contracts": "^0.13.1",
    "prompt": "*"
  }
  "eris": {
		"package_id": "XXXXX"
		"environment": "ASDF=1234"
		"chain_name": "simplechain"
		"chain_id": "YYYYYY"
		"chain_types": ["mint"]
  }
}
`
}

func writeEmptyPkgJson() error {
	if _, err := os.Stat(emptyPkg); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(emptyPkg), 0755); err != nil {
			return err
		}
	}
	return ioutil.WriteFile(emptyPkg, []byte{}, 0644)
}

func checkLinks(do *definitions.Do) error {
	if do.LocalCompiler {
		notFound := true
		for _, srv := range do.ServicesSlice {
			if srv == "compilers" {
				notFound = false
			}
		}
		if notFound == true {
			return fmt.Errorf("local compiler not placed into do.ServicesSlice")
		}
	}
	for _, dep := range do.ServicesSlice {
		match := false
		for _, link := range do.Service.Links {
			if strings.HasSuffix(link, ":"+dep) {
				match = true
			}
		}
		if !match {
			return fmt.Errorf("chain or service not properly linked: %s", dep)
		}
	}

	return nil
}

func create(t *testing.T, chain string) {
	doMake := definitions.NowDo()
	doMake.Name = chain
	doMake.ChainType = "simplechain"
	if err := chains.MakeChain(doMake); err != nil {
		t.Fatalf("expected a chain to be made, got %v", err)
	}

	do := definitions.NowDo()
	do.Name = chain
	do.Operations.PublishAllPorts = true
	do.Path = filepath.Join(config.ChainsPath, chain)
	if err := chains.StartChain(do); err != nil {
		t.Fatalf("expected a new chain to be created, got %v", err)
	}
}

func kill(t *testing.T, chain string) {
	do := definitions.NowDo()
	do.Operations.Args, do.Rm, do.RmD = []string{"keys"}, true, true
	if err := services.KillService(do); err != nil {
		t.Fatalf("killing keys service failed: %v", err)
	}

	do = definitions.NowDo()
	do.Name, do.File, do.RmD, do.RmHF, do.Force = chain, true, true, true, true
	if err := chains.RemoveChain(do); err != nil {
		t.Fatalf("killing chain failed: %v", err)
	}
}

func exec(t *testing.T, name string, args []string) string {
	do := definitions.NowDo()
	do.Name = name + "_tmp_"
	do.Operations.Args = args
	buf, err := data.ExecData(do)
	if err != nil {
		t.Fatalf("expected %s to execute command [%s], got %v", name, strings.Join(args, " "), err)
	}

	return buf.String()
}
