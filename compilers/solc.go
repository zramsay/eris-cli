package compilers

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/monax/cli/log"
	//"github.com/monax/cli/util"
)

//The following represents solidity outputs from the compiler that we're interested in
type SolcReturn struct {
	Warning   string
	Error     error
	Version   string                `json:"version"`
	Contracts map[string]*SolcItems `json:"contracts"`
}

//The key return items to enable unmarshalling from the returns from the compiler
type SolcItems struct {
	//Note: There will be more fields than this in future versions, this is just a base set of requirements
	Bin     string `json:"bin"`
	Abi     string `json:"abi"`
	DevDoc  string `json:"devdoc"`
	UserDoc string `json:"userdoc"`
	Opcodes string `json:"opcodes"`
	Asm     string `json:"asm"`
}

//This is a template to define our inputs for the compiler image
type SolcTemplate struct {
	// (Optional) what to get in the output, can be any combination of [abi, bin, clone-bin, bin-runtime, userdoc, devdoc, asm]
	// abi: application binary interface. Necessary for interaction with contracts.
	// bin: binary bytecode. Necessary for creating and deploying and interacting with contracts.
	// clone-bin: Cloned contracts binary.
	// bin-runtime: Runtime binaries for contracts.
	// userdoc: natspec for users.
	// devdoc: natspec for devs.
	// asm: assembly opcodes.
	CombinedOutput []string `mapstructure:"combined-json" yaml:"combined-json"`
	// (Optional) Direct string of library address mappings.
	//  Syntax: <libraryName>:<address>
	//  Address is interpreted as a hex string optionally prefixed by 0x.
	Libraries []string `mapstructure:"libraries" yaml:"libraries"`
	// (Optional) Remappings, see https://solidity.readthedocs.io/en/latest/layout-of-source-files.html#use-in-actual-compilers
	// Syntax: <remoteName>=<localName>
	Remappings []string `mapstructure:"remappings" yaml:"remappings"`
	// (Optional) Whether or not to get a gas estimation. Default false.
	// Cannot get abi, binaries and documentations while enabled.
	GasEstimate bool `mapstructure:"gas-estimate" yaml:"gas-estimate"`
	// (Optional) if true, optimizes solidity code
	Optimize bool `mapstructure:"optimize" yaml:"optimize"`
	// (Optional) the number of optimization runs to run on solidity
	OptimizeRuns uint64 `mapstructure:"optimize-runs" yaml:"optimize-runs"`
	// (Optional) For anything else we may have missed
	Exec string `mapstructure:"exec" yaml:"exec"`
}

// Compiles a series of files using the solidity compiler
func (s *SolcTemplate) Compile(files []string, version string) (Return, error) {
	solcExecute := []string{"solc"}
	solReturn := &SolcReturn{}
	image := "ethereum/solc:" + version

	//check files for .bin extension for linking addresses
	//separate .sol and .bin files
	//link .bins separately
	solFiles, binFiles, err := s.sortFiles(files)
	if err != nil {
		return Return{}, err
	}

	if len(binFiles) > 0 {
		solcExecute = append(solcExecute, append([]string{"--link", "--libraries", strings.Join(s.Libraries, ",")}, binFiles...)...)
		log.Warn(solcExecute)
		output, err := ExecuteCompilerCommand(image, solcExecute)
		//Parse output into a return
		if err != nil {
			if err.Error() != "Compiler error." {
				return Return{}, err
			}
			solReturn.Error = errors.New(strings.TrimSpace(string(output)))
			return Return{*solReturn}, nil
		}

		if len(solFiles) == 0 {
			return Return{}, nil
		}
		solcExecute = []string{"solc"}
	}

	//assemble command
	switch {
	case s.Exec != "":
		solcExecute = append(solcExecute, s.Exec)
	case s.GasEstimate:
		solcExecute = append(solcExecute, "--gas")
	default:
		if len(s.CombinedOutput) > 0 {
			solcExecute = append(solcExecute, "--combined-json", strings.Join(s.CombinedOutput, ","))
		}
		if len(s.Libraries) > 0 {
			solcExecute = append(solcExecute, "--libraries", strings.Join(s.Libraries, ","))
		}
		if len(s.Remappings) > 0 {
			solcExecute = append(solcExecute, strings.Join(s.Remappings, " "))
		}
		if s.Optimize {
			solcExecute = append(solcExecute, "--optimize")

			if s.OptimizeRuns != 0 {
				solcExecute = append(solcExecute, "--optimize-runs", strconv.FormatUint(s.OptimizeRuns, 10))
			}
		}
	}
	solcExecute = append(solcExecute, solFiles...)
	//Execute command
	log.Warn(solcExecute)
	output, err := ExecuteCompilerCommand(image, solcExecute)
	//Parse output into a return
	if err != nil {
		if err.Error() != "Compiler error." {
			return Return{}, err
		}
		solReturn.Error = errors.New(strings.TrimSpace(string(output)))
		return Return{*solReturn}, nil
	}
	trimmedOutput := strings.TrimSpace(string(output))
	jsonBeginsCertainly := strings.Index(trimmedOutput, `{"contracts":`)

	if jsonBeginsCertainly > 0 {
		solReturn.Warning = trimmedOutput[:jsonBeginsCertainly]
		trimmedOutput = trimmedOutput[jsonBeginsCertainly:]
	}

	log.WithField("Json: ", string(output)).Info("Command Output")
	if err = json.Unmarshal([]byte(trimmedOutput), solReturn); err != nil {
		return Return{}, err
	}

	return Return{*solReturn}, nil
}

// A utility function to sort .sol and .bin files into separate slices
func (s *SolcTemplate) sortFiles(files []string) ([]string, []string, error) {
	var solFiles []string
	var binFiles []string
	for _, file := range files {
		switch path.Ext(file) {
		case ".sol":
			solFiles = append(solFiles, file)
		case ".bin":
			binFiles = append(binFiles, file)
		default:
			return nil, nil, fmt.Errorf("Unexpected file extension found during compilation for solc: %v", file)
		}
	}
	return solFiles, binFiles, nil
}
