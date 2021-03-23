package ascii

import "unsafe"

// EqualFold is a version of bytes.EqualFold designed to work on ASCII input
// instead of UTF-8.
//
// When the program has guarantees that the input is composed of ASCII
// characters only, it allows for greater optimizations.
func EqualFold(a, b []byte) bool {
	return EqualFoldString(unsafeString(a), unsafeString(b))
}

func HasPrefixFold(s, prefix []byte) bool {
	return len(s) >= len(prefix) && EqualFold(s, prefix)
}

func HasSuffixFold(s, prefix []byte) bool {
	return len(s) >= len(prefix) && EqualFold(s[len(s)-len(prefix):], prefix)
}

// EqualFoldString is a version of strings.EqualFold designed to work on ASCII
// input instead of UTF-8.
//
// When the program has guarantees that the input is composed of ASCII
// characters only, it allows for greater optimizations.
func EqualFoldString(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	n := uintptr(len(a))
	p := *(*unsafe.Pointer)(unsafe.Pointer(&a))
	q := *(*unsafe.Pointer)(unsafe.Pointer(&b))
	r := unsafe.Pointer(uintptr(p) + n)

	if n >= 8 {
		k := (n / 8) * 8
		x := unsafe.Pointer(uintptr(p) + k)

		for uintptr(p) < uintptr(x) {
			const mask = 0xDFDFDFDFDFDFDFDF

			if (*(*uint64)(p) & mask) != (*(*uint64)(q) & mask) {
				return false
			}

			p = unsafe.Pointer(uintptr(p) + 8)
			q = unsafe.Pointer(uintptr(q) + 8)
		}
	}

	if (uintptr(r) - uintptr(p)) >= 4 {
		const mask = 0xDFDFDFDF

		if (*(*uint32)(p) & mask) != (*(*uint32)(q) & mask) {
			return false
		}

		p = unsafe.Pointer(uintptr(p) + 4)
		q = unsafe.Pointer(uintptr(q) + 4)
	}

	if (uintptr(r) - uintptr(p)) >= 2 {
		const mask = 0xDFDF

		if (*(*uint16)(p) & mask) != (*(*uint16)(q) & mask) {
			return false
		}

		p = unsafe.Pointer(uintptr(p) + 2)
		q = unsafe.Pointer(uintptr(q) + 2)
	}

	if uintptr(p) < uintptr(r) {
		const mask = 0xDF
		return (*(*uint8)(p) & mask) == (*(*uint8)(q) & mask)
	}

	return true
}

func HasPrefixFoldString(s, prefix string) bool {
	return len(s) >= len(prefix) && EqualFoldString(s[:len(prefix)], prefix)
}

func HasSuffixFoldString(s, prefix string) bool {
	return len(s) >= len(prefix) && EqualFoldString(s[len(s)-len(prefix):], prefix)
}
