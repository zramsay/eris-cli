package mock

// Basic imports
import (
	"bytes"
	"encoding/hex"
	"net/http"
	"os"
	"runtime"
	"testing"

	consensus_types "github.com/eris-ltd/eris-db/consensus/types"

	account "github.com/eris-ltd/eris-db/account"
	core_types "github.com/eris-ltd/eris-db/core/types"
	event "github.com/eris-ltd/eris-db/event"
	rpc "github.com/eris-ltd/eris-db/rpc"
	rpc_v0 "github.com/eris-ltd/eris-db/rpc/v0"
	server "github.com/eris-ltd/eris-db/server"
	td "github.com/eris-ltd/eris-db/test/testdata/testdata"
	"github.com/eris-ltd/eris-db/txs"

	"github.com/eris-ltd/eris-db/rpc/v0/shared"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/log15"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log15.Root().SetHandler(log15.LvlFilterHandler(
		log15.LvlWarn,
		log15.StreamHandler(os.Stdout, log15.TerminalFormat()),
	))
	gin.SetMode(gin.ReleaseMode)
}

type MockSuite struct {
	suite.Suite
	baseDir      string
	serveProcess *server.ServeProcess
	codec        rpc.Codec
	sUrl         string
	testData     *td.TestData
}

func (mockSuite *MockSuite) SetupSuite() {
	gin.SetMode(gin.ReleaseMode)
	// Load the supporting objects.
	testData := td.LoadTestData()
	pipe := NewMockPipe(testData)
	codec := &rpc_v0.TCodec{}
	evtSubs := event.NewEventSubscriptions(pipe.Events())
	// The server
	restServer := rpc_v0.NewRestServer(codec, pipe, evtSubs)
	sConf := server.DefaultServerConfig()
	sConf.Bind.Port = 31402
	// Create a server process.
	proc, _ := server.NewServeProcess(sConf, restServer)
	err := proc.Start()
	if err != nil {
		panic(err)
	}
	mockSuite.serveProcess = proc
	mockSuite.codec = rpc_v0.NewTCodec()
	mockSuite.testData = testData
	mockSuite.sUrl = "http://localhost:31402"
}

func (mockSuite *MockSuite) TearDownSuite() {
	sec := mockSuite.serveProcess.StopEventChannel()
	mockSuite.serveProcess.Stop(0)
	<-sec
}

// ********************************************* Accounts *********************************************

func (mockSuite *MockSuite) TestGetAccounts() {
	resp := mockSuite.get("/accounts")
	ret := &core_types.AccountList{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetAccounts.Output, ret)
}

func (mockSuite *MockSuite) TestGetAccount() {
	addr := hex.EncodeToString(mockSuite.testData.GetAccount.Input.Address)
	resp := mockSuite.get("/accounts/" + addr)
	ret := &account.Account{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetAccount.Output, ret)
}

func (mockSuite *MockSuite) TestGetStorage() {
	addr := hex.EncodeToString(mockSuite.testData.GetStorage.Input.Address)
	resp := mockSuite.get("/accounts/" + addr + "/storage")
	ret := &core_types.Storage{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetStorage.Output, ret)
}

func (mockSuite *MockSuite) TestGetStorageAt() {
	addr := hex.EncodeToString(mockSuite.testData.GetStorageAt.Input.Address)
	key := hex.EncodeToString(mockSuite.testData.GetStorageAt.Input.Key)
	resp := mockSuite.get("/accounts/" + addr + "/storage/" + key)
	ret := &core_types.StorageItem{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetStorageAt.Output, ret)
}

// ********************************************* Blockchain *********************************************

func (mockSuite *MockSuite) TestGetBlockchainInfo() {
	resp := mockSuite.get("/blockchain")
	ret := &core_types.BlockchainInfo{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetBlockchainInfo.Output, ret)
}

func (mockSuite *MockSuite) TestGetChainId() {
	resp := mockSuite.get("/blockchain/chain_id")
	ret := &core_types.ChainId{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetChainId.Output, ret)
}

func (mockSuite *MockSuite) TestGetGenesisHash() {
	resp := mockSuite.get("/blockchain/genesis_hash")
	ret := &core_types.GenesisHash{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetGenesisHash.Output, ret)
}

func (mockSuite *MockSuite) TestLatestBlockHeight() {
	resp := mockSuite.get("/blockchain/latest_block_height")
	ret := &core_types.LatestBlockHeight{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetLatestBlockHeight.Output, ret)
}

func (mockSuite *MockSuite) TestBlocks() {
	resp := mockSuite.get("/blockchain/blocks")
	ret := &core_types.Blocks{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetBlocks.Output, ret)
}

// ********************************************* Consensus *********************************************

