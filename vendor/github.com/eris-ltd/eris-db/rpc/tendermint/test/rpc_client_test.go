// +build integration

// Space above here matters
package test

import (
	"bytes"
	"fmt"
	"testing"

	"golang.org/x/crypto/ripemd160"

	"time"

	consensus_types "github.com/eris-ltd/eris-db/consensus/types"
	edbcli "github.com/eris-ltd/eris-db/rpc/tendermint/client"
	"github.com/eris-ltd/eris-db/txs"
	"github.com/stretchr/testify/assert"
	tm_common "github.com/tendermint/go-common"
	rpcclient "github.com/tendermint/go-rpc/client"
	_ "github.com/tendermint/tendermint/config/tendermint_test"
)

// When run with `-test.short` we only run:
// TestHTTPStatus, TestHTTPBroadcast, TestJSONStatus, TestJSONBroadcast, TestWSConnect, TestWSSend

// Note: the reason that we have tests implemented in tests.go is I believe
// due to weirdness with go-wire's interface registration, and those global
// registrations not being available within a *_test.go runtime context.
func testWithAllClients(t *testing.T,
	testFunction func(*testing.T, string, rpcclient.Client)) {
	for clientName, client := range clients {
		testFunction(t, clientName, client)
	}
}

//--------------------------------------------------------------------------------
func TestStatus(t *testing.T) {
	testWithAllClients(t, func(t *testing.T, clientName string, client rpcclient.Client) {
		resp, err := edbcli.Status(client)
		assert.NoError(t, err)
		fmt.Println(resp)
		if resp.NodeInfo.Network != chainID {
			t.Fatal(fmt.Errorf("ChainID mismatch: got %s expected %s",
				resp.NodeInfo.Network, chainID))
		}
	})
}

func TestBroadcastTx(t *testing.T) {
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client rpcclient.Client) {
		// Avoid duplicate Tx in mempool
		amt := hashString(clientName) % 1000
		toAddr := user[1].Address
		tx := makeDefaultSendTxSigned(t, client, toAddr, amt)
		//receipt := broadcastTx(t, client, tx)
		receipt, err := broadcastTxAndWaitForBlock(t, client, wsc, tx)
		assert.NoError(t, err)
		if receipt.CreatesContract > 0 {
			t.Fatal("This tx does not create a contract")
		}
		if len(receipt.TxHash) == 0 {
			t.Fatal("Failed to compute tx hash")
		}
		n, errp := new(int), new(error)
		buf := new(bytes.Buffer)
		hasher := ripemd160.New()
		tx.WriteSignBytes(chainID, buf, n, errp)
		assert.NoError(t, *errp)
		txSignBytes := buf.Bytes()
		hasher.Write(txSignBytes)
		txHashExpected := hasher.Sum(nil)
		if bytes.Compare(receipt.TxHash, txHashExpected) != 0 {
			t.Fatalf("The receipt hash '%x' does not equal the ripemd160 hash of the "+
				"transaction signed bytes calculated in the test: '%x'",
				receipt.TxHash, txHashExpected)
		}
	})
}

func TestGetAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testWithAllClients(t, func(t *testing.T, clientName string, client rpcclient.Client) {
		acc := getAccount(t, client, user[0].Address)
		if acc == nil {
			t.Fatal("Account was nil")
		}
		if bytes.Compare(acc.Address, user[0].Address) != 0 {
			t.Fatalf("Failed to get correct account. Got %x, expected %x", acc.Address,
				user[0].Address)
		}
	})
}

func TestGetStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client rpcclient.Client) {
		eid := txs.EventStringNewBlock()
		subscribe(t, wsc, eid)
		defer func() {
			unsubscribe(t, wsc, eid)
			wsc.Stop()
		}()

		amt, gasLim, fee := int64(1100), int64(1000), int64(1000)
		code := []byte{0x60, 0x5, 0x60, 0x1, 0x55}
		// Call with nil address will create a contract
		tx := makeDefaultCallTx(t, client, nil, code, amt, gasLim, fee)
		receipt, err := broadcastTxAndWaitForBlock(t, client, wsc, tx)
		assert.NoError(t, err)
		assert.Equal(t, uint8(1), receipt.CreatesContract, "This transaction should"+
			" create a contract")
		assert.NotEqual(t, 0, len(receipt.TxHash), "Receipt should contain a"+
			" transaction hash")
		contractAddr := receipt.ContractAddr
		assert.NotEqual(t, 0, len(contractAddr), "Transactions claims to have"+
			" created a contract but the contract address is empty")

		v := getStorage(t, client, contractAddr, []byte{0x1})
		got := tm_common.LeftPadWord256(v)
		expected := tm_common.LeftPadWord256([]byte{0x5})
		if got.Compare(expected) != 0 {
			t.Fatalf("Wrong storage value. Got %x, expected %x", got.Bytes(),
				expected.Bytes())
		}
	})
}

