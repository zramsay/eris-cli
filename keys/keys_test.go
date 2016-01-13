package keys

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	srv "github.com/eris-ltd/eris-cli/services"
	tests "github.com/eris-ltd/eris-cli/testutils"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	logger "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

var DEAD bool

func fatal(t *testing.T, err error) {
	if !DEAD {
		tests.TestsTearDown()
		DEAD = true
		panic(err)
	}
}

func TestMain(m *testing.M) {
	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(testsInit())

	exitCode := m.Run()

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		tests.IfExit(tests.TestsTearDown())
	}

	os.Exit(exitCode)
}

func TestGenerateKey(t *testing.T) {
	testStartKeys(t)
	defer testKillService(t, "keys", true)

	address := testsGenAKey()

	lsOut := new(bytes.Buffer)
	config.GlobalConfig.Writer = lsOut
	do := def.NowDo()
	do.Name = "keys"
	do.Operations.Interactive = false
	do.Operations.ContainerNumber = 1
	path := path.Join(ErisContainerRoot, "keys", "data")

	do.Operations.Args = []string{"ls", path}
	if err := srv.ExecService(do); err != nil {
		fatal(t, err)
	}

	lsOutBytes := lsOut.Bytes()

	output := util.TrimString(string(lsOutBytes))

	if address != output {
		fatal(t, fmt.Errorf("Expected (%s), got (%s)\n", address, output))
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
		fatal(t, err)
	}

	pubBytes := pub.Bytes()
	pubkey := util.TrimString(string(pubBytes))

	key := new(bytes.Buffer)
	config.GlobalConfig.Writer = key
	doKey := def.NowDo()
	doKey.Address = doPub.Address
	if err := ConvertKey(doKey); err != nil {
		fatal(t, err)
	}

	converted := regexp.MustCompile(`"pub_key":\[1,"([^"]+)"\]`).FindStringSubmatch(key.String())[1]

	if converted != pubkey {
		fatal(t, fmt.Errorf("Expected (%s), got (%s)\n", pubkey, converted))
	}
}

func TestExportKeySingle(t *testing.T) {
	testStartKeys(t)
	defer testKillService(t, "keys", true) //or some clean function

	address := testsGenAKey()

	catOut := new(bytes.Buffer)
	config.GlobalConfig.Writer = catOut
	do := def.NowDo()
	do.Name = "keys"
	do.Operations.Interactive = false
	do.Operations.ContainerNumber = 1
	keyPath := path.Join(ErisContainerRoot, "keys", "data", address, address)

	//cat container contents of new key
	do.Operations.Args = []string{"cat", keyPath}
	if err := srv.ExecService(do); err != nil {
		fatal(t, err)
	}

	catOutBytes := catOut.Bytes()
	keyInCont := util.TrimString(string(catOutBytes))

	doExp := def.NowDo()
	doExp.Address = address
	doExp.Destination = filepath.Join(KeysPath, "data") //is default

	//export
	if err := ExportKey(doExp); err != nil {
		fatal(t, err)
	}

	//cat host contents
	key, err := ioutil.ReadFile(filepath.Join(doExp.Destination, address, address))
	if err != nil {
		fatal(t, err)
	}

	keyOnHost := util.TrimString(string(key))
	if keyInCont != keyOnHost {
		fatal(t, fmt.Errorf("Expected (%s), got (%s)\n", keyInCont, keyOnHost))
	}
}

func TestImportKeySingle(t *testing.T) {
	testStartKeys(t)
	defer testKillService(t, "keys", true) //or some clean function

	address := testsGenAKey()

	//export it
	doExp := def.NowDo()
	doExp.Address = address
	doExp.Destination = filepath.Join(KeysPath, "data") //is default

	if err := ExportKey(doExp); err != nil {
		fatal(t, err)
	}

	key, err := ioutil.ReadFile(filepath.Join(doExp.Destination, address, address))
	if err != nil {
		fatal(t, err)
	}
	//key b4 import
	keyOnHost := util.TrimString(string(key))

	//rm key that was generated before import
	doRm := def.NowDo()
	doRm.Name = "keys"
	doRm.Operations.Interactive = false
	doRm.Operations.ContainerNumber = 1
	keyPath := path.Join(ErisContainerRoot, "keys", "data", address)

	doRm.Operations.Args = []string{"rm", "-rf", keyPath}
	if err := srv.ExecService(doRm); err != nil {
		fatal(t, err)
	}

	doImp := def.NowDo()
	doImp.Address = address
	//doImp.Destination // set in function

	if err := ImportKey(doImp); err != nil {
		fatal(t, err)
	}

	catOut := new(bytes.Buffer)
	config.GlobalConfig.Writer = catOut
	doCat := def.NowDo()
	doCat.Name = "keys"
	doCat.Operations.Interactive = false
	doCat.Operations.ContainerNumber = 1
	keyPathCat := path.Join(ErisContainerRoot, "keys", "data", address, address)

	//cat container contents of new key
	doCat.Operations.Args = []string{"cat", keyPathCat}
	if err := srv.ExecService(doCat); err != nil {
		fatal(t, err)
	}

	catOutBytes := catOut.Bytes()
	keyInCont := util.TrimString(string(catOutBytes))

	if keyOnHost != keyInCont {
		fatal(t, fmt.Errorf("Expected (%s), got (%s)\n", keyOnHost, keyInCont))
	}
}

func TestConvertKey(t *testing.T) {
	// tested in TestGetPubKey
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

func testsInit() error {
	if err := tests.TestsInit("keys"); err != nil {
		return err
	}
	return nil
}

func testStartKeys(t *testing.T) {
	serviceName := "keys"
	do := def.NowDo()
	do.Operations.Args = []string{serviceName}
	do.Operations.ContainerNumber = 1
	log.WithField("=>", fmt.Sprintf("%s:%d", serviceName, do.Operations.ContainerNumber)).Debug("Starting service (via tests)")
	e := srv.StartService(do)
	if e != nil {
		log.Infof("Error starting service: %v", e)
		t.Fail()
	}

	testExistAndRun(t, serviceName, 1, true, true)
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
		log.Error(e)
		fatal(t, e)
	}
	testExistAndRun(t, serviceName, 1, !wipe, false)
	testNumbersExistAndRun(t, serviceName, 0, 0)
}

func testExistAndRun(t *testing.T, servName string, containerNumber int, toExist, toRun bool) {
	if err := tests.TestExistAndRun(servName, "service", containerNumber, toExist, toRun); err != nil {
		fatal(t, nil)
	}
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
		fatal(t, nil)
	}

	if run != containerRun {
		log.WithFields(log.Fields{
			"service":  servName,
			"expected": containerExist,
			"got":      run,
		}).Error("Wrong number of running containers")
		fatal(t, nil)
	}

	log.Info("All good")
}
