package jobs

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/monax/cli/log"
)

// ------------------------------------------------------------------------
// Testing Jobs
// ------------------------------------------------------------------------

// aka. Simulated Call.
type QueryContract struct {
	// (Optional, if account job or global account set) address of the account from which to send (the
	// public key for the account must be available to monax-keys)
	Source string `mapstructure:"source" yaml:"source"`
	// (Required) address of the contract which should be called
	Destination string `mapstructure:"destination" yaml:"destination"`
	// (Required) data which should be called. will use the abi tooling under the hood to formalize the
	// transaction. QueryContract will usually be used with "accessor" functions in contracts
	Function string `mapstructure:"function" yaml:"function"`
	// (Optional) data to be used in the function arguments. Will use the abi tooling under the hood to formalize the
	// transaction.
	Data interface{} `mapstructure:"data" yaml:"data"`
	// (Optional) location of the abi file to use (can be relative path or in abi path)
	// deployed contracts save ABI artifacts in the abi folder as *both* the name of the contract
	// and the address where the contract was deployed to
	ABI string `mapstructure:"abi" yaml:"abi"`
}

func (qContract *QueryContract) PreProcess(jobs *Jobs) (err error) {
	qContract.Source, _, err = preProcessString(qContract.Source, jobs)
	if err != nil {
		return err
	}
	qContract.Destination, _, err = preProcessString(qContract.Destination, jobs)
	if err != nil {
		return err
	}
	qContract.Function, _, err = preProcessString(qContract.Function, jobs)
	if err != nil {
		return err
	}
	qContract.Data, err = preProcessInterface(qContract.Data, jobs)
	if err != nil {
		return err
	}
	return
}

type QueryAccount struct {
	// (Required) address of the account which should be queried
	Account string `mapstructure:"account" yaml:"account"`
	// (Required) field which should be queried. If users are trying to query the permissions of the
	// account one can get either the `permissions.base` which will return the base permission of the
	// account, or one can get the `permissions.set` which will return the setBit of the account.
	Field string `mapstructure:"field" yaml:"field"`
}

func (qAccount *QueryAccount) PreProcess(jobs *Jobs) (err error) {
	qAccount.Field, _, err = preProcessString(qAccount.Field, jobs)
	if err != nil {
		return err
	}
	qAccount.Account, _, err = preProcessString(qAccount.Account, jobs)
	if err != nil {
		return err
	}
	return
}

func (qAccount *QueryAccount) Execute(jobs *Jobs) (*JobResults, error) {
	addrBytes, err := hex.DecodeString(qAccount.Account)
	if err != nil {
		return nil, fmt.Errorf("Account Addr %s is improper hex: %v", qAccount.Account, err)
	}

	if r, err := jobs.NodeClient.GetAccount(addrBytes); err != nil {
		return nil, err
	} else if r == nil {
		return nil, fmt.Errorf("Account %s does not exist", qAccount.Account)
	} else {
		var result Type
		invalidAccount := fmt.Errorf("Invalid account field queried %v", qAccount.Field)

		switch {
		case strings.Contains(qAccount.Field, "permissions"):
			permissions := strings.Split(qAccount.Field, ".")
			if len(permissions) > 2 {
				return nil, invalidAccount
			}
			switch permissions[1] {
			case "roles":
				//[rj] the second field is cheating...for now...need to update asseerts to handle arrays of strings, or create a helper to turn this into an interface
				result = Type{strings.Join(r.Permissions.Roles, ","), strings.Join(r.Permissions.Roles, ",")}
			case "base", "perms":
				result = Type{strconv.Itoa(int(r.Permissions.Base.Perms)), r.Permissions.Base.Perms}
			case "set":
				result = Type{strconv.Itoa(int(r.Permissions.Base.SetBit)), r.Permissions.Base.SetBit}
			default:
				return nil, invalidAccount
			}
		case qAccount.Field == "balance":
			result = Type{strconv.Itoa(int(r.Balance)), r.Balance}
		default:
			return nil, invalidAccount
		}

		return &JobResults{result, nil}, nil
	}
}

