package afec

import (
	"fmt"
	"net"

	"sync/atomic"
)

type Afec struct {
	recv
	send
}

func NewAfec(rawConn net.Conn) *Afec {
	var a = &afec{
		rawConn: rawConn,
	}

	var f = &Afec{}
	f.send = newSend(a)
	f.recv = newRecv(a, 4)
	return f
}

type afec struct {
	rawConn net.Conn

	recvCnt, lossCnt atomic.Uint32
}

func (a *afec) pl() (pl float64) {
	l, r := a.lossCnt.Swap(0), a.recvCnt.Swap(0)

	if r == 0 || l == 0 {
		return 0
	} else {
		return float64(l) / float64(l+r)
	}
}

// dynamic fec algorithm
//
//	limit：
//	parityBlocks==0, dataBlocks == 1
//	parityBlocks==1, dataBlocks ∈ [1, 63]
//	parityBlocks> 1, dataBlocks 1
//
//go:noinline
func (a *afec) fec(pl float64) (dataBlocks, parityBlocks uint8) {
	if debug {
		defer func() {
			var ok = false
			if parityBlocks == 0 && dataBlocks == 1 {
				ok = true
			} else if parityBlocks == 1 && (dataBlocks >= 1 && dataBlocks <= 63) {
				ok = true
			} else if parityBlocks > 1 && dataBlocks == 1 {
				ok = true
			}
			if !ok {
				panic(fmt.Sprintf("pl %f dataBlocks %d parityBlocks %d", pl, dataBlocks, parityBlocks))
			}
		}()
	}

	return DefaultDataBlocks, 1
}

const DefaultDataBlocks = 16
