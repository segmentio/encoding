// +build !amd64

package ascii

import "unsafe"

//go:nosplit
func valid16(s *byte, n uintptr) int {
	p := unsafe.Pointer(s)
	i := uintptr(0)

	for n > 0 {
		lo := *(*uint64)(unsafe.Pointer(uintptr(p) + i))
		hi := *(*uint64)(unsafe.Pointer(uintptr(p) + i + 8))

		if (lo&0x8080808080808080) != 0 || (hi&0x8080808080808080) != 0 {
			return 0
		}

		i += 16
		n--
	}

	return 1
}
