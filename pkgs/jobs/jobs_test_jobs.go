package jobs

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/pkgs/abi"
	"github.com/monax/monax/util"

	"github.com/hyperledger/burrow/client"
	"github.com/hyperledger/burrow/logging/loggers"
)

func QueryContractJob(query *definitions.QueryContract, do *definitions.Do) (string, []*definitions.Variable, error) {
	// Preprocess variables. We don't preprocess data as it is processed by ReadAbiFormulateCall
	query.Source, _ = util.PreProcess(query.Source, do)
	query.Destination, _ = util.PreProcess(query.Destination, do)
	query.ABI, _ = util.PreProcess(query.ABI, do)

	var queryDataArray []string
	var err error
	query.Function, queryDataArray, err = util.PreProcessInputData(query.Function, query.Data, do, false)
	if err != nil {
		return "", make([]*definitions.Variable, 0), err
	}
	// Set the from and the to
	fromAddrBytes, err := hex.DecodeString(query.Source)
	if err != nil {
		return "", make([]*definitions.Variable, 0), err
	}
	toAddrBytes, err := hex.DecodeString(query.Destination)
	if err != nil {
		return "", make([]*definitions.Variable, 0), err
	}

	// Get the packed data from the ABI functions
	var data string
	var packedBytes []byte
	if query.ABI == "" {
		packedBytes, err = abi.ReadAbiFormulateCall(query.Destination, query.Function, queryDataArray, do)
		data = hex.EncodeToString(packedBytes)
	} else {
		packedBytes, err = abi.ReadAbiFormulateCall(query.ABI, query.Function, queryDataArray, do)
		data = hex.EncodeToString(packedBytes)
	}
	if err != nil {
		var str, err = util.ABIErrorHandler(do, err, nil, query)
		return str, make([]*definitions.Variable, 0), err
	}
	dataBytes, err := hex.DecodeString(data)
	if err != nil {
		return "", make([]*definitions.Variable, 0), err
	}

	// Call the client
	nodeClient := client.NewBurrowNodeClient(do.ChainURL, loggers.NewNoopInfoTraceLogger())
	result, _, err := nodeClient.QueryContract(fromAddrBytes, toAddrBytes, dataBytes)
	if err != nil {
		return "", make([]*definitions.Variable, 0), err
	}

	// Formally process the return
	log.WithField("res", result).Debug("Decoding Raw Result")
	if query.ABI == "" {
		log.WithField("abi", query.Destination).Debug()
		query.Variables, err = abi.ReadAndDecodeContractReturn(query.Destination, query.Function, result, do)
	} else {
		log.WithField("abi", query.ABI).Debug()
		query.Variables, err = abi.ReadAndDecodeContractReturn(query.ABI, query.Function, result, do)
	}
	if err != nil {
		return "", make([]*definitions.Variable, 0), err
	}

	result2 := util.GetReturnValue(query.Variables)
	// Finalize
	if result2 != "" {
		log.WithField("=>", result2).Warn("Return Value")
	} else {
		log.Debug("No return.")
	}
	return result2, query.Variables, nil
}

func QueryAccountJob(query *definitions.QueryAccount, do *definitions.Do) (string, error) {
	// Preprocess variables
	query.Account, _ = util.PreProcess(query.Account, do)
	query.Field, _ = util.PreProcess(query.Field, do)

	// Perform Query
	arg := fmt.Sprintf("%s:%s", query.Account, query.Field)
	log.WithField("=>", arg).Info("Querying Account")

	result, err := util.AccountsInfo(query.Account, query.Field, do)
	if err != nil {
		return "", err
	}

	// Result
	if result != "" {
		log.WithField("=>", result).Warn("Return Value")
	} else {
		log.Debug("No return.")
	}
	return result, nil
}

func QueryNameJob(query *definitions.QueryName, do *definitions.Do) (string, error) {
	// Preprocess variables
	query.Name, _ = util.PreProcess(query.Name, do)
	query.Field, _ = util.PreProcess(query.Field, do)

	// Peform query
	log.WithFields(log.Fields{
		"name":  query.Name,
		"field": query.Field,
	}).Info("Querying")
	result, err := util.NamesInfo(query.Name, query.Field, do)
	if err != nil {
		return "", err
	}

	if result != "" {
		log.WithField("=>", result).Warn("Return Value")
	} else {
		log.Debug("No return.")
	}
	return result, nil
}

