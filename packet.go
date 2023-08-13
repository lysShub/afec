package afec

const (
	PackHdrSize  = 4
	MiniPackSize = PackHdrSize + 1
)

type Pack []byte

/*
	Upack{
		Payload: {
			data(nB)
		}

		Tail-Header: {
			Tail-Val(2b)      : indicate tail zero replace value, 1 or 2
			Group-Len(8b)     : data block length in a group
			Group-Idx(1B)     : group id, dentify the group, cycle 0-255
			Lossy-Perc(1B)    : loss percentage peer link
			Group-Flag(1B)    : flag, always not 0
		}
	}
*/

type Flag uint8

func (f Flag) IsParity() bool {
	return f == ParityBlock
}

const (
	_ Flag = iota
	ParityBlock
)

func (f Pack) Valid() bool {
	return len(f) >= MiniPackSize && f.Flag() != 0
}

func (f Pack) TailVal() uint8 {
	_ = f[MiniPackSize]
	return f[len(f)-1] >> 6
}

func (f Pack) SetTailVal(v uint8) {
	_ = f[MiniPackSize]
	f[len(f)-1] |= v << 6
}

func (f Pack) GroupDataLen() uint8 {
	_ = f[MiniPackSize]
	return f[len(f)-4] & 0b00111111
}

func (f Pack) SetGroupDataLen(n uint8) {
	_ = f[MiniPackSize]
	f[len(f)-4] |= n & 0b00111111
}

func (f Pack) GroupIdx() uint8 {
	_ = f[MiniPackSize]
	return f[len(f)-3]
}

func (f Pack) SetGroupIdx(u uint8) {
	_ = f[MiniPackSize]
	f[len(f)-3] = u
}

func (f Pack) Flag() Flag {
	_ = f[MiniPackSize]
	return Flag(f[len(f)-2])
}

func (f Pack) SetFlag(flag Flag) {
	_ = f[MiniPackSize]
	f[len(f)-2] = byte(flag)
}

func (f Pack) PL() Float8 {
	_ = f[MiniPackSize]
	return Float8(f[len(f)-1])
}

func (f Pack) SetPL(pl float64) {
	_ = f[MiniPackSize]
	f[len(f)-1] = byte(NewFloat8(pl))
}
