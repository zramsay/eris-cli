package jobs

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/monax/cli/compilers"
	"github.com/monax/cli/log"
	"github.com/monax/cli/pkgs/abi"
	"github.com/monax/cli/version"

	"github.com/hyperledger/burrow/client/rpc"
)

// ------------------------------------------------------------------------
// Contracts Jobs
// ------------------------------------------------------------------------

type Compile struct {
	// embedded interface to allow us to access the Compile function regardless of which language is chosen. This gets set in the preprocessing stage.
	compilers.Compiler
	// List of files with which you would like to compile with. Optional setting if creating a plugin compiler setting for a deploy job.
	Files []string `mapstructure:"files" yaml:"files"`
	// (Optional) the version of the compiler to use, if left blank, defaults to whatever the default is in compilers.toml, if that's blank, defaults to "latest" tag of compilers image
	Version string `mapstructure:"version" yaml:"version"`
	// One of the following fields is required.

	// Solidity Compiler
	Solc *compilers.SolcTemplate `mapstructure:"solc" yaml:"solc"`
	//Use defaults
	UseDefault bool
}

func (compile *Compile) PreProcess(jobs *Jobs) (err error) {
	//Normal preprocessing
	if compile.Version, _, err = preProcessString(compile.Version, jobs); err != nil {
		return err
	}

	for i, file := range compile.Files {
		if compile.Files[i], _, err = preProcessString(file, jobs); err != nil {
			return err
		}
	}

	switch {
	case !reflect.DeepEqual(compile.Solc, compilers.SolcTemplate{}):
		compile.Compiler = compile.Solc
	default:
		return fmt.Errorf("Could not find compiler to use")
	}
	return nil
}

func (compile *Compile) Execute(jobs *Jobs) (*JobResults, error) {
	compileReturn, err := compile.Compile(compile.Files, compile.Version)
	if err != nil {
		return &JobResults{}, err
	}

	switch compile.Compiler.(type) {
	case *compilers.SolcTemplate:

		if compileReturn.Warning != "" {
			log.Warn(compileReturn.Warning)
		}
		if compileReturn.Error != nil {
			log.Warn("There was an error in your contracts.")
			return &JobResults{}, compileReturn.Error
		}

		returnMessage, err := json.Marshal(compileReturn)
		if err != nil {
			return &JobResults{}, err
		}

		var namedResults map[string]Type

		for objectName, result := range compileReturn.Contracts {
			stringResult, err := json.Marshal(result)
			if err != nil {
				return &JobResults{}, err
			}
			// while we're here, let's write these abis and bins to their directories
			if err := ioutil.WriteFile(filepath.Join(jobs.AbiPath, objectName+".abi"), []byte(result.Abi), 0664); err != nil {
				return &JobResults{}, err
			}

			if err := ioutil.WriteFile(filepath.Join(jobs.BinPath, objectName+".bin"), []byte(result.Bin), 0664); err != nil {
				return &JobResults{}, err
			}

			namedResults[objectName] = Type{string(stringResult), result}
		}

		return &JobResults{
			FullResult: Type{
				StringResult: string(returnMessage),
				ActualResult: compileReturn,
			},
			NamedResults: namedResults,
		}, nil
	default:
		return &JobResults{}, fmt.Errorf("Invalid compiler type")
	}
}

