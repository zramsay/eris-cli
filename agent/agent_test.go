package agent

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/keys"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/testutil"
)

var (
	// incoming post request args
	// auth added in SetBundlePath
	bundleInfo = map[string]string{
		"groupId":  "io.monax",
		"bundleId": "marmoty-contracts",
		"version":  "2.1.2", //don't use ints since path joining
	}
	hash      = "QmdbzmNH1iDg2H86Uk3USuJ2vaugvwU7HubvCxMC2fUykm"
	chainName = "agent-test"
	address   = "C7A4F01D58FC60429A3330CED519BBE14563FFA4"

	// build the path
	installPath = SetTarballPath(bundleInfo) // where the tarball is dropped

	// expected from deploying idi.sol = > to test XXX
	chainCode = "60606040526000357C01000000000000000000000000000000000000000000000000000000009004806360FE47B11460415780636D4CE63C14605757603F565B005B605560048080359060200190919050506078565B005B606260048050506086565B6040518082815260200191505060405180910390F35B806000600050819055505B50565B600060006000505490506094565B9056"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	testutil.IfExit(testutil.Init(testutil.Pull{
		Images:   []string{"data", "pm", "keys"},
		Services: []string{"keys"},
	}))

	exitCode := m.Run()
	testutil.IfExit(testutil.TearDown())
	os.Exit(exitCode)
}

// Ensure url format satisfies required schema
// TODO parse each kinds of payload (chains, dowload, install)
func TestParsePayload(t *testing.T) {
	toTest := map[string]string{
		"groupId":   bundleInfo["groupId"],  // needed to buildpath
		"bundleId":  bundleInfo["bundleId"], // ibid
		"version":   bundleInfo["version"],  // ibid
		"hash":      hash,                   // hash of tarball
		"chainName": chainName,              // chain to deploy
		"address":   address,                // account to deploy from
	}

	rawUrlInstall := fmt.Sprintf("https://localhost:17552/install?groupId=%s&bundleId=%s&version=%s&hash=%s&chainName=%s&address=%s", toTest["groupId"], toTest["bundleId"], toTest["version"], toTest["hash"], toTest["chainName"], toTest["address"])

	requiredArguments := []string{"groupId", "bundleId", "version", "hash", "chainName", "address"}

	parsed, err := ParseURL(requiredArguments, rawUrlInstall)
	if err != nil {
		t.Fatalf("error parsing url (%s): err:\n%v", rawUrlInstall, err)
	}

	if !reflect.DeepEqual(toTest, parsed) {
		t.Fatalf("toTest (%v) does not equal parsed (%v)", toTest, parsed)
	}
}

// the test that matters!
func TestDeployContract(t *testing.T) {
	defer testutil.RemoveAllContainers()
	start(t, "keys", false)
	defer kill(t, "keys", true)

	testSetupChain(t, chainName) // has test for IsChainRunning
	defer testStopChain(t, chainName)

	doKey := definitions.NowDo()
	doKey.Container = true
	doKey.Host = false
	//there should only be one key
	keys, err := keys.ListKeys(doKey)
	if err != nil {
		t.Fatalf("err listing keys: %v\n", err)
	}
	address := keys[0]

	// testGetTarballFromIPFS(t) (replaced by using a tar'd bundle directly!)
	// defer kill(t, "ipfs", true)

	testMakeABundle(t) // untar's the bundle into installPath for deployment

	if err := DeployContractBundle(installPath, chainName, address); err != nil {
		t.Fatalf("error deploying contract bundle: %v\n", err)
	}

	// TODO check that chainCode matches expected code
}

func testMakeABundle(t *testing.T) {
	// write two files to config.AppsPath/idi
	idiPath := filepath.Join(config.AppsPath, "idi")
	if err := os.MkdirAll(idiPath, 0777); err != nil {

	}

	// make the two files we need
	generalContract, err := os.Create(filepath.Join(idiPath, "idi.sol"))
	defer generalContract.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	epmYammer, err := os.Create(filepath.Join(idiPath, "epm.yaml"))
	defer epmYammer.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// [zr] we could pull these from GH ...?
	// write to them
	idiSol := `
contract IdisContractsFTW {
  uint storedData;

  function set(uint x) {
    storedData = x;
  }

  function get() constant returns (uint retVal) {
    return storedData;
  }
}
`
	_, err = io.WriteString(generalContract, idiSol)
	if err != nil {
		t.Fatalf("%v", err)
	}

	epmYaml := `
jobs:

- name: setStorageBase
  job:
    set:
      val: 5

- name: deployStorageK
  job:
    deploy:
      contract: idi.sol
      wait: true

- name: setStorage
  job:
    call:
      destination: $deployStorageK
      data: set $setStorageBase
      wait: true

- name: queryStorage
  job:
    query-contract:
      destination: $deployStorageK
      data: get

- name: assertStorage
  job:
    assert:
      key: $queryStorage
      relation: eq
      val: $setStorageBase
`
	_, err = io.WriteString(epmYammer, epmYaml)
	if err != nil {
		t.Fatalf("%v", err)
	}

	tarballName := "addMeToIPFSeventually.tar.gz"
	// os.Exec the tar function on them
	if err := os.Chdir(config.AppsPath); err != nil {
		t.Fatalf("%v", err)
	}
	targs := []string{"-cvzf", tarballName, "-C", idiPath, "."}

	stdOut, err := exec.Command("tar", targs...).CombinedOutput()
	if err != nil {
		t.Fatalf("error with tar:%v\n%s", err, string(stdOut))
	}
	log.Warn(string(stdOut))

	from := filepath.Join(config.AppsPath, tarballName)
	to := installPath

	if err := UnpackTarball(from, to); err != nil {
		t.Fatalf("%v", err)
	}
}

func testSetupChain(t *testing.T, chainName string) {
	doCh := definitions.NowDo()
	doCh.Name = chainName
	doCh.ChainType = "simplechain"

	if err := chains.MakeChain(doCh); err != nil {
		t.Fatalf("error making chain: %v\n", err)
	}

	doCh.Path = filepath.Join(config.ChainsPath, chainName)
	if err := chains.StartChain(doCh); err != nil {
		t.Fatalf("error new-ing chain: %v\n", err)
	}

	if !IsChainRunning(doCh.Name) {
		t.Fatal("chainName is not running")
	}
}

func testStopChain(t *testing.T, chainName string) {
	doCh := definitions.NowDo()
	doCh.Name = chainName
	doCh.Rm = true
	doCh.Force = true
	if err := chains.StopChain(doCh); err != nil {
		t.Fatalf("error killing chain: %v", err)
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

//TODO set this up right
/*
func _TestAuthenticateUser(t *testing.T) {
	user := "sire"
	if !AuthenticateUser(user) {
		t.Fatalf("permissioned denied")
	}
}

func _TestAuthenticateAgent(t *testing.T) {
	// is the name of the agent registered with eris?
	// similar to above, or duplicate??
}*/

// these next two parts are hacky
// get the idi tarball from IPFS
// not currently used
/*func testGetTarballFromIPFS(t *testing.T) {
	start(t, "ipfs", false)
	time.Sleep(time.Second * 1)
	//wake up ipfs
	//if err := testPutTarBalltoIPFS(t); err != nil {
	//	t.Fatalf("error waking up ipfs: %v\n", err)
	//}

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
}*/
