package jobs

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/pkgs/abi"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/util"

	compilers "github.com/eris-ltd/eris-compilers/perform"

	"github.com/eris-ltd/eris-db/client"
	"github.com/eris-ltd/eris-db/client/core"
	"github.com/eris-ltd/eris-db/keys"
	"github.com/eris-ltd/eris-db/txs"
)

func PackageDeployJob(pkgDeploy *definitions.PackageDeploy, do *definitions.Do) (string, error) {
	// todo
	// note: we should have this talking to a new client package...add that to the todo
	var result string

	return result, nil
}

func DeployJob(deploy *definitions.Deploy, do *definitions.Do) (result string, err error) {
	// Preprocess variables
	deploy.Source, _ = util.PreProcess(deploy.Source, do)
	deploy.Contract, _ = util.PreProcess(deploy.Contract, do)
	deploy.Instance, _ = util.PreProcess(deploy.Instance, do)
	deploy.Libraries, _ = util.PreProcessLibs(deploy.Libraries, do)
	deploy.Amount, _ = util.PreProcess(deploy.Amount, do)
	deploy.Nonce, _ = util.PreProcess(deploy.Nonce, do)
	deploy.Fee, _ = util.PreProcess(deploy.Fee, do)
	deploy.Gas, _ = util.PreProcess(deploy.Gas, do)

	// trim the extension
	contractName := strings.TrimSuffix(deploy.Contract, filepath.Ext(deploy.Contract))

	// Use defaults
	deploy.Source = useDefault(deploy.Source, do.Package.Account)
	deploy.Instance = useDefault(deploy.Instance, contractName)
	deploy.Amount = useDefault(deploy.Amount, do.DefaultAmount)
	deploy.Fee = useDefault(deploy.Fee, do.DefaultFee)
	deploy.Gas = useDefault(deploy.Gas, do.DefaultGas)

	// assemble contract
	var contractPath string
	if _, err := os.Stat(deploy.Contract); err != nil {
		return "", err
	}
	contractPath = deploy.Contract
	log.WithField("=>", contractPath).Info("Contract path")

	// use the proper compiler
	if do.Compiler != "" {
		log.WithField("=>", do.Compiler).Info("Setting compiler path")
	}

	// Don't use pubKey if account override
	var oldKey string
	if deploy.Source != do.Package.Account {
		oldKey = do.PublicKey
		do.PublicKey = ""
	}

	// compile
	if filepath.Ext(deploy.Contract) == ".bin" {
		log.Info("Binary file detected. Using binary deploy sequence.")
		// binary deploy sequence
		contractCode, err := ioutil.ReadFile(contractPath)
		if err != nil {
			return "could not read binary file", err
		}
		tx, err := deployRaw(do, deploy, contractName, string(contractCode))
		if err != nil {
			return "could not deploy binary contract", err
		}
		result, err := deployFinalize(do, tx)
		if err != nil {
			return "", fmt.Errorf("Error finalizing contract deploy from path %s: %v", contractPath, err)
		}
		return result, err
	} else {
		// normal compilation/deploy sequence
		resp, err := compilers.RequestCompile(do.Compiler, contractPath, false, deploy.Libraries)

		if err != nil {
			log.Errorln("Error compiling contracts: Compilers error:")
			return "", err
		} else if resp.Error != "" {
			log.Errorln("Error compiling contracts: Language error:")
			return "", fmt.Errorf("%v", resp.Error)
		} else if resp.Warning != "" {
			log.WithField("Warning", resp.Warning).Warn("Warning Generated during Contract Compilation")
		}
		// loop through objects returned from compiler
		switch {
		case len(resp.Objects) == 1:
			log.WithField("path", contractPath).Info("Deploying the only contract in file")
			response := resp.Objects[0]
			if response.Bytecode != "" {
				result, err = deployContract(deploy, do, response, contractPath)
				if err != nil {
					return "", err
				}
			}
		case deploy.Instance == "all":
			log.WithField("path", contractPath).Info("Deploying all contracts")
			var baseObj string
			for _, response := range resp.Objects {
				if response.Bytecode == "" {
					continue
				}
				result, err = deployContract(deploy, do, response, contractPath)
				if err != nil {
					return "", err
				}
				if strings.ToLower(response.Objectname) == strings.ToLower(strings.TrimSuffix(filepath.Base(deploy.Contract), filepath.Ext(filepath.Base(deploy.Contract)))) {
					baseObj = result
				}
			}
			if baseObj != "" {
				result = baseObj
			}
		default:
			log.WithField("contract", deploy.Instance).Info("Deploying a single contract")
			for _, response := range resp.Objects {
				if response.Bytecode == "" {
					continue
				}
				if strings.ToLower(response.Objectname) == strings.ToLower(deploy.Instance) {
					result, err = deployContract(deploy, do, response, contractPath)
					if err != nil {
						return "", err
					}
				}
			}
		}
	}

	// Don't use pubKey if account override
	if deploy.Source != do.Package.Account {
		do.PublicKey = oldKey
	}

	return result, nil
}

