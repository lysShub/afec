package afec

type send struct {
	*afec

	// todo：
	// 可以用一个atomic表示，write的时候只需执行一次

	groupIdx uint8

	dataBlocks, parityBlocks uint8
	sendLen                  uint8

	parityBlock Pack
}

func newSend(a *afec) send {
	var s = send{
		afec:     a,
		groupIdx: 1,
	}
	s.dataBlocks, s.parityBlocks = s.algo(0)
	return s
}

func (s *send) next(dataBlocks, parityBlocks uint8) {
	if parityBlocks > dataBlocks {
		if parityBlocks%dataBlocks > 0 {
			parityBlocks = 1
		}
		parityBlocks += parityBlocks / dataBlocks
		dataBlocks = 1
	}

	s.groupIdx++
	s.dataBlocks, s.parityBlocks = dataBlocks, parityBlocks
	s.sendLen = 0
	s.parityBlock = s.parityBlock[:0]
}

func (s *send) tail() bool { return s.sendLen >= s.dataBlocks }

func (s *send) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}

	pl := s.pl()
	var p Pack
	{
		if n := len(b) + HdrSize; n > cap(b) {
			p = make([]byte, n)
			copy(p, b)
		} else {
			p = b[:n]
		}

		p.SetGroupDataLen(s.dataBlocks)
		p.SetGroupIdx(s.groupIdx)
		p.SetPL(pl)
		p.SetFlag(0)
		if s.parityBlocks > 0 {
			s.parityBlock = s.parityBlock.Xor(p)
		}
		s.sendLen++
	}

	// 可以拆分
	n, err = s.rawConn.Write(p)
	if err != nil {
		return 0, err
	}

	if s.tail() {
		if s.parityBlocks > 0 {
			// pb := f.parityBlock[:len(f.parityBlock)+3]
			pb := append(s.parityBlock, 0, 0, 0)
			pb.SetGroupIdx(s.groupIdx)
			pb.SetPL(pl)
			pb.SetFlag(ParityBlock)
			for i := uint8(0); i < s.parityBlocks; i++ {
				n, err = s.rawConn.Write(pb)
				if err != nil {
					return 0, err
				}
			}
		}

		s.next(s.algo(pl))
	}

	return n, nil
}
