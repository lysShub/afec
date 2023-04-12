package fec

func Xor(in, parity []byte) []byte {
	n := len(in)
	if n > len(parity) {
		// must cap(parity) > len(in)
		parity = parity[:len(in)]
	}

	for i, b := range in {
		parity[i] = parity[i] ^ b
	}

	return parity
}