type Deploy struct {
	// (Optional, if account job or global account set) address of the account from which to send (the
	// public key for the account must be available to monax-keys)
	Source string `mapstructure:"source" yaml:"source"`
	// (Required) the filepath to the contract file. this should be relative to the current path **or**
	// relative to the contracts path established via the --contracts-path flag or the $EPM_CONTRACTS_PATH
	// environment variable. If contract has a "bin" file extension then it will not be sent to the
	// compilers but rather will just be sent to the chain. Note, if you use a "call" job after deploying
	// a binary contract then you will be **required** to utilize an abi field in the call job.
	Contract string `mapstructure:"contract" yaml:"contract"`
	// (Optional) the name of contract to instantiate (it has to be one of the contracts present)
	// in the file defined in Contract above.
	// When none is provided, the system will choose the contract with the same name as that file.
	// use "all" to override and deploy all contracts in order. if "all" is selected the result
	// of the job will default to the address of the contract which was deployed that matches
	// the name of the file (or the last one deployed if there are no matching names; not the "last"
	// one deployed" strategy is non-deterministic and should not be used).
	Instance string `mapstructure:"instance" yaml:"instance"`
	// (Optional) list of Name:Address separated by commas of libraries (see solc --help)
	Libraries []string `mapstructure:"libraries" yaml:"libraries"`
	// (Optional) additional arguments to send along with the contract code
	Data []interface{} `mapstructure:"data" yaml:"data"`
	// (Optional) a job plugin for a compile job
	CompilerStub interface{} `mapstructure:"compiler" yaml:"compiler"`
	// (Optional) location of the abi file to use (will search relative to abi path)
	// deployed contracts save ABI artifacts in the abi folder as *both* the name of the contract
	// and the address where the contract was deployed to
	ABI string `mapstructure:"abi" yaml:"abi"`
	// (Optional) amount of tokens to send to the contract which will (after deployment) reside in the
	// contract's account
	Amount string `mapstructure:"amount" yaml:"amount"`
	// (Optional) validators' fee
	Fee string `mapstructure:"fee" yaml:"fee"`
	// (Optional) amount of gas which should be sent along with the contract deployment transaction
	Gas string `mapstructure:"gas" yaml:"gas"`
	// (Optional, advanced only) nonce to use when monax-keys signs the transaction (do not use unless you
	// know what you're doing)
	Nonce string `mapstructure:"nonce" yaml:"nonce"`
	// our stored compiler after preprocessing
	compiler *Compile
}

func (deploy *Deploy) PreProcess(jobs *Jobs) (err error) {
	deploy.Source, _, err = preProcessString(deploy.Source, jobs)
	if err != nil {
		return err
	}
	deploy.Contract, _, err = preProcessString(deploy.Contract, jobs)
	if err != nil {
		return err
	}

	// first check to see whether or not the contract exists in the pwd
	if _, err := os.Stat(deploy.Contract); os.IsNotExist(err) {
		// if it doesn't exist, check in the contract path now
		contractPathFile := filepath.Join(jobs.ContractPath, deploy.Contract)
		if _, err = os.Stat(contractPathFile); os.IsNotExist(err) {
			return err
		}
		deploy.Contract = contractPathFile
	}

	deploy.Instance, _, err = preProcessString(deploy.Instance, jobs)
	if err != nil {
		return err
	}

	for i, data := range deploy.Data {
		if deploy.Data[i], err = preProcessInterface(data, jobs); err != nil {
			return err
		}
	}

	buf, err := ioutil.ReadFile(filepath.Join(jobs.AbiPath, deploy.ABI))
	if err != nil {
		return err
	}
	deploy.ABI = string(buf)

	deploy.Amount, _, err = preProcessString(deploy.Amount, jobs)
	if err != nil {
		return err
	}
	deploy.Amount = useDefault(deploy.Amount, jobs.DefaultAmount)

	deploy.Fee, _, err = preProcessString(deploy.Fee, jobs)
	if err != nil {
		return err
	}
	deploy.Fee = useDefault(deploy.Fee, jobs.DefaultFee)

	deploy.Gas, _, err = preProcessString(deploy.Gas, jobs)
	if err != nil {
		return err
	}
	deploy.Gas = useDefault(deploy.Gas, jobs.DefaultGas)

	deploy.Nonce, _, err = preProcessString(deploy.Nonce, jobs)
	if err != nil {
		return err
	}

	return deploy.selectCompiler(jobs)
}