func TestCallCode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testWithAllClients(t, func(t *testing.T, clientName string, client rpcclient.Client) {
		// add two integers and return the result
		code := []byte{0x60, 0x5, 0x60, 0x6, 0x1, 0x60, 0x0, 0x52, 0x60, 0x20, 0x60,
			0x0, 0xf3}
		data := []byte{}
		expected := []byte{0xb}
		callCode(t, client, user[0].PubKey.Address(), code, data, expected)

		// pass two ints as calldata, add, and return the result
		code = []byte{0x60, 0x0, 0x35, 0x60, 0x20, 0x35, 0x1, 0x60, 0x0, 0x52, 0x60,
			0x20, 0x60, 0x0, 0xf3}
		data = append(tm_common.LeftPadWord256([]byte{0x5}).Bytes(),
			tm_common.LeftPadWord256([]byte{0x6}).Bytes()...)
		expected = []byte{0xb}
		callCode(t, client, user[0].PubKey.Address(), code, data, expected)
	})
}

func TestCallContract(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client rpcclient.Client) {
		eid := txs.EventStringNewBlock()
		subscribe(t, wsc, eid)
		defer func() {
			unsubscribe(t, wsc, eid)
			wsc.Stop()
		}()

		// create the contract
		amt, gasLim, fee := int64(6969), int64(1000), int64(1000)
		code, _, _ := simpleContract()
		tx := makeDefaultCallTx(t, client, nil, code, amt, gasLim, fee)
		receipt, err := broadcastTxAndWaitForBlock(t, client, wsc, tx)
		assert.NoError(t, err)
		if err != nil {
			t.Fatalf("Problem broadcasting transaction: %v", err)
		}
		assert.Equal(t, uint8(1), receipt.CreatesContract, "This transaction should"+
			" create a contract")
		assert.NotEqual(t, 0, len(receipt.TxHash), "Receipt should contain a"+
			" transaction hash")
		contractAddr := receipt.ContractAddr
		assert.NotEqual(t, 0, len(contractAddr), "Transactions claims to have"+
			" created a contract but the contract address is empty")

		// run a call through the contract
		data := []byte{}
		expected := []byte{0xb}
		callContract(t, client, user[0].PubKey.Address(), contractAddr, data, expected)
	})
}

func TestNameReg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client rpcclient.Client) {

		txs.MinNameRegistrationPeriod = 1

		// register a new name, check if its there
		// since entries ought to be unique and these run against different clients, we append the client
		name := "ye_old_domain_name_" + clientName
		const data = "if not now, when"
		fee := int64(1000)
		numDesiredBlocks := int64(2)
		amt := fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier*txs.NameBaseCost(name, data)

		tx := makeDefaultNameTx(t, client, name, data, amt, fee)
		// verify the name by both using the event and by checking get_name
		subscribeAndWaitForNext(t, wsc, txs.EventStringNameReg(name),
			func() {
				broadcastTxAndWaitForBlock(t, client, wsc, tx)
			},
			func(eid string, eventData txs.EventData) (bool, error) {
				eventDataTx := asEventDataTx(t, eventData)
				tx, ok := eventDataTx.Tx.(*txs.NameTx)
				if !ok {
					t.Fatalf("Could not convert %v to *NameTx", eventDataTx)
				}
				assert.Equal(t, name, tx.Name)
				assert.Equal(t, data, tx.Data)
				return true, nil
			})
		mempoolCount = 0

		entry := getNameRegEntry(t, client, name)
		assert.Equal(t, data, entry.Data)
		assert.Equal(t, user[0].Address, entry.Owner)

		// update the data as the owner, make sure still there
		numDesiredBlocks = int64(5)
		const updatedData = "these are amongst the things I wish to bestow upon " +
			"the youth of generations come: a safe supply of honey, and a better " +
			"money. For what else shall they need"
		amt = fee + numDesiredBlocks*txs.NameByteCostMultiplier*
			txs.NameBlockCostMultiplier*txs.NameBaseCost(name, updatedData)
		tx = makeDefaultNameTx(t, client, name, updatedData, amt, fee)
		broadcastTxAndWaitForBlock(t, client, wsc, tx)
		mempoolCount = 0
		entry = getNameRegEntry(t, client, name)

		assert.Equal(t, updatedData, entry.Data)

		// try to update as non owner, should fail
		tx = txs.NewNameTxWithNonce(user[1].PubKey, name, "never mind", amt, fee,
			getNonce(t, client, user[1].Address)+1)
		tx.Sign(chainID, user[1])

		_, err := broadcastTxAndWaitForBlock(t, client, wsc, tx)
		assert.Error(t, err, "Expected error when updating someone else's unexpired"+
			" name registry entry")
		if err != nil {
			assert.Contains(t, err.Error(), "permission denied", "Error should be "+
				"permission denied")
		}

		// Wait a couple of blocks to make sure name registration expires
		waitNBlocks(t, wsc, 3)

		//now the entry should be expired, so we can update as non owner
		const data2 = "this is not my beautiful house"
		tx = txs.NewNameTxWithNonce(user[1].PubKey, name, data2, amt, fee,
			getNonce(t, client, user[1].Address)+1)
		tx.Sign(chainID, user[1])
		_, err = broadcastTxAndWaitForBlock(t, client, wsc, tx)
		assert.NoError(t, err, "Should be able to update a previously expired name"+
			" registry entry as a different address")
		mempoolCount = 0
		entry = getNameRegEntry(t, client, name)
		assert.Equal(t, data2, entry.Data)
		assert.Equal(t, user[1].Address, entry.Owner)
	})
}

