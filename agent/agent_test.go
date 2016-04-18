package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/files"
	"github.com/eris-ltd/eris-cli/keys"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/tests"

	log "github.com/Sirupsen/logrus"
	"github.com/eris-ltd/common/go/common"
	logger "github.com/eris-ltd/common/go/log"
)

var (
	// incoming post request args
	// auth added in SetBundlePath
	bundleInfo = map[string]string{
		"groupId":  "com.erisindustries",
		"bundleId": "marmoty-contracts",
		"version":  "2.1.2", //don't use ints since path joining
		"dirName":  "idi",
	}
	hash      = "QmdbzmNH1iDg2H86Uk3USuJ2vaugvwU7HubvCxMC2fUykm"
	chainName = "agent-test"
	address   = "C7A4F01D58FC60429A3330CED519BBE14563FFA4"

	// build the path
	//TODO make a test for this func!

	installPath   = SetTarballPath(bundleInfo)                        // where the tarball is dropped
	contractsPath = filepath.Join(installPath, bundleInfo["dirName"]) // the dir in which contracts are to be deployed

	// expected from deploying idi.sol = > to test XXX
	chainCode = "60606040526000357C01000000000000000000000000000000000000000000000000000000009004806360FE47B11460415780636D4CE63C14605757603F565B005B605560048080359060200190919050506078565B005B606260048050506086565B6040518082815260200191505060405180910390F35B806000600050819055505B50565B600060006000505490506094565B9056"
)

//TODO use :defer tests.RemoveAllContainers()
func TestMain(m *testing.M) {
	log.SetFormatter(logger.ErisFormatter{})

	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit("agent"))

	exitCode := m.Run()
	tests.IfExit(tests.TestsTearDown())
	os.Exit(exitCode)
}

// ensure url format satifies required schema
func TestParsePayload(t *testing.T) {

	//TODO update
	toTest := map[string]string{
		"groupId":   bundleInfo["groupId"],  // needed to buildpath
		"bundleId":  bundleInfo["bundleId"], // ibid
		"version":   bundleInfo["version"],  // ibid
		"dirName":   bundleInfo["dirName"],  // ibid
		"hash":      hash,                   // hash of tarball
		"chainName": chainName,              // chain to deploy
		"address":   address,                // account to deploy from
	}

	rawUrl := fmt.Sprintf("https://localhost:17552/install?groupId=%s&bundleId=%s&version=%s&dirName=%s&hash=%s&chainName=%s&address=%s", toTest["groupId"], toTest["bundleId"], toTest["version"], toTest["dirName"], toTest["hash"], toTest["chainName"], toTest["address"])

	parsed, err := ParseURL(rawUrl)
	if err != nil {
		t.Fatalf("error parsing url (%s): err:\n%v", rawUrl, err)
	}

	if !reflect.DeepEqual(toTest, parsed) {
		t.Fatalf("toTest (%v) does not equal parsed (%v)", toTest, parsed)
	}
}

// the test that matters!
func TestDeployContract(t *testing.T) {
	start(t, "keys", false)
	defer kill(t, "keys", true)

	testSetupChain(t, chainName) // has test for IsChainRunning
	defer testKillChain(t, chainName)

	// testGetTarballFromIPFS(t)
	// defer kill(t, "ipfs", true)

	doKey := definitions.NowDo()
	doKey.Container = true
	doKey.Host = false
	//there should only be one key
	if err := keys.ListKeys(doKey); err != nil {
		t.Fatalf("err listing keys: %v\n", err)
	}
	address := strings.Split(doKey.Result, ",")[0]

	// hack but fmt ipfs
	// assume wd = agent/
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error getting wd: %v\n", err)
	}

	contractsPath = filepath.Join(wd, bundleInfo["dirName"])

	if err := DeployContractBundle(contractsPath, chainName, address); err != nil {
		t.Fatalf("error deploying contract bundle: %v\n", err)
	}

	// TODO check that chainCode matches expected code
}

// setup fake server & his install endpoint ... ?
func TestInstallAgent(t *testing.T) {
}

