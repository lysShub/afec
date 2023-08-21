package afec

import "io"

// 256%rgroups must equal 0
const rgroups = 4

type recv struct {
	*afec

	groups        [rgroups]groups
	directReadIdx int8 // >=0 valid
}

func newRecv(a *afec) recv {
	return recv{
		afec: a,
	}
}

func (f *recv) Read(b []byte) (n int, err error) {
	if f.directReadIdx > 0 {
		idx := f.directReadIdx
		f.directReadIdx = -1
		return f.groups[idx].readDirect(b)
	}

	for {
		n, err = f.rawConn.Read(b)
		if err != nil {
			return 0, err
		} else if n < MiniPackSize {
			// todo: maybe return error
			continue
		}
		p := Pack(b[:n])
		gidx := p.GroupIdx()
		xorLen := len(p) - PackHdrSize + 1
		g := &f.groups[gidx%rgroups]

		if g.groupIdx != gidx {
			if g.needReconstruct() {
				f.directReadIdx = int8(gidx % rgroups)

				return g.reconstruct(b)
			} else {
				f.lossCnt.Add(uint32(g.dataBlocks - g.recvLen))

				g.groupIdx = gidx
				g.parityBlock = cpyclr(p[:xorLen], g.parityBlock)
				if !p.Flag().IsParity() {
					g.recvLen = 1
					g.dataBlocks = p.GroupDataLen()

					return len(p) - PackHdrSize, nil
				}
			}
		} else {
			// the block alone same group
			g.parityBlock = g.parityBlock.Xor(p[:xorLen])

			if !p.Flag().IsParity() {
				g.recvLen += 1

				return len(p) - PackHdrSize, nil
			}
		}
	}
}

type groups struct {
	groupIdx   uint8
	dataBlocks uint8
	recvLen    uint8

	parityBlock Pack
}

func (g *groups) needReconstruct() bool {
	return g.recvLen+1 == g.dataBlocks
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
