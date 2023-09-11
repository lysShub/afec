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
	s.dataBlocks, s.parityBlocks = a.fec(0)
	return s
}

func (s *send) nextGroup(dataBlocks, parityBlocks uint8) {
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
	var p Pack = b[:len(b)+HdrSize]
	{
		p.setGid(s.groupIdx)
		p.setPL(pl)
		p.setGlen(s.dataBlocks)
		p.setDataType()
		if s.parityBlocks > 0 {
			s.parityBlock = xor(s.parityBlock, p)
		}
		s.sendLen++
	}

	n, err = s.rawConn.Write(p)
	if err != nil {
		return 0, err
	}

	if s.tail() {
		if s.parityBlocks > 0 {
			s.parityBlock.setGid(s.groupIdx)
			s.parityBlock.setPL(pl)
			s.parityBlock.setGlen(s.dataBlocks)
			s.parityBlock.setParityType()
			for i := uint8(0); i < s.parityBlocks; i++ {
				n, err = s.rawConn.Write(s.parityBlock)
				if err != nil {
					return 0, err
				}
			}
		}

		s.nextGroup(s.fec(pl))
	}

	return n, nil
}