func QueryValsJob(query *definitions.QueryVals, do *definitions.Do) (string, error) {
	var result string

	// Preprocess variables
	query.Field, _ = util.PreProcess(query.Field, do)

	// Peform query
	log.WithField("=>", query.Field).Info("Querying Vals")
	result, err := util.ValidatorsInfo(query.Field, do)
	if err != nil {
		return "", err
	}

	if result != "" {
		log.WithField("=>", result).Warn("Return Value")
	} else {
		log.Debug("No return.")
	}
	return result, nil
}

func AssertJob(assertion *definitions.Assert, do *definitions.Do) (string, error) {
	var result string
	// Preprocess variables
	assertion.Key, _ = util.PreProcess(assertion.Key, do)
	assertion.Relation, _ = util.PreProcess(assertion.Relation, do)
	assertion.Value, _ = util.PreProcess(assertion.Value, do)

	// Switch on relation
	log.WithFields(log.Fields{
		"key":      assertion.Key,
		"relation": assertion.Relation,
		"value":    assertion.Value,
	}).Info("Assertion =>")

	switch assertion.Relation {
	case "==", "eq":
		/*log.Debug("Compare", strings.Compare(assertion.Key, assertion.Value))
		log.Debug("UTF8?: ", utf8.ValidString(assertion.Key))
		log.Debug("UTF8?: ", utf8.ValidString(assertion.Value))
		log.Debug("UTF8?: ", utf8.RuneCountInString(assertion.Key))
		log.Debug("UTF8?: ", utf8.RuneCountInString(assertion.Value))*/
		if assertion.Key == assertion.Value {
			return assertPass("==", assertion.Key, assertion.Value)
		} else {
			return assertFail("==", assertion.Key, assertion.Value)
		}
	case "!=", "ne":
		if assertion.Key != assertion.Value {
			return assertPass("!=", assertion.Key, assertion.Value)
		} else {
			return assertFail("!=", assertion.Key, assertion.Value)
		}
	case ">", "gt":
		k, v, err := bulkConvert(assertion.Key, assertion.Value)
		if err != nil {
			return convFail()
		}
		if k > v {
			return assertPass(">", assertion.Key, assertion.Value)
		} else {
			return assertFail(">", assertion.Key, assertion.Value)
		}
	case ">=", "ge":
		k, v, err := bulkConvert(assertion.Key, assertion.Value)
		if err != nil {
			return convFail()
		}
		if k >= v {
			return assertPass(">=", assertion.Key, assertion.Value)
		} else {
			return assertFail(">=", assertion.Key, assertion.Value)
		}
	case "<", "lt":
		k, v, err := bulkConvert(assertion.Key, assertion.Value)
		if err != nil {
			return convFail()
		}
		if k < v {
			return assertPass("<", assertion.Key, assertion.Value)
		} else {
			return assertFail("<", assertion.Key, assertion.Value)
		}
	case "<=", "le":
		k, v, err := bulkConvert(assertion.Key, assertion.Value)
		if err != nil {
			return convFail()
		}
		if k <= v {
			return assertPass("<=", assertion.Key, assertion.Value)
		} else {
			return assertFail("<=", assertion.Key, assertion.Value)
		}
	default:
		return "", fmt.Errorf("Error: Bad assert relation: \"%s\" is not a valid relation. See documentation for more information.", assertion.Relation)
	}

	return result, nil
}

func bulkConvert(key, value string) (int, int, error) {
	k, err := strconv.Atoi(key)
	if err != nil {
		return 0, 0, err
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, 0, err
	}
	return k, v, nil
}

func assertPass(typ, key, val string) (string, error) {
	log.WithField("=>", fmt.Sprintf("%s %s %s", key, typ, val)).Warn("Assertion Succeeded")
	return "passed", nil
}

func assertFail(typ, key, val string) (string, error) {
	log.WithField("=>", fmt.Sprintf("%s %s %s", key, typ, val)).Warn("Assertion Failed")
	return "failed", fmt.Errorf("assertion failed")
}

func convFail() (string, error) {
	return "", fmt.Errorf("The Key of your assertion cannot be converted into an integer.\nFor string conversions please use the equal or not equal relations.")
}
