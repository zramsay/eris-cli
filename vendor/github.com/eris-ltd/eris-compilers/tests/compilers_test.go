package compilersTest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-compilers/perform"
	"github.com/eris-ltd/eris-compilers/definitions"
	"github.com/eris-ltd/eris-compilers/util"

	"github.com/eris-ltd/eris-cli/config"
)

func TestRequestCreation(t *testing.T) {
	var err error
	contractCode := `pragma solidity ^0.4.0;

contract c {
    function f() {
        uint8[5] memory foo3 = [1, 1, 1, 1, 1];
    }
}`
	var testMap = map[string]*definitions.IncludedFiles{
		"27fbf28c5dfb221f98526c587c5762cdf4025e85809c71ba871caa2ca42a9d85.sol": {
			ObjectNames: []string{"c"},
			Script:      []byte(contractCode),
		},
	}

	req, err := perform.CreateRequest("simpleContract.sol", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if req.Libraries != "" {
		t.Errorf("Expected empty libraries, got ", req.Libraries)
	}
	if req.Language != "sol" {
		t.Errorf("Expected Solidity file, got ", req.Language)
	}
	if req.Optimize != false {
		t.Errorf("Expected false optimize, got true")
	}
	if !reflect.DeepEqual(req.Includes, testMap) {
		t.Errorf("Got incorrect Includes map, expected %v, got %v", testMap, req.Includes)
	}

}

func TestServerSingle(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(perform.CompileHandler))
	defer testServer.Close()

	expectedSolcResponse := definitions.BlankSolcResponse()

	actualOutput, err := exec.Command("solc", "--combined-json", "bin,abi", "simpleContract.sol").Output()
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSpace(string(actualOutput))
	err = json.Unmarshal([]byte(output), expectedSolcResponse)

	respItemArray := make([]perform.ResponseItem, 0)

	for contract, item := range expectedSolcResponse.Contracts {
		respItem := perform.ResponseItem{
			Objectname: strings.TrimSpace(contract),
			Bytecode:   strings.TrimSpace(item.Bin),
			ABI:        strings.TrimSpace(item.Abi),
		}
		respItemArray = append(respItemArray, respItem)
	}
	expectedResponse := &perform.Response{
		Objects: respItemArray,
		Warning: "",
		Version: "",
		Error:   "",
	}
	util.ClearCache(config.SolcScratchPath)
	t.Log(testServer.URL)
	resp, err := perform.RequestCompile(testServer.URL, "simpleContract.sol", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedResponse, resp) {
		t.Errorf("Got incorrect response, expected %v, \n\n got %v", expectedResponse, resp)
	}

	util.ClearCache(config.SolcScratchPath)
}

func TestServerMulti(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(perform.CompileHandler))
	defer testServer.Close()
	util.ClearCache(config.SolcScratchPath)
	expectedSolcResponse := definitions.BlankSolcResponse()

	actualOutput, err := exec.Command("solc", "--combined-json", "bin,abi", "contractImport1.sol").Output()
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSpace(string(actualOutput))
	err = json.Unmarshal([]byte(output), expectedSolcResponse)

	respItemArray := make([]perform.ResponseItem, 0)

	for contract, item := range expectedSolcResponse.Contracts {
		respItem := perform.ResponseItem{
			Objectname: strings.TrimSpace(contract),
			Bytecode:   strings.TrimSpace(item.Bin),
			ABI:        strings.TrimSpace(item.Abi),
		}
		respItemArray = append(respItemArray, respItem)
	}
	expectedResponse := &perform.Response{
		Objects: respItemArray,
		Warning: "",
		Version: "",
		Error:   "",
	}
	util.ClearCache(config.SolcScratchPath)
	t.Log(testServer.URL)
	resp, err := perform.RequestCompile(testServer.URL, "contractImport1.sol", false, "")
	if err != nil {
		t.Fatal(err)
	}
	allClear := true
	for _, object := range expectedResponse.Objects {
		if !contains(resp.Objects, object) {
			allClear = false
		}
	}
	if !allClear {
		t.Errorf("Got incorrect response, expected %v, \n\n got %v", expectedResponse, resp)
	}
	util.ClearCache(config.SolcScratchPath)
}

