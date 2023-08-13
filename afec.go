package afec

import (
	"errors"
	"fmt"
	"net"

	"github.com/tmthrgd/go-memset"
)

// Algo a-fec algorithm, dataBlocks can't be 0.
type Algo func(pl float64) (dataBlocks, parityBlocks uint8)

type Afec struct {
	// a datagram conn
	rawConn net.Conn

	// for write
	algo      Algo
	globalInc cnt[uint8] // global group id cycle increase
	groupsInc cnt[uint8] // next mutiGroups index, 确保数据包交叉分配在不同组中
	wgroups   [maxMixGroups]wgroupState

	// read helper
	ring       Ring
	fecReGhash uint8
	fecReFlag  bool //
	derep      Derep

	// for read
	readies    uint8
	groupsHead cnt[uint8]
	rgroups    [maxMixGroups + 1]rgroupState

	// statistic
	pl *Statistic
}

const maxMixGroups = 5

// wgroupState write group state
type wgroupState struct {
	gIdx       uint8 // this group id
	dataBlocks uint8 // this group data length
	sendLen    uint8
	working    bool // this group is working

	parityBlock Pack // record this group parity data
	tmpDataBuff Pack // use for memcpy
}

func (w *wgroupState) tail() bool {

	return w.sendLen >= w.dataBlocks
}

func (w *wgroupState) Xor(p Pack) {
	// Header field Group-Len should be xor.
	w.parityBlock = xor(p[:len(p)-PackHdrSize-1], w.parityBlock)
}

type rgroupState struct {
	gIdx       uint8
	dataBlocks uint8
	recvLen    uint8
	state      uint8
	working    bool

	tmpPack Pack
}

type rstate uint8

const (
	idle  rstate = iota // 空闲, 没有有效数据
	work                // 正在接收
	ready               // 接收到dataBlocks个数据
)

func (g *rgroupState) init() {}

func (g *rgroupState) fini() {}

func (g *rgroupState) reconstruct(b []byte) (int, error) {
	return 0, nil
}

func (g *rgroupState) xor(p Pack) {
	g.tmpPack = xor(p[:len(p)-PackHdrSize-1], g.tmpPack)
}

// NewAfec
func NewAfec(conn net.Conn, algo Algo) *Afec {
	var f = &Afec{
		rawConn: conn,

		pl: &Statistic{},
	}

	return f
}

func (f *Afec) updateParity() {
	/*
		todo:
			实现有缺陷, 这样每个组还是一个ParityBlock

			如果冗余比小于1：
				每个组一个ParityBlock，只需调整每个组的DataBlocks个数
			如果冗余比大于等于1：
				不再需要ParityBlock, 只需要重复发送DataBlocks

		如果冗余比是1.2, 可不可以重复发送一个DataBlock，也将数据包放入一个
	*/

	// 在每个组结束时，更新groupState

	datas, parities := f.algo(f.pl.PL())
	if datas == 0 {
		datas = 1
	}
	if parities > 5 {
		factor := 5 / float64(parities)
		datas = uint8(factor * float64(datas))
		parities = 5
	}

	/*
		更新策略：
			因为algo中得出的Parities决定了mutiGroup的大小，所以
				如果new-parities小于等于当前的working-groups数量，将不会更新, 新的block-datas也不会被设置
				如果new-parities大于当前的working-groups数量，将会新增delta个new-datas的group
	*/

	workings := uint8(0)
	for _, g := range f.wgroups {
		if g.working {
			workings++
		}
	}

	for i := uint8(0); i < workings; {
		if !f.wgroups[i].working {
			f.wgroups[i].gIdx = f.globalInc.Inc()
			f.wgroups[i].working = true
			f.wgroups[i].dataBlocks = datas
			i++
		}
	}
}

func (f *Afec) nextWriteBlock() *wgroupState {
	// must can found
	for {
		idx := f.groupsInc.Inc()
		if f.wgroups[idx].working {
			f.wgroups[idx].sendLen++
			return &f.wgroups[idx]
		}
	}
}

func (f *Afec) Write(b []byte) (n int, err error) {
	/*

	 */

	if len(b) == 0 {
		return 0, nil
	}

	gs := f.nextWriteBlock()
	var p = f.getPack(gs, b)

	if n, err = f.rawConn.Write(p); err != nil {
		return 0, err // TODO: rollback state
	}
	gs.Xor(p)
	if gs.tail() {
		// TODO: no parity
		if err = f.writeParity(gs); err != nil {
			return 0, err
		}
		f.updateParity()
	}

	return n, nil
}

