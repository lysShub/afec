package afec

import (
	"testing"
)

func TestXor(t *testing.T) {

	var group = make([][]byte, 3)
	group[0] = []byte{1, 2, 3, 4, 5}
	group[1] = []byte{1, 1, 1, 0}
	group[2] = []byte{2, 2, 2, 2}

	var parity = []byte{}

	for _, b := range group {
		parity = xor(b, parity)
	}

	// group[1] loss

	var recover = []byte{}

	recover = xor(group[0], recover)
	recover = xor(group[2], recover)
	recover = xor(recover, parity)

	t.Log(recover)
}

func TestSwap(t *testing.T) {
	{
		var a = []byte{1, 2, 3}
		var b = []byte{1, 1, 1, 1}

		a1, b1 := swap(a, b)

		t.Log(a1)
		t.Log(b1)
	}
}
