// +build !amd64

package ascii

import (
	"unsafe"
)

//go:nosplit
func validPrint16(s *byte, n uintptr) int {
	p := unsafe.Pointer(s)
	i := uintptr(0)

	for n > 0 {
		x := *(*uint64)(unsafe.Pointer(uintptr(p) + i))
		y := *(*uint64)(unsafe.Pointer(uintptr(p) + i + 8))

		if hasLess64(x, 0x20) || hasMore64(x, 0x7e) || hasLess64(y, 0x20) || hasMore64(y, 0x7e) {
			return 0
		}

		i += 16
		n--
	}

	return 1
}
