package afec

import (
	"encoding/binary"
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

func TestXxx(t *testing.T) {
	// tzc: tail zero compression

	// 结尾为N个0, 用xx0表示
	// Deprecated: 解码的时候失灵
}

func encode(b []byte) []byte {
	zeros := tailZeros(b)

	if zeros == 0 {
		return b
	} else {
		i := len(b) - zeros

		binary.LittleEndian.PutUint16(b[i:], uint16(zeros))
		return b[:i+2]
	}
}
