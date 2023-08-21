package afec

import "github.com/tmthrgd/go-memset"

// xor a[i] = a[i] xor b[i]
//
// if len(a) != len(b), by len(b); so a maybe alloc new memory.
func xor(a, b []byte) []byte {
	if len(a) < len(b) {
		if cap(a) < len(b) {
			tmp := make([]byte, cap(b))
			copy(tmp, a)
			a = tmp[:len(b)]
		} else {
			memset.Memset(a[len(a):len(b)], 0)
			a = a[:len(b)]
		}
	}

	for i, v := range b {
		a[i] = a[i] ^ v
	}
	return a
}

func swap(a, b []byte) ([]byte, []byte) {
	na, nb := len(a), len(b)
	if na != nb {
		delta := na - nb
		if delta > 0 {
			if nb+delta > cap(b) {
				tmp := make([]byte, nb+delta)
				copy(tmp, b)
				b = tmp[:nb]
			} else {
				memset.Memset(b[nb:nb+delta], 0)
			}
			b = b[:na]
		} else {
			delta = -delta
			if na+delta > cap(a) {
				tmp := make([]byte, na+delta)
				copy(tmp, a)
				a = tmp[:nb]
			} else {
				memset.Memset(a[na:na+delta], 0)
			}
			a = a[:nb]
		}
	}

	a = xor(a, b)
	b = xor(b, a)
	a = xor(a, b)
	return a[:nb], b[:na]
}

// cpyclr, copy src to dst
func cpyclr(src, dst []byte) []byte {
	n := len(src)
	if cap(dst) < n {
		dst = make([]byte, cap(src))
	} else if len(dst) > n {
		memset.Memset(dst[n:], 0)
	}
	copy(dst[:n], src)
	return dst[:n]
}