func TestBlockchainInfo(t *testing.T) {
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client rpcclient.Client) {
		nBlocks := 4
		waitNBlocks(t, wsc, nBlocks)

		resp, err := edbcli.BlockchainInfo(client, 0, 0)
		if err != nil {
			t.Fatalf("Failed to get blockchain info: %v", err)
		}
		//TODO: [Silas] reintroduce this when Tendermint changes logic to fire
		// NewBlock after saving a block
		// see https://github.com/tendermint/tendermint/issues/273
		//assert.Equal(t, 4, resp.LastHeight, "Last height should be 4 after waiting for first 4 blocks")
		assert.True(t, nBlocks <= len(resp.BlockMetas),
			"Should see at least 4 BlockMetas after waiting for first 4 blocks")

		lastBlockHash := resp.BlockMetas[nBlocks-1].Hash
		for i := nBlocks - 2; i >= 0; i-- {
			assert.Equal(t, lastBlockHash, resp.BlockMetas[i].Header.LastBlockHash,
				"Blockchain should be a hash tree!")
			lastBlockHash = resp.BlockMetas[i].Hash
		}

		resp, err = edbcli.BlockchainInfo(client, 1, 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(resp.BlockMetas),
			"Should see 2 BlockMetas after extracting 2 blocks")
	})
}

func TestListUnconfirmedTxs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client rpcclient.Client) {
		amt, gasLim, fee := int64(1100), int64(1000), int64(1000)
		code := []byte{0x60, 0x5, 0x60, 0x1, 0x55}
		// Call with nil address will create a contract
		tx := makeDefaultCallTx(t, client, []byte{}, code, amt, gasLim, fee)
		txChan := make(chan []txs.Tx)

		// We want to catch the Tx in mempool before it gets reaped by tendermint
		// consensus. We should be able to do this almost always if we broadcast our
		// transaction immediately after a block has been committed. There is about
		// 1 second between blocks, and we will have the lock on Reap
		// So we wait for a block here
		waitNBlocks(t, wsc, 1)

		go func() {
			for {
				resp, err := edbcli.ListUnconfirmedTxs(client)
				assert.NoError(t, err)
				if resp.N > 0 {
					txChan <- resp.Txs
				}
			}
		}()

		runThenWaitForBlock(t, wsc, nextBlockPredicateFn(), func() {
			broadcastTx(t, client, tx)
			select {
			case <-time.After(time.Second * timeoutSeconds):
				t.Fatal("Timeout out waiting for unconfirmed transactions to appear")
			case transactions := <-txChan:
				assert.Len(t, transactions, 1,
					"There should only be a single transaction in the mempool during "+
						"this test (previous txs should have made it into a block)")
				assert.Contains(t, transactions, tx,
					"Transaction should be returned by ListUnconfirmedTxs")
			}
		})
	})
}

func TestGetBlock(t *testing.T) {
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client rpcclient.Client) {
		waitNBlocks(t, wsc, 3)
		resp, err := edbcli.GetBlock(client, 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, resp.Block.Height)
		assert.Equal(t, 2, resp.BlockMeta.Header.Height)
	})
}

func TestListValidators(t *testing.T) {
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client rpcclient.Client) {
		waitNBlocks(t, wsc, 3)
		resp, err := edbcli.ListValidators(client)
		assert.NoError(t, err)
		assert.Len(t, resp.BondedValidators, 1)
		validator := resp.BondedValidators[0].(*consensus_types.TendermintValidator)
		assert.Equal(t, genesisDoc.Validators[0].PubKey, validator.PubKey)
	})
}

func TestDumpConsensusState(t *testing.T) {
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client rpcclient.Client) {
		waitNBlocks(t, wsc, 3)
		resp, err := edbcli.DumpConsensusState(client)
		assert.NoError(t, err)
		startTime := resp.ConsensusState.StartTime
		// TODO: uncomment lines involving commitTime when
		// https://github.com/tendermint/tendermint/issues/277 is fixed in Tendermint
		//commitTime := resp.ConsensusState.CommitTime
		assert.NotZero(t, startTime)
		//assert.NotZero(t, commitTime)
		//assert.True(t, commitTime.Unix() > startTime.Unix(),
		//	"Commit time %v should be later than start time %v", commitTime, startTime)
		assert.Equal(t, uint8(1), resp.ConsensusState.Step)
	})
}

func asEventDataTx(t *testing.T, eventData txs.EventData) txs.EventDataTx {
	eventDataTx, ok := eventData.(txs.EventDataTx)
	if !ok {
		t.Fatalf("Expected eventData to be EventDataTx was %v", eventData)
	}
	return eventDataTx
}
