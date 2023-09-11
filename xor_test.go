package afec

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXor(t *testing.T) {

	var group = make([][]byte, 3)
	group[0] = []byte{1, 2, 3, 4, 5}
	group[1] = []byte{1, 1, 1, 0}
	group[2] = []byte{2, 2, 2, 2}

	{
		var parity = []byte{1, 1, 1, 1, 1}

		for _, b := range group {
			parity = xor(b, parity)
		}

		t.Log(parity) // [3 0 1 7 4]
	}

	{
		var parity = []byte{0, 0, 0, 0, 0}

		for _, b := range group {
			parity = xor(b, parity)
		}

		t.Log(parity) // [2 1 0 6 5]
	}

}

func TestSwap(t *testing.T) {
	{
		var a = []byte{1, 2, 3}
		var b = []byte{1, 1, 1, 1}

		a1, b1 := swap(a, b)

		require.Equal(t, []byte{1, 1, 1, 1}, a1)
		require.Equal(t, []byte{1, 2, 3}, b1)
	}

	{
		var a = make([]byte, 2, 4)
		a[0], a[1] = 9, 9
		var _ = append(a, 11, 11)
		var b = []byte{1, 2, 3, 4}

		rawswap(a[:4], b)

		require.Equal(t, []byte{1, 2, 3, 4}, a[:4])
		require.Equal(t, []byte{9, 9, 11, 11}, b)
	}
}
