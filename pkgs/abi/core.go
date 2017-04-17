package abi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	ethAbi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

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

func FormatAndPackInputs(reference ethAbi.ABI, function string, inputs ...interface{}) ([]byte, error) {
	if func2Pack, ok := reference.Methods[function]; !ok {
		return nil, fmt.Errorf("Invalid method called for contract, doesn't exist")
	} else {
		if len(inputs) != len(func2Pack.Inputs) {
			return nil, fmt.Errorf("Invalid number of inputs called for this function, expected %v got %v", len(func2Pack.Inputs), len(inputs))
		}
		for i, expectedInput := range func2Pack.Inputs {
			var err error
			inputs[i], err = convertToPackingType(inputs[i], expectedInput.Type)
			if err != nil {
				return nil, err
			}
		}
		return reference.Pack(function, inputs)
	}
}

func convertSlice(from []interface{}, to ethAbi.Type) (interface{}, error) {
	if !to.IsSlice {
		return nil, fmt.Errorf("Attempting to convert to non slice type")
	} else if to.SliceSize != -1 && len(from) != to.SliceSize {
		return nil, fmt.Errorf("Length of array does not match, expected %v got %v", to.SliceSize, len(from))
	}
	for i, typ := range from {
		var err error
		from[i], err = convertToPackingType(typ, *to.Elem)
		if err != nil {
			return nil, err
		}
	}
	return from, nil
}

func convertToPackingType(from interface{}, to ethAbi.Type) (interface{}, error) {
	if to.IsSlice || to.IsArray && to.T != ethAbi.BytesTy && to.T != ethAbi.FixedBytesTy {
		if typ, ok := from.([]interface{}); !ok {
			return nil, fmt.Errorf("Unexpected non slice type during type conversion, please reformat your run file to use an array/slice.")
		} else {
			return convertSlice(typ, to)
		}
	} else {
		switch to.T {
		case ethAbi.IntTy, ethAbi.UintTy:
			var signed bool = to.T == ethAbi.IntTy
			if typ, ok := from.(int); !ok {
				return nil, fmt.Errorf("Unexpected non integer type during type conversion, please reformat your run file to use an integer.")
			} else {
				switch to.Size {
				case 8:
					if signed {
						return int8(typ), nil
					}
					return uint8(typ), nil
				case 16:
					if signed {
						return int16(typ), nil
					}
					return uint16(typ), nil
				case 32:
					if signed {
						return int32(typ), nil
					}
					return uint32(typ), nil
				case 64:
					if signed {
						return int64(typ), nil
					}
					return uint64(typ), nil
				default:
					big := common.Big0
					if signed {
						return big.SetInt64(int64(typ)), nil
					}
					return big.SetUint64(uint64(typ)), nil
				}
			}
		case ethAbi.BoolTy:
			if typ, ok := from.(bool); !ok {
				return nil, fmt.Errorf("Unexpected non bool type during type conversion, please reformat your run file to use a bool.")
			} else {
				return typ, nil
			}
		case ethAbi.StringTy:
			if typ, ok := from.(string); !ok {
				return nil, fmt.Errorf("Unexpected non string type during type conversion, please reformat your run file to use a string.")
			} else {
				return typ, nil
			}
		case ethAbi.AddressTy:
			if typ, ok := from.(string); !ok {
				return nil, fmt.Errorf("Unexpected non string type during type conversion, please reformat your run file to use a string.")
			} else {
				return common.HexToAddress(typ), nil
			}
		case ethAbi.FunctionTy:
			if typ, ok := from.(string); !ok {
				return nil, fmt.Errorf("Unexpected non string type during type conversion, please reformat your run file to use a string.")
			} else {
				if len(typ) != 24 {
					return nil, fmt.Errorf("Expected function signature to be address + 4 byte function signature. Got %v bytes.", len(typ))
				} else {
					return common.Hex2Bytes(typ), nil
				}
			}
		case ethAbi.BytesTy:
			if typ, ok := from.(string); !ok {
				return nil, fmt.Errorf("Unexpected non string type during type conversion, please reformat your run file to use a string.")
			} else {
				return common.Hex2Bytes(typ), nil
			}
		default:
			return nil, fmt.Errorf("Invalid type during type conversion.")
		}
	}
}

func CreateBlankSlate(reference ethAbi.ABI, function string) ([]interface{}, ethAbi.Method, error) {
	if func2Unpack, ok := reference.Methods[function]; !ok {
		return nil, ethAbi.Method{}, fmt.Errorf("Invalid method called for contract, doesn't exist")
	} else {
		var outputs []interface{}
		for i, output := range func2Unpack.Outputs {
			outputs[i] = output.Type.Type
		}
		return outputs, func2Unpack, nil
	}
}

func ConvertUnpackedToJobTypes(from interface{}, reference ethAbi.Type) (string, interface{}, error) {
	if reference.IsSlice || reference.IsArray && reference.T != ethAbi.FixedBytesTy && reference.T != ethAbi.BytesTy {
		var normalSliceString = func(i interface{}) string {
			buf := new(bytes.Buffer)
			json.NewEncoder(buf).Encode(i)
			return fmt.Sprintf(buf.String())
		}
		// convert to yaml createable types, ignoring string and bool because those are accounted for already
		sliceVal := reflect.ValueOf(from)
		var stored []interface{}
		for i := 0; i < sliceVal.Len(); i++ {
			if _, typ, err := ConvertUnpackedToJobTypes(sliceVal.Index(i).Interface(), *reference.Elem); err != nil {
				stored = append(stored, typ)
			} else {
				return "", nil, fmt.Errorf("Error in converting slice: %v", err)
			}

		}
		return normalSliceString(stored), stored, nil
	} else {
		switch reference.T {
		case ethAbi.UintTy, ethAbi.IntTy:
			switch typ := from.(type) {
			case int8, int16, int32, int64:
				return fmt.Sprintf("%v", from), int(typ.(int)), nil
			case uint8, uint16, uint32, uint64:
				return fmt.Sprintf("%v", from), int(typ.(uint)), nil
			case *big.Int:
				val := typ.Int64()
				if val == 0 {
					val := typ.Uint64()
					return typ.String(), int(val), nil
				}
				return typ.String(), int(val), nil
			default:
				return "", nil, fmt.Errorf("Could not find integer type to convert.")
			}
		case ethAbi.StringTy:
			return from.(string), from.(string), nil
		case ethAbi.BoolTy:
			return fmt.Sprintf("%v", from), from.(bool), nil
		case ethAbi.AddressTy:
			return from.(common.Address).Str(), from.(common.Address).Str(), nil
		case ethAbi.BytesTy, ethAbi.FixedBytesTy:
			return string(bytes.Trim(from.([]byte), "\x00")[:]), string(bytes.Trim(from.([]byte), "\x00")[:]), nil
		default:
			return "", nil, fmt.Errorf("Could not find type to convert.")
		}
	}

}
