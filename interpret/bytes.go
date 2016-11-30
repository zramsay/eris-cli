package interpret

import (
	"bytes"
	"encoding/binary"
)

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

// Copy bytes
//
// Returns an exact copy of the provided bytes
func CopyBytes(b []byte) (copiedBytes []byte) {
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)

	return
}

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