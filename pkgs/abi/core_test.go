package abi

import (
	"bytes"
	//"encoding/json"
	"fmt"
	"testing"

	"github.com/monax/cli/log"

	"github.com/ethereum/go-ethereum/common"
)

//To Test:
//Bools, Arrays, Addresses, Hashes
//Test Packing different things
//After that, should be good to go

// quick helper padding
func pad(input []byte, size int, left bool) []byte {
	if left {
		return common.LeftPadBytes(input, size)
	}
	return common.RightPadBytes(input, size)
}

func TestPacker(t *testing.T) {
	for _, test := range []struct {
		ABI            string
		args           []interface{}
		name           string
		expectedOutput []byte
	}{
		{
			`[{"constant":false,"inputs":[{"name":"","type":"uint256"}],"name":"UInt","outputs":[],"payable":false,"type":"function"}]`,
			[]interface{}{int(1)},
			"UInt",
			pad([]byte{1}, 32, true),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"int256"}],"name":"Int","outputs":[],"payable":false,"type":"function"}]`,
			[]interface{}{int(-1)},
			"Int",
			[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"bool"}],"name":"Bool","outputs":[],"payable":false,"type":"function"}]`,
			[]interface{}{bool(true)},
			"Bool",
			pad([]byte{1}, 32, true),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"string"}],"name":"String","outputs":[],"payable":false,"type":"function"}]`,
			[]interface{}{string("marmots")},
			"String",
			append(common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000007"), pad([]byte("marmots"), 32, false)...),
		},
		{
			`[{"constant":false,"inputs":[{"name":"x","type":"bytes32"}],"name":"Bytes32","outputs":[],"payable":false,"type":"function"}]`,
			[]interface{}{string("deadbeef")},
			"Bytes32",
			pad([]byte("deadbeef"), 32, false),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"uint8"}],"name":"UInt8","outputs":[],"payable":false,"type":"function"}]`,
			[]interface{}{int(1)},
			"UInt8",
			pad([]byte{1}, 32, true),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"int8"}],"name":"Int8","outputs":[],"payable":false,"type":"function"}]`,
			[]interface{}{int(-1)},
			"Int8",
			[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"uint256"},{"name":"","type":"uint256"}],"name":"multiPackUInts","outputs":[],"payable":false,"type":"function"}]`,
			[]interface{}{int(1), int(1)},
			"multiPackUInts",
			append(pad([]byte{1}, 32, true), pad([]byte{1}, 32, true)...),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"bool"},{"name":"","type":"bool"}],"name":"multiPackBools","outputs":[],"payable":false,"type":"function"}]`,
			[]interface{}{bool(false), bool(false)},
			"multiPackBools",
			append(pad([]byte{0}, 32, true), pad([]byte{0}, 32, true)...),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"int256"},{"name":"","type":"int256"}],"name":"multiPackInts","outputs":[],"payable":false,"type":"function"}]`,
			[]interface{}{int(-1), int(-1)},
			"multiPackInts",
			[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		},

		{
			`[{"constant":false,"inputs":[{"name":"","type":"string"},{"name":"","type":"string"}],"name":"multiPackStrings","outputs":[],"payable":false,"type":"function"}]`,
			[]interface{}{"hello", "world"},
			"multiPackStrings",
			append(
				common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000005"),
				append(pad([]byte("hello"), 32, false),
					append(common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000005"),
						pad([]byte("world"), 32, false)...)...)...,
			),
		},
		{
			`[{"constant":false,"inputs":[],"name":"arrayOfBytes32Pack","inputs":[{"name":"","type":"bytes32[3]"}],"payable":false,"type":"function"}]`,
			[]interface{}{[]string{"den", "of", "marmots"}},
			"arrayOfBytes32Pack",
			append(
				pad([]byte("den"), 32, false),
				append(pad([]byte("of"), 32, false), pad([]byte("marmots"), 32, false)...)...,
			),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"uint256[3]"}],"name":"arrayOfUIntsPack","outputs":[],"payable":false,"type":"function"}]`,
			[]interface{}{[]int{1, 2, 3}},
			"arrayOfUIntsPack",
			common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"int256[3]"}],"name":"arrayOfIntsPack","outputs":[],"payable":false,"type":"function"}]`,
			[]interface{}{[]int{-1, -2, -3}},
			"arrayOfIntsPack",
			[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 254, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 253},
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"bool[2]"}],"name":"arrayOfBoolsPack","outputs":[],"payable":false,"type":"function"}]`,
			[]interface{}{[]bool{true, true}},
			"arrayOfBoolsPack",
			append(pad([]byte{1}, 32, true), pad([]byte{0}, 32, true)...),
		},
	} {
		log.SetLevel(log.DebugLevel)
		fmt.Println(test.name)
		abiStruct, err := MakeAbi(test.ABI)
		if err != nil {
			t.Errorf("Incorrect ABI: ", err)
		}
		if output, err := FormatAndPackInputs(abiStruct, test.name, test.args); err != nil {
			t.Errorf("Unexpected error in %v: %v", test.name, err)
		} else {
			if bytes.Compare(output[4:], test.expectedOutput) != 0 {
				t.Errorf("Incorrect output in %v,\n\t expected %v,\n\t got \t%v", test.name, test.expectedOutput, output[4:])
			}
		}
	}
}

func TestUnpacker(t *testing.T) {
	for _, test := range []struct {
		abi                  string
		packed               []byte
		function             string
		expectedStringOutput string
		expectedActualOutput interface{}
	}{
		{
			`[{"constant":true,"inputs":[],"name":"String","outputs":[{"name":"","type":"string"}],"payable":false,"type":"function"}]`,
			append(pad(common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000020"), 32, true), append(pad(common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000005"), 32, true), pad([]byte("Hello"), 32, false)...)...),
			"String",
			"hello",
			"hello",
		},
		{
			`[{"constant":true,"inputs":[],"name":"UInt","outputs":[{"name":"","type":"uint256"}],"payable":false,"type":"function"}]`,
			common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001"),
			"UInt",
			"1",
			int(1),
		},
		{
			`[{"constant":false,"inputs":[],"name":"Int","outputs":[{"name":"retVal","type":"int256"}],"payable":false,"type":"function"}]`,
			[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
			"Int",
			"-1",
			int(-1),
		},
		{
			`[{"constant":true,"inputs":[],"name":"Bool","outputs":[{"name":"","type":"bool"}],"payable":false,"type":"function"}]`,
			common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001"),
			"Bool",
			"true",
			true,
		},
		{
			`[{"constant":true,"inputs":[],"name":"Address","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"}]`,
			common.Hex2Bytes("0000000000000000000000001040E6521541DAB4E7EE57F21226DD17CE9F0FB7"),
			"Address",
			"1040E6521541DAB4E7EE57F21226DD17CE9F0FB7",
			"1040E6521541DAB4E7EE57F21226DD17CE9F0FB7",
		},
		{
			`[{"constant":false,"inputs":[],"name":"Bytes32","outputs":[{"name":"retBytes","type":"bytes32"}],"payable":false,"type":"function"}]`,
			pad([]byte("marmatoshi"), 32, true),
			"Bytes32",
			"marmatoshi",
			"marmatoshi",
		},
		{
			`[{"constant":false,"inputs":[],"name":"multiReturnUIntInt","outputs":[{"name":"","type":"uint256"},{"name":"","type":"int256"}],"payable":false,"type":"function"}]`,
			append(
				common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001"),
				[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}...,
			),
			"multiReturnUIntInt",
			"[1, -1]",
			[]int{1, -1},
		},
		{
			`[{"constant":false,"inputs":[],"name":"multiReturnMixed","outputs":[{"name":"","type":"string"},{"name":"","type":"uint256"}],"payable":false,"type":"function"}]`,
			append(
				common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000001"),
				append(common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000005"), pad([]byte("Hello"), 32, false)...)...,
			),
			"multiReturnMixed",
			`("Hello", 1)`,
			[]interface{}{"Hello", 1},
		},
		{
			`[{"constant":false,"inputs":[],"name":"multiPackBytes32","outputs":[{"name":"","type":"bytes32"},{"name":"","type":"bytes32"},{"name":"","type":"bytes32"}],"payable":false,"type":"function"}]`,
			append(
				pad([]byte("den"), 32, true),
				append(pad([]byte("of"), 32, true), pad([]byte("marmots"), 32, true)...)...,
			),
			"multiPackBytes32",
			`["den","of","marmots"]`,
			[]string{"den", "of", "marmots"},
		},
		{
			`[{"constant":false,"inputs":[],"name":"arrayReturnBytes32","outputs":[{"name":"","type":"bytes32[3]"}],"payable":false,"type":"function"}]`,
			append(
				pad([]byte("den"), 32, true),
				append(pad([]byte("of"), 32, true), pad([]byte("marmots"), 32, true)...)...,
			),
			"arrayReturnBytes32",
			"[den,of,marmots]",
			[]string{"den", "of", "marmots"},
		},
		{
			`[{"constant":false,"inputs":[],"name":"arrayReturnUInt","outputs":[{"name":"","type":"uint256[3]"}],"payable":false,"type":"function"}]`,
			common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"),
			"arrayReturnUInt",
			"[1,2,3]",
			[]int{1, 2, 3},
		},
		{
			`[{"constant":false,"inputs":[],"name":"arrayReturnInt","outputs":[{"name":"","type":"int256[2]"}],"payable":false,"type":"function"}]`,
			[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 253, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 254},
			"arrayReturnInt",
			"[-3,-2]",
			[]int{-3, -2},
		},
	} {
		//t.Log(test.name)
		contractAbi, err := MakeAbi(test.abi)
		if err != nil {
			t.Fatal(err)
		}
		toUnpackInto, method, err := CreateBlankSlate(contractAbi, test.function)
		if err != nil {
			t.Fatal(err)
		}
		err = contractAbi.Unpack(&toUnpackInto, test.function, test.packed)
		if err != nil {
			t.Fatal(err)
		}
		// get names of the types, get string results, get actual results, return them.
		fullStringResults := []string{"("}
		for i, methodOutput := range method.Outputs {
			_, _, err := ConvertUnpackedToJobTypes(toUnpackInto[i], methodOutput.Type)
			if err != nil {
				t.Fatal(err)
			}
		}
		fullStringResults = append(fullStringResults, ")")
	}
}
