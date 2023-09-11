//go:build !debug
// +build !debug

package afec

const debug = false

func isEmpty(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}