func (f *Afec) getPack(gs *wgroupState, b []byte) (p Pack) {
	// handel tail zero
	tailVal := uint8(0)
	if zeros := tailZeros(b); zeros > 0 {
		if b[len(b)-1-zeros] != 1 {
			tailVal = 1
		} else {
			tailVal = 2
		}
		for i := 0; i < zeros; i++ {
			b[len(b)-1-i] = tailVal
		}
	}

	if cap(b) >= len(b)+PackHdrSize {
		p = Pack(b[:len(b)+PackHdrSize])
	} else {
		for cap(gs.tmpDataBuff) < len(b)+PackHdrSize {
			gs.tmpDataBuff = append(gs.tmpDataBuff, 0)
		}

		copy(gs.tmpDataBuff[0:], b)
		p = Pack(gs.tmpDataBuff[0 : len(b)+PackHdrSize])
	}

	p.SetTailVal(tailVal)
	p.SetGroupDataLen(gs.dataBlocks)
	p.SetGroupIdx(gs.gIdx)
	p.SetPL(f.pl.PL())
	p.SetFlag(0)
	return p
}

func (f *Afec) writeParity(gs *wgroupState) error {
	n := len(gs.parityBlock) + PackHdrSize
	for cap(gs.parityBlock) < n {
		gs.parityBlock = append(gs.parityBlock, 0)
	}
	gs.parityBlock = gs.parityBlock[:n]

	// can't set ParityBlock's TailVal and DataLen
	gs.parityBlock.SetGroupIdx(gs.gIdx)
	gs.parityBlock.SetPL(f.pl.PL())
	gs.parityBlock.SetFlag(ParityBlock)

	_, err := f.rawConn.Write(gs.parityBlock)

	memset.Memset(gs.parityBlock, 0)
	gs.parityBlock = gs.parityBlock[:0]
	return err
}

func (s *Afec) nextReadBlock(gIdx uint8) *rgroupState {
	idleIdx, minIdx := -1, -1
	minGidx := uint8(0xff)
	for i := 0; i < maxMixGroups+1; i++ {
		if s.rgroups[i].working {
			if s.rgroups[i].gIdx == gIdx {
				return &s.rgroups[i]
			}
		} else {
			idleIdx = i
		}
		if s.rgroups[i].gIdx < minGidx {
			minGidx = s.rgroups[i].gIdx
			minIdx = i
		}
	}

	if idleIdx != -1 {
		s.rgroups[idleIdx].init()
		return &s.rgroups[idleIdx]
	} else {
		// discard group
		s.rgroups[minIdx].fini()
		return &s.rgroups[minIdx]
	}
}

func (s *Afec) Read(b []byte) (n int, err error) {
	/*
		Reconstruct
	*/
	// read Reconstruct
	if s.readies > 0 {
		for i := 0; i < maxMixGroups+1; i++ {
			if s.rgroups[i].state == 0 /* ready */ {
				return s.rgroups[i].reconstruct(b)
			}
		}
	}

	n, err = s.rawConn.Read(b)
	if err != nil {
		return 0, err
	}

	p := Pack(b[:n])
	if !p.Valid() {
		return 0, errors.New("can't parse pack")
	}

	g := s.nextReadBlock(p.GroupIdx())
	g.xor(p)
	g.recvLen++

	return 0, nil
}

func (s *Afec) Read1(b []byte) (n int, err error) {

start:
	if s.fecReFlag {
		s.fecReFlag = false

		n, err = s.ring.GetGroup(s.fecReGhash).Read(b)
		if err != nil {
			return 0, nil
		} else {
			s.pl.Stat(Pack(b[:n]).GroupIdx())
			return n - MiniPackSize, nil
		}
	} else {

		if n, err = s.rawConn.Read(b); err != nil {
			return 0, err
		} else if n < MiniPackSize {
			return 0, errors.New("xxx")
		}

		p := Pack(b[:n])
		if !p.Flag().IsParity() {
			if p.GroupDataLen() > 1 {
				// fec
				rec, skip := s.ring.GetGroup(p.GroupIdx()).Do1(p)
				if rec {
					s.fecReGhash = p.GroupIdx()
					s.fecReFlag = true
				}

				if skip {
					goto start
				} else {
					s.pl.Stat(p.GroupIdx())
					return n - MiniPackSize, nil
				}
			} else {
				// 非fec, 重复发包
				if s.derep.Skip(p.GroupIdx()) {
					goto start
				}

				s.pl.Stat(p.GroupIdx())
				return n - MiniPackSize, nil
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