// defines rules of a plugin job compiler selection for a deploy job
func (deploy *Deploy) selectCompiler(jobs *Jobs) error {
	if deploy.CompilerStub != nil {
		compiler, err := preProcessPluginJob(deploy.CompilerStub, jobs)
		if err != nil {
			return err
		}
		switch compiler := compiler.(type) {
		case *Compile:

			compiler.Files = append(compiler.Files, deploy.Contract)

			switch compilerType := compiler.Compiler.(type) {
			case *compilers.SolcTemplate:
				compilerType.Libraries = append(compilerType.Libraries, deploy.Libraries...)
				compiler.Compiler = compilerType
			default:
				return fmt.Errorf("Could not find compiler to use")
			}

			deploy.compiler = compiler

		default:
			return fmt.Errorf("Invalid preprocessing of compiler")
		}

	} else {
		compiler := &Compile{}
		compiler.Compiler = &compilers.SolcTemplate{
			CombinedOutput: []string{"bin", "abi"},
			Libraries:      deploy.Libraries,
		}
		compiler.Files = []string{deploy.Contract}
		compiler.Version = version.SOLC_VERSION
		deploy.compiler = compiler
	}

	return nil
}

func (deploy *Deploy) Execute(jobs *Jobs) (*JobResults, error) {
	var namedResults map[string]Type
	// switch context of the compiler
	switch deploy.compiler.Compiler.(type) {
	case *compilers.SolcTemplate:
		// execute compilation
		results, err := deploy.compiler.Execute(jobs)
		if err != nil {
			return &JobResults{}, fmt.Errorf("Compile job: %v", err)
		}
		// begin deployment
		if deploy.Instance == "all" || deploy.Instance == "" || filepath.Ext(deploy.Contract) == ".bin" {
			log.Info("Deploying all contracts")
			for name, contract := range results.NamedResults {
				solcItem, ok := contract.ActualResult.(compilers.SolcItems)
				if !ok {
					return &JobResults{}, fmt.Errorf("Couldn't get the needed solc items from your compile job, did you remember to include a combined-json field with bin and abi?")
				}
				// TODO: Encapsulate the following in a function of some kind and use instead
				binary := []byte(solcItem.Bin)
				if binary == nil {
					continue
				}
				log.Warn("Deploying contract: ", name, contract)
				// create ABI
				var abiSource string
				if deploy.ABI == "" && solcItem.Abi == "" && filepath.Ext(deploy.Contract) != ".bin" {
					return &JobResults{}, fmt.Errorf("Couldn't get the needed abi from your compile job, can you provide it through the abi field in your deploy job?")
				} else if deploy.ABI != "" {
					abiSource = deploy.ABI
				} else {
					abiSource = solcItem.Abi
				}
				contractAbi, err := abi.MakeAbi(abiSource)
				if err != nil {
					return &JobResults{}, err
				}

				// format data
				constructorData, err := abi.FormatAndPackInputs(contractAbi, "", deploy.Data)
				if err != nil {
					return &JobResults{}, err
				}
				// append to binary
				contractCode := fmt.Sprintf("%X", append(binary, constructorData...))
				// call to deploy binary
				tx, err := rpc.Call(jobs.NodeClient, jobs.KeyClient, jobs.PublicKey, deploy.Source, "", deploy.Amount, deploy.Nonce, deploy.Gas, deploy.Fee, contractCode)
				if err != nil {
					return &JobResults{}, err
				}

				result, err := txFinalize(tx, jobs, Address)
				if err != nil {
					return &JobResults{}, err
				}
				// store the resulting address of the contract via a mapping of contract name -> address
				namedResults[name] = result.FullResult
				// store the abi in a address -> abi mapping
				jobs.AbiMap[result.FullResult.StringResult] = abiSource
				// store the functions of said contract in a named result (address + function signature) in the format name.function->function sig
				for methodName, method := range contractAbi.Methods {
					namedResults[name+"."+methodName] = Type{ActualResult: append(result.FullResult.ActualResult.([]byte), method.Id()...), StringResult: string(append(result.FullResult.ActualResult.([]byte), method.Id()...))}
				}
			}
			// breaking change... if instance isn't specified, then you need to call your destination by the contract name that you're talking to.
			// we need to demand precision here of users so that they don't screw themselves over.
			return &JobResults{NamedResults: namedResults}, nil
		} else {
			if solcItem, ok := results.NamedResults[deploy.Instance].ActualResult.(compilers.SolcItems); ok {
				log.WithField("=>", deploy.Instance).Warn("Deploying single contract")
				// TODO: Encapsulate the following in a function of some kind and use instead
				binary := []byte(solcItem.Bin)
				if binary == nil {
					return &JobResults{}, nil
				}
				// create ABI
				var abiSource string
				if deploy.ABI == "" && solcItem.Abi == "" && filepath.Ext(deploy.Contract) != ".bin" {
					return &JobResults{}, fmt.Errorf("Couldn't get the needed abi from your compile job, can you provide it through the abi field in your deploy job?")
				} else if deploy.ABI != "" {
					abiSource = deploy.ABI
				} else {
					abiSource = solcItem.Abi
				}
				contractAbi, err := abi.MakeAbi(abiSource)
				if err != nil {
					return &JobResults{}, err
				}

				// format data
				constructorData, err := abi.FormatAndPackInputs(contractAbi, "", deploy.Data)
				if err != nil {
					return &JobResults{}, err
				}
				// append to binary
				contractCode := fmt.Sprintf("%X", append(binary, constructorData...))
				// call to deploy binary
				tx, err := rpc.Call(jobs.NodeClient, jobs.KeyClient, jobs.PublicKey, deploy.Source, "", deploy.Amount, deploy.Nonce, deploy.Gas, deploy.Fee, contractCode)
				if err != nil {
					return &JobResults{}, err
				}
				result, err := txFinalize(tx, jobs, Address)
				if err != nil {
					return &JobResults{}, err
				}
				// store the resulting address of the contract via a mapping of contract name -> address
				namedResults[deploy.Instance] = result.FullResult
				// store the abi in a address -> abi mapping
				jobs.AbiMap[result.FullResult.StringResult] = abiSource
				// store the functions of said contract in a named result (address + function signature) in the format function->function sig
				for methodName, method := range contractAbi.Methods {
					namedResults[methodName] = Type{ActualResult: append(result.FullResult.ActualResult.([]byte), method.Id()...), StringResult: string(append(result.FullResult.ActualResult.([]byte), method.Id()...))}
				}
				// store the base result
				return &JobResults{FullResult: result.FullResult, NamedResults: namedResults}, nil
			} else {
				return &JobResults{}, fmt.Errorf("Could not acquire requested instance named %v", deploy.Instance)
			}
		}
	default:
		return &JobResults{}, fmt.Errorf("Invalid compiler used in execution process")
	}

}