// TODO [rj] refactor to remove [contractPath] from functions signature => only used in a single error throw.
func deployContract(deploy *definitions.Deploy, do *definitions.Do, compilersResponse compilers.ResponseItem, contractPath string) (string, error) {
	log.WithField("=>", string(compilersResponse.ABI)).Debug("ABI Specification (From Compilers)")
	contractCode := compilersResponse.Bytecode

	// Save ABI
	if _, err := os.Stat(do.ABIPath); os.IsNotExist(err) {
		if err := os.Mkdir(do.ABIPath, 0775); err != nil {
			return "", err
		}
	}

	// saving contract/library abi
	var abiLocation string
	if compilersResponse.Objectname != "" {
		abiLocation = filepath.Join(do.ABIPath, compilersResponse.Objectname)
		log.WithField("=>", abiLocation).Warn("Saving ABI")
		if err := ioutil.WriteFile(abiLocation, []byte(compilersResponse.ABI), 0664); err != nil {
			return "", err
		}
	} else {
		log.Debug("Objectname from compilers is blank. Not saving abi.")
	}

	// saving binary
	if deploy.SaveBinary {
		contractDir := filepath.Dir(deploy.Contract)
		contractName := filepath.Join(contractDir, fmt.Sprintf("%s.bin", strings.TrimSuffix(deploy.Contract, filepath.Ext(deploy.Contract))))
		log.WithField("=>", contractName).Warn("Saving Binary")
		if err := ioutil.WriteFile(contractName, []byte(contractCode), 0664); err != nil {
			return "", err
		}
	} else {
		log.Debug("Not saving binary.")
	}

	// additional data may be sent along with the contract
	// these are naively added to the end of the contract code using standard
	// mint packing

	if deploy.Data != nil {
		_, callDataArray, err := util.PreProcessInputData(compilersResponse.Objectname, deploy.Data, do, true)
		if err != nil {
			return "", err
		}
		packedBytes, err := abi.ReadAbiFormulateCall(compilersResponse.Objectname, "", callDataArray, do)
		if err != nil {
			return "", err
		}
		callData := hex.EncodeToString(packedBytes)
		contractCode = contractCode + callData
	}

	tx, err := deployRaw(do, deploy, compilersResponse.Objectname, contractCode)
	if err != nil {
		return "", err
	}

	// Sign, broadcast, display
	result, err := deployFinalize(do, tx)
	if err != nil {
		return "", fmt.Errorf("Error finalizing contract deploy %s: %v", contractPath, err)
	}

	// saving contract/library abi at abi/address
	if result != "" {
		abiLocation := filepath.Join(do.ABIPath, result)
		log.WithField("=>", abiLocation).Debug("Saving ABI")
		if err := ioutil.WriteFile(abiLocation, []byte(compilersResponse.ABI), 0664); err != nil {
			return "", err
		}
	} else {
		// we shouldn't reach this point because we should have an error before this.
		log.Error("The contract did not deploy. Unable to save abi to abi/contractAddress.")
	}

	return result, err
}

func deployRaw(do *definitions.Do, deploy *definitions.Deploy, contractName, contractCode string) (*txs.CallTx, error) {

	// Deploy contract
	log.WithFields(log.Fields{
		"name": contractName,
	}).Warn("Deploying Contract")

	log.WithFields(log.Fields{
		"source": deploy.Source,
		"code":   contractCode,
	}).Info()

	erisNodeClient := client.NewErisNodeClient(do.ChainURL)
	erisKeyClient := keys.NewErisKeyClient(do.Signer)
	tx, err := core.Call(erisNodeClient, erisKeyClient, do.PublicKey, deploy.Source, "", deploy.Amount, deploy.Nonce, deploy.Gas, deploy.Fee, contractCode)
	if err != nil {
		return &txs.CallTx{}, fmt.Errorf("Error deploying contract %s: %v", contractName, err)
	}

	return tx, err
}

