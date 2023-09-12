package afec

import (
	"fmt"
)

// xor a[i] = a[i] xor b[i]
func xor(a, b []byte) []byte {
	n := max(len(a), len(b))
	if cap(a) < len(b) {
		tmp := make([]byte, n)
		copy(tmp, a)
		a = tmp
	}

	if debug && !isEmpty(a[len(a):n]) {
		panic(fmt.Sprintf("expect zero: % X", a[len(a):n]))
	}

	rawxor(a[:n], b[:n])
	return a[:n]
}

// rawxor a[i] = a[i] ^ b[i]
func rawxor(a, b []byte) {
	if debug && len(a) != len(b) {
		panic(fmt.Errorf("%d %d", len(a), len(b)))
	}

	for i, v := range b {
		a[i] = a[i] ^ v
	}
}

func swap(a, b []byte) ([]byte, []byte) {
	na, nb := len(a), len(b)
	if na != nb {
		n := max(na, nb)
		if na < n {
			if cap(a) >= n {
				a = a[:n]
			} else {
				tmp := make([]byte, n)
				copy(tmp, a)
				a = tmp
			}
		} else {
			if cap(b) >= n {
				b = b[:n]
			} else {
				tmp := make([]byte, n)
				copy(tmp, b)
				b = tmp
			}
		}
	}

	rawxor(a, b)
	rawxor(b, a)
	rawxor(a, b)
	return a[:nb], b[:na]
}

func rawswap(a, b []byte) {
	rawxor(a, b)
	rawxor(b, a)
	rawxor(a, b)
}

// cpyclr  copy src to dst and clear dst't tail
func cpyclr(src, dst []byte) []byte {
	if cap(dst) < len(src) {
		dst = make([]byte, len(src))
	} else {
		dst = clrtail(dst, len(src))
	}

	copy(dst, src)

	return dst
}

func clrtail(b []byte, n int) []byte {
	if n < len(b) {
		clear(b[n:])
	}
	return b[:n]
}

func grow(p []byte, to int) []byte {
	if cap(p) > to {
		return p[:to]
	} else {
		tmp := make([]byte, to)
		copy(tmp, p)
		return tmp
	}
}
