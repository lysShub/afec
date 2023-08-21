package afec

import (
	"net"

	"sync/atomic"
)

// Algo a-fec algorithm, dataBlocks can't be 0.
type Algo func(pl float64) (dataBlocks, parityBlocks uint8)

type afec struct {
	rawConn net.Conn
	algo    Algo

	recvCnt, lossCnt atomic.Uint32
}

func (a *afec) pl() (pl float64) {
	l, r := a.lossCnt.Load(), a.recvCnt.Load()
	a.lossCnt.Store(0)
	a.lossCnt.Store(0)

	if r == 0 && l == 0 {
		return 0
	} else {
		return float64(l) / float64(l+r)
	}
}

type Afec struct {
	recv
	send
}

func NewAfec(rawConn net.Conn, algo Algo) *Afec {
	var a = &afec{
		rawConn: rawConn,
		algo:    algo,
	}

	var f = &Afec{}
	f.send = newSend(a)
	f.recv = newRecv(a)
	return f
}
