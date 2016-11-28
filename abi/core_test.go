package abi

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	pm "github.com/eris-ltd/eris-cli/definitions"
)

//To Test:
//Bools, Arrays, Addresses, Hashes
//Test Packing different things
//After that, should be good to go

func TestPacker(t *testing.T) {
	for _, test := range []struct {
		ABI            string
		args           []string
		name           string
		expectedOutput []byte
	}{
		{
			`[{"constant":false,"inputs":[{"name":"","type":"uint256"}],"name":"UInt","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"1"},
			"UInt",
			pad([]byte{1}, 32, true),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"uint256"},{"name":"","type":"uint256"}],"name":"multiPack","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"1", "1"},
			"multiPack",
			append(pad([]byte{1}, 32, true), pad([]byte{1}, 32, true)...),
		},
		{
			`[{"constant":false,"inputs":[{"name":"x","type":"bytes32"}],"name":"setBytes","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"marmatoshi"},
			"setBytes",
			pad([]byte("marmatoshi"), 32, false),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"uint8"},{"name":"","type":"uint8"}],"name":"smallInts","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"1", "1"},
			"smallInts",
			append(pad([]byte{1}, 32, true), pad([]byte{1}, 32, true)...),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"string"},{"name":"","type":"string"}],"name":"multiPackStrings","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"hello", "world"},
			"multiPackStrings",
			append(
				Hex2Bytes("000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000005"),
				append(pad([]byte("hello"), 32, false),
					append(Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000005"),
						pad([]byte("world"), 32, false)...)...)...,
			),
		},
		{
			`[{"constant":false,"inputs":[],"name":"getBytes","inputs":[{"name":"","type":"bytes32[3]"}],"payable":false,"type":"function"}]`,
			[]string{`[den,of,marmots]`},
			"getBytes",
			append(
				pad([]byte("den"), 32, false),
				append(pad([]byte("of"), 32, false), pad([]byte("marmots"), 32, false)...)...,
			),
		},
		{
			`[{"constant":false,"inputs":[{"name":"","type":"uint256[3]"}],"name":"arrayPack","outputs":[],"payable":false,"type":"function"}]`,
			[]string{"[1,2,3]"},
			"arrayPack",
			Hex2Bytes("000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"),
		},
	} {
		t.Log(test.args)
		fmt.Println(test.name)
		abiStruct, err := JSON(strings.NewReader(test.ABI))
		if err != nil {
			t.Errorf("Incorrect ABI: ", err)
		}
		if output, err := Packer(test.ABI, test.name, test.args...); err != nil {
			t.Error("Unexpected error in ", test.name, ": ", err)
		} else {
			if bytes.Compare(output, append(abiStruct.Methods[test.name].Id(), test.expectedOutput...)) != 0 {
				t.Errorf("Incorrect output, expected %v, got %v", test.expectedOutput, output)
			}
		}
	}
}

func TestUnpacker(t *testing.T) {
	for _, test := range []struct {
		abi            string
		packed         []byte
		name           string
		expectedOutput []pm.Variable
	}{
		{
			`[{"constant":true,"inputs":[],"name":"x","outputs":[{"name":"","type":"string"}],"payable":false,"type":"function"}]`,
			append(pad(Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000020"), 32, true), append(pad(Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000005"), 32, true), pad([]byte("Hello"), 32, false)...)...),
			"x",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "Hello",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"get","outputs":[{"name":"retVal","type":"int256"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"x","type":"int256"}],"name":"set","outputs":[],"payable":false,"type":"function"}]`,
			[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
			"get",
			[]pm.Variable{
				{
					Name:  "retVal",
					Value: "-1",
				},
			},
		},
		{
			`[{"constant":true,"inputs":[],"name":"Bool","outputs":[{"name":"","type":"bool"}],"payable":false,"type":"function"}]`,
			Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001"),
			"Bool",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "true",
				},
			},
		},
		{
			`[{"constant":true,"inputs":[],"name":"addr","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"}]`,
			Hex2Bytes("0000000000000000000000001040E6521541DAB4E7EE57F21226DD17CE9F0FB7"),
			"addr",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "1040E6521541DAB4E7EE57F21226DD17CE9F0FB7",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"getBytes","outputs":[{"name":"retBytes","type":"bytes32"}],"payable":false,"type":"function"}]`,
			pad([]byte("marmatoshi"), 32, true),
			"getBytes",
			[]pm.Variable{
				{
					Name:  "retBytes",
					Value: "marmatoshi",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"multiReturn","outputs":[{"name":"","type":"uint256"},{"name":"","type":"int256"}],"payable":false,"type":"function"}]`,
			append(
				Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001"),
				[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}...,
			),
			"multiReturn",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "1",
				},
				{
					Name:  "1",
					Value: "-1",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"multiReturnMixed","outputs":[{"name":"","type":"string"},{"name":"","type":"uint256"}],"payable":false,"type":"function"}]`,
			append(
				Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000001"),
				append(Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000005"), pad([]byte("Hello"), 32, false)...)...,
			),
			"multiReturnMixed",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "Hello",
				},
				{
					Name:  "1",
					Value: "1",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"getBytes","outputs":[{"name":"","type":"bytes32"},{"name":"","type":"bytes32"},{"name":"","type":"bytes32"}],"payable":false,"type":"function"}]`,
			append(
				pad([]byte("den"), 32, true),
				append(pad([]byte("of"), 32, true), pad([]byte("marmots"), 32, true)...)...,
			),
			"getBytes",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "den",
				},
				{
					Name:  "1",
					Value: "of",
				},
				{
					Name:  "2",
					Value: "marmots",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"getBytesSlice","outputs":[{"name":"","type":"bytes32[3]"}],"payable":false,"type":"function"}]`,
			append(
				pad([]byte("den"), 32, true),
				append(pad([]byte("of"), 32, true), pad([]byte("marmots"), 32, true)...)...,
			),
			"getBytesSlice",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "[den,of,marmots]",
				},
			},
		},
		{
			`[{"constant":false,"inputs":[],"name":"arrayReturn","outputs":[{"name":"","type":"uint256[3]"}],"payable":false,"type":"function"}]`,
			Hex2Bytes("000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"),
			"arrayReturn",
			[]pm.Variable{
				{
					Name:  "0",
					Value: "[1,2,3]",
				},
			},
		},
	} {
		//t.Log(test.name)
		//t.Log(test.packed)
		output, err := Unpacker(test.abi, test.name, test.packed)
		if err != nil {
			t.Errorf("Unpacker failed: %v", err)
		}
		for i, expectedOutput := range test.expectedOutput {

			if output[i].Name != expectedOutput.Name {
				t.Errorf("Unpacker failed: Incorrect Name, got %v expected %v", output[i].Name, expectedOutput.Name)
			}
			//t.Log("Test: ", output[i].Value)
			//t.Log("Test: ", expectedOutput.Value)
			if strings.Compare(output[i].Value, expectedOutput.Value) != 0 {
				t.Errorf("Unpacker failed: Incorrect value, got %v expected %v", output[i].Value, expectedOutput.Value)
			}
		}
	}
}
