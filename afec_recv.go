package afec

import (
	"fmt"
	"io"
)

// require 256%rgroups == 0
const rgroups = uint8(4)

func init() {
	if rgroups > 127 || 256%uint16(rgroups) != 0 {
		panic("invalid rgroups")
	}
}

type recv struct {
	*afec

	rgs           uint8
	groups        []rgroup
	directReadIdx int8 // if >=0 valid
}

func newRecv(a *afec, rgs uint8) recv {
	if 256%uint16(rgs) != 0 ||
		rgs == 0 ||
		rgs > 127 {
		panic("invalid rgs")
	}

	return recv{
		afec:   a,
		rgs:    rgs,
		groups: make([]rgroup, rgs),
	}
}

func (r *recv) Read(b []byte) (n int, err error) {
	if r.directReadIdx > 0 {

		// 初始化组
		g := &r.groups[r.directReadIdx]
		g.groupLen = g.restore.glen()
		g.groupIdx = g.restore.gid()
		g.recvLen = 1

		if g.restore.isDataType() {
			r.directReadIdx = -1
			g.dataLen = 1
			return g.readDirect(b)
		}
	}

	var p Pack = b[: len(b)+HdrSize : len(b)+HdrSize]

	var maxLen int
	for {
		n, err = r.rawConn.Read(p[:cap(p)])
		if err != nil {
			return 0, err
		}
		maxLen = max(maxLen, n)

		if p = p[:n]; !p.valid() {
			return 0, fmt.Errorf("invalid pack: % X", p)
		}
		gidx := p.gid()
		g := &r.groups[gidx%rgroups]

		if g.groupIdx == gidx {
			if g.recvLen > g.groupLen { // 可以取等
				continue
			}

			g.restore = xor(g.restore, p)
		} else {
			if g.dataLen != g.groupLen {
				if g.recvLen == g.groupLen &&
					g.dataLen+1 == g.groupLen {

					// 恢复丢失数据包，此时p可能是任意类型数据包
					r.directReadIdx = int8(gidx % rgroups)
					clear(p[n:max(n, maxLen)])
					return g.reconstruct(p)
				} else {
					r.lossCnt.Add(uint32(g.groupLen - g.dataLen))
				}
			}

			// 组的第一个包，需要设置g.restore，并初始化参数
			g.restore = cpyclr(p, g.restore)
			g.groupLen = p.glen()
			g.groupIdx = gidx
			g.recvLen = 0
			g.dataLen = 0
		}

		g.recvLen += 1
		if p.isDataType() {
			g.dataLen += 1

			n = len(p) - HdrSize

			// 确保 p[n:] 的数据为 0
			if maxLen > n {
				clear(p[n:maxLen])
			}

			return n, nil
		}
	}
}

type rgroup struct {
	groupIdx uint8
	groupLen uint8
	recvLen  uint8
	dataLen  uint8

	restore Pack
}

// reconstruct 恢复丢失的包
func (g *rgroup) reconstruct(p Pack) (n int, err error) {
	if cap(p) < len(g.restore) {
		return 0, io.ErrShortBuffer
	} else {
		m := max(len(p), len(g.restore))
		if len(g.restore) < m {
			n = len(g.restore)
			g.restore = grow(g.restore, m)

			if debug && !isEmpty(g.restore[n:m]) {
				panic(fmt.Sprintf("expect zero: % X", g.restore[n:m]))
			}

			rawswap(p, g.restore)
			clear(p[n:])
			p = p[:n]
		} else {
			n = len(p)
			p = p[:m]

			if debug && !isEmpty(p[n:m]) {
				panic(fmt.Sprintf("expect zero: % X", p[n:m]))
			}

			rawswap(p, g.restore)
			clear(g.restore[n:])
			g.restore = g.restore[:n]
		}
	}

	// 解决尾0问题
	for n = len(p) - 1; n >= 0; n-- {
		if p[n] != 0 {
			break
		}
	}
	p = p[:n-HdrSize+1]

	p.clearTail()
	return len(p), nil
}

func (g *rgroup) readDirect(b []byte) (n int, err error) {
	n = len(g.restore) - HdrSize
	if len(b) < n {
		return 0, io.ErrShortBuffer
	}

	return copy(b[:n], g.restore[:n]), nil
}
