package jobs

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/monax/cli/compilers"
	"github.com/monax/cli/log"
)

// ------------------------------------------------------------------------
// Contracts Jobs
// ------------------------------------------------------------------------

type PackageDeploy struct {
	// TODO
}

type Compile struct {
	// embedded interface to allow us to access the Compile function regardless of which language is chosen. This gets set in the preprocessing stage.
	compilers.Compiler
	// List of files with which you would like to compile with. Optional setting if creating a plugin compiler setting for a deploy job.
	Files []string `mapstructure:"files" yaml:"files"`
	// (Optional) the version of the compiler to use, if left blank, defaults to whatever the default is in compilers.toml, if that's blank, defaults to "latest" tag of compilers image
	Version string `mapstructure:"version" yaml:"version"`
	// One of the following fields is required.

	// Solidity Compiler
	Solc compilers.SolcTemplate `mapstructure:"solc" yaml:"solc"`
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
	case compilers.SolcTemplate:

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
	Data         []interface{} `mapstructure:"data" yaml:"data"`
	CompilerStub interface{}   `mapstructure:"compiler" yaml:"compiler"`
	// (Optional) amount of tokens to send to the contract which will (after deployment) reside in the
	// contract's account
	Amount string `mapstructure:"amount" yaml:"amount"`
	// (Optional) validators' fee
	Fee string `mapstructure:"fee" yaml:"fee"`
	// (Optional) amount of gas which should be sent along with the contract deployment transaction
	Gas string `mapstructure:"gas" yaml:"gas"`
	// (Optional) after compiling the contract save the binary in filename.bin in same directory
	// where the *.sol or *.se file is located. This will speed up subsequent installs
	SaveBinary bool `mapstructure:"save" yaml:"save"`
	// (Optional, advanced only) nonce to use when monax-keys signs the transaction (do not use unless you
	// know what you're doing)
	Nonce          string `mapstructure:"nonce" yaml:"nonce"`
	DeployBinary   bool
	DeployBinSuite bool
	Compiler       *Compile
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
	deploy.Instance, _, err = preProcessString(deploy.Instance, jobs)
	if err != nil {
		return err
	}

	for i, data := range deploy.Data {
		if deploy.Data[i], err = preProcessInterface(data, jobs); err != nil {
			return err
		}
	}

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

	// skip compile jobs if the contract we're attempting to deploy resides in the form of a binary,
	// if the special bin folder is specified then we
	// otherwise go and figure out the compilation either via a compile job or spin one up if one isn't specified
	if deploy.Contract == jobs.BinPath {
		deploy.DeployBinSuite = true
	} else if filepath.Ext(deploy.Contract) != ".bin" {
		deploy.DeployBinary = true
	} else {
		if deploy.CompilerStub != nil {
			compiler, err := preProcessPluginJob(deploy.CompilerStub, jobs)
			if err != nil {
				return err
			}
			switch compiler := compiler.(type) {
			case *Compile:
				compiler.Files = append(compiler.Files, deploy.Contract)
				switch compilerType := compiler.Compiler.(type) {
				case compilers.SolcTemplate:
					compilerType.Libraries = append(compilerType.Libraries, deploy.Libraries...)
					compiler.Compiler = compilerType
				default:
					return fmt.Errorf("Could not find compiler to use")
				}
				deploy.Compiler = compiler
			default:
				return fmt.Errorf("Invalid preprocessing of compiler")
			}

		} else {
			compiler := &Compile{}
			compiler.Compiler = compilers.SolcTemplate{
				CombinedOutput: []string{"bin", "abi"},
				Libraries:      deploy.Libraries,
			}
			compiler.Files = []string{deploy.Contract}
			compiler.Version = "stable"
			deploy.Compiler = compiler
		}
	}

	return nil
}

func (deploy *Deploy) Execute(jobs *Jobs) (*JobResults, error) {
	switch deploy.Compiler.Compiler.(type) {
	case compilers.SolcTemplate:
		results, err := deploy.Compiler.Execute(jobs)
		if err != nil {
			return &JobResults{}, fmt.Errorf("Compiler error: %v", err)
		}

		if deploy.Instance == "all" || deploy.Instance == "" {
			log.Info("Deploying all contracts")
			for name, contract := range results.NamedResults {
				log.Warn("Deploying contract: ", name, contract)

			}
		} else {
			if object, ok := results.NamedResults[deploy.Instance]; ok {
				log.Warn(object)
			} else {
				return &JobResults{}, fmt.Errorf("Could not acquire requested instance named %v", deploy.Instance)
			}
		}
		return &JobResults{}, fmt.Errorf("placeholder...to be gotten rid of")

	default:
		return &JobResults{}, fmt.Errorf("Invalid compiler used in execution process")
	}
}

type Call struct {
	// (Optional, if account job or global account set) address of the account from which to send (the
	// public key for the account must be available to monax-keys)
	Source string `mapstructure:"source" yaml:"source"`
	// (Required) address of the contract which should be called
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
	// (Optional) location of the abi file to use (can be relative path or in abi path)
	// deployed contracts save ABI artifacts in the abi folder as *both* the name of the contract
	// and the address where the contract was deployed to
	ABI string `mapstructure:"abi" yaml:"abi"`
	// (Optional) by default the call job will "store" the return from the contract as the
	// result of the job. If you would like to store the transaction hash instead of the
	// return from the call job as the result of the call job then select "tx" on the save
	// variable. Anything other than "tx" in this field will use the default.
	Save string `mapstructure:"save" yaml:"save"`
}

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
