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

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	logger "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

func TestMain(m *testing.M) {
	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit("keys"))

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
		t.Fatalf("Expected one key, got (%v)\n", len(output))
	}

	if address != output[0] {
		t.Fatalf("Expected (%s), got (%s)\n", address, output[0])
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
		t.Fatalf("error getting pubkey: %v\n", err)
	}

	pubkey := util.TrimString(pub.String())

	key := new(bytes.Buffer)
	config.GlobalConfig.Writer = key
	doKey := def.NowDo()
	doKey.Address = doPub.Address
	if err := ConvertKey(doKey); err != nil {
		t.Fatalf("error converting key: %v\n", err)
	}

	converted := regexp.MustCompile(`"pub_key":\[1,"([^"]+)"\]`).FindStringSubmatch(key.String())[1]

	if converted != pubkey {
		t.Fatalf("Expected (%s), got (%s)\n", pubkey, converted)
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
		t.Fatalf("error exec-ing: %v\n", err)
	}

	keyInCont := util.TrimString(catOut.String())

	doExp := def.NowDo()
	doExp.Address = address
	doExp.Destination = filepath.Join(KeysPath, "data") //is default

	//export
	if err := ExportKey(doExp); err != nil {
		t.Fatalf("error exporting: %v\n", err)
	}

	//cat host contents
	key, err := ioutil.ReadFile(filepath.Join(doExp.Destination, address, address))
	if err != nil {
		t.Fatalf("error reading file: %v\n", err)
	}

	keyOnHost := util.TrimString(string(key))
	if keyInCont != keyOnHost {
		t.Fatalf("Expected (%s), got (%s)\n", keyInCont, keyOnHost)
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
		t.Fatalf("error exporting key: %v\n", err)
	}

	key, err := ioutil.ReadFile(filepath.Join(doExp.Destination, address, address))
	if err != nil {
		t.Fatalf("error reading file: %v\n", err)
	}
	//key b4 import
	keyOnHost := util.TrimString(string(key))

	//rm key that was generated before import
	keyPath := path.Join(ErisContainerRoot, "keys", "data", address)

	if _, err := srv.ExecHandler("keys", []string{"rm", "-rf", keyPath}); err != nil {
		t.Fatalf("error exec-ing: %v\n", err)
	}

	doImp := def.NowDo()
	doImp.Address = address
	//doImp.Destination // set in function
	doImp.Source = filepath.Join(KeysPath, "data")

	if err := ImportKey(doImp); err != nil {
		t.Fatalf("error importing key: %v\n", err)
	}

	keyPathCat := path.Join(ErisContainerRoot, "keys", "data", address, address)

	//cat container contents of new key
	catOut, err := srv.ExecHandler("keys", []string{"cat", keyPathCat})
	if err != nil {
		t.Fatalf("error exec-ing: %v\n", err)
	}

	keyInCont := util.TrimString(catOut.String())

	if keyOnHost != keyInCont {
		t.Fatalf("Expected (%s), got (%s)\n", keyOnHost, keyInCont)
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
	addrs[testsGenAKey()] = true

	output := testListKeys("container")

	i := 0
	for _, out := range output {
		if addrs[util.TrimString(out)] == true {
			i++
		}
	}

	if i != 3 {
		t.Fatalf("Expected 3 keys, got (%v)\n", i)
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
		t.Fatalf("error exporting key: %v\n", err)
	}

	doExp.Address = addr1
	if err := ExportKey(doExp); err != nil {
		t.Fatalf("error exporting key: %v\n", err)
	}

	output := testListKeys("host")

	i := 0
	for _, out := range output {
		if addrs[util.TrimString(out)] == true {
			i++
		}
	}

	if i != 2 {
		t.Fatalf("Expected 2 keys, got (%v)\n", i)
	}
}

//
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

	res := strings.Split(do.Result, ",")

	return res
}

//returns an addr for tests
func testsGenAKey() string {
	addr := new(bytes.Buffer)
	config.GlobalConfig.Writer = addr
	doGen := def.NowDo()
	tests.IfExit(GenerateKey(doGen))

	addrBytes := addr.Bytes()
	address := util.TrimString(string(addrBytes))
	return address
}

func testStartKeys(t *testing.T) {
	serviceName := "keys"
	do := def.NowDo()
	do.Operations.Args = []string{serviceName}
	log.WithField("=>", serviceName).Debug("Starting service (via tests)")
	e := srv.StartService(do)
	if e != nil {
		t.Fatalf("Error starting service: %v", e)
	}

	testExistAndRun(t, serviceName, true, true)
	testNumbersExistAndRun(t, serviceName, 1, 1)
}

func testKillService(t *testing.T, serviceName string, wipe bool) {
	log.WithField("=>", serviceName).Debug("Stopping service (from tests)")

	do := def.NowDo()
	do.Name = serviceName
	do.Operations.Args = []string{serviceName}
	if wipe {
		do.Rm = true
		do.RmD = true
	}
	e := srv.KillService(do)
	if e != nil {
		t.Fatalf("error killing services: %v\n", e)
	}
	testExistAndRun(t, serviceName, !wipe, false)
	testNumbersExistAndRun(t, serviceName, 0, 0)
}

func testExistAndRun(t *testing.T, servName string, toExist, toRun bool) {
	tests.IfExit(tests.TestExistAndRun(servName, "service", toExist, toRun))
}

func testNumbersExistAndRun(t *testing.T, servName string, containerExist, containerRun int) {
	log.WithFields(log.Fields{
		"=>":        servName,
		"existing#": containerExist,
		"running#":  containerRun,
	}).Info("Checking number of containers for")
	log.WithField("=>", servName).Debug("Checking existing containers for")
	exist := util.HowManyContainersExisting(servName, "service")
	log.WithField("=>", servName).Debug("Checking running containers for")
	run := util.HowManyContainersRunning(servName, "service")

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
