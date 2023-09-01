package afec

import (
	"crypto/rand"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Mock UDP link, can mock delay、loss package、reverse.etc, by set pack transmit period.
// delay larger than 5s mean loss pack
func NewMockUDPConn(delayFn func() time.Duration, mss ...int) (c, s net.Conn) {
	var timeout = time.Second * 5

	var m int = 64
	if len(mss) != 0 {
		m = mss[0]
	}

	m1 := &simplexUDPLink{m: &sync.Mutex{}, timeout: timeout, delayFn: delayFn, mss: m}
	m1.c = sync.NewCond(m1.m)
	m2 := &simplexUDPLink{m: &sync.Mutex{}, timeout: timeout, delayFn: delayFn, mss: m}
	m2.c = sync.NewCond(m2.m)
	return &MockUDPConn{a: m1, b: m2}, &MockUDPConn{a: m2, b: m1}
}

type MockUDPConn struct {
	a, b *simplexUDPLink
}

func (c *MockUDPConn) Read(b []byte) (n int, err error)  { return c.a.Read(b) }
func (c *MockUDPConn) Write(b []byte) (n int, err error) { return c.b.Write(b) }
func (c *MockUDPConn) Close() error {
	c.a.Close()
	c.b.Close()
	return nil
}
func (c *MockUDPConn) LocalAddr() net.Addr                { return nil }
func (C *MockUDPConn) RemoteAddr() net.Addr               { return nil }
func (c *MockUDPConn) SetDeadline(t time.Time) error      { return nil }
func (c *MockUDPConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *MockUDPConn) SetWriteDeadline(t time.Time) error { return nil }

type simplexUDPLink struct {
	mss int

	tail, head *pkg // head is newest data

	timeout time.Duration

	delayFn func() time.Duration

	closed bool
	m      *sync.Mutex
	c      *sync.Cond
}

var errClosed = errors.New("use of closed network connection")

type pkg struct {
	da   []byte
	t    time.Time // be read time
	size int
	next *pkg
	prev *pkg
}

func (l *simplexUDPLink) Read(b []byte) (n int, err error) {

	for {
		l.m.Lock()

		if l.head == nil {
			l.c.Wait()
			if l.closed {
				l.m.Unlock()
				return 0, errClosed
			} else {
				l.m.Unlock()
			}
		} else {
			dt := time.Until(l.head.t)
			n = copy(b, l.head.da[:l.head.size])

			l.head = l.head.prev
			if l.head == nil {
				l.tail = nil
			}
			l.m.Unlock()

			time.Sleep(dt)
			return
		}

	}
}
func (l *simplexUDPLink) Write(b []byte) (n int, err error) {
	l.m.Lock()

	if l.closed {
		l.m.Unlock()
		return 0, errClosed
	}

	if D := l.delayFn(); D > l.timeout {
		l.m.Unlock()
		return len(b), nil // loss pack
	} else {
		T := time.Now().Add(D)
		E := &pkg{t: T, da: make([]byte, l.mss)}
		E.size = copy(E.da[:], b)

		if l.tail == nil {
			l.tail = E
			l.head = E
		} else {
			var e *pkg
			for e = l.tail; e != nil && e.t.UnixNano() > T.UnixNano(); e = e.next {
			}

			if e == nil {
				E.prev = l.head
				l.head.next = E
				l.head = E
			} else {

				if e.prev != nil {
					e.prev.next = E
					E.prev = e.prev

					E.next = e
					e.prev = E
				} else {
					E.next = l.tail
					l.tail.prev = E

					l.tail = E
				}
			}
		}
	}
	l.m.Unlock()

	l.c.Signal()
	return len(b), nil
}
func (l *simplexUDPLink) Close() error {
	l.m.Lock()
	l.closed = true
	l.m.Unlock()
	l.c.Signal()
	return nil
}
func (l *simplexUDPLink) LocalAddr() net.Addr                { return nil }
func (l *simplexUDPLink) RemoteAddr() net.Addr               { return nil }
func (l *simplexUDPLink) SetDeadline(t time.Time) error      { return nil }
func (l *simplexUDPLink) SetReadDeadline(t time.Time) error  { return nil }
func (l *simplexUDPLink) SetWriteDeadline(t time.Time) error { return nil }

/*
*
*    test mock udp
*
 */
func Test_MockUDPConn_Base(t *testing.T) {
	var f = func() time.Duration { return 0 }

	s, r := NewMockUDPConn(f)
	go func() { // Ping-Pong
		var rda = make([]byte, 8)
		for {
			n, err := s.Read(rda)
			require.NoError(t, err)
			_, err = s.Write(rda[:n])
			require.NoError(t, err)
		}
	}()

	time.Sleep(time.Second)

	var sda, rda = make([]byte, 8), make([]byte, 8)
	for i := 0; i < 10; i++ {
		rand.Read(sda)
		l1, err := r.Write(sda)
		require.NoError(t, err)

		l2, err := r.Read(rda)
		require.NoError(t, err)

		require.Equal(t, sda[:l1], rda[:l2])
	}
}

func Test_MockUDPConn_Delay(t *testing.T) {
	i := 0

	var f = func() time.Duration {
		return time.Millisecond * time.Duration(int(i)*10)
	}

	s, c := NewMockUDPConn(f)
	var start = time.Now()

	for ; i < 32; i++ {
		c.Write([]byte{byte(i)})
	}

	var lv byte = 0
	var rda = make([]byte, 8)
	for j := 0; j < 32; j++ {
		n, err := s.Read(rda)
		require.NoError(t, err)
		require.Equal(t, n, 1)

		if !(lv <= rda[0]) {
			t.Fatal(lv, rda[0])
		}
		lv = rda[0]
	}

	d := time.Since(start)
	if d < time.Millisecond*(320-30) || d > time.Millisecond*(320+30) {
		t.Log(d)
		t.FailNow()
	}

}

func Test_MockUDPConn_Reverse(t *testing.T) {
	var i = 31
	var f = func() time.Duration {
		var r = time.Millisecond * 100 * time.Duration(i)
		i--
		return r
	}

	s, r := NewMockUDPConn(f)

	go func() {
		for j := 31; j >= 0; j-- {
			s.Write([]byte{byte(j)})
		}
	}()

	var rs []byte = make([]byte, 0, 32)
	var rda = make([]byte, 8)
	for i := 0; i < 32; i++ {
		n, err := r.Read(rda)
		require.NoError(t, err)
		require.Equal(t, n, 1)

		rs = append(rs, rda[0])
	}

	require.Equal(t, rs, []byte{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	})
}

func Test_MockUDPConn_LossPacket(t *testing.T) {
	var i = 31
	var f = func() time.Duration {
		var r = time.Millisecond * 100 * time.Duration(i)
		if i%5 == 0 {
			r = time.Minute
		}
		i--
		return r
	}

	s, r := NewMockUDPConn(f)
	for j := 31; j >= 0; j-- {
		s.Write([]byte{byte(j)})
	}

	var rs []byte = make([]byte, 0, 32-7)
	var rda = make([]byte, 8)
	for i := 0; i < 32-7; i++ {
		n, err := r.Read(rda)
		require.NoError(t, err)
		require.Equal(t, n, 1)

		rs = append(rs, rda[0])
	}

	require.Equal(t, rs, []byte{
		1, 2, 3, 4, 6, 7, 8, 9, 11, 12, 13, 14, 16, 17, 18, 19, 21, 22, 23, 24, 26, 27, 28, 29, 31,
	})
}