func testIsAgentRunning(t *testing.T) bool {
	return true
}

func testSetupChain(t *testing.T, chainName string) {
	doCh := definitions.NowDo()
	doCh.Name = chainName
	doCh.ChainType = "simplechain"
	doCh.AccountTypes = []string{"Full:1"}

	if err := chains.MakeChain(doCh); err != nil {
		t.Fatalf("error making chain: %v\n", err)
	}

	doCh.Path = filepath.Join(common.ChainsPath, chainName)
	if err := chains.NewChain(doCh); err != nil {
		t.Fatalf("error new-ing chain: %v\n", err)
	}

	if !IsChainRunning(doCh.Name) {
		t.Fatal("chainName is not running")
	}
}

// these next two parts are hacky
// get the idi tarball from IPFS
func testGetTarballFromIPFS(t *testing.T) {
	start(t, "ipfs", false)
	time.Sleep(time.Second * 1)
	//wake up ipfs
	if err := testPutTarBalltoIPFS(t); err != nil {
		t.Fatalf("error waking up ipfs: %v\n", err)
	}

	passed := false
	for i := 0; i < 8; i++ { //usually needs 3-4
		_, err := GetTarballFromIPFS(hash, installPath)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		} else {
			passed = true
			break
		}
	}

	if !passed {
		_, err := GetTarballFromIPFS(hash, installPath)
		if err != nil {
			t.Fatalf("error getting test ball to IPFS: %v\n", err)
		}
	}
}

//ugh
func testPutTarBalltoIPFS(t *testing.T) error {
	do := definitions.NowDo()
	//TODO better solution
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error getting wd: %v\n", err)
	}
	do.Name = filepath.Join(wd, "operate.go")

	passed := false
	for i := 0; i < 6; i++ { //usually needs 3-4
		if err := files.PutFiles(do); err != nil {
			time.Sleep(2 * time.Second)
			continue
		} else {
			passed = true
			break
		}
	}

	if !passed {
		if err := files.PutFiles(do); err != nil {
			t.Fatalf("error putting test ball to IPFS: %v\n", err)

		}
	}
	return nil
}

func testKillChain(t *testing.T, chainName string) {
	doCh := definitions.NowDo()
	doCh.Name = chainName
	doCh.Rm = true
	doCh.Force = true
	if err := chains.KillChain(doCh); err != nil {
		t.Fatal("error killing chain: %v\n", err)
	}
}

func start(t *testing.T, serviceName string, publishAll bool) {
	do := definitions.NowDo()
	do.Operations.Args = []string{serviceName}
	do.Operations.PublishAllPorts = publishAll
	if err := services.StartService(do); err != nil {
		t.Fatalf("expected service to start, got %v", err)
	}
}

func kill(t *testing.T, serviceName string, wipe bool) {
	do := definitions.NowDo()
	do.Name = serviceName
	do.Operations.Args = []string{serviceName}
	if wipe {
		do.Rm = true
		do.RmD = true
	}
	if err := services.KillService(do); err != nil {
		t.Fatalf("expected service to be stopped, got %v", err)
	}
}

func testStartAgent(t *testing.T) {
	doAgent := definitions.NowDo()

	tests.IfExit(StartAgent(doAgent))

	if !testIsAgentRunning(t) {
		t.Fatalf("expected running agent, agent is not running")
	}
}

func testStopAgent(t *testing.T, agentName string) {
	doAgent := definitions.NowDo()
	doAgent.Name = agentName

	tests.IfExit(StopAgent(doAgent))

	if testIsAgentRunning(t) {
		t.Fatalf("expected no agent running, found running agent")
	}

}

///TODO! deal with all this
func _TestAuthenticateUser(t *testing.T) {
	//TODO set this up right
	user := "sire"
	if !AuthenticateUser(user) {
		t.Fatalf("permissioned denied")
	}
}

// XXX ignoring this feature for now
func _TestAuthenticateAgent(t *testing.T) {
	// is the name of the agent registered with eris?
	// similar to above, or duplicate??
}
