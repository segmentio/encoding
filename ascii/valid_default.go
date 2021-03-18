// +build !amd64

package ascii

import (
	"unsafe"
)

//go:nosplit
func validPrint(s string) int {
	p := *(*unsafe.Pointer)(unsafe.Pointer(&s))
	n := uintptr(len(s))
	i := uintptr(0)

	if n == 0 {
		return 1
	}

	for (n - i) >= 8 {
		x := *(*uint64)(unsafe.Pointer(uintptr(p) + i))
		if hasLess64(x, 0x20) || hasMore64(x, 0x7e) {
			return 0
		}
		i += 8
	}

	if (n - i) >= 4 {
		x := *(*uint32)(unsafe.Pointer(uintptr(p) + i))
		if hasLess32(x, 0x20) || hasMore32(x, 0x7e) {
			return 0
		}
		i += 4
	}

	var x uint32
	switch n - i {
	case 3:
		x = 0x20000000 | uint32(*(*uint8)(unsafe.Pointer(uintptr(p) + i))) | uint32(*(*uint16)(unsafe.Pointer(uintptr(p) + i + 1)))<<8
	case 2:
		x = 0x20200000 | uint32(*(*uint16)(unsafe.Pointer(uintptr(p) + i)))
	case 1:
		x = 0x20202000 | uint32(*(*uint8)(unsafe.Pointer(uintptr(p) + i)))
	default:
		return 1
	}
	if hasLess32(x, 0x20) || hasMore32(x, 0x7e) {
		return 0
	}
	return 1
}
