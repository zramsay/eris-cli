package compilers

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/monax/cli/log"
	"github.com/monax/cli/util"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.InfoLevel)
	util.DockerConnect(false, "monax")
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestSolcCompilerNormal(t *testing.T) {

	var solFile string = `pragma solidity >= 0.0.0;
	contract main {
		uint a;
		function f() {
			a = 1;
		}
	}`
	file, err := os.Create("simpleContract.sol")
	defer os.Remove("simpleContract.sol")
	if err != nil {
		t.Fatal(err)
	}
	file.WriteString(solFile)
	template := &SolcTemplate{
		CombinedOutput: []string{"bin", "abi"},
	}

	solReturn, err := template.Compile([]string{"simpleContract.sol"}, "stable")
	if err != nil {
		t.Fatal(err)
	}

	if solReturn.Error != nil || solReturn.Warning != "" || len(solReturn.Contracts) != 1 {
		t.Fatalf("Expected no errors or warnings and expected contract items. Got %v for errors, %v for warnings, and %v for contract items", solReturn.Error, solReturn.Warning, solReturn.Contracts)
	}
}

func TestSolcCompilerError(t *testing.T) {
	var solFile string = `pragma solidity >= 0.0.0;
	contract main {
		uint a;
		function f() {
			a = 1;
		}
	`
	file, err := os.Create("faultyContract.sol")
	defer os.Remove("faultyContract.sol")
	if err != nil {
		t.Fatal(err)
	}
	file.WriteString(solFile)
	template := &SolcTemplate{
		CombinedOutput: []string{"bin", "abi"},
	}

	solReturn, err := template.Compile([]string{"faultyContract.sol"}, "stable")
	if err != nil {
		t.Fatal(err)
	}
	if solReturn.Error == nil {
		t.Fatal("Expected an error, got nil.")
	}
}

func TestSolcCompilerWarning(t *testing.T) {
	var solFile string = `contract main {
		uint a;
		function f() {
			a = 1;
		}
	}`
	file, err := os.Create("simpleContract.sol")
	defer os.Remove("simpleContract.sol")
	if err != nil {
		t.Fatal(err)
	}
	file.WriteString(solFile)
	template := &SolcTemplate{
		CombinedOutput: []string{"bin", "abi"},
	}

	solReturn, err := template.Compile([]string{"simpleContract.sol"}, "stable")
	if err != nil {
		t.Fatal(err)
	}
	if solReturn.Warning == "" {
		t.Fatal("Expected a warning.")
	}
}

func TestLinkingBinaries(t *testing.T) {
	var solFile string = `pragma solidity >=0.0.0;

library Set {
  struct Data { mapping(uint => bool) flags; }
  function insert(Data storage self, uint value)
      returns (bool)
  {
      if (self.flags[value])
          return false; // already there
      self.flags[value] = true;
      return true;
  }

  function remove(Data storage self, uint value)
      returns (bool)
  {
      if (!self.flags[value])
          return false; // not there
      self.flags[value] = false;
      return true;
  }

  function contains(Data storage self, uint value)
      returns (bool)
  {
      return self.flags[value];
  }
}

contract C {
    Set.Data knownValues;
    function register(uint value) {
        if (!Set.insert(knownValues, value))
            throw;
    }
}`
	file, err := os.Create("simpleLibrary.sol")
	defer os.Remove("simpleLibrary.sol")
	if err != nil {
		t.Fatal(err)
	}
	file.WriteString(solFile)
	template := &SolcTemplate{
		CombinedOutput: []string{"bin"},
	}

	solReturn, err := template.Compile([]string{"simpleLibrary.sol"}, "stable")
	if err != nil {
		t.Fatal(err)
	}

	if solReturn.Error != nil || solReturn.Warning != "" || len(solReturn.Contracts) != 2 {
		t.Fatalf("Expected no errors or warnings and expected contract items. Got %v for errors, %v for warnings, and %v for contract items", solReturn.Error, solReturn.Warning, solReturn.Contracts)
	}
	// note: When solc upgrades to 0.4.10, will need to add "simpleLibrary.sol:" to beginning of this string
	template.Libraries = []string{"Set:0x692a70d2e424a56d2c6c27aa97d1a86395877b3a"}
	binFile, err := os.Create("C.bin")
	defer os.Remove("C.bin")
	if err != nil {
		t.Fatal(err)
	}
	binFile.WriteString(solReturn.Contracts["C"].Bin)
	_, err = template.Compile([]string{"./C.bin"}, "stable")
	if err != nil {
		t.Fatal(err)
	}
	output, err := ioutil.ReadFile("C.bin")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(output))
	if strings.Contains(string(output), "_") {
		t.Fatal("Expected binaries to link, but they did not")
	}
}