type Call struct {
	// (Optional, if account job or global account set) address of the account from which to send (the
	// public key for the account must be available to monax-keys)
	Source string `mapstructure:"source" yaml:"source"`
	// (Required) the contract which should be called
	Destination string `mapstructure:"destination" yaml:"destination"`
	// (Required unless testing fallback function) function inside the contract to be called
	Function string `mapstructure:"function" yaml:"function"`
	// (Optional) data which should be called. will use the abi tooling under the hood to formalize the
	// transaction
	Data []interface{} `mapstructure:"data" yaml:"data"`
	// (Optional) amount of tokens to send to the contract
	Amount string `mapstructure:"amount" yaml:"amount"`
	// (Optional) validators' fee
	Fee string `mapstructure:"fee" yaml:"fee"`
	// (Optional) amount of gas which should be sent along with the call transaction
	Gas string `mapstructure:"gas" yaml:"gas"`
	// (Optional, advanced only) nonce to use when monax-keys signs the transaction (do not use unless you
	// know what you're doing)
	Nonce string `mapstructure:"nonce" yaml:"nonce"`
	// (Optional) location of the abi file to use (will search relative to abi path)
	// deployed contracts save ABI artifacts in the abi folder as *both* the name of the contract
	// and the address where the contract was deployed to
	ABI string `mapstructure:"abi" yaml:"abi"`
	// (Optional) by default the call job will "store" the return from the contract as the
	// result of the job.
	Save string `mapstructure:"tx-return" yaml:"tx-return"`
}

// Note: save jobs_output.json as jobs_output_<chain_ID>.json and concatenate outputs if chainID is the same
// If not, save it to a different file.

