package afec

const (
	PackHdrSize  = 4
	MiniPackSize = PackHdrSize + 1
)

type Pack []byte

/*

	// xor要包含Group-Len，因为Group-Len始终不为0

	Upack{
		Payload: {
			data(nB)
		}

		Tailer: {
			Group-Len(1B)     : data block length in a group
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

func (p Pack) Valid() bool {
	return len(p) >= MiniPackSize
}

func (p Pack) GroupDataLen() uint8 {
	_ = p[PackHdrSize]
	return p[len(p)-4]
}

func (p Pack) SetGroupDataLen(n uint8) {
	_ = p[PackHdrSize]
	p[len(p)-4] = n
}

func (p Pack) GroupIdx() uint8 {
	_ = p[PackHdrSize]
	return p[len(p)-3]
}

func (p Pack) SetGroupIdx(u uint8) {
	_ = p[PackHdrSize]
	p[len(p)-3] = u
}

func (p Pack) PL() Float8 {
	_ = p[PackHdrSize]
	return Float8(p[len(p)-2])
}

func (p Pack) SetPL(pl float64) {
	_ = p[PackHdrSize]
	p[len(p)-2] = byte(NewFloat8(pl))
}

func (p Pack) Flag() Flag {
	_ = p[PackHdrSize]
	return Flag(p[len(p)-1])
}

func (p Pack) SetFlag(flag Flag) {
	_ = p[PackHdrSize]
	p[len(p)-1] = byte(flag)
}

// Xor src xor to p
func (p Pack) Xor(src Pack) Pack {
	return xor(p, src[:len(src)-PackHdrSize+1])
}
