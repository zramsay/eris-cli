package abi

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/util"

	ethAbi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

func ReadAbiFormulateCall(abiLocation string, funcName string, args []string, do *definitions.Do) ([]byte, error) {
	abiSpecBytes, err := util.ReadAbi(do.ABIPath, abiLocation)
	if err != nil {
		return []byte{}, err
	}
	log.WithField("=>", string(abiSpecBytes)).Debug("ABI Specification (Formulate)")
	log.WithFields(log.Fields{
		"function":  funcName,
		"arguments": fmt.Sprintf("%v", args),
	}).Debug("Packing Call via ABI")

	return Packer(abiSpecBytes, funcName, args...)
}

func ReadAndDecodeContractReturn(abiLocation, funcName string, resultRaw []byte, do *definitions.Do) ([]*definitions.Variable, error) {
	abiSpecBytes, err := util.ReadAbi(do.ABIPath, abiLocation)
	if err != nil {
		return nil, err
	}
	log.WithField("=>", abiSpecBytes).Debug("ABI Specification (Decode)")

	// Unpack the result
	return Unpacker(abiSpecBytes, funcName, resultRaw)
}

func MakeAbi(abiData string) (ethAbi.ABI, error) {
	if len(abiData) == 0 {
		return ethAbi.ABI{}, nil
	}

	abiSpec, err := ethAbi.JSON(strings.NewReader(abiData))
	if err != nil {
		return ethAbi.ABI{}, err
	}

	return abiSpec, nil
}

//Convenience Packing Functions
func Packer(abiData, funcName string, args ...string) ([]byte, error) {
	abiSpec, err := MakeAbi(abiData)
	if err != nil {
		return nil, err
	}

	packedTypes, err := getPackingTypes(abiSpec, funcName, args...)
	if err != nil {
		return nil, err
	}

	packedBytes, err := abiSpec.Pack(funcName, packedTypes...)
	if err != nil {
		return nil, err
	}

	return packedBytes, nil
}

func getPackingTypes(abiSpec ethAbi.ABI, methodName string, args ...string) ([]interface{}, error) {
	var method ethAbi.Method
	if methodName == "" {
		method = abiSpec.Constructor
	} else {
		var exist bool
		method, exist = abiSpec.Methods[methodName]
		if !exist {
			return nil, fmt.Errorf("method '%s' not found", methodName)
		}
	}
	var values []interface{}
	if len(args) != len(method.Inputs) {
		return nil, fmt.Errorf("Invalid number of arguments asked to be packed, expected %v, got %v", len(method.Inputs), len(args))
	}
	for i, input := range method.Inputs { //loop through and get string vals packed into proper types
		inputType := input.Type
		val, err := packInterfaceValue(inputType, args[i])
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}
	return values, nil
}

