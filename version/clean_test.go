package version

import (
	"fmt"
	"os"
	"path"

	"testing"

	chn "github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	def "github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/services"
	tests "github.com/eris-ltd/eris-cli/testings"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
)

//XXX this is the test for clean. It's in here to prevent import cycle
// because clean is imported by testings/testing_utils.go

var DEAD bool
var chainName string = "chain_to_clean"
var serviceName string = "ipfs"

var do *def.Do

func fatal(t *testing.T, err error) {
	if !DEAD {
		log.Flush()
		testsTearDown()
		DEAD = true
		panic(err)
	}
}

func TestMain(m *testing.M) {
	var logLevel log.LogLevel

	logLevel = 0
	// logLevel = 1
	// logLevel = 3

	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	testsInit()
	logger.Infoln("Test init completed. Starting main test sequence now.\n")

	var exitCode int
	defer func() {
		logger.Infoln("Commensing with Tests Tear Down.")
		if err := testsTearDown(); err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}
		os.Exit(exitCode)

	}()

	exitCode = m.Run()
}

func TestCleanDefault(t *testing.T) {
	//	do.Yes = true
	//start one of each chain, service, data
	//run clean
	// check that they don't exist
}

func TestCleanRmImages(t *testing.T) {
	//	do.Yes = true
	//	do.Images = true
	//pull some test images ??
	//remove them
	//ensure they gone
}

func TestCleanRmDir(t *testing.T) {
	//	do.Yes = true
	//	do.RmD = true
}

func testsInit() error {
	if err := tests.TestsInit("clean"); err != nil {
		return err
	}
	return nil

	// lay a chain service def
	testNewChain(chainName)
	//start ipfs

	return nil
}

//get rid of some of these (from /chains/chains_test.go
//use some from /services/services_test.go
func testStartChain(t *testing.T, chain string) {
	do := def.NowDo()
	do.Name = chain
	do.Operations.ContainerNumber = 1
	do.Operations.PublishAllPorts = true
	logger.Infof("Starting chain (from tests) =>\t%s\n", do.Name)
	if e := chn.StartChain(do); e != nil {
		logger.Errorln(e)
		fatal(t, nil)
	}
	//testExistAndRun(t, chain, true, true)
}

func testNewChain(chain string) {
	do := def.NowDo()
	do.GenesisFile = path.Join(common.BlockchainsPath, "default", "genesis.json")
	do.Name = chain
	do.Operations.ContainerNumber = 1
	do.Operations.PublishAllPorts = true
	logger.Infof("Creating chain (from tests) =>\t%s\n", chain)
	tests.IfExit(chn.NewChain(do))

	// remove the data container
	do.Args = []string{chain}
	tests.IfExit(data.RmData(do))
}

func testKillChain(t *testing.T, chain string) {
	// log.SetLoggers(2, os.Stdout, os.Stderr)
	//testExistAndRun(t, chain, true, true)

	do := def.NowDo()
	do.Args, do.Rm, do.RmD = []string{"keys"}, true, true
	logger.Infof("Killing keys (from tests) =>\n%s\n", do.Name)
	if e := services.KillService(do); e != nil {
		fatal(t, e)
	}

	do = def.NowDo()
	do.Name, do.Rm, do.RmD = chain, true, true
	logger.Infof("Stopping chain (from tests) =>\t%s\n", do.Name)
	if e := chn.KillChain(do); e != nil {
		fatal(t, e)
	}
	//testExistAndRun(t, chain, false, false)
}

func testsTearDown() error {
	DEAD = true
	killService("keys")
	testKillChain(nil, chainName)
	log.Flush()
	return tests.TestsTearDown()
}

//func test kill chain

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
