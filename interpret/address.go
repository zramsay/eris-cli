package interpret

import (
	"encoding/hex"
)

const (
	AddressLength = 20
)
/////////// Address
type (
	Address [AddressLength]byte
)

func AddressStringToBytes(addr string) []byte {
	var slice []byte
	for i := 0; i < len(addr); i++ {
		a, _ := hex.DecodeString(addr[i : i+2])
		slice = append(slice, a[0])
		i++
	}
	return slice
}

func (a Address) Str() string   { return string(a[:]) }
func (a Address) Bytes() []byte { return a[:] }

// Sets the address to the value of b. If b is larger than len(a) it will panic
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

func HexToAddress(s string) Address { return BytesToAddress(FromHex(s)) }

func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}