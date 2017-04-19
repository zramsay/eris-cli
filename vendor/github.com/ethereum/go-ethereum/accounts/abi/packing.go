// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package abi

import (
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

// packBytesSlice packs the given bytes as [L, V] as the canonical representation
// bytes slice
func packBytesSlice(bytes []byte, l int) []byte {
	len := packNum(reflect.ValueOf(l))
	return append(len, common.RightPadBytes(bytes, (l+31)/32*32)...)
}

// packElement packs the given reflect value according to the abi specification in
// t.
func packElement(t Type, reflectValue reflect.Value) []byte {
	switch t.T {
	case IntTy, UintTy:
		return packNum(reflectValue)
	case StringTy:
		return packBytesSlice([]byte(reflectValue.String()), reflectValue.Len())
	case AddressTy:
		if reflectValue.Kind() == reflect.Array {
			reflectValue = mustArrayToByteSlice(reflectValue)
		}

		return common.LeftPadBytes(reflectValue.Bytes(), 32)
	case BoolTy:
		fmt.Printf("Bool val: %v\n", reflectValue.Bool())
		if reflectValue.Bool() == true {
			return math.PaddedBigBytes(common.Big1, 32)
		} else {
			fmt.Println("Padded value: ", math.PaddedBigBytes(common.Big0, 32))
			return math.PaddedBigBytes(common.Big0, 32)
		}
	case BytesTy:
		if reflectValue.Kind() == reflect.Array {
			reflectValue = mustArrayToByteSlice(reflectValue)
		}
		return packBytesSlice(reflectValue.Bytes(), reflectValue.Len())
	case FixedBytesTy, FunctionTy:
		if reflectValue.Kind() == reflect.Array {
			reflectValue = mustArrayToByteSlice(reflectValue)
		}

		return common.RightPadBytes(reflectValue.Bytes(), 32)
	}
	panic("abi: fatal error")
}
