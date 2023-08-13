package afec

import (
	"math"
)

// Float8 stores [0,1) in 8 bits
//
// map to f(x) = e^(x) (x in (-6,0])
type Float8 uint8

const step = 6.0 / 256

func NewFloat8(val float64) Float8 {
	var r Float8
	r.Put(val)
	return r
}

func (f *Float8) Put(val float64) {
	const stepE = 1 / step
	if val > 0.9975 {
		val = 0.9975
	} else if val < 0 {
		val = 0
	}

	val = 1 - val

	val = -math.Log(val) * stepE
	*f = Float8(val)
}

func (f Float8) Get() float64 {
	return 1 - math.Pow(math.E, -float64(f)*step)
}
