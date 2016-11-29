package interpret

const (
	HashLength    = 32	
)

type (
	Hash    [HashLength]byte
)

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

// Sets the hash to the value of b. If b is larger than len(h) it will panic
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}