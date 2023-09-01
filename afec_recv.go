package afec

import "io"

// 256%rgroups must equal 0
const rgroups = 4

type recv struct {
	*afec

	groups        [rgroups]groups
	directReadIdx int8 // if >=0 valid
}

func newRecv(a *afec) recv {
	return recv{
		afec: a,
	}
}

func (r *recv) Read(b []byte) (n int, err error) {
	if r.directReadIdx > 0 {
		idx := r.directReadIdx
		r.directReadIdx = -1
		return r.groups[idx].readDirect(b)
	}

	for {
		n, err = r.rawConn.Read(b)
		if err != nil {
			return 0, err
		} else if n < MiniPackSize {
			// todo: maybe return error
			continue
		}
		p := Pack(b[:n])
		gidx := p.GroupIdx()
		g := &r.groups[gidx%rgroups]

		if g.groupIdx != gidx {
			// belong a new group

			// todo: reconstruct and update pl by packet-len
			if g.needReconstruct() {
				r.directReadIdx = int8(gidx % rgroups)

				return g.reconstruct(b)
			} else {
				if debug && g.recvLen > g.groupLen {
					panic(int(g.groupLen) - int(g.recvLen))
				}
				r.lossCnt.Add(uint32(g.groupLen - g.recvLen))

				g.groupIdx = gidx
				g.parityBlock = cpyclr(p[:len(p)-HdrSize+1], g.parityBlock)
				if !p.Flag().IsParity() {
					g.recvLen = 1
					g.groupLen = p.GroupDataLen()

					return len(p) - HdrSize, nil
				}
			}
		} else {
			g.parityBlock = g.parityBlock.Xor(p)

			if !p.Flag().IsParity() {
				g.recvLen += 1

				return len(p) - HdrSize, nil
			}
		}
	}
}

type groups struct {
	groupIdx uint8
	groupLen uint8
	recvLen  uint8

	parityBlock Pack
}

func (g *groups) needReconstruct() bool {
	return g.recvLen+1 == g.groupLen
}

func (g *groups) reconstruct(p Pack) (n int, err error) {
	if len(p) < len(g.parityBlock)-1 {
		return 0, io.ErrShortBuffer
	}

	p, g.parityBlock = swap(p, g.parityBlock)

	for n = len(p) - 1; n >= 0; n-- {
		if p[n] != 0 {
			break
		}
	}
	return n, nil
}

func (g *groups) readDirect(b []byte) (n int, err error) {
	n = len(g.parityBlock) - 1
	if len(b) < n {
		return 0, io.ErrShortBuffer
	}
	copy(b[:n], g.parityBlock[:n])
	return n, nil
}
