package afec

type float8 uint8

// float8(0b10000000).get()
const prec = float64(0.0040)

func newFloat8(v float64) (r float8) {
	r.put(v)
	return r
}

func (f *float8) put(v float64) {
	var r float8
	for i := 0; i < 8; i++ {
		v = v * 2
		if v >= 1 {
			r |= float8(1) << i
			v = v - 1
		}
	}
	*f = r
}

func (f float8) get() (r float64) {
	for i := 0; i < 8; i++ {
		if f&(float8(1)<<i) != 0 {
			r += float64(1) / (float64(uint(1) << (i + 1)))
		}
	}
	return r
}