func TestLinkingBinariesAndNormalCompileMixed(t *testing.T) {
	var solFile string = `pragma solidity >=0.0.0;

library Set {
  struct Data { mapping(uint => bool) flags; }
  function insert(Data storage self, uint value)
      returns (bool)
  {
      if (self.flags[value])
          return false; // already there
      self.flags[value] = true;
      return true;
  }

  function remove(Data storage self, uint value)
      returns (bool)
  {
      if (!self.flags[value])
          return false; // not there
      self.flags[value] = false;
      return true;
  }

  function contains(Data storage self, uint value)
      returns (bool)
  {
      return self.flags[value];
  }
}

contract C {
    Set.Data knownValues;
    function register(uint value) {
        if (!Set.insert(knownValues, value))
            throw;
    }
}`
	file, err := os.Create("simpleLibrary.sol")
	defer os.Remove("simpleLibrary.sol")
	if err != nil {
		t.Fatal(err)
	}
	file.WriteString(solFile)
	template := &SolcTemplate{
		CombinedOutput: []string{"bin"},
	}

	solReturn, err := template.Compile([]string{"simpleLibrary.sol"}, "stable")
	if err != nil {
		t.Fatal(err)
	}

	if solReturn.Error != nil || solReturn.Warning != "" || len(solReturn.Contracts) != 2 {
		t.Fatalf("Expected no errors or warnings and expected contract items. Got %v for errors, %v for warnings, and %v for contract items", solReturn.Error, solReturn.Warning, solReturn.Contracts)
	}
	// note: When solc upgrades to 0.4.10, will need to add "simpleLibrary.sol:" to beginning of this string
	template.Libraries = []string{"Set:0x692a70d2e424a56d2c6c27aa97d1a86395877b3a"}
	binFile, err := os.Create("C.bin")
	defer os.Remove("C.bin")
	if err != nil {
		t.Fatal(err)
	}
	binFile.WriteString(solReturn.Contracts["C"].Bin)

	solOutput, err := template.Compile([]string{"./C.bin", "simpleLibrary.sol"}, "stable")
	if err != nil {
		t.Fatal(err)
	}
	binOutput, err := ioutil.ReadFile("C.bin")
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(binOutput), "_") {
		t.Fatal("Expected binaries to link, but they did not")
	}

	if solOutput.Error != nil || solOutput.Warning != "" || len(solOutput.Contracts) != 2 {
		t.Fatalf("Expected no errors or warnings and expected contract items. Got %v for errors, %v for warnings, and %v for contract items", solReturn.Error, solReturn.Warning, solReturn.Contracts)
	}
}

func TestMultipleFilesCompiling(t *testing.T) {
	var solFile1 string = `pragma solidity >=0.0.0;

library Set {
  struct Data { mapping(uint => bool) flags; }
  function insert(Data storage self, uint value)
      returns (bool)
  {
      if (self.flags[value])
          return false; // already there
      self.flags[value] = true;
      return true;
  }

  function remove(Data storage self, uint value)
      returns (bool)
  {
      if (!self.flags[value])
          return false; // not there
      self.flags[value] = false;
      return true;
  }

  function contains(Data storage self, uint value)
      returns (bool)
  {
      return self.flags[value];
  }
}`

	var solFile2 string = `pragma solidity >=0.0.0;
import "./set.sol";

contract C {
    Set.Data knownValues;
    function register(uint value) {
        if (!Set.insert(knownValues, value))
            throw;
    }
}`
	set, err := os.Create("Set.sol")
	defer os.Remove("Set.sol")
	if err != nil {
		t.Fatal(err)
	}
	set.WriteString(solFile1)

	c, err := os.Create("C.sol")
	defer os.Remove("C.sol")
	if err != nil {
		t.Fatal(err)
	}
	c.WriteString(solFile2)
	template := &SolcTemplate{
		CombinedOutput: []string{"bin", "abi"},
	}

	solReturn, err := template.Compile([]string{"C.sol"}, "stable")
	if err != nil {
		t.Fatal(err)
	}

	if solReturn.Error != nil || solReturn.Warning != "" || len(solReturn.Contracts) != 2 {
		t.Fatalf("Expected no errors or warnings and expected contract items. Got %v for errors, %v for warnings, and %v for contract items", solReturn.Error, solReturn.Warning, solReturn.Contracts)
	}
}

func TestRemappings(t *testing.T) {

}

func TestDefaultCompilerUnmarshalling(t *testing.T) {

}

func TestPullingDifferentVersions(t *testing.T) {

}

func TestPullingInvalidVersions(t *testing.T) {

}

func TestDefaultCompiling(t *testing.T) {

}
