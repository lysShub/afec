package afec

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Correct(t *testing.T) {
	type suit struct {
		send  [][]byte
		recv  [][]byte
		delay []time.Duration
		algo  Algo
		msg   string
	}

	var tail_zero_suits = []suit{
		{
			send: [][]byte{
				{1, 2, 3},
				{1},
				{2, 3, 0, 0},
			},
			algo: func(pl float64) (dataBlocks, parityBlocks uint8) { return 2, 1 },
			msg:  "test-zero-tail-case1",
		},
		{
			send: [][]byte{
				{1, 2, 3, 0},
				{1, 0},
				{2, 3, 0, 0},
			},
			algo: func(pl float64) (dataBlocks, parityBlocks uint8) { return 2, 1 },
			msg:  "test-zero-tail-case2",
		},
		{
			send: [][]byte{
				{1, 2, 3},
				{1},
				{0, 0, 0, 0},
			},
			algo: func(pl float64) (dataBlocks, parityBlocks uint8) { return 2, 1 },
			msg:  "test-zero-pack-case1",
		},
		{
			send: [][]byte{
				{0, 0, 0},
				{0},
				{0, 0, 0, 0},
			},
			algo: func(pl float64) (dataBlocks, parityBlocks uint8) { return 2, 1 },
			msg:  "test-zero-pack-case1",
		},
	}

	suits := []suit{}
	suits = append(suits, tail_zero_suits...)
	for _, suit := range suits {
		if suit.recv == nil {
			suit.recv = suit.send
		}

		c, s := NewMockUDPConn(func() func() time.Duration {
			var i int
			return func() time.Duration {
				i++
				if i > len(suit.delay) {
					return 0
				}
				return suit.delay[i-1]
			}
		}())

		go func() { // sender
			conn := NewAfec(c, suit.algo)
			for _, p := range suit.send {
				_, err := conn.Write(p)
				require.NoError(t, err)
			}
		}()

		{ // recv
			conn := NewAfec(s, suit.algo)

			var b = make([]byte, 1532)
			for _, r := range suit.recv {
				b = b[:cap(b)]
				n, err := conn.Read(b)
				require.NoError(t, err)
				require.Equal(t, b[:n], r)
			}
		}
	}
}

// func TestMock(t *testing.T) {
// 	c, s := NewMockUDPConn(func(data []byte) time.Duration {
// 		return 0
// 	})

// 	go func() {

// 		c.Write([]byte{1, 2, 3, 1, 2, 3, 1, 2, 3})

// 	}()

// 	b := make([]byte, 6)
// 	n, err := s.Read(b)
// 	t.Log(n, err)
// }
