package afec

import (
	"fmt"
)

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
			clear(a[len(a):len(b)])
			a = a[:len(b)]
		}
	}

	rawxor(a, b)
	return a
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

// cpyclr  copy src to dst
func cpyclr(src, dst []byte) []byte {
	n := len(src)
	if cap(dst) < n {
		dst = make([]byte, cap(src))
	} else if len(dst) > n {
		clear(dst[n:])
	}
	copy(dst[:n], src)
	return dst[:n]
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
