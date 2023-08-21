package afec

import (
	"net"
	"testing"
	"time"
)

func TestSample(t *testing.T) {
	c, s := NewMockUDPConn(func(data []byte) time.Duration {
		return 0
	})
	var algo Algo = func(pl float64) (dataBlocks uint8, parityBlocks uint8) {
		return 2, 1
	}

	go func(rawConn net.Conn, algo Algo) { // sender
		conn := NewAfec(rawConn, algo)

		conn.Write([]byte{1, 1, 1, 1})
		conn.Write([]byte{2, 2})
		conn.Write([]byte{3, 3, 0, 0})
	}(c, algo)

	go func(rawCon net.Conn, algo Algo) { // recver
		conn := NewAfec(rawCon, algo)

		var b = make([]byte, 1532)
		for {
			n, err := conn.Read(b)
			t.Log(n, err)
		}
	}(s, algo)

	time.Sleep(time.Minute)
}

func TestMock(t *testing.T) {
	c, s := NewMockUDPConn(func(data []byte) time.Duration {
		return 0
	})

	go func() {

		c.Write([]byte{1, 2, 3, 1, 2, 3, 1, 2, 3})

	}()

	b := make([]byte, 6)
	n, err := s.Read(b)
	t.Log(n, err)
}