type QueryName struct {
	// (Required) name which should be queried
	Name string `mapstructure:"name" yaml:"name"`
	// (Required) field which should be queried (generally will be "data" to get the registered "name")
	Field string `mapstructure:"field" yaml:"field"`
}

func (qName *QueryName) PreProcess(jobs *Jobs) (err error) {
	qName.Field, _, err = preProcessString(qName.Field, jobs)
	if err != nil {
		return err
	}
	qName.Name, _, err = preProcessString(qName.Name, jobs)
	if err != nil {
		return err
	}
	return
}

func (qName *QueryName) Execute(jobs *Jobs) (*JobResults, error) {
	owner, data, expirationBlock, err := jobs.NodeClient.GetName(qName.Name)
	if err != nil {
		return nil, err
	}
	var result Type
	switch qName.Field {
	case "name":
		result = Type{qName.Name, qName.Name}
	case "owner":
		result = Type{string(owner), owner}
	case "data":
		result = Type{data, data}
	case "expires":
		result = Type{strconv.Itoa(expirationBlock), expirationBlock}
	default:
		return nil, fmt.Errorf("Field %s not recognized", qName.Field)
	}
	return &JobResults{result, nil}, nil
}

type QueryVals struct {
	// (Required) should be of the set ["bonded_validators" or "unbonding_validators"] and it will
	// return a comma separated listing of the addresses which fall into one of those categories
	Field string `mapstructure:"field" yaml:"field"`
}

func (qVals *QueryVals) PreProcess(jobs *Jobs) (err error) {
	qVals.Field, _, err = preProcessString(qVals.Field, jobs)
	if err != nil {
		return err
	}
	if qVals.Field != "bonded_validators" && qVals.Field != "unbonded_validators" {
		return fmt.Errorf("Invalid value passed in, expected bonded_validators or unbonded_validators, got %v", qVals.Field)
	}
	return
}

func (qVals *QueryVals) Execute(jobs *Jobs) (*JobResults, error) {
	// Peform query
	log.WithField("=>", qVals.Field).Info("Querying Vals")
	_, bondedValidators, unbondingValidators, err := jobs.NodeClient.ListValidators()
	if err != nil {
		return nil, err
	}

	vals := []string{}
	switch qVals.Field {
	case "bonded_validators":
		for _, v := range bondedValidators {
			vals = append(vals, string(v.Address()))
		}
	case "unbonding_validators":
		for _, v := range unbondingValidators {
			vals = append(vals, string(v.Address()))
		}
	default:
		return nil, fmt.Errorf("Field %s not recognized", qVals.Field)
	}

	return &JobResults{
		FullResult:   Type{strings.Join(vals, ","), vals},
		NamedResults: nil,
	}, nil
}

type Assert struct {
	// (Required) key which should be used for the assertion. This is usually known as the "expected"
	// value in most testing suites
	Key interface{} `mapstructure:"key" yaml:"key"`
	// (Required) must be of the set ["eq", "ne", "ge", "gt", "le", "lt", "==", "!=", ">=", ">", "<=", "<"]
	// establishes the relation to be tested by the assertion. If a strings key:value pair is being used
	// only the equals or not-equals relations may be used as the key:value will try to be converted to
	// ints for the remainder of the relations. if strings are passed to them then the job runner will return an
	// error
	Relation string `mapstructure:"relation" yaml:"relation"`
	// (Required) value which should be used for the assertion. This is usually known as the "given"
	// value in most testing suites. Generally it will be a variable expansion from one of the query
	// jobs.
	Value interface{} `mapstructure:"val" yaml:"val"`
}

