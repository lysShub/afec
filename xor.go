package afec

func xor(in, parity []byte) []byte {
	n := len(in)
	for n > len(parity) {
		parity = append(parity, 0)
	}

	for i, b := range in {
		parity[i] = parity[i] ^ b
	}

	return parity
}

func tailZeros(b []byte) (n int) {
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] == 0 {
			n++
		} else {
			return
		}
	}
	return
}
