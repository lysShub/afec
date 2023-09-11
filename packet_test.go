package afec

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPacket(t *testing.T) {
	var mk uint8 = 0b10000000

	{
		var p = Pack(make([]byte, MiniPackSize))
		p.setDataType()
		p.setGlen(3)

		require.Equal(t, p[3], mk+(uint8(1)<<6)+3)
		require.Equal(t, true, p.isDataType())
		require.Equal(t, uint8(3), p.glen())
	}

	{
		var p = Pack(make([]byte, MiniPackSize))
		p.setGlen(3)
		p.setDataType()

		require.Equal(t, p[3], mk+(uint8(1)<<6)+3)
		require.Equal(t, true, p.isDataType())
		require.Equal(t, uint8(3), p.glen())
	}

	{
		var p = Pack(make([]byte, MiniPackSize))
		p.setDataType()
		p.setGlen(3)
		p.setGlen(7)

		require.Equal(t, p[3], mk+(uint8(1)<<6)+7)
		require.Equal(t, true, p.isDataType())
		require.Equal(t, uint8(7), p.glen())
	}

	{
		var p = Pack(make([]byte, MiniPackSize))
		p.setGlen(3)
		p.setDataType()
		p.setGlen(7)
		p.setParityType()

		require.Equal(t, p[3], mk+(uint8(0)<<6)+7)
		require.Equal(t, false, p.isDataType())
		require.Equal(t, uint8(7), p.glen())
	}

	{
		var p = Pack(make([]byte, MiniPackSize))
		p.setGid(111)
		p.setPL(0.082)

		require.Equal(t, uint8(111), p.gid())
		require.InDelta(t, 0.082, p.pl().get(), prec)
	}
}
