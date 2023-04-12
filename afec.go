package afec

import (
	"errors"
	"fmt"
	"net"

	"github.com/lysShub/afec/fec"
	"github.com/tmthrgd/go-memset"
)

// udp conn use fec

const MTU = 1450

type Afec struct {
	// config
	dataShards, parityShards uint8 // groupShards = groupDataShards + 1(groupParityShard)
	rawConn                  net.Conn

	setFec func()

	// write helper
	groupInc, groupIdx uint8
	blockBuff          Upack
	parityBuff         Upack

	// read helper
	ring       Ring
	fecReGhash uint8
	fecReFlag  bool
	derep      Derep

	// statistic
	pl *Statistic
}

// NewFudp
func NewFudp(conn net.Conn) *Afec {
	return &Afec{
		dataShards:   16,
		parityShards: 1,
		rawConn:      conn,

		groupInc:   1,
		blockBuff:  make(Upack, 0, 65536),
		parityBuff: make(Upack, 0, 65536),

		pl: &Statistic{},
	}
}

// SetFEC 设置fec的纠错能力。
// dataShards, parityShards中必须有一个值是1。
func (f *Afec) SetFEC(dataShards, parityShards uint8) error {
	if dataShards != 1 && parityShards != 1 {
		return errors.New("xxx")
	}

	f.setFec = func() {
		f.dataShards = dataShards
		f.parityShards = parityShards
	}
	return nil
}

func (f *Afec) wstate() (ghash uint8, tail bool) {
	if f.groupIdx == 0 && f.setFec != nil {
		f.setFec()
		f.setFec = nil
	}

	ghash = f.groupInc

	f.groupIdx = (f.groupIdx + 1) % uint8(f.dataShards)
	if f.groupIdx == 0 { // group end
		tail = true
		f.groupInc += 1
	}

	return ghash, tail
}

func (f *Afec) Write(b []byte) (n int, err error) {
	var p Upack
	if cap(b) > len(b)+HdrSize {
		p = Upack(b[:len(b)+HdrSize])

	} else {
		p = Upack(f.blockBuff[:len(b)+HdrSize])
		copy(p[0:], b[0:])
	}
	ghash, tail := f.wstate()
	p.SetFlag(Data)
	p.SetGroupInc(ghash)
	p.SetGroupDataLen(f.dataShards)

	if f.dataShards > f.parityShards {
		// fec; dataShards>1, parityShards==1
		f.parityBuff = fec.Xor(p, f.parityBuff)
		if _, err = f.rawConn.Write(p); err != nil {
			return 0, err
		}

		if tail {
			f.parityBuff.SetFlag(DataGroupTail)
			f.parityBuff.SetGroupInc(ghash)
			f.parityBuff.SetGroupDataLen(f.dataShards)
			if _, err = f.rawConn.Write(f.parityBuff); err != nil {
				return 0, err
			}

			memset.Memset(f.parityBuff, 0)
			f.parityBuff = f.parityBuff[:0]
		}
	} else {
		// repeat; dataShards==1; parityShards>=1
		for i := uint8(0); i < f.dataShards+f.parityShards; i++ {

			if _, err = f.rawConn.Write(p); err != nil {
				return 0, err
			}
		}
	}
	return 0, nil
}

func (s *Afec) Read(b []byte) (n int, err error) {
start:
	if s.fecReFlag {
		s.fecReFlag = false

		n, err = s.ring.GetGroup(s.fecReGhash).Read(b)
		if err != nil {
			return 0, nil
		} else {
			s.pl.Stat(Upack(b[:n]).GroupInc())
			return n - HdrSize, nil
		}
	} else {

		if n, err = s.rawConn.Read(b); err != nil {
			return 0, err
		} else if n < HdrSize {
			return 0, errors.New("xxx")
		}

		p := Upack(b[:n])
		if p.Flag().HasData() {
			if p.GroupDataLen() > 1 {
				// fec
				rec, skip := s.ring.GetGroup(p.GroupInc()).Do1(p)
				if rec {
					s.fecReGhash = p.GroupInc()
					s.fecReFlag = true
				}

				if skip {
					goto start
				} else {
					s.pl.Stat(p.GroupInc())
					return n - HdrSize, nil
				}
			} else {
				// 非fec, 重复发包
				if s.derep.Skip(p.GroupInc()) {
					goto start
				}

				s.pl.Stat(p.GroupInc())
				return n - HdrSize, nil
			}
		} else {
			// do else
			fmt.Println("do else")
		}
	}

	return 0, nil
}

// LP loss packet percent, [0,100].
func (f *Afec) LP() uint8 {
	return 0
}

// Ping RTT delay, milli second.
func (f *Afec) Ping() int {

	return 0
}

// Speed origin transmit speed, B/s.
func (f *Afec) Speed() (up, dn int) {
	return 0, 0
}
