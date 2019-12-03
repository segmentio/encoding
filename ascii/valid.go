package ascii

import "unsafe"

// Valid returns true if b contains only ASCII characters.
func Valid(b []byte) bool {
	return valid(unsafe.Pointer(&b), uintptr(len(b)))
}

// ValidString returns true if s contains only ASCII characters.
func ValidString(s string) bool {
	return valid(unsafe.Pointer(&s), uintptr(len(s)))
}

//go:nosplit
func valid(s unsafe.Pointer, n uintptr) bool {
	if n == 0 {
		return true
	}

	i := uintptr(0)
	p := *(*unsafe.Pointer)(s)

	for (n - i) >= 8 {
		if ((*(*uint64)(unsafe.Pointer(uintptr(p) + i))) & 0x8080808080808080) != 0 {
			return false
		}
		i += 8
	}

	if (n - i) >= 4 {
		if ((*(*uint32)(unsafe.Pointer(uintptr(p) + i))) & 0x80808080) != 0 {
			return false
		}
		i += 4
	}

	if (n - i) >= 2 {
		if ((*(*uint16)(unsafe.Pointer(uintptr(p) + i))) & 0x8080) != 0 {
			return false
		}
		i += 2
	}

	if (n - i) >= 1 {
		if ((*(*uint8)(unsafe.Pointer(uintptr(p) + i))) & 0x80) != 0 {
			return false
		}
	}

	return true
}