func TestLocalMulti(t *testing.T) {
	util.ClearCache(config.SolcScratchPath)
	expectedSolcResponse := definitions.BlankSolcResponse()

	actualOutput, err := exec.Command("solc", "--combined-json", "bin,abi", "contractImport1.sol").Output()
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSpace(string(actualOutput))
	err = json.Unmarshal([]byte(output), expectedSolcResponse)

	respItemArray := make([]perform.ResponseItem, 0)

	for contract, item := range expectedSolcResponse.Contracts {
		respItem := perform.ResponseItem{
			Objectname: strings.TrimSpace(contract),
			Bytecode:   strings.TrimSpace(item.Bin),
			ABI:        strings.TrimSpace(item.Abi),
		}
		respItemArray = append(respItemArray, respItem)
	}
	expectedResponse := &perform.Response{
		Objects: respItemArray,
		Warning: "",
		Version: "",
		Error:   "",
	}
	util.ClearCache(config.SolcScratchPath)
	resp, err := perform.RequestCompile("", "contractImport1.sol", false, "")
	if err != nil {
		t.Fatal(err)
	}
	allClear := true
	for _, object := range expectedResponse.Objects {
		if !contains(resp.Objects, object) {
			allClear = false
		}
	}
	if !allClear {
		t.Errorf("Got incorrect response, expected %v, \n\n got %v", expectedResponse, resp)
	}
	util.ClearCache(config.SolcScratchPath)
}

func TestLocalSingle(t *testing.T) {
	util.ClearCache(config.SolcScratchPath)
	expectedSolcResponse := definitions.BlankSolcResponse()

	actualOutput, err := exec.Command("solc", "--combined-json", "bin,abi", "simpleContract.sol").Output()
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSpace(string(actualOutput))
	err = json.Unmarshal([]byte(output), expectedSolcResponse)

	respItemArray := make([]perform.ResponseItem, 0)

	for contract, item := range expectedSolcResponse.Contracts {
		respItem := perform.ResponseItem{
			Objectname: strings.TrimSpace(contract),
			Bytecode:   strings.TrimSpace(item.Bin),
			ABI:        strings.TrimSpace(item.Abi),
		}
		respItemArray = append(respItemArray, respItem)
	}
	expectedResponse := &perform.Response{
		Objects: respItemArray,
		Warning: "",
		Version: "",
		Error:   "",
	}
	util.ClearCache(config.SolcScratchPath)
	resp, err := perform.RequestCompile("", "simpleContract.sol", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedResponse, resp) {
		t.Errorf("Got incorrect response, expected %v, \n\n got %v", expectedResponse, resp)
	}
	util.ClearCache(config.SolcScratchPath)
}

func TestFaultyContract(t *testing.T) {
	util.ClearCache(config.SolcScratchPath)
	var expectedSolcResponse perform.Response

	actualOutput, err := exec.Command("solc", "--combined-json", "bin,abi", "faultyContract.sol").CombinedOutput()
	err = json.Unmarshal(actualOutput, expectedSolcResponse)
	t.Log(expectedSolcResponse.Error)
	resp, err := perform.RequestCompile("", "faultyContract.sol", false, "")
	t.Log(resp.Error)
	if err != nil {
		if expectedSolcResponse.Error != resp.Error {
			t.Errorf("Expected %v got %v", expectedSolcResponse.Error, resp.Error)
		}
	}
	output := strings.TrimSpace(string(actualOutput))
	err = json.Unmarshal([]byte(output), expectedSolcResponse)
}

func contains(s []perform.ResponseItem, e perform.ResponseItem) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