func CallJob(call *definitions.Call, do *definitions.Do) (string, []*definitions.Variable, error) {
	var err error
	var callData string
	var callDataArray []string
	// Preprocess variables
	call.Source, _ = util.PreProcess(call.Source, do)
	call.Destination, _ = util.PreProcess(call.Destination, do)
	//todo: find a way to call the fallback function here
	call.Function, callDataArray, err = util.PreProcessInputData(call.Function, call.Data, do, false)
	if err != nil {
		return "", make([]*definitions.Variable, 0), err
	}
	call.Function, _ = util.PreProcess(call.Function, do)
	call.Amount, _ = util.PreProcess(call.Amount, do)
	call.Nonce, _ = util.PreProcess(call.Nonce, do)
	call.Fee, _ = util.PreProcess(call.Fee, do)
	call.Gas, _ = util.PreProcess(call.Gas, do)
	call.ABI, _ = util.PreProcess(call.ABI, do)

	// Use default
	call.Source = useDefault(call.Source, do.Package.Account)
	call.Amount = useDefault(call.Amount, do.DefaultAmount)
	call.Fee = useDefault(call.Fee, do.DefaultFee)
	call.Gas = useDefault(call.Gas, do.DefaultGas)

	// formulate call
	var packedBytes []byte
	if call.ABI == "" {
		packedBytes, err = abi.ReadAbiFormulateCall(call.Destination, call.Function, callDataArray, do)
		callData = hex.EncodeToString(packedBytes)
	} else {
		packedBytes, err = abi.ReadAbiFormulateCall(call.ABI, call.Function, callDataArray, do)
		callData = hex.EncodeToString(packedBytes)
	}
	if err != nil {
		if call.Function == "()" {
			log.Warn("Calling the fallback function")
		} else {
			var str, err = util.ABIErrorHandler(do, err, call, nil)
			return str, make([]*definitions.Variable, 0), err
		}
	}

	// Don't use pubKey if account override
	var oldKey string
	if call.Source != do.Package.Account {
		oldKey = do.PublicKey
		do.PublicKey = ""
	}

	log.WithFields(log.Fields{
		"destination": call.Destination,
		"function":    call.Function,
		"data":        callData,
	}).Info("Calling")

	erisNodeClient := client.NewErisNodeClient(do.ChainURL)
	erisKeyClient := keys.NewErisKeyClient(do.Signer)
	tx, err := core.Call(erisNodeClient, erisKeyClient, do.PublicKey, call.Source, call.Destination, call.Amount, call.Nonce, call.Gas, call.Fee, callData)
	if err != nil {
		return "", make([]*definitions.Variable, 0), err
	}

	// Don't use pubKey if account override
	if call.Source != do.Package.Account {
		do.PublicKey = oldKey
	}

	// Sign, broadcast, display

	res, err := core.SignAndBroadcast(do.ChainID, erisNodeClient, erisKeyClient, tx, true, true, true)
	if err != nil {
		var str, err = util.MintChainErrorHandler(do, err)
		return str, make([]*definitions.Variable, 0), err
	}

	txResult := res.Return
	var result string
	log.Debug(txResult)

	// Formally process the return
	if txResult != nil {
		log.WithField("=>", result).Debug("Decoding Raw Result")
		if call.ABI == "" {
			call.Variables, err = abi.ReadAndDecodeContractReturn(call.Destination, call.Function, txResult, do)
		} else {
			call.Variables, err = abi.ReadAndDecodeContractReturn(call.ABI, call.Function, txResult, do)
		}
		if err != nil {
			return "", make([]*definitions.Variable, 0), err
		}
		log.WithField("=>", call.Variables).Debug("call variables:")
		result = util.GetReturnValue(call.Variables)
		if result != "" {
			log.WithField("=>", result).Warn("Return Value")
		} else {
			log.Debug("No return.")
		}
	} else {
		log.Debug("No return from contract.")
	}

	if call.Save == "tx" {
		log.Info("Saving tx hash instead of contract return")
		result = fmt.Sprintf("%X", res.Hash)
	}

	return result, call.Variables, nil
}

func deployFinalize(do *definitions.Do, tx interface{}) (string, error) {
	var result string

	erisNodeClient := client.NewErisNodeClient(do.ChainURL)
	erisKeyClient := keys.NewErisKeyClient(do.Signer)
	res, err := core.SignAndBroadcast(do.ChainID, erisNodeClient, erisKeyClient, tx.(txs.Tx), true, true, true)
	if err != nil {
		return util.MintChainErrorHandler(do, err)
	}

	if err := util.ReadTxSignAndBroadcast(res, err); err != nil {
		log.Error("ERROR =>")
		return "", err
	}

	result = fmt.Sprintf("%X", res.Address)
	return result, nil
}