func (call *Call) PreProcess(jobs *Jobs) (err error) {
	call.Source, _, err = preProcessString(call.Source, jobs)
	if err != nil {
		return err
	}
	call.Destination, _, err = preProcessString(call.Destination, jobs)
	if err != nil {
		return err
	}
	call.Function, _, err = preProcessString(call.Function, jobs)
	if err != nil {
		return err
	}

	for i, data := range call.Data {
		if call.Data[i], err = preProcessInterface(data, jobs); err != nil {
			return err
		}
	}

	if call.ABI != "" {
		buf, err := ioutil.ReadFile(filepath.Join(jobs.AbiPath, call.ABI))
		if err != nil {
			return err
		}
		call.ABI = string(buf)
	}

	call.Amount, _, err = preProcessString(call.Amount, jobs)
	if err != nil {
		return err
	}
	call.Amount = useDefault(call.Amount, jobs.DefaultAmount)
	call.Fee, _, err = preProcessString(call.Fee, jobs)
	if err != nil {
		return err
	}
	call.Fee = useDefault(call.Fee, jobs.DefaultFee)
	call.Gas, _, err = preProcessString(call.Gas, jobs)
	if err != nil {
		return err
	}
	call.Gas = useDefault(call.Gas, jobs.DefaultGas)

	call.Nonce, _, err = preProcessString(call.Nonce, jobs)
	if err != nil {
		return err
	}
	call.Save, _, err = preProcessString(call.Save, jobs)
	if err != nil {
		return err
	}
	return nil
}

func (call *Call) Execute(jobs *Jobs) (*JobResults, error) {

	var namedResults map[string]Type
	var abiSource string
	if abi, ok := jobs.AbiMap[call.Destination]; !ok {
		if call.ABI == "" {
			return &JobResults{}, fmt.Errorf("Couldn't get the needed abi from your job results, can you provide it through the abi field in your call job?")
		} else {
			abiSource = call.ABI
		}
	} else if call.ABI != "" {
		abiSource = call.ABI
	} else {
		abiSource = abi
	}

	contractAbi, err := abi.MakeAbi(abiSource)
	if err != nil {
		return &JobResults{}, err
	}

	if call.Function == "()" {
		log.Warn("Calling the fallback function")
	}
	// format data
	callData, err := abi.FormatAndPackInputs(contractAbi, call.Function, call.Data)
	if err != nil {
		if call.Function == "()" {
			log.Warn("Calling the fallback function")
		} else {
			return &JobResults{}, err
		}
	}

	// create call
	tx, err := rpc.Call(jobs.NodeClient, jobs.KeyClient, jobs.PublicKey, call.Source, call.Destination, call.Amount, call.Nonce, call.Gas, call.Fee, hex.EncodeToString(callData))
	if err != nil {
		return &JobResults{}, err
	}
	result, err := txFinalize(tx, jobs, Return)
	if err != nil {
		return &JobResults{}, err
	}

	toUnpackInto, method, err := abi.CreateBlankSlate(contractAbi, call.Function)
	if err != nil {
		return &JobResults{}, err
	}
	err = contractAbi.Unpack(&toUnpackInto, call.Function, result.FullResult.ActualResult.([]byte))
	if err != nil {
		return &JobResults{}, err
	}
	// get names of the types, get string results, get actual results, return them.
	fullStringResults := []string{"("}
	for i, methodOutput := range method.Outputs {
		strResult, actualResult, err := abi.ConvertUnpackedToJobTypes(toUnpackInto[i], methodOutput.Type)
		if err != nil {
			return &JobResults{}, err
		}
		fullStringResults = append(fullStringResults, strResult+", ")
		if methodOutput.Name == "" {
			methodOutput.Name = strconv.FormatInt(int64(i), 10)
		}
		namedResults[methodOutput.Name] = Type{ActualResult: actualResult, StringResult: strResult}
	}
	fullStringResults = append(fullStringResults, ")")
	return &JobResults{FullResult: Type{StringResult: strings.Join(fullStringResults, ""), ActualResult: strings.Join(fullStringResults, "")}, NamedResults: namedResults}, nil
}
