package jobs

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/pkgs/abi"
	"github.com/monax/monax/util"

	compilers "github.com/monax/compilers/perform"

	"github.com/hyperledger/burrow/client"
	"github.com/hyperledger/burrow/client/rpc"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/txs"
)

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
		if _, secErr := os.Stat(filepath.Join(do.BinPath, deploy.Contract)); secErr != nil {
			return "", fmt.Errorf("Could not find contract in %v or in binary path %v", deploy.Contract, do.BinPath)
		}
	}

	// Don't use pubKey if account override
	var oldKey string
	if deploy.Source != do.Package.Account {
		oldKey = do.PublicKey
		do.PublicKey = ""
	}

	// compile
	if filepath.Ext(deploy.Contract) == ".bin" {
		contractPath = filepath.Join(do.BinPath, deploy.Contract)
		log.Info("Binary file detected. Using binary deploy sequence.")
		log.WithField("=>", contractPath).Info("Binary path")
		binaryResponse, err := compilers.RequestBinaryLinkage(do.Compiler+"/binaries", contractPath, deploy.Libraries)
		if err != nil {
			return "", fmt.Errorf("Something went wrong with your binary deployment: %v", err)
		}
		if binaryResponse.Error != "" {
			return "", fmt.Errorf("Something went wrong when you were trying to link your binaries: %v", binaryResponse.Error)
		}
		contractCode := binaryResponse.Binary

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
		contractPath = deploy.Contract
		log.WithField("=>", contractPath).Info("Contract path")
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
			log.WithField("=>", response.ABI).Info("Abi")
			log.WithField("=>", response.Bytecode).Info("Bin")
			if response.Bytecode != "" {
				result, err = deployContract(deploy, do, response)
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
				result, err = deployContract(deploy, do, response)
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
					result, err = deployContract(deploy, do, response)
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
func deployContract(deploy *definitions.Deploy, do *definitions.Do, compilersResponse compilers.ResponseItem) (string, error) {
	log.WithField("=>", string(compilersResponse.ABI)).Debug("ABI Specification (From Compilers)")
	contractCode := compilersResponse.Bytecode

	// Save ABI
	if _, err := os.Stat(do.ABIPath); os.IsNotExist(err) {
		if err := os.Mkdir(do.ABIPath, 0775); err != nil {
			return "", err
		}
	}
	if _, err := os.Stat(do.BinPath); os.IsNotExist(err) {
		if err := os.Mkdir(do.BinPath, 0775); err != nil {
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
		return "", fmt.Errorf("Error finalizing contract deploy %s: %v", deploy.Contract, err)
	}

	// saving contract/library abi at abi/address
	if result != "" {
		abiLocation := filepath.Join(do.ABIPath, result)
		log.WithField("=>", abiLocation).Debug("Saving ABI")
		if err := ioutil.WriteFile(abiLocation, []byte(compilersResponse.ABI), 0664); err != nil {
			return "", err
		}
		// saving binary
		if deploy.SaveBinary {
			contractName := filepath.Join(do.BinPath, fmt.Sprintf("%s.bin", compilersResponse.Objectname))
			log.WithField("=>", contractName).Warn("Saving Binary")
			if err := ioutil.WriteFile(contractName, []byte(contractCode), 0664); err != nil {
				return "", err
			}
		} else {
			log.Debug("Not saving binary.")
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
		"source":    deploy.Source,
		"code":      contractCode,
		"chain-url": do.ChainURL,
	}).Info()

	monaxNodeClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
	monaxKeyClient := keys.NewBurrowKeyClient(do.Signer, loggers.NewNoopInfoTraceLogger())
	tx, err := rpc.Call(monaxNodeClient, monaxKeyClient, do.PublicKey, deploy.Source, "", deploy.Amount, deploy.Nonce, deploy.Gas, deploy.Fee, contractCode)
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

	monaxNodeClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
	monaxKeyClient := keys.NewBurrowKeyClient(do.Signer, loggers.NewNoopInfoTraceLogger())
	tx, err := rpc.Call(monaxNodeClient, monaxKeyClient, do.PublicKey, call.Source, call.Destination, call.Amount, call.Nonce, call.Gas, call.Fee, callData)
	if err != nil {
		return "", make([]*definitions.Variable, 0), err
	}

	// Don't use pubKey if account override
	if call.Source != do.Package.Account {
		do.PublicKey = oldKey
	}

	// Sign, broadcast, display

	res, err := rpc.SignAndBroadcast(do.ChainID, monaxNodeClient, monaxKeyClient, tx, true, true, true)
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

	monaxNodeClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
	monaxKeyClient := keys.NewBurrowKeyClient(do.Signer, loggers.NewNoopInfoTraceLogger())
	res, err := rpc.SignAndBroadcast(do.ChainID, monaxNodeClient, monaxKeyClient, tx.(txs.Tx), true, true, true)
	if err != nil {
		return util.MintChainErrorHandler(do, err)
	}

	if err := util.ReadTxSignAndBroadcast(res, err); err != nil {
		return "", err
	}

	result = fmt.Sprintf("%X", res.Address)
	return result, nil
}
