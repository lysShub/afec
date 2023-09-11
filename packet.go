package afec

const (
	HdrSize      = 3
	MiniPackSize = HdrSize + 1
)

type Pack []byte

/*

	// xor要包含Group-Len，因为Group-Len始终不为0

	Upack{
		Payload: {
			data(nB)
		}

		Tailer: {
			Group-Idx(1B)     : group id, dentify the group, cycle 0-255
			Lossy-Perc(1B)    : loss percentage peer link
			Group-Len(6b)     : data block length in a group
			Block-Type(1b)    : block type, 1 data-block, 0 parity-block
			Tail-Mark(1b)     : tail mark, data-block alway 1 （似乎是无效的，因为 Block-Type）
		}
	}

	异或时包括包头
	发送Parity时重置包头
	恢复后Block将包括包头
	（
	  恢复的Block头的参数是不可靠的、因为Parity主动重
	  置包头的值，只需确定最后一位不为0即可
	 ）
*/

func (p Pack) valid() bool {
	return len(p) >= MiniPackSize && p[len(p)-1]&0b10000000 != 0
}

func (p Pack) clearTail() {
	_ = p[HdrSize]
	p[len(p)-3], p[len(p)-2], p[len(p)-1] = 0, 0, 0
}

func (p Pack) gid() uint8 {
	_ = p[HdrSize]
	return p[len(p)-3]
}

// setGid set group id
func (p Pack) setGid(u uint8) {
	_ = p[HdrSize]
	p[len(p)-3] = u
}

func (p Pack) pl() float8 {
	_ = p[HdrSize]
	return float8(p[len(p)-2])
}

func (p Pack) setPL(pl float64) {
	_ = p[HdrSize]
	p[len(p)-2] = byte(newFloat8(pl))
}

func (p Pack) glen() uint8 {
	_ = p[HdrSize]
	return p[len(p)-1] & 0b00111111
}

// setGlen set group data block len
func (p Pack) setGlen(n uint8) {
	_ = p[HdrSize]
	p[len(p)-1] = (p[len(p)-1] & 0b11000000) + (n & 0b00111111)
}

// setDataType set block type data-block
func (p Pack) setDataType() {
	_ = p[HdrSize]
	// set tail-mark same time
	p[len(p)-1] = (p[len(p)-1] | 0b11000000)
}

// setParityType set block type prity-block
func (p Pack) setParityType() {
	_ = p[HdrSize]
	p[len(p)-1] = (p[len(p)-1] & 0b10111111)
}

func (p Pack) isDataType() bool {
	_ = p[HdrSize]
	return p[len(p)-1]&0b01000000 != 0
}

func (p Pack) grow(to int) Pack {
	tmp := make([]byte, to)
	copy(tmp, p)
	return Pack(tmp)
}
