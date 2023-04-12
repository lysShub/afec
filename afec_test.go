package afec

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Fudp_Fec_Base(t *testing.T) {
	var fecList = [][2]uint8{
		{2, 1},
		{1, 1},
		{1, 2},
	}

	for _, fec := range fecList {
		dataShards, parityShards := fec[0], fec[1]

		t.Logf("%d/%d", dataShards, parityShards)

		var c, s = NewMockUDPConn(func(data []byte) time.Duration { return time.Millisecond * 50 })

		go func() {
			cf := NewFudp(c)
			cf.SetFEC(uint8(dataShards), uint8(parityShards))

			for i := 4; i < 10; i++ {
				var b = []byte{}
				for j := 0; j < i; j++ {
					b = append(b, byte(j))
				}

				_, err := cf.Write(b)
				require.NoError(t, err)
			}
		}()

		sf := NewFudp(s)

		for i := 4; i < 10; i++ {
			var b = []byte{}
			for j := 0; j < i; j++ {
				b = append(b, byte(j))
			}

			var rb = make([]byte, 64)
			n, err := sf.Read(rb)
			require.NoError(t, err)
			require.Equal(t, b, rb[:n])
		}
	}
}

func Test_Fudp_Fec_Loss(t *testing.T) {
	var fecList = [][2]uint8{
		{2, 1},
		{1, 1},
		{1, 2},
	}

	for _, fec := range fecList {
		dataShards, parityShards := fec[0], fec[1]

		t.Logf("%d/%d", dataShards, parityShards)

		var c, s = NewMockUDPConn(func(data []byte) time.Duration { return time.Millisecond * 50 })

		go func() {
			cf := NewFudp(c)
			cf.SetFEC(uint8(dataShards), uint8(parityShards))

			for i := 4; i < 10; i++ {
				var b = []byte{}
				for j := 0; j < i; j++ {
					b = append(b, byte(j))
				}

				_, err := cf.Write(b)
				require.NoError(t, err)
			}
		}()

		sf := NewFudp(s)
		for i := 4; i < 10; i++ {
			var b = []byte{}
			for j := 0; j < i; j++ {
				b = append(b, byte(j))
			}

			var rb = make([]byte, 64)
			n, err := sf.Read(rb)
			require.NoError(t, err)
			require.Equal(t, b, rb[:n])
		}
	}
}
