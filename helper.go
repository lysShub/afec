package afec

import (
	"errors"

	"github.com/tmthrgd/go-memset"
)

const W = 64

type Ring [W]group

func (r *Ring) GetGroup(ghash uint8) *group {
	return &r[ghash%W]
}

type group struct {
	lossShard           []byte
	parityShards        []byte
	groupDataLen, count uint8
	groupInc            uint8
	finied              bool
}

func (g *group) Do(u Pack) (rec bool) {
	// 将增加组路由
	if g.groupInc != u.GroupIdx() {
		g.reset()
		g.groupInc = u.GroupIdx()
		g.groupDataLen = u.GroupDataLen()
	}
	if len(u) > cap(g.lossShard) {
		g.grow(cap(g.lossShard) - len(u))
	}

	// TODO: 通过GroupDataLen判断
	if u.Flag() == 0 /* DataGroupTail */ { // use counter
		n := copy(g.parityShards[0:cap(g.parityShards)], u)
		g.parityShards = g.parityShards[:n]
	} else {
		g.lossShard = xor(u, g.lossShard)
	}
	g.count += 1

	if g.count == g.groupDataLen &&
		len(g.parityShards) != 0 {

		g.lossShard = xor(u, g.lossShard)
		n := trim(g.lossShard)
		g.lossShard = g.lossShard[:n]
		return true
	}
	return false
}

func (g *group) Do1(u Pack) (rec, skip bool) {
	if g.groupInc != u.GroupIdx() {
		g.reset()
		g.groupInc = u.GroupIdx()
		g.groupDataLen = u.GroupDataLen()
	} else {
		if g.finied {
			return false, true
		}
	}
	if len(u) > cap(g.lossShard) {
		g.grow(cap(g.lossShard) - len(u))
	}

	if u.Flag() == 0 /* DataGroupTail */ { // use counter
		n := copy(g.parityShards[0:cap(g.parityShards)], u)
		g.parityShards = g.parityShards[:n]
		skip = true
	} else {
		g.lossShard = xor(u, g.lossShard)
	}
	g.count += 1

	if g.count == g.groupDataLen {
		g.finied = true

		if len(g.parityShards) != 0 {
			g.lossShard = xor(g.lossShard, g.parityShards)
			n := trim(g.lossShard)
			g.lossShard = g.lossShard[:n]
			rec = true
		}
	}

	return rec, skip
}

func (g *group) Read(b Pack) (int, error) {
	if len(g.lossShard) == 0 {
		return 0, nil
	} else {
		if len(b) < len(g.lossShard) {
			return 0, errors.New("msg buff too small")
		}
		defer func() {
			g.lossShard = g.lossShard[:0]
		}()
		return copy(b[0:], g.lossShard[0:]), nil
	}
}

func (g *group) reset() {
	g.finied = false
	memset.Memset(g.parityShards, 0)
	g.parityShards = g.parityShards[:0]
	memset.Memset(g.lossShard, 0)
	g.lossShard = g.lossShard[:0]
}

func (g *group) grow(delta int) {
	delta = delta - delta%128 + 128
	n := delta + cap(g.lossShard)

	t1 := make([]byte, 0, n)
	n1 := copy(t1[0:len(g.lossShard)], g.lossShard)
	g.lossShard = t1[:n1]

	t2 := make([]byte, 0, n)
	n2 := copy(t2[0:len(g.parityShards)], g.parityShards)
	g.parityShards = t2[:n2]
}

func trim(b []byte) int {
	n := len(b) - 1
	for i := n; i >= 0; i-- {
		if b[i] != 0 {
			return i
		}
	}
	return 0
}

// limit heap
type Derep [W]uint8

func (d *Derep) Skip(ghash uint8) bool {
	if d[ghash%W] != ghash {
		d[ghash%W] = ghash
		return false
	} else {
		return true
	}
}

type Statistic struct {
	b          [W]uint8
	loss, recv int
	prevPL     float64
}

func (s *Statistic) Stat(ghash uint8) {
	if s.dist(s.b[ghash%W], ghash) > W {
		s.loss += 1
	} else {
		s.recv += 1
	}
	s.b[ghash%W] = ghash

	s.reset()
}

func (s *Statistic) dist(x, y uint8) uint8 {
	if x > y {
		return min(x-y, 0xff+y-x)
	} else {
		return min(y-x, 0xff+x-y)
	}
}

func (s *Statistic) reset() {
	if s.loss > 0 && s.loss+s.recv > 100 {
		s.prevPL = float64(s.loss) / float64(s.recv)
		s.loss, s.recv = 0, 0
	}
}

func (s *Statistic) PL() float64 {
	s.reset()
	return s.prevPL
}

// cycle counter
type cnt[T uint8 | int] struct {
	val T
	max T
}

func NewInc[T uint8 | int](max T) *cnt[T] {
	return &cnt[T]{max: max}
}

func (i *cnt[T]) Val() T {
	return i.val % i.max
}

func (i *cnt[T]) Inc() (old T) {
	old = i.Val()
	i.val = (i.val + 1) % i.max
	return
}

func (i *cnt[T]) Dec() (old T) {
	old = i.Val()
	i.val = (i.val - 1) % i.max
	return
}

func (i *cnt[T]) Set(v T) {
	i.val = v
}
