package afec

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
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
		{0.082, 0.082},
	}

	for i, s := range suits {
		f := newFloat8(s.data)
		r := f.get()

		require.InDelta(t, s.want, r, prec, i)
	}
}

func BenchmarkEncode(b *testing.B) {

	f8 := newFloat8(0)
	for i := 0; i < b.N; i++ {
		f8.put(0.832)
	}
}

func BenchmarkDecode(b *testing.B) {
	f8 := newFloat8(0.832)
	for i := 0; i < b.N; i++ {
		f8.get()
	}
}
