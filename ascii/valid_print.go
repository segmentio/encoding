//go:generate go run valid_print_asm.go -out valid_print_amd64.s -stubs valid_print_amd64.go
package ascii

import "unsafe"

// Valid returns true if b contains only printable ASCII characters.
func ValidPrint(b []byte) bool {
	return len(b) == 0 || validPrint(unsafe.Pointer(&b), uintptr(len(b)))
}

// ValidString returns true if s contains only printable ASCII characters.
func ValidPrintString(s string) bool {
	return len(s) == 0 || validPrint(unsafe.Pointer(&s), uintptr(len(s)))
}

// ValidBytes returns true if b is an ASCII character.
func ValidPrintByte(b byte) bool {
	return 0x20 <= b && b <= 0x7e
}

// ValidBytes returns true if b is an ASCII character.
func ValidPrintRune(r rune) bool {
	return 0x20 <= r && r <= 0x7e
}

//go:nosplit
func validPrint(s unsafe.Pointer, n uintptr) bool {
	p := *(*unsafe.Pointer)(s)
	i := uintptr(0)

	if n >= 16 {
		if asm.validPrint16((*byte)(p), n/16) == 0 {
			return false
		}
		i = ((n / 16) * 16)
	}

	if (n - i) >= 8 {
		x := *(*uint64)(unsafe.Pointer(uintptr(p) + i))
		if hasLess64(x, 0x20) || hasMore64(x, 0x7e) {
			return false
		}
		i += 8
	}

	if (n - i) >= 4 {
		x := *(*uint32)(unsafe.Pointer(uintptr(p) + i))
		if hasLess32(x, 0x20) || hasMore32(x, 0x7e) {
			return false
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
		return true
	}
	return !(hasLess32(x, 0x20) || hasMore32(x, 0x7e))
}

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
