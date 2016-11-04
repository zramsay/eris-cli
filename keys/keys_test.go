package keys

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/testutil"
	"github.com/eris-ltd/eris-cli/util"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	testutil.IfExit(testutil.Init(testutil.Pull{
		Images:   []string{"keys", "data"},
		Services: []string{"keys"},
	}))

	exitCode := m.Run()
	testutil.IfExit(testutil.TearDown())
	os.Exit(exitCode)
}

func TestGenerateKey(t *testing.T) {
	testStartKeys(t)
	defer testKillService(t, "keys", true)

	address := testsGenAKey()

	output := testListKeys("container")

	if len(output) != 1 {
		t.Fatalf("Expected one key, got (%v)", len(output))
	}

	if address != output[0] {
		t.Fatalf("Expected (%s), got (%s)", address, output[0])
	}
}

func TestExportKeySingle(t *testing.T) {
	testStartKeys(t)
	defer testKillService(t, "keys", true)

	address := testsGenAKey()

	keyPath := path.Join(config.ErisContainerRoot, "keys", "data", address, address)

	//cat container contents of new key
	catOut, err := services.ExecHandler("keys", []string{"cat", keyPath})
	if err != nil {
		t.Fatalf("error exec-ing: %v", err)
	}

	keyInCont := strings.TrimSpace(catOut.String())

	doExp := definitions.NowDo()
	doExp.Address = address
	doExp.Destination = filepath.Join(config.KeysPath, "data") //is default

	//export
	if err := ExportKey(doExp); err != nil {
		t.Fatalf("error exporting: %v", err)
	}

	//cat host contents
	key, err := ioutil.ReadFile(filepath.Join(doExp.Destination, address, address))
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}

	keyOnHost := strings.TrimSpace(string(key))
	if keyInCont != keyOnHost {
		t.Fatalf("Expected (%s), got (%s)", keyInCont, keyOnHost)
	}
}

func TestImportKeyAll(t *testing.T) {
	testStartKeys(t)

	// gen some keys
	addrs := make(map[string]bool)
	addrs[testsGenAKey()] = true
	addrs[testsGenAKey()] = true

	// export them to host
	doExp := definitions.NowDo()
	doExp.All = true

	//export
	if err := ExportKey(doExp); err != nil {
		t.Fatalf("error exporting: %v", err)
	}

	// kill container
	testKillService(t, "keys", true)

	// start keys
	testStartKeys(t)
	defer testKillService(t, "keys", true)

	// eris keys import --all
	doImp := definitions.NowDo()
	doImp.All = true

	if err := ImportKey(doImp); err != nil {
		t.Fatalf("error exporting: %v", err)
	}

	// check that they in container
	output := testListKeys("container")

	i := 0
	for _, out := range output {
		if addrs[strings.TrimSpace(out)] == true {
			i++
		}
	}

	if i != 2 {
		t.Fatalf("Expected 2 keys, got (%v)", i)
	}
}

func TestExportKeyAll(t *testing.T) {
	testStartKeys(t)
	defer testKillService(t, "keys", true)
	// gen some keys
	addrs := make(map[string]bool)
	addrs[testsGenAKey()] = true
	addrs[testsGenAKey()] = true

	// export them all
	doExp := definitions.NowDo()
	doExp.All = true
	//export
	if err := ExportKey(doExp); err != nil {
		t.Fatalf("error exporting: %v", err)
	}
	// check that they on host
	output := testListKeys("host")

	i := 0
	for _, out := range output {
		if addrs[strings.TrimSpace(out)] == true {
			i++
		}
	}

	if i != 2 {
		t.Fatalf("Expected 2 keys, got (%v)", i)
	}
}

func TestImportKeySingle(t *testing.T) {
	testStartKeys(t)
	defer testKillService(t, "keys", true)

	address := testsGenAKey()

	//export it
	doExp := definitions.NowDo()
	doExp.Address = address

	if err := ExportKey(doExp); err != nil {
		t.Fatalf("error exporting key: %v", err)
	}

	key, err := ioutil.ReadFile(filepath.Join(doExp.Destination, address, address))
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}
	//key b4 import
	keyOnHost := strings.TrimSpace(string(key))

	//rm key that was generated before import
	keyPath := path.Join(config.ErisContainerRoot, "keys", "data", address)

	if _, err := services.ExecHandler("keys", []string{"rm", "-rf", keyPath}); err != nil {
		t.Fatalf("error exec-ing: %v", err)
	}

	doImp := definitions.NowDo()
	doImp.Address = address

	if err := ImportKey(doImp); err != nil {
		t.Fatalf("error importing key: %v", err)
	}

	keyPathCat := path.Join(config.ErisContainerRoot, "keys", "data", address, address)

	//cat container contents of new key
	catOut, err := services.ExecHandler("keys", []string{"cat", keyPathCat})
	if err != nil {
		t.Fatalf("error exec-ing: %v", err)
	}

	keyInCont := strings.TrimSpace(catOut.String())

	if keyOnHost != keyInCont {
		t.Fatalf("Expected (%s), got (%s)", keyOnHost, keyInCont)
	}
}

