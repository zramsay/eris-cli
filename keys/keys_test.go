package keys

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/services"
	"github.com/monax/monax/testutil"
	"github.com/monax/monax/util"
)

func TestMain(m *testing.M) {
	// log.SetLevel(log.ErrorLevel)
	log.SetLevel(log.WarnLevel)
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

func TestStartKeys(t *testing.T) {
	_, err := InitKeyClient()
	defer testKillService(t, "keys", true)
	if err != nil {
		t.Fatalf("Could not initialize key client, got err %v", err)
	}
	//Run multiple attempts at initialization
	_, err = InitKeyClient()
	if err != nil {
		t.Fatalf("Could not initialize second key client, got err %v", err)
	}
	_, err = InitKeyClient()
	if err != nil {
		t.Fatalf("Could not initialize third key client, got err %v", err)
	}
	_, err = InitKeyClient()
	if err != nil {
		t.Fatalf("Could not initialize fourth key client, got err %v", err)
	}

	testExistAndRun(t, "keys", true, true)
	testNumbersExistAndRun(t, "keys", true, true)
}

func TestGenerateKey(t *testing.T) {
	keyClient, err := InitKeyClient()
	defer testKillService(t, "keys", true)
	if err != nil {
		t.Fatalf("Could not initialize key client, got err %v", err)
	}
	//todo: clean this test up to be made from a test struct/loop
	//Try without saving the key
	address, err := testsGenAKey(keyClient, false, "", "")
	if err != nil {
		t.Fatalf("Unexpected error in key generation: %v", err)
	}
	if address == "" {
		t.Fatalf("Expected generated key, but got empty output")
	}
	//See if saved on the container
	output := testListKeys(keyClient, "container")

	if len(output) != 1 {
		t.Fatalf("Expected one key, got (%v)", len(output))
	}

	if address != output[0] {
		t.Fatalf("Expected (%s), got (%s)", address, output[0])
	}

	//Try saving the key
	address, err = testsGenAKey(keyClient, true, "", "")
	if err != nil {
		t.Fatalf("Unexpected error in key generation: %v", err)
	}
	if address == "" {
		t.Fatalf("Expected generated key, but got empty output")
	}

	//See if saved on the host
	output = testListKeys(keyClient, "host")

	if len(output) != 1 {
		t.Fatalf("Expected one key, got (%v)", len(output))
	}

	if address != output[0] {
		t.Fatalf("Expected (%s), got (%s)", address, output[0])
	}

	// Todo: implement password and change this
	_, err = testsGenAKey(keyClient, true, "", "marmot")
	if err == nil {
		t.Fatal("Expected error for password usage in key generation. Got none.")
	}
}

func TestExportKeySingle(t *testing.T) {
	keyClient, err := InitKeyClient()
	defer testKillService(t, "keys", true)
	if err != nil {
		t.Fatalf("Could not initialize key client, got err %v", err)
	}

	address, err := testsGenAKey(keyClient, false, "", "")
	if err != nil {
		t.Fatalf("Unexpected error in key generation: %v", err)
	}

	keyPath := path.Join(config.MonaxContainerRoot, "keys", "data", address, address)

	//cat container contents of new key
	catOut, err := services.ExecHandler("keys", []string{"cat", keyPath})
	if err != nil {
		t.Fatalf("error exec-ing: %v", err)
	}

	keyInCont := strings.TrimSpace(catOut.String())

	//export
	if err := keyClient.ExportKey(address, false); err != nil {
		t.Fatalf("error exporting: %v", err)
	}

	//cat host contents
	key, err := ioutil.ReadFile(filepath.Join(filepath.Join(config.KeysPath, "data"), address, address))
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}

	keyOnHost := strings.TrimSpace(string(key))
	if keyInCont != keyOnHost {
		t.Fatalf("Expected (%s), got (%s)", keyInCont, keyOnHost)
	}
}

func TestImportKeySingle(t *testing.T) {
	keyClient, err := InitKeyClient()
	defer testKillService(t, "keys", true)
	if err != nil {
		t.Fatalf("Could not initialize key client, got err %v", err)
	}
	// automatically exported when we save
	address, err := testsGenAKey(keyClient, true, "", "")
	if err != nil {
		t.Fatalf("Unexpected error in key generation: %v", err)
	}

	key, err := ioutil.ReadFile(filepath.Join(config.KeysPath, "data", address, address))
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}
	//key before import
	keyOnHost := strings.TrimSpace(string(key))

	//rm key that was generated before import
	keyPath := path.Join(config.MonaxContainerRoot, "keys", "data", address)

	if _, err := services.ExecHandler("keys", []string{"rm", "-rf", keyPath}); err != nil {
		t.Fatalf("error exec-ing: %v", err)
	}

	if err := keyClient.ImportKey(address, false); err != nil {
		t.Fatalf("error importing key: %v", err)
	}

	keyPathCat := path.Join(config.MonaxContainerRoot, "keys", "data", address, address)

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