func packInterfaceValue(typ ethAbi.Type, val string) (interface{}, error) {
	if typ.IsArray || typ.IsSlice {

		//check for fixed byte types and bytes types
		if typ.T == ethAbi.BytesTy {
			bytez := bytes.NewBufferString(val)
			return common.RightPadBytes(bytez.Bytes(), bytez.Len()%32), nil
		} else if typ.T == ethAbi.FixedBytesTy {
			bytez := bytes.NewBufferString(val)
			return common.RightPadBytes(bytez.Bytes(), typ.SliceSize), nil
		} else if typ.Elem.T == ethAbi.BytesTy || typ.Elem.T == ethAbi.FixedBytesTy {
			val = strings.Trim(val, "[]")
			arr := strings.Split(val, ",")
			var sliceOfFixedBytes [][]byte
			for _, str := range arr {
				bytez := bytes.NewBufferString(str)
				sliceOfFixedBytes = append(sliceOfFixedBytes, common.RightPadBytes(bytez.Bytes(), 32))
			}
			return sliceOfFixedBytes, nil
		} else {
			val = strings.Trim(val, "[]")
			arr := strings.Split(val, ",")
			var values interface{}

			for i := 0; i < typ.SliceSize; i++ {
				value, err := packInterfaceValue(*typ.Elem, arr[i])
				if err != nil {
					return nil, err
				}
				if value == nil {
					return nil, nil
				}
				//var bigIntValue = (*big.Int)(nil)
				switch value := value.(type) {
				case string:
					var ok bool
					if values, ok = values.([]string); ok {
						fmt.Printf("n=%#v\n", value)
					}
					values = append(values.([]string), value)
				case bool:
					var ok bool
					if values, ok = values.([]bool); ok {
						fmt.Printf("n=%#v\n", value)
					}
					values = append(values.([]bool), value)
				case uint8:
					var ok bool
					if values, ok = values.([]uint8); ok {
						fmt.Printf("n=%#v\n", value)
					}
					values = append(values.([]uint8), value)
				case uint16:
					var ok bool
					if values, ok = values.([]uint16); ok {
						fmt.Printf("n=%#v\n", value)
					}
					values = append(values.([]uint16), value)
				case uint32:
					var ok bool
					if values, ok = values.([]uint32); ok {
						fmt.Printf("n=%#v\n", value)
					}
					values = append(values.([]uint32), value)
				case uint64:
					var ok bool
					if values, ok = values.([]uint64); ok {
						fmt.Printf("n=%#v\n", value)
					}
					values = append(values.([]uint64), value)
				case int8:
					var ok bool
					if values, ok = values.([]int8); ok {
						fmt.Printf("n=%#v\n", value)
					}
					values = append(values.([]int8), value)
				case int16:
					var ok bool
					if values, ok = values.([]int16); ok {
						fmt.Printf("n=%#v\n", value)
					}
					values = append(values.([]int16), value)
				case int32:
					var ok bool
					if values, ok = values.([]uint32); ok {
						fmt.Printf("n=%#v\n", value)
					}
					values = append(values.([]int32), value)
				case int64:
					var ok bool
					if values, ok = values.([]uint64); ok {
						fmt.Printf("n=%#v\n", value)
					}
					values = append(values.([]int64), value)
				case *big.Int:
					var ok bool
					if values, ok = values.([]*big.Int); ok {
						fmt.Printf("n=%#v\n", value)
					}

					values = append(values.([]*big.Int), value)
				case common.Address:
					var ok bool
					if values, ok = values.([]common.Address); ok {
						fmt.Printf("n=%#v\n", value)
					}

					values = append(values.([]common.Address), value)
				}
			}
			return values, nil
		}
	} else {
		switch typ.T {
		case ethAbi.IntTy:
			switch typ.Size {
			case 8:
				val, err := strconv.ParseInt(val, 10, 8)
				if err != nil {
					return nil, err
				}
				return int8(val), nil
			case 16:
				val, err := strconv.ParseInt(val, 10, 16)
				if err != nil {
					return nil, err
				}
				return int16(val), nil
			case 32:
				val, err := strconv.ParseInt(val, 10, 32)
				if err != nil {
					return nil, err
				}
				return int32(val), nil
			case 64:
				val, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					return nil, err
				}
				return int64(val), nil
			default:
				val, set := big.NewInt(0).SetString(val, 10)
				if !set {
					return nil, fmt.Errorf("Could not set to big int")
				}
				return val, nil
			}
		case ethAbi.UintTy:
			switch typ.Size {
			case 8:
				val, err := strconv.ParseUint(val, 10, 8)
				if err != nil {
					return nil, err
				}
				return uint8(val), nil
			case 16:
				val, err := strconv.ParseUint(val, 10, 16)
				if err != nil {
					return nil, err
				}
				return uint16(val), nil
			case 32:
				val, err := strconv.ParseUint(val, 10, 32)
				if err != nil {
					return nil, err
				}
				return uint32(val), nil
			case 64:
				val, err := strconv.ParseUint(val, 10, 64)
				if err != nil {
					return nil, err
				}
				return uint64(val), nil
			default:
				val, set := big.NewInt(0).SetString(val, 10)
				if !set {
					return nil, fmt.Errorf("Could not set to big int")
				}
				return val, nil
			}
		case ethAbi.BoolTy:
			return strconv.ParseBool(val)
		case ethAbi.StringTy:
			return val, nil
		case ethAbi.AddressTy:
			return common.HexToAddress(val), nil
		default:
			return nil, fmt.Errorf("Could not get valid type from input")
		}
	}
}

