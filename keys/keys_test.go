package keys

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	srv "github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/tests"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit(tests.ConnectAndPull))

	exitCode := m.Run()
	tests.IfExit(tests.TestsTearDown())
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

func TestGetPubKey(t *testing.T) {
	testStartKeys(t)
	defer testKillService(t, "keys", true)

	doPub := def.NowDo()
	doPub.Address = testsGenAKey()

	pub := new(bytes.Buffer)
	config.GlobalConfig.Writer = pub
	if err := GetPubKey(doPub); err != nil {
		t.Fatalf("error getting pubkey: %v", err)
	}

	pubkey := util.TrimString(pub.String())

	key := new(bytes.Buffer)
	config.GlobalConfig.Writer = key
	doKey := def.NowDo()
	doKey.Address = doPub.Address
	if err := ConvertKey(doKey); err != nil {
		t.Fatalf("error converting key: %v", err)
	}

	converted := regexp.MustCompile(`"pub_key":\[1,"([^"]+)"\]`).FindStringSubmatch(key.String())[1]

	if converted != pubkey {
		t.Fatalf("Expected (%s), got (%s)", pubkey, converted)
	}
}

func TestExportKeySingle(t *testing.T) {
	testStartKeys(t)
	defer testKillService(t, "keys", true)

	address := testsGenAKey()

	keyPath := path.Join(ErisContainerRoot, "keys", "data", address, address)

	//cat container contents of new key
	catOut, err := srv.ExecHandler("keys", []string{"cat", keyPath})
	if err != nil {
		t.Fatalf("error exec-ing: %v", err)
	}

	keyInCont := util.TrimString(catOut.String())

	doExp := def.NowDo()
	doExp.Address = address
	doExp.Destination = filepath.Join(KeysPath, "data") //is default

	//export
	if err := ExportKey(doExp); err != nil {
		t.Fatalf("error exporting: %v", err)
	}

	//cat host contents
	key, err := ioutil.ReadFile(filepath.Join(doExp.Destination, address, address))
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}

	keyOnHost := util.TrimString(string(key))
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
	doExp := def.NowDo()
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
	doImp := def.NowDo()
	doImp.All = true

	if err := ImportKey(doImp); err != nil {
		t.Fatalf("error exporting: %v", err)
	}

	// check that they in container
	output := testListKeys("container")

	i := 0
	for _, out := range output {
		if addrs[util.TrimString(out)] == true {
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
	doExp := def.NowDo()
	doExp.All = true
	//export
	if err := ExportKey(doExp); err != nil {
		t.Fatalf("error exporting: %v", err)
	}
	// check that they on host
	output := testListKeys("host")

	i := 0
	for _, out := range output {
		if addrs[util.TrimString(out)] == true {
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
	doExp := def.NowDo()
	doExp.Address = address
	doExp.Destination = filepath.Join(KeysPath, "data") //is default set by flag

	if err := ExportKey(doExp); err != nil {
		t.Fatalf("error exporting key: %v", err)
	}

	key, err := ioutil.ReadFile(filepath.Join(doExp.Destination, address, address))
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}
	//key b4 import
	keyOnHost := util.TrimString(string(key))

	//rm key that was generated before import
	keyPath := path.Join(ErisContainerRoot, "keys", "data", address)

	if _, err := srv.ExecHandler("keys", []string{"rm", "-rf", keyPath}); err != nil {
		t.Fatalf("error exec-ing: %v", err)
	}

	doImp := def.NowDo()
	doImp.Address = address
	//doImp.Destination // set in function
	doImp.Source = filepath.Join(KeysPath, "data")

	if err := ImportKey(doImp); err != nil {
		t.Fatalf("error importing key: %v", err)
	}

	keyPathCat := path.Join(ErisContainerRoot, "keys", "data", address, address)

	//cat container contents of new key
	catOut, err := srv.ExecHandler("keys", []string{"cat", keyPathCat})
	if err != nil {
		t.Fatalf("error exec-ing: %v", err)
	}

	keyInCont := util.TrimString(catOut.String())

	if keyOnHost != keyInCont {
		t.Fatalf("Expected (%s), got (%s)", keyOnHost, keyInCont)
	}
}

func TestConvertKey(t *testing.T) {
	// tested in TestGetPubKey
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
		if addrs[util.TrimString(out)] == true {
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

	doExp := def.NowDo()
	doExp.Address = addr0
	doExp.Destination = filepath.Join(KeysPath, "data") //is default

	if err := ExportKey(doExp); err != nil {
		t.Fatalf("error exporting key: %v", err)
	}

	doExp.Address = addr1
	if err := ExportKey(doExp); err != nil {
		t.Fatalf("error exporting key: %v", err)
	}

	output := testListKeys("host")

	i := 0
	for _, out := range output {
		if addrs[util.TrimString(out)] == true {
			i++
		}
	}

	if i != 2 {
		t.Fatalf("Expected 2 keys, got (%v)", i)
	}
}

func testListKeys(typ string) []string {
	do := def.NowDo()

	if typ == "container" {
		do.Container = true
		do.Host = false
	} else if typ == "host" {
		do.Container = false
		do.Host = true
	}

	if err := ListKeys(do); err != nil {
		tests.IfExit(err)
	}

	return strings.Split(do.Result, ",")
}

//returns an addr for tests
func testsGenAKey() string {
	addr := new(bytes.Buffer)
	config.GlobalConfig.Writer = addr
	doGen := def.NowDo()
	tests.IfExit(GenerateKey(doGen))

	addrBytes := addr.Bytes()
	return util.TrimString(string(addrBytes))
}

func testStartKeys(t *testing.T) {
	serviceName := "keys"
	do := def.NowDo()
	do.Operations.Args = []string{serviceName}
	e := srv.StartService(do)
	if e != nil {
		t.Fatalf("Error starting service: %v", e)
	}

	testExistAndRun(t, serviceName, true, true)
	testNumbersExistAndRun(t, serviceName, true, true)
}

func testKillService(t *testing.T, serviceName string, wipe bool) {
	do := def.NowDo()
	do.Name = serviceName
	do.Operations.Args = []string{serviceName}
	if wipe {
		do.Rm = true
		do.RmD = true
	}
	e := srv.KillService(do)
	if e != nil {
		t.Fatalf("error killing service: %v", e)
	}
	testExistAndRun(t, serviceName, !wipe, false)
	testNumbersExistAndRun(t, serviceName, false, false)
}

func testExistAndRun(t *testing.T, servName string, toExist, toRun bool) {
	tests.IfExit(tests.TestExistAndRun(servName, "service", toExist, toRun))
}

func testNumbersExistAndRun(t *testing.T, servName string, containerExist, containerRun bool) {
	log.WithFields(log.Fields{
		"=>":        servName,
		"existing#": containerExist,
		"running#":  containerRun,
	}).Info("Checking number of containers for")
	log.WithField("=>", servName).Debug("Checking existing containers for")
	exist := util.Exists(def.TypeService, servName)
	log.WithField("=>", servName).Debug("Checking running containers for")
	run := util.Running(def.TypeService, servName)

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
