package afec

import (
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Fudp_Fec_Base(t *testing.T) {
	var fecList = [][2]uint8{
		{2, 1},
		{1, 1},
		{1, 2},
	}

	for _, fec := range fecList {
		dataShards, parityShards := fec[0], fec[1]
		_ = float64(parityShards) / float64(dataShards+parityShards)

		t.Logf("%d/%d", dataShards, parityShards)

		var c, s = NewMockUDPConn(func(data []byte) time.Duration { return time.Millisecond * 50 })

		go func() {
			cf := NewAfec(c, nil)

			for i := 4; i < 10; i++ {
				var b = []byte{}
				for j := 0; j < i; j++ {
					b = append(b, byte(j))
				}

				_, err := cf.Write(b)
				require.NoError(t, err)
			}
		}()

		sf := NewAfec(s, nil)

		for i := 4; i < 10; i++ {
			var b = []byte{}
			for j := 0; j < i; j++ {
				b = append(b, byte(j))
			}

			var rb = make([]byte, 64)
			n, err := sf.Read(rb)
			require.NoError(t, err)
			require.Equal(t, b, rb[:n])
		}
	}
}

func Test_Fudp_Fec_Loss(t *testing.T) {
	var fecList = [][2]uint8{
		{2, 1},
		{1, 1},
		{1, 2},
	}

	for _, fec := range fecList {
		dataShards, parityShards := fec[0], fec[1]
		_ = float64(parityShards) / float64(dataShards+parityShards)

		t.Logf("%d/%d", dataShards, parityShards)

		var c, s = NewMockUDPConn(func(data []byte) time.Duration { return time.Millisecond * 50 })

		go func() {
			cf := NewAfec(c, nil)

			for i := 4; i < 10; i++ {
				var b = []byte{}
				for j := 0; j < i; j++ {
					b = append(b, byte(j))
				}

				_, err := cf.Write(b)
				require.NoError(t, err)
			}
		}()

		sf := NewAfec(s, nil)
		for i := 4; i < 10; i++ {
			var b = []byte{}
			for j := 0; j < i; j++ {
				b = append(b, byte(j))
			}

			var rb = make([]byte, 64)
			n, err := sf.Read(rb)
			require.NoError(t, err)
			require.Equal(t, b, rb[:n])
		}
	}
}

func TestCreate2(t *testing.T) {
	// var delta []float64

	// for i := float64(0); i < 1; i = i + 0.00001 {

	// 	f8 := Float8(0)
	// 	f8.Put(i)

	// 	delta = append(delta, f8.Get())
	// }

	// fh, err := os.OpenFile("a.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer fh.Close()
	// for _, v := range delta {
	// 	b := []byte{}
	// 	b = append(b, []byte(strconv.FormatFloat(v, 'f', 6, 64))...)

	// 	fh.Write(append(b, '\n'))
	// }
}

func Test_All_Profile(t *testing.T) {
	// 遍历所有支持的冗余度
	var tmp []float64

	for d := 1; d < 128; /* max is 256 */ d++ {
		for p := 0; p < 5; p++ {
			tmp = append(tmp, float64(p)/float64(d+p))
		}
	}

	sort.Float64s(tmp)

	var r []float64 = []float64{tmp[0]}
	for _, v := range tmp {
		if v == r[len(r)-1] {
			continue
		} else {
			r = append(r, v)
		}
	}

	fh, err := os.OpenFile(`D:\OneDrive\code\go\afec\a.txt`, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}

	for _, v := range r {
		fh.WriteString(fmt.Sprintf("%v\n", v))
	}
}

func Test_All_Profile2(t *testing.T) {
	// 遍历所有支持的冗余度
	var tmp []float64

	for d := 1; d < 128; /* max is 256 */ d++ {
		for p := 0; p < 2; p++ {
			tmp = append(tmp, float64(p)/float64(d+p))
		}
	}

	sort.Float64s(tmp)

	var r []float64 = []float64{tmp[0]}
	for _, v := range tmp {
		if v == r[len(r)-1] {
			continue
		} else {
			r = append(r, v)
		}
	}

	fh, err := os.OpenFile(`D:\OneDrive\code\go\afec\b.txt`, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}

	for _, v := range r {
		fh.WriteString(fmt.Sprintf("%v\n", v))
	}
}
