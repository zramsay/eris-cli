package chains

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	def "github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

var erisDir string = path.Join(os.TempDir(), "eris")
var chainName string = "my_testing_chain_dot_com" // :( [csk]-> :)
var hash string

var DEAD bool // XXX: don't double panic (TODO: Flushing twice blocks)

func fatal(t *testing.T, err error) {
	if !DEAD {
		testsTearDown()
		DEAD = true
		panic(err)
	}
}

func TestMain(m *testing.M) {
	runtime.GOMAXPROCS(1)
	var logLevel log.LogLevel
	var err error

	logLevel = 0
	// logLevel = 1
	// logLevel = 3

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	testsInit()
	logger.Infoln("Test init completed. Starting main test sequence now.\n")

	var exitCode int
	defer func() {
		logger.Infoln("Commensing with Tests Tear Down.")
		if err = testsTearDown(); err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}
		os.Exit(exitCode)

	}()

	exitCode = m.Run()
}

func TestKnownChain(t *testing.T) {
	do := def.NowDo()
	ifExit(ListKnown(do))

	k := strings.Split(do.Result, "\n") // tests output formatting.

	// apparently these have extra space
	if strings.TrimSpace(k[0]) != chainName {
		logger.Debugf("Result =>\t\t%s\n", do.Result)
		ifExit(fmt.Errorf("Unexpected chain definition file. Got %s, expected %s.", k[0], chainName))
	}
}

func TestChainGraduate(t *testing.T) {
	do := def.NowDo()
	do.Name = chainName
	logger.Infof("Graduate chain (from tests) =>\t%s\n", do.Name)
	if err := GraduateChain(do); err != nil {
		fatal(t, err)
	}

	srvDef, err := loaders.LoadServiceDefinition(chainName, false, 1)
	if err != nil {
		fatal(t, err)
	}

	image := "eris/erisdb:" + version.VERSION
	if srvDef.Service.Image != image {
		fatal(t, fmt.Errorf("FAILURE: improper service image on GRADUATE. expected: %s\tgot: %s\n", image, srvDef.Service.Image))
	}

	if srvDef.Service.Command != loaders.ErisChainStart {
		fatal(t, fmt.Errorf("FAILURE: improper service command on GRADUATE. expected: %s\tgot: %s\n", loaders.ErisChainStart, srvDef.Service.Command))
	}

	if !srvDef.Service.AutoData {
		fatal(t, fmt.Errorf("FAILURE: improper service autodata on GRADUATE. expected: %t\tgot: %t\n", true, srvDef.Service.AutoData))
	}

	if len(srvDef.Dependencies.Services) != 1 {
		fatal(t, fmt.Errorf("FAILURE: improper service deps on GRADUATE. expected: [\"keys\"]\tgot: %s\n", srvDef.Dependencies.Services))
	}
}

func TestLoadChainDefinition(t *testing.T) {
	var e error
	logger.Infof("Load chain def (from tests) =>\t%s\n", chainName)
	chn, e := loaders.LoadChainDefinition(chainName, false, 1)
	if e != nil {
		fatal(t, e)
	}

	if chn.Service.Name != chainName {
		fatal(t, fmt.Errorf("FAILURE: improper service name on LOAD. expected: %s\tgot: %s", chainName, chn.Service.Name))
	}

	if !chn.Service.AutoData {
		fatal(t, fmt.Errorf("FAILURE: data_container not properly read on LOAD."))
	}

	if chn.Operations.DataContainerName == "" {
		fatal(t, fmt.Errorf("FAILURE: data_container_name not set."))
	}
}

func TestStartKillChain(t *testing.T) {
	testStartChain(t, chainName)
	testKillChain(t, chainName)
}

