package common

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

//-------------------------------------------------------
// hex and ints

// keeps N bytes of the conversion
func NumberToBytes(num interface{}, N int) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, num)
	if err != nil {
		// TODO: get this guy a return error?
	}
	//fmt.Println("btyes!", buf.Bytes())
	if buf.Len() > N {
		return buf.Bytes()[buf.Len()-N:]
	}
	return buf.Bytes()
}

// s can be string, hex, or int.
// returns properly formatted 32byte hex value
func Coerce2Hex(s string) string {
	//fmt.Println("coercing to hex:", s)
	// is int?
	i, err := strconv.Atoi(s)
	if err == nil {
		return "0x" + hex.EncodeToString(NumberToBytes(int32(i), i/256+1))
	}
	// is already prefixed hex?
	if len(s) > 1 && s[:2] == "0x" {
		if len(s)%2 == 0 {
			return s
		}
		return "0x0" + s[2:]
	}
	// is unprefixed hex?
	if len(s) > 32 {
		return "0x" + s
	}
	pad := strings.Repeat("\x00", (32-len(s))) + s
	ret := "0x" + hex.EncodeToString([]byte(pad))
	//fmt.Println("result:", ret)
	return ret
}

func CoerceHexAndPad(aa string, padright bool) string {
	if !IsHex(aa) {
		//first try and convert to int
		n, err := strconv.Atoi(aa)
		if err != nil {
			// right pad strings
			if padright {
				aa = "0x" + fmt.Sprintf("%x", aa) + fmt.Sprintf("%0"+strconv.Itoa(64-len(aa)*2)+"s", "")
			} else {
				aa = "0x" + fmt.Sprintf("%x", aa)
			}
		} else {
			aa = "0x" + fmt.Sprintf("%x", n)
		}
	}
	return aa
}

func IsHex(s string) bool {
	if len(s) < 2 {
		return false
	}
	if s[:2] == "0x" {
		return true
	}
	return false
}

func AddHex(s string) string {
	if len(s) < 2 {
		return "0x" + s
	}

	if s[:2] != "0x" {
		return "0x" + s
	}

	return s
}

func StripHex(s string) string {
	if len(s) > 1 {
		if s[:2] == "0x" {
			s = s[2:]
			if len(s)%2 != 0 {
				s = "0" + s
			}
			return s
		}
	}
	return s
}

func StripZeros(s string) string {
	i := 0
	for ; i < len(s); i++ {
		if s[i] != '0' {
			break
		}
	}
	return s[i:]
}

func StripOnes(s string) string {
	i := 0
	for ; i < len(s); i++ {
		if s[i:i+1] != "01" {
			break
		}
	}
	return s[i:]
}

func Bytes2Hex(d []byte) string {
	return hex.EncodeToString(d)
}

func RightPadBytes(slice []byte, l int) []byte {
	if l < len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[0:len(slice)], slice)

	return padded
}

func LeftPadBytes(slice []byte, l int) []byte {
	if l < len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)

	return padded
}

func LeftPadString(str string, l int) string {
	if l < len(str) {
		return str
	}

	zeros := Bytes2Hex(make([]byte, (l-len(str))/2))

	return zeros + str

}

func RightPadString(str string, l int) string {
	if l < len(str) {
		return str
	}

	zeros := Bytes2Hex(make([]byte, (l-len(str))/2))

	return str + zeros

}

func UnLeftPadBytes(slice []byte) []byte {
	var l int
	for i, b := range slice {
		if b != byte(0) {
			l = i
			break
		}
	}
	unpadded := make([]byte, len(slice)-l)
	copy(unpadded, slice[l:])

	return unpadded
}

func UnRightPadBytes(slice []byte) []byte {
	var l int
	for i, b := range slice {
		if b == byte(0) {
			l = i
			break
		}
	}
	unpadded := make([]byte, l)
	copy(unpadded, slice[:l])

	return unpadded
}

func Address(slice []byte) (addr []byte) {
	if len(slice) < 20 {
		addr = LeftPadBytes(slice, 20)
	} else if len(slice) > 20 {
		addr = slice[len(slice)-20:]
	} else {
		addr = slice
	}

	addr = CopyBytes(addr)

	return
}

func AddressStringToBytes(addr string) []byte {
	var slice []byte
	for i := 0; i < len(addr); i++ {
		a, _ := hex.DecodeString(addr[i : i+2])
		slice = append(slice, a[0])
		i++
	}
	return slice
}

// Copy bytes
//
// Returns an exact copy of the provided bytes
func CopyBytes(b []byte) (copiedBytes []byte) {
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)

	return
}