func (assert *Assert) PreProcess(jobs *Jobs) (err error) {
	failed := "Assertion Failed!: " //useful appendage for errors

	convertString := func(toChange string, reference interface{}) (interface{}, error) {
		//go func for converting underlying string types to the reference type
		//something to note: if strconv fails, the strconv error will not be reported
		//due to it being handled with the invalid conversion error. If something is fishy in failing,
		//please do try to get this error to return.
		switch reference.(type) {
		case bool:
			return strconv.ParseBool(toChange)
		case int:
			if strings.HasPrefix(toChange, "0x") {
				return strconv.ParseInt(toChange[2:], 16, 0)
			}
			return strconv.Atoi(toChange)
		case []interface{}:
			var changed []interface{}
			err = json.NewDecoder(strings.NewReader(toChange)).Decode(&changed)
			return changed, err
		default:
			return nil, fmt.Errorf(failed+"Do not have conversion for string %v to type %T", toChange, reference)
		}
	}

	// initial preprocess
	if assert.Key, err = preProcessInterface(assert.Key, jobs); err != nil {
		return fmt.Errorf("Assertion Failed!: %v", err)
	}
	if assert.Relation, _, err = preProcessString(assert.Relation, jobs); err != nil {
		return fmt.Errorf("Assertion Failed!: %v", err)
	}
	if assert.Value, err = preProcessInterface(assert.Value, jobs); err != nil {
		return fmt.Errorf("Assertion Failed!: %v", err)
	}

	// catch invalid types early
	keyString := assert.Key.(Type).StringResult
	valString := assert.Value.(Type).StringResult

	keyString = strings.Trim(keyString, " \n\t")
	valString = strings.Trim(valString, " \n\t")

	keyType := assert.Key.(Type).ActualResult
	valType := assert.Value.(Type).ActualResult

	// second round of preprocessing
	switch assert.Relation {
	case ">", "gt", ">=", "ge", "<", "lt", "<=", "le":
		switch keyType.(type) {
		default:
			return fmt.Errorf(failed+"Cannot use key type %T in ordered comparison.", keyType)
		case string:
			keyType, err = strconv.Atoi(keyType.(string))
			if err != nil {
				return err
			}
		case int, int64, uint, uint64:
			break
		}

		switch valType.(type) {
		default:
			return fmt.Errorf(failed+"Cannot use val type %T in ordered comparison.", valType)
		case string:
			valType, err = strconv.Atoi(valType.(string))
			if err != nil {
				return err
			}
		case int, int64, uint, uint64:
			break
		}

	case "==", "eq", "!=", "ne":
		invalidConversion := fmt.Errorf(failed+"Cannot convert key type %T to val type %T", keyType, valType)
		defer func() { //since reflect.Convert panicks if it can't be converted, we have this handler here
			if recover() != nil {
				err = invalidConversion
			}
		}()
		valReflectType := reflect.TypeOf(valType)
		keyReflectType := reflect.TypeOf(keyType)
		if valReflectType != keyReflectType {

			switch valType.(type) {
			case []interface{}:
				switch keyType.(type) {
				case string:
					keyType, err = convertString(keyType.(string), valType)
				case bool, int:
					return invalidConversion
				default:
					keyVal := reflect.ValueOf(keyType)
					keyType = keyVal.Convert(valReflectType)
				}
			case int:
				switch keyType.(type) {
				case string:
					keyType, err = convertString(keyType.(string), valType)
				case bool:
					if valType.(int) != 0 && valType.(int) != 1 {
						return invalidConversion
					}
					valType = valType != 0
				default:
					keyVal := reflect.ValueOf(keyType)
					keyType = keyVal.Convert(valReflectType)
				}
			case bool:
				switch keyType.(type) {
				case int:
					if keyType.(int) != 0 && keyType.(int) != 1 {
						return invalidConversion
					}
					keyType = keyType != 0
				case []byte:
					buf := bytes.NewBuffer(keyType.([]byte))
					if keyType, err = binary.ReadVarint(buf); err != nil {
						return err
					} else if keyType.(int) != 0 && keyType.(int) != 1 {
						return invalidConversion
					}
					keyType = keyType != 0
				case string:
					keyType, err = convertString(keyType.(string), valType)
				default:
					return invalidConversion
				}
			case []byte:
				switch keyType.(type) {
				case string:
					valType = string(valType.([]byte))
				case bool:
					valType = valType.([]byte)[len(valType.([]byte))-1] != 0
				default:
					return invalidConversion
				}
			case string:
				switch keyType.(type) {
				case int, bool, []interface{}:
					valType, err = convertString(valType.(string), keyType)
				default:
					keyVal := reflect.ValueOf(keyType)
					keyType = keyVal.Convert(valReflectType)
				}
			default:
				valVal := reflect.ValueOf(valType)
				valType = valVal.Convert(keyReflectType)
			}
			if err != nil {
				return invalidConversion
			}
		}
	default:
		return fmt.Errorf(failed+"Invalid assertion relation %v", assert.Relation)
	}
	assert.Key = Type{StringResult: keyString, ActualResult: keyType}
	assert.Value = Type{StringResult: valString, ActualResult: valType}
	return err
}