func TestImportKeyAll(t *testing.T) {
	keyClient, err := InitKeyClient()
	if err != nil {
		t.Fatalf("Could not initialize key client, got err %v", err)
	}

	// gen some keys, and export them to the host
	addrs := make(map[string]bool)
	addr1, err := testsGenAKey(keyClient, true, "", "")
	if err != nil {
		t.Fatalf("Unexpected error in key generation: %v", err)
	}
	addr2, err := testsGenAKey(keyClient, true, "", "")
	if err != nil {
		t.Fatalf("Unexpected error in key generation: %v", err)
	}
	addrs[addr1] = true
	addrs[addr2] = true

	// kill container
	testKillService(t, "keys", true)

	// start keys
	keyClient, err = InitKeyClient()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	defer testKillService(t, "keys", true)

	if err := keyClient.ImportKey("", true); err != nil {
		t.Fatalf("error exporting: %v", err)
	}

	// check that they are in the container
	output := testListKeys(keyClient, "container")

	i := 0
	for _, out := range output {
		if addrs[strings.TrimSpace(out)] {
			i++
		}
	}

	if i != 2 {
		t.Fatalf("Expected 2 keys, got (%v)", i)
	}
}

func TestExportKeyAll(t *testing.T) {
	keyClient, err := InitKeyClient()
	defer testKillService(t, "keys", true)
	if err != nil {
		t.Fatalf("Could not initialize key client, got err %v", err)
	}
	// gen some keys
	addrs := make(map[string]bool)
	addr1, err := testsGenAKey(keyClient, false, "", "")
	if err != nil {
		t.Fatalf("Unexpected error in key generation: %v", err)
	}
	addr2, err := testsGenAKey(keyClient, false, "", "")
	if err != nil {
		t.Fatalf("Unexpected error in key generation: %v", err)
	}
	addrs[addr1] = true
	addrs[addr2] = true

	//export
	if err := keyClient.ExportKey("", true); err != nil {
		t.Fatalf("error exporting: %v", err)
	}
	// check that they on host
	output := testListKeys(keyClient, "host")

	i := 0
	for _, out := range output {
		if addrs[strings.TrimSpace(out)] {
			i++
		}
	}

	if i != 2 {
		t.Fatalf("Expected 2 keys, got (%v)", i)
	}
}

func TestListKeyContainer(t *testing.T) {
	keyClient, err := InitKeyClient()
	defer testKillService(t, "keys", true)
	if err != nil {
		t.Fatalf("Could not initialize key client, got err %v", err)
	}

	// gen some keys
	addrs := make(map[string]bool)
	addr1, err := testsGenAKey(keyClient, false, "", "")
	if err != nil {
		t.Fatalf("Unexpected error in key generation: %v", err)
	}
	addr2, err := testsGenAKey(keyClient, false, "", "")
	if err != nil {
		t.Fatalf("Unexpected error in key generation: %v", err)
	}
	addrs[addr1] = true
	addrs[addr2] = true

	output := testListKeys(keyClient, "container")

	i := 0
	for _, out := range output {
		if addrs[strings.TrimSpace(out)] {
			i++
		}
	}

	if i != 2 {
		t.Fatalf("Expected 2 keys, got (%v)", i)
	}
}

func TestListKeyHost(t *testing.T) {
	keyClient, err := InitKeyClient()
	defer testKillService(t, "keys", true)
	if err != nil {
		t.Fatalf("Could not initialize key client, got err %v", err)
	}

	// gen some keys
	addrs := make(map[string]bool)
	addr1, err := testsGenAKey(keyClient, true, "", "")
	if err != nil {
		t.Fatalf("Unexpected error in key generation: %v", err)
	}
	addr2, err := testsGenAKey(keyClient, true, "", "")
	if err != nil {
		t.Fatalf("Unexpected error in key generation: %v", err)
	}
	addrs[addr1] = true
	addrs[addr2] = true

	output := testListKeys(keyClient, "host")

	i := 0
	for _, out := range output {
		if addrs[strings.TrimSpace(out)] {
			i++
		}
	}

	if i != 2 {
		t.Fatalf("Expected 2 keys, got (%v)", i)
	}
}

func TestKeyConvert(t *testing.T) {
	keyClient, err := InitKeyClient()
	defer testKillService(t, "keys", true)
	if err != nil {
		t.Fatalf("Could not initialize key client, got err %v", err)
	}
	tendermintKey, err := testsGenAKey(keyClient, true, "ed25519,ripemd160", "")
	if err != nil {
		t.Fatalf("Unexpected error in key generation: %v", err)
	}
	_, err = keyClient.Convert(tendermintKey, "")
	if err != nil {
		t.Fatalf("Unexpected error during conversion of address %v to priv validator.json", tendermintKey)
	}
}

func testListKeys(keys *KeyClient, typ string) []string {
	var container, host bool

	if typ == "container" {
		container = true
		host = false
	} else if typ == "host" {
		container = false
		host = true
	}

	result, err := keys.ListKeys(host, container, false)
	if err != nil {
		testutil.IfExit(err)
	}

	return result
}

func testsGenAKey(keys *KeyClient, save bool, keyType, password string) (string, error) {
	return keys.GenerateKey(save, true, keyType, password)
}

func testExistAndRun(t *testing.T, servName string, toExist, toRun bool) {
	testutil.IfExit(testutil.ExistAndRun(servName, "service", toExist, toRun))
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