func Unpacker(abiData, name string, data []byte) ([]*definitions.Variable, error) {

	abiSpec, err := MakeAbi(abiData)
	if err != nil {
		return []*definitions.Variable{}, err
	}

	numArgs, err := numReturns(abiSpec, name)
	if err != nil {
		return nil, err
	}

	if numArgs == 0 {
		return nil, nil
	} else if numArgs == 1 {
		var unpacked interface{}
		err = abiSpec.Unpack(&unpacked, name, data)
		if err != nil {
			return []*definitions.Variable{}, err
		}
		return formatUnpackedReturn(abiSpec, name, unpacked)
	} else {
		var unpacked []interface{}
		err = abiSpec.Unpack(&unpacked, name, data)
		if err != nil {
			return []*definitions.Variable{}, err
		}
		return formatUnpackedReturn(abiSpec, name, unpacked)
	}

}

func numReturns(abiSpec ethAbi.ABI, methodName string) (uint, error) {
	method, exist := abiSpec.Methods[methodName]
	if !exist {
		if methodName == "()" {
			return 0, nil
		}
		return 0, fmt.Errorf("method '%s' not found", methodName)
	}
	if len(method.Outputs) == 0 {
		log.Debug("Empty output, nothing to interface to")
		return 0, nil
	} else if len(method.Outputs) == 1 {
		return 1, nil
	} else {
		return 2, nil
	}
}

func formatUnpackedReturn(abiSpec ethAbi.ABI, methodName string, values ...interface{}) ([]*definitions.Variable, error) {
	var returnVars []*definitions.Variable
	method, exist := abiSpec.Methods[methodName]
	if !exist {
		return nil, fmt.Errorf("method '%s' not found", methodName)
	}

	if len(method.Outputs) > 1 {
		slice := reflect.ValueOf(reflect.ValueOf(values).Index(0).Interface())
		for i, output := range method.Outputs {
			arg, err := getStringValue(slice.Index(i).Interface(), output.Type)
			if err != nil {
				return nil, err
			}
			var name string
			if len(output.Name) > 0 {
				name = output.Name
			} else {
				nameNum := i
				name = strconv.Itoa(nameNum)
			}
			returnVar := &definitions.Variable{
				Name:  name,
				Value: arg,
			}
			returnVars = append(returnVars, returnVar)
		}
	} else {
		value := values[0]
		output := method.Outputs[0]
		arg, err := getStringValue(value, output.Type)
		if err != nil {
			return nil, err
		}
		var name string
		if len(output.Name) > 0 {
			name = output.Name
		} else {
			nameNum := 0
			name = strconv.Itoa(nameNum)
		}
		returnVar := &definitions.Variable{
			Name:  name,
			Value: arg,
		}
		returnVars = append(returnVars, returnVar)
	}
	return returnVars, nil
}

func getStringValue(value interface{}, typ ethAbi.Type) (string, error) {

	if typ.IsSlice || typ.IsArray {
		if typ.T == ethAbi.BytesTy || typ.T == ethAbi.FixedBytesTy {
			return string(bytes.Trim(value.([]byte), "\x00")[:]), nil
		}
		var val []string
		if typ.Elem.T == ethAbi.FixedBytesTy {
			byteVals := reflect.ValueOf(value)
			for i := 0; i < byteVals.Len(); i++ {
				val = append(val, string(bytes.Trim(byteVals.Index(i).Interface().([]byte), "\x00")[:]))
			}
		} else {
			values := reflect.ValueOf(value)
			for i := 0; i < typ.SliceSize; i++ {
				underlyingValue, err := getStringValue(values.Index(i).Interface(), *typ.Elem)
				if err != nil {
					return "", err
				}
				val = append(val, underlyingValue)
			}
		}
		StringVal := strings.Join(val, ",")
		StringVal = strings.Join([]string{"[", StringVal, "]"}, "")
		return StringVal, nil
	} else {
		switch typ.T {
		case ethAbi.IntTy:
			switch typ.Size {
			case 8, 16, 32, 64:
				return fmt.Sprintf("%v", value), nil
			default:
				return math.S256(value.(*big.Int)).String(), nil
			}
		case ethAbi.UintTy:
			switch typ.Size {
			case 8, 16, 32, 64:
				return fmt.Sprintf("%v", value), nil
			default:
				return math.U256(value.(*big.Int)).String(), nil
			}
		case ethAbi.BoolTy:
			return strconv.FormatBool(value.(bool)), nil
		case ethAbi.StringTy:
			return value.(string), nil
		case ethAbi.AddressTy:
			return strings.ToUpper(common.Bytes2Hex(value.(common.Address).Bytes())), nil
		default:
			return "", fmt.Errorf("Could not unpack value %v", value)
		}
	}
}