func (assert *Assert) Execute(jobs *Jobs) (*JobResults, error) {
	// first class functions for passing and failing
	pass := func() (*JobResults, error) {
		log.Warn("Assertion Passed!")
		return &JobResults{}, nil
	}
	fail := func() (*JobResults, error) {
		return &JobResults{}, fmt.Errorf("Assertion Failed!")
	}
	isOrderedInt := func(i interface{}) (int, bool) {
		switch i.(type) {
		case int:
			return i.(int), true
		case int64:
			return int(i.(int64)), true
		default:
			return 0, false
		}
	}
	// Switch on relation
	stringKey := assert.Key.(Type).StringResult
	stringVal := assert.Value.(Type).StringResult
	typeKey := assert.Key.(Type).ActualResult
	typeVal := assert.Value.(Type).ActualResult

	log.WithFields(log.Fields{
		"key":      stringKey,
		"relation": assert.Relation,
		"value":    stringVal,
	}).Error("Assertion =>")

	keyVal := reflect.ValueOf(typeKey)
	valVal := reflect.ValueOf(typeVal)
	switch assert.Relation {
	case "==", "eq":
		switch typeKey.(type) {
		default:
			if stringKey == stringVal || keyVal.Interface() == valVal.Interface() {
				return pass()
			} else {
				return fail()
			}
		case []interface{}:
			if stringKey == stringVal || reflect.DeepEqual(typeKey, typeVal) {
				return pass()
			} else {
				return fail()
			}
		}

	case "!=", "ne":
		switch typeKey.(type) {
		default:
			if stringKey != stringVal || keyVal.Interface() != valVal.Interface() {
				return pass()
			} else {
				return fail()
			}
		case []interface{}:
			if stringKey != stringVal || !reflect.DeepEqual(typeKey, typeVal) {
				return pass()
			} else {
				return fail()
			}
		}

	case ">=", "ge", ">", "gt", "<", "lt", "<=", "le":
		if newTypeKey, ok := isOrderedInt(typeKey); ok {
			typeKey = newTypeKey
		} else {
			typeKey = typeKey.(string)
		}
		if newTypeVal, ok := isOrderedInt(typeKey); ok {
			typeVal = newTypeVal
		} else {
			typeVal = typeVal.(string)
		}
		switch assert.Relation {
		case ">=", "ge":
			if stringKey >= stringVal || typeKey.(int) >= typeVal.(int) {
				return pass()
			} else {
				return fail()
			}
		case "<", "lt":
			if stringKey < stringVal || typeKey.(int) < typeVal.(int) {
				return pass()
			} else {
				return fail()
			}
		case "<=", "le":
			if stringKey == stringVal || typeKey.(int) == typeVal.(int) {
				return pass()
			} else {
				return fail()
			}
		case ">", "gt":
			if stringKey > stringVal || typeKey.(int) > typeVal.(int) {
				return pass()
			} else {
				return fail()
			}
		default:
			return &JobResults{}, fmt.Errorf("The marmots are very confused as to how you got here.")
		}
	default:
		return &JobResults{}, fmt.Errorf("Improper relation detected.")
	}
}
