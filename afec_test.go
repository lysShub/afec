package afec

import (
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/require"
)

func Test_Base(t *testing.T) {
	monkey.Patch((*afec).fec, func(_ *afec, pl float64) (dataBlocks, parityBlocks uint8) {
		return 1, 1
	})
	require.True(t, debug)

	type suit struct {
		send  [][]byte
		recv  [][]byte
		delay []time.Duration
		msg   string
	}

	var tail_zero_suits = []suit{
		{
			send: [][]byte{
				{1, 2, 3},
				{1},
				{2, 3, 0, 0},
			},
			msg: "test-zero-tail-case1",
		},
		{
			send: [][]byte{
				{1, 2, 3, 0},
				{1, 0},
				{2, 3, 0, 0},
			},
			msg: "test-zero-tail-case2",
		},
		{
			send: [][]byte{
				{1, 2, 3},
				{1},
				{0, 0, 0, 0},
			},
			msg: "test-zero-pack-case3",
		},
		{
			send: [][]byte{
				{0, 0, 0},
				{0},
				{0, 0, 0, 0},
			},
			msg: "test-zero-pack-case4",
		},
	}

	var reconstruct_case = []suit{
		{
			send: [][]byte{
				{0, 0, 0},
				{1}, {2}, {3},
				{0, 0, 0, 0},
			},
			delay: []time.Duration{time.Minute},
			recv: [][]byte{
				{1}, {2}, {3},
				{0, 0, 0},
				{0, 0, 0, 0},
			},
			msg: "test-reconstruct-case1",
		},
		{
			send: [][]byte{
				{0, 0, 0},
				{1}, {2}, {3},
				{0, 0, 0, 0},
			},
			delay: []time.Duration{time.Minute, 0, time.Minute},
			recv: [][]byte{
				{2}, {3},
				{0, 0, 0},
				{0, 0, 0, 0},
			},
			msg: "test-reconstruct-case2",
		},
		{
			send: [][]byte{
				{0, 0, 0},
				{1}, {2}, {3},
				{0, 0, 0, 0},
				{9},
			},
			delay: []time.Duration{time.Minute, 0, time.Minute},
			recv: [][]byte{
				{2}, {3},
				{0, 0, 0},
				{0, 0, 0, 0},
				{1},
				{9},
			},
			msg: "test-reconstruct-case3",
		},
	}

	suits := []suit{}
	suits = append(suits, tail_zero_suits...)
	suits = append(suits, reconstruct_case...)
	for _, suit := range suits {
		if suit.recv == nil {
			suit.recv = suit.send
		}

		if suit.msg == `test-reconstruct-case2` {
			print()
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
			conn := NewAfec(c)
			for _, p := range suit.send {
				_, err := conn.Write(gcap(p))
				require.NoError(t, err, suit.msg)
			}
		}()

		{ // recver
			conn := NewAfec(s)

			var b = make([]byte, 1532)
			for _, r := range suit.recv {
				b = b[:1500]
				n, err := conn.Read(b)
				require.NoError(t, err, suit.msg)
				require.Equal(t, b[:n], r, suit.msg)
				require.True(t, isEmpty(b[n:cap(b)]), suit.msg)
			}
		}
	}
}

func gcap(b []byte) []byte {
	tmp := make([]byte, len(b), len(b)+HdrSize)
	copy(tmp, b)
	return tmp
}
