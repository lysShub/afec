package afec

const HdrSize = 4

type Upack []byte

/*
    结构:
	Upack{
		PAYLOAD: {
			data(nB): 最多可以负载BLOCK-Size*8
		}

		TAIL: {
			Group-Inc(1B)    : group id, 表明所属组; 循环inc
			Shard-Len(1B)     : FEC组中的data shard个数, FEC最小值2
			Shard-Flag(1B)    : 标志属性, 非0
		}
	}
*/

type Flag uint8

func (f Flag) HasData() bool {
	return f == Data || f == DataGroupTail
}

const (
	Data Flag = iota
	DataGroupTail
)

func (f Upack) Flag() Flag {
	_ = f[len(f)-1]
	return Flag(f[len(f)-1])
}

func (f Upack) SetFlag(flag Flag) {
	_ = f[len(f)-1]
	f[len(f)-1] = byte(flag)
}

// data shards in a group, fec > 2, rep = 1
func (f Upack) GroupDataLen() uint8 {
	_ = f[len(f)-2]
	return f[len(f)-2]
}

func (f Upack) SetGroupDataLen(n uint8) {
	_ = f[len(f)-2]
	f[len(f)-2] = n
}

func (f Upack) GroupInc() uint8 {
	_ = f[len(f)-3]
	return f[len(f)-3]
}

func (f Upack) SetGroupInc(u uint8) {
	_ = f[len(f)-3]
	f[len(f)-3] = u
}