func TestExecChain(t *testing.T) {
	/*	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have exec privileges (due to their driver). Skipping test.")
		return
	}*/

	testStartChain(t, chainName)
	defer testKillChain(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Args = strings.Fields("ls -la /home/eris/.eris/blockchains")
	do.Interactive = false
	logger.Infof("Exec-ing chain (from tests) =>\t%s\n", do.Name)
	e := ExecChain(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}
}

// eris chains new --dir _ -g _
// the default chain_id is my_tests, so should be overwritten
func TestChainsNewDirGen(t *testing.T) {
	chainID := "testChainsNewDirGen"
	myDir := path.Join(common.DataContainersPath, chainID)
	if err := os.MkdirAll(myDir, 0700); err != nil {
		fatal(t, err)
	}
	contents := "this is a file in the directory\n"
	if err := ioutil.WriteFile(path.Join(myDir, "file.file"), []byte(contents), 0600); err != nil {
		fatal(t, err)
	}

	do := def.NowDo()
	do.GenesisFile = path.Join(common.BlockchainsPath, "default", "genesis.json")
	do.Name = chainID
	do.Path = myDir
	do.Operations.ContainerNumber = 1
	logger.Infof("Creating chain (from tests) =>\t%s\n", do.Name)
	ifExit(NewChain(do))

	// remove the data container
	defer removeChainContainer(t, chainID, do.Operations.ContainerNumber)

	// verify the contents of file.file - swap config writer with bytes.Buffer
	// TODO: functions for facilitating this
	do.Name = util.DataContainersName(do.Name, do.Operations.ContainerNumber)
	oldWriter := config.GlobalConfig.Writer
	newWriter := new(bytes.Buffer)
	config.GlobalConfig.Writer = newWriter
	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/blockchains/%s/file.file", chainID)}
	b, err := perform.DockerRunVolumesFromContainer(do.Name, false, args)
	if err != nil {
		fatal(t, err)
	}

	config.GlobalConfig.Writer = oldWriter
	result := trimResult(string(b))
	contents = trimResult(contents)
	if result != contents {
		fatal(t, fmt.Errorf("file not faithfully copied. Got: %s \n Expected: %s", result, contents))
	}

	// verify the chain_id got swapped in the genesis.json
	// TODO: functions for facilitating this
	oldWriter = config.GlobalConfig.Writer
	newWriter = new(bytes.Buffer)
	config.GlobalConfig.Writer = newWriter
	args = []string{"cat", fmt.Sprintf("/home/eris/.eris/blockchains/%s/genesis.json", chainID)} //, "|", "jq", ".chain_id"}
	b, err = perform.DockerRunVolumesFromContainer(do.Name, false, args)
	if err != nil {
		fatal(t, err)
	}

	config.GlobalConfig.Writer = oldWriter
	result = string(b)

	s := struct {
		ChainID string `json:"chain_id"`
	}{}
	if err := json.Unmarshal([]byte(result), &s); err != nil {
		fatal(t, err)
	}
	if s.ChainID != chainID {
		fatal(t, fmt.Errorf("ChainID mismatch: got %s, expected %s", s.ChainID, chainID))
	}
}

// eris chains new -c _ -csv _
func TestChainsNewConfigAndCSV(t *testing.T) {
	chainID := "testChainsNewConfigAndCSV"
	do := def.NowDo()
	do.Name = chainID
	do.ConfigFile = path.Join(common.BlockchainsPath, "default", "config.toml")
	do.CSV = path.Join(common.BlockchainsPath, "default", "genesis.csv")
	do.Operations.ContainerNumber = 1
	logger.Infof("Creating chain (from tests) =>\t%s\n", do.Name)
	ifExit(NewChain(do))
	b, err := ioutil.ReadFile(do.ConfigFile)
	if err != nil {
		fatal(t, err)
	}

	fmt.Println("CONFIG CONFIG CONFIG:", string(b))

	// remove the data container
	defer removeChainContainer(t, chainID, do.Operations.ContainerNumber)

	// verify the contents of config.toml
	do.Name = util.DataContainersName(do.Name, do.Operations.ContainerNumber)
	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/blockchains/%s/config.toml", chainID)}
	result := trimResult(string(runContainer(t, do.Name, args)))
	contents := trimResult(ini.DefChainConfig())
	if result != contents {
		fatal(t, fmt.Errorf("config not properly copied. Got: %s \n Expected: %s", result, contents))
	}

	// verify the contents of genesis.json (should have the validator from the csv)
	args = []string{"cat", fmt.Sprintf("/home/eris/.eris/blockchains/%s/genesis.json", chainID)}
	result = string(runContainer(t, do.Name, args))
	var found bool
	for _, s := range strings.Split(result, "\n") {
		if strings.Contains(s, ini.DefaultPubKeys[0]) {
			found = true
			break
		}
	}
	if !found {
		fatal(t, fmt.Errorf("Did not find pubkey %s in genesis.json: %s", ini.DefaultPubKeys[0], result))
	}
}

// eris chains new --options
func TestChainsNewConfigOpts(t *testing.T) {
	// XXX: need to use a different chainID or remove the local tmp/eris/data/chainID dir with each test!
	chainID := "testChainsNewConfigOpts"
	do := def.NowDo()

	do.Name = chainID
	do.ConfigOpts = []string{"moniker=satoshi", "p2p=1.1.1.1:42", "fast-sync=true"}
	do.Operations.ContainerNumber = 1
	logger.Infof("Creating chain (from tests) =>\t%s\n", do.Name)
	ifExit(NewChain(do))

	// remove the data container
	defer removeChainContainer(t, chainID, do.Operations.ContainerNumber)

	// verify the contents of config.toml
	do.Name = util.DataContainersName(do.Name, do.Operations.ContainerNumber)
	args := []string{"cat", fmt.Sprintf("/home/eris/.eris/blockchains/%s/config.toml", chainID)}
	result := string(runContainer(t, do.Name, args))

	spl := strings.Split(result, "\n")
	var found bool
	for _, s := range spl {
		if ensureTomlValue(t, s, "moniker", "satoshi") {
			found = true
		}
		if ensureTomlValue(t, s, "node_laddr", "1.1.1.1:42") {
			found = true
		}
		if ensureTomlValue(t, s, "fast_sync", "true") {
			found = true
		}
	}
	if !found {
		fatal(t, fmt.Errorf("failed to find fields: %s", result))
	}
}

func TestLogsChain(t *testing.T) {
	testStartChain(t, chainName)
	defer testKillChain(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Follow = false
	do.Tail = "all"
	logger.Infof("Get chain logs (from tests) =>\t%s:%s\n", do.Name, do.Tail)
	e := LogsChain(do)
	if e != nil {
		fatal(t, e)
	}
}

func TestUpdateChain(t *testing.T) {
	testStartChain(t, chainName)
	defer testKillChain(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.SkipPull = true
	logger.Infof("Updating chain (from tests) =>\t%s\n", do.Name)
	if e := UpdateChain(do); e != nil {
		fatal(t, e)
	}

	testExistAndRun(t, chainName, true, true)
}

func TestInspectChain(t *testing.T) {
	testStartChain(t, chainName)
	defer testKillChain(t, chainName)

	do := def.NowDo()
	do.Name = chainName
	do.Args = []string{"name"}
	do.Operations.ContainerNumber = 1
	logger.Debugf("Inspect chain (via tests) =>\t%s:%v\n", chainName, do.Args)
	if e := InspectChain(do); e != nil {
		fatal(t, fmt.Errorf("Error inspecting chain =>\t%v\n", e))
	}
	// log.SetLoggers(0, os.Stdout, os.Stderr)
}

func TestRenameChain(t *testing.T) {
	oldName := chainName
	newName := "niahctset"
	testStartChain(t, oldName)
	defer testKillChain(t, oldName)

	do := def.NowDo()
	do.Name = oldName
	do.NewName = newName
	logger.Infof("Renaming chain (from tests) =>\t%s:%s\n", do.Name, do.NewName)
	if e := RenameChain(do); e != nil {
		fatal(t, e)
	}

	testExistAndRun(t, newName, true, true)

	do = def.NowDo()
	do.Name = newName
	do.NewName = chainName
	logger.Infof("Renaming chain (from tests) =>\t%s:%s\n", do.Name, do.NewName)
	if e := RenameChain(do); e != nil {
		fatal(t, e)
	}

	testExistAndRun(t, chainName, true, true)
}

// TODO: finish this....
// func TestServiceWithChainDependencies(t *testing.T) {
// 	do := definitions.NowDo()
// 	do.Name = "keys"
// 	do.Args = []string{"eris/keys"}
// 	err := services.NewService(do)
// 	if err != nil {
// 		logger.Errorln(err)
// 		t.FailNow()
// 	}

// 	services.TestCatService(t)

// }

func TestRmChain(t *testing.T) {
	testStartChain(t, chainName)

	do := def.NowDo()
	do.Args, do.Rm, do.RmD = []string{"keys"}, true, true
	logger.Infof("Removing keys (from tests) =>\n%s\n", do.Name)
	if e := services.KillService(do); e != nil {
		fatal(t, e)
	}

	do = def.NowDo()
	do.Name, do.Rm, do.RmD = chainName, false, false
	logger.Infof("Stopping chain (from tests) =>\t%s\n", do.Name)
	if e := KillChain(do); e != nil {
		fatal(t, e)
	}
	testExistAndRun(t, chainName, true, false)

	do = def.NowDo()
	do.Name = chainName
	do.RmD = true
	logger.Infof("Removing chain (from tests) =>\n%s\n", do.Name)
	if e := RmChain(do); e != nil {
		fatal(t, e)
	}

	testExistAndRun(t, chainName, false, false)
}

//------------------------------------------------------------------
// testing utils

func testStartChain(t *testing.T, chain string) {
	do := def.NowDo()
	do.Name = chain
	do.Operations.ContainerNumber = 1
	logger.Infof("Starting chain (from tests) =>\t%s\n", do.Name)
	if e := StartChain(do); e != nil {
		logger.Errorln(e)
		fatal(t, nil)
	}
	testExistAndRun(t, chain, true, true)
}

func testKillChain(t *testing.T, chain string) {
	// log.SetLoggers(2, os.Stdout, os.Stderr)
	testExistAndRun(t, chain, true, true)

	do := def.NowDo()
	do.Args, do.Rm, do.RmD = []string{"keys"}, true, true
	logger.Infof("Killing keys (from tests) =>\n%s\n", do.Name)
	if e := services.KillService(do); e != nil {
		fatal(t, e)
	}

	do = def.NowDo()
	do.Name, do.Rm, do.RmD = chain, true, true
	logger.Infof("Stopping chain (from tests) =>\t%s\n", do.Name)
	if e := KillChain(do); e != nil {
		fatal(t, e)
	}
	testExistAndRun(t, chain, false, false)
}

func testExistAndRun(t *testing.T, chainName string, toExist, toRun bool) {
	var exist, run bool
	logger.Infof("\nTesting whether (%s) is running? (%t) and existing? (%t)\n", chainName, toRun, toExist)
	chainName = util.ChainContainersName(chainName, 1) // not worried about containerNumbers, deal with multiple containers in services tests

	do := def.NowDo()
	do.Quiet = true
	do.Args = []string{"testing"}
	if err := ListExisting(do); err != nil {
		fatal(t, err)
	}
	res := strings.Split(do.Result, "\n")
	for _, r := range res {
		logger.Debugf("Existing =>\t\t\t%s\n", r)
		if r == util.ContainersShortName(chainName) {
			exist = true
		}
	}

	do = def.NowDo()
	do.Quiet = true
	do.Args = []string{"testing"}
	if err := ListRunning(do); err != nil {
		fatal(t, err)
	}
	logger.Debugln("RUNNING RESULT:", do.Result)
	res = strings.Split(do.Result, "\n")
	for _, r := range res {
		logger.Debugf("Running =>\t\t\t%s\n", r)
		if r == util.ContainersShortName(chainName) {
			run = true
		}
	}

	if toExist != exist {
		if toExist {
			logger.Infof("Could not find an existing =>\t%s\n", chainName)
		} else {
			logger.Infof("Found an existing instance of %s when I shouldn't have\n", chainName)
		}
		fatal(t, nil)
	}

	if toRun != run {
		if toRun {
			logger.Infof("Could not find a running =>\t%s\n", chainName)
		} else {
			logger.Infof("Found a running instance of %s when I shouldn't have\n", chainName)
		}
		fatal(t, nil)
	}

	logger.Debugln("")
}

func testsInit() error {
	var err error
	// TODO: make a reader/pipe so we can see what is written from tests.
	config.GlobalConfig, err = config.SetGlobalObject(os.Stdout, os.Stderr)
	if err != nil {
		ifExit(fmt.Errorf("TRAGIC. Could not set global config.\n"))
	}

	// common is initialized on import so
	// we have to manually override these
	// variables to ensure that the tests
	// run correctly.
	config.ChangeErisDir(erisDir)

	// init dockerClient
	util.DockerConnect(false, "eris-test-nyc2-1.8.1")

	// this dumps the ipfs service def into the temp dir which
	// has been set as the erisRoot
	do := def.NowDo()
	do.Pull = true
	do.Services = true
	do.Actions = true
	if err := ini.Initialize(do); err != nil {
		ifExit(fmt.Errorf("TRAGIC. Could not initialize the eris dir.\n"))
	}

	// lay a chain service def
	testNewChain(chainName)

	return nil
}

func testNewChain(chain string) {
	do := def.NowDo()
	do.GenesisFile = path.Join(common.BlockchainsPath, "default", "genesis.json")
	do.Name = chain
	do.Operations.ContainerNumber = 1
	logger.Infof("Creating chain (from tests) =>\t%s\n", chain)
	ifExit(NewChain(do))

	// remove the data container
	do.Args = []string{chain}
	ifExit(data.RmData(do))
}

func removeChainContainer(t *testing.T, chainID string, cNum int) {
	do := def.NowDo()
	do.Name = chainID
	do.Rm, do.Force, do.RmD = true, true, true
	do.Operations.ContainerNumber = cNum
	if err := KillChain(do); err != nil {
		fatal(t, err)
	}
}

func testsTearDown() error {
	DEAD = true
	killService("keys")
	testKillChain(nil, chainName)
	log.Flush()
	return os.RemoveAll(erisDir)
}

func killService(name string) {
	do := def.NowDo()
	do.Name = name
	do.Args = []string{name}
	do.Rm, do.RmD, do.Force = true, true, true
	if e := services.KillService(do); e != nil {
		logger.Errorln(e)
		fatal(nil, e)
	}
}

func runContainer(t *testing.T, name string, args []string) []byte {
	oldWriter := config.GlobalConfig.Writer
	newWriter := new(bytes.Buffer)
	config.GlobalConfig.Writer = newWriter
	b, err := perform.DockerRunVolumesFromContainer(name, false, args)
	if err != nil {
		fatal(t, err)
	}
	logger.Debugf("Container ran =>\t\t%s:%v\n", name, args)
	config.GlobalConfig.Writer = oldWriter
	return b
}

func ensureTomlValue(t *testing.T, s, field, value string) bool {
	if strings.Contains(s, field) {
		if !strings.Contains(s, value) {
			fatal(t, fmt.Errorf("Expected %s to be %s. Got: %s", field, value, s))
		}
		return true
	}
	return false
}

func ifExit(err error) {
	if err != nil {
		logger.Errorln(err)
		log.Flush()
		testsTearDown()
		os.Exit(1)
	}
}

func trimResult(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\n")
	spl := strings.Split(s, "\n")
	for i, t := range spl {
		t = strings.TrimSpace(t)
		spl[i] = t
	}
	return strings.Trim(strings.Join(spl, "\n"), "\n")
}
