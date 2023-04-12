package afec

import (
	"testing"
)

func TestDerep(t *testing.T) {

	var d Derep

	var r []bool
	for i := 0; i < 0xff*2; i++ {
		r = append(r, d.Skip(uint8(i%0xff)))
	}

	t.Error(r)
}

func TestStatistic(t *testing.T) {
	var s = &Statistic{}
	r := s.dist(254, 5)
	t.Log(r)
}
