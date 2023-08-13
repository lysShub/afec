package afec

import (
	"math"
	"testing"
)

func TestFloat8(t *testing.T) {
	var suits = []struct {
		data float64
		want float64
	}{
		{-1.1, 0},
		{-1, 0},
		{-math.SmallestNonzeroFloat64, 0},
		{0.0, 0},
		{math.SmallestNonzeroFloat64, 0},
		{0.999, 0.9975},
		{1, 0.9975},
		{1.0001, 0.9975},
		{1.1, 0.9975},
		{math.NaN(), 0},
		{math.Inf(-1), 0},
		{math.Inf(1), 0.9975},
	}

	for _, s := range suits {
		f := NewFloat8(s.data)
		r := f.Get()

		delta := math.Abs(r - s.want)

		if delta > 1-0.9975 {
			t.Errorf("data=%v, want=%v, got=%v, delta=%v", s.data, s.want, r, delta)
		}
	}
}