// TODO: re-enable these when implemented
//func (mockSuite *MockSuite) TestGetConsensusState() {
//	resp := mockSuite.get("/consensus")
//	ret := &core_types.ConsensusState{}
//	errD := mockSuite.codec.Decode(ret, resp.Body)
//	mockSuite.NoError(errD)
//	ret.StartTime = ""
//	mockSuite.Equal(mockSuite.testData.GetConsensusState.Output, ret)
//}
//
//func (mockSuite *MockSuite) TestGetValidators() {
//	resp := mockSuite.get("/consensus/validators")
//	ret := &core_types.ValidatorList{}
//	errD := mockSuite.codec.Decode(ret, resp.Body)
//	mockSuite.NoError(errD)
//	mockSuite.Equal(mockSuite.testData.GetValidators.Output, ret)
//}

// ********************************************* NameReg *********************************************

func (mockSuite *MockSuite) TestGetNameRegEntry() {
	resp := mockSuite.get("/namereg/" + mockSuite.testData.GetNameRegEntry.Input.Name)
	ret := &core_types.NameRegEntry{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetNameRegEntry.Output, ret)
}

func (mockSuite *MockSuite) TestGetNameRegEntries() {
	resp := mockSuite.get("/namereg")
	ret := &core_types.ResultListNames{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetNameRegEntries.Output, ret)
}

// ********************************************* Network *********************************************

func (mockSuite *MockSuite) TestGetNetworkInfo() {
	resp := mockSuite.get("/network")
	ret := &shared.NetworkInfo{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetNetworkInfo.Output, ret)
}

func (mockSuite *MockSuite) TestGetClientVersion() {
	resp := mockSuite.get("/network/client_version")
	ret := &core_types.ClientVersion{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetClientVersion.Output, ret)
}

func (mockSuite *MockSuite) TestGetMoniker() {
	resp := mockSuite.get("/network/moniker")
	ret := &core_types.Moniker{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetMoniker.Output, ret)
}

func (mockSuite *MockSuite) TestIsListening() {
	resp := mockSuite.get("/network/listening")
	ret := &core_types.Listening{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.IsListening.Output, ret)
}

func (mockSuite *MockSuite) TestGetListeners() {
	resp := mockSuite.get("/network/listeners")
	ret := &core_types.Listeners{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetListeners.Output, ret)
}

func (mockSuite *MockSuite) TestGetPeers() {
	resp := mockSuite.get("/network/peers")
	ret := []*consensus_types.Peer{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetPeers.Output, ret)
}

/*
func (mockSuite *MockSuite) TestGetPeer() {
	addr := mockSuite.testData.GetPeer.Input.Address
	resp := mockSuite.get("/network/peer/" + addr)
	ret := []*core_types.Peer{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetPeers.Output)
}
*/

// ********************************************* Transactions *********************************************

func (mockSuite *MockSuite) TestTransactCreate() {
	resp := mockSuite.postJson("/unsafe/txpool", mockSuite.testData.TransactCreate.Input)
	ret := &txs.Receipt{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.TransactCreate.Output, ret)
}

func (mockSuite *MockSuite) TestTransact() {
	resp := mockSuite.postJson("/unsafe/txpool", mockSuite.testData.Transact.Input)
	ret := &txs.Receipt{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.Transact.Output, ret)
}

func (mockSuite *MockSuite) TestTransactNameReg() {
	resp := mockSuite.postJson("/unsafe/namereg/txpool", mockSuite.testData.TransactNameReg.Input)
	ret := &txs.Receipt{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.TransactNameReg.Output, ret)
}

func (mockSuite *MockSuite) TestGetUnconfirmedTxs() {
	resp := mockSuite.get("/txpool")
	ret := &txs.UnconfirmedTxs{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.GetUnconfirmedTxs.Output, ret)
}

func (mockSuite *MockSuite) TestCallCode() {
	resp := mockSuite.postJson("/codecalls", mockSuite.testData.CallCode.Input)
	ret := &core_types.Call{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.CallCode.Output, ret)
}

func (mockSuite *MockSuite) TestCall() {
	resp := mockSuite.postJson("/calls", mockSuite.testData.Call.Input)
	ret := &core_types.Call{}
	errD := mockSuite.codec.Decode(ret, resp.Body)
	mockSuite.NoError(errD)
	mockSuite.Equal(mockSuite.testData.CallCode.Output, ret)
}

// ********************************************* Utilities *********************************************

func (mockSuite *MockSuite) get(endpoint string) *http.Response {
	resp, errG := http.Get(mockSuite.sUrl + endpoint)
	mockSuite.NoError(errG)
	mockSuite.Equal(200, resp.StatusCode)
	return resp
}

func (mockSuite *MockSuite) postJson(endpoint string, v interface{}) *http.Response {
	bts, errE := mockSuite.codec.EncodeBytes(v)
	mockSuite.NoError(errE)
	resp, errP := http.Post(mockSuite.sUrl+endpoint, "application/json", bytes.NewBuffer(bts))
	mockSuite.NoError(errP)
	mockSuite.Equal(200, resp.StatusCode)
	return resp
}

// ********************************************* Entrypoint *********************************************

func TestMockSuite(t *testing.T) {
	suite.Run(t, &MockSuite{})
}