func TestListKeyContainer(t *testing.T) {
	testStartKeys(t)
	defer testKillService(t, "keys", true)

	addrs := make(map[string]bool)
	addrs[testsGenAKey()] = true
	addrs[testsGenAKey()] = true

	output := testListKeys("container")

	i := 0
	for _, out := range output {
		if addrs[strings.TrimSpace(out)] == true {
			i++
		}
	}

	if i != 2 {
		t.Fatalf("Expected 2 keys, got (%v)", i)
	}
}

func TestListKeyHost(t *testing.T) {
	testStartKeys(t)
	defer testKillService(t, "keys", true)

	addr0 := testsGenAKey()
	addr1 := testsGenAKey()

	addrs := make(map[string]bool)
	addrs[addr0] = true
	addrs[addr1] = true

	doExp := definitions.NowDo()
	doExp.All = true

	if err := ExportKey(doExp); err != nil {
		t.Fatalf("error exporting key: %v", err)
	}

	output := testListKeys("host")

	i := 0
	for _, out := range output {
		if addrs[strings.TrimSpace(out)] == true {
			i++
		}
	}

	if i != 2 {
		t.Fatalf("Expected 2 keys, got (%v)", i)
	}
}

func testListKeys(typ string) []string {
	do := definitions.NowDo()

	if typ == "container" {
		do.Container = true
		do.Host = false
	} else if typ == "host" {
		do.Container = false
		do.Host = true
	}

	result, err := ListKeys(do)
	if err != nil {
		testutil.IfExit(err)
	}

	return result
}

func testsGenAKey() string {
	addr := new(bytes.Buffer)
	config.Global.Writer = addr
	doGen := definitions.NowDo()
	testutil.IfExit(GenerateKey(doGen))

	addrBytes := addr.Bytes()
	return strings.TrimSpace(string(addrBytes))
}

func testStartKeys(t *testing.T) {
	serviceName := "keys"
	do := definitions.NowDo()
	do.Operations.Args = []string{serviceName}
	e := services.StartService(do)
	if e != nil {
		t.Fatalf("Error starting service: %v", e)
	}

	testExistAndRun(t, serviceName, true, true)
	testNumbersExistAndRun(t, serviceName, true, true)
}

func testKillService(t *testing.T, serviceName string, wipe bool) {
	do := definitions.NowDo()
	do.Name = serviceName
	do.Operations.Args = []string{serviceName}
	if wipe {
		do.Rm = true
		do.RmD = true
	}
	e := services.KillService(do)
	if e != nil {
		t.Fatalf("error killing service: %v", e)
	}
	testExistAndRun(t, serviceName, !wipe, false)
	testNumbersExistAndRun(t, serviceName, false, false)
}

func testExistAndRun(t *testing.T, servName string, toExist, toRun bool) {
	testutil.IfExit(testutil.ExistAndRun(servName, "service", toExist, toRun))
}

func testNumbersExistAndRun(t *testing.T, servName string, containerExist, containerRun bool) {
	log.WithFields(log.Fields{
		"=>":        servName,
		"existing#": containerExist,
		"running#":  containerRun,
	}).Info("Checking number of containers for")
	log.WithField("=>", servName).Debug("Checking existing containers for")
	exist := util.Exists(definitions.TypeService, servName)
	log.WithField("=>", servName).Debug("Checking running containers for")
	run := util.Running(definitions.TypeService, servName)

	if exist != containerExist {
		log.WithFields(log.Fields{
			"service":  servName,
			"expected": containerExist,
			"got":      exist,
		}).Error("Wrong number of existing containers")
		t.Fatalf("Bad failure")
	}

	if run != containerRun {
		log.WithFields(log.Fields{
			"service":  servName,
			"expected": containerExist,
			"got":      run,
		}).Error("Wrong number of running containers")
		t.Fatalf("Bad failure")
	}

	log.Info("All good")
}
