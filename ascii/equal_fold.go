//go:generate go run equal_fold_asm.go -out equal_fold_amd64.s -stubs equal_fold_amd64.go
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
	c := byte(0)

	// Pre-check to avoid the other tests that would all evaluate to false.
	// For very small strings, this helps reduce the processing overhead.
	if n >= 8 {
		// If there is more than 32 bytes to copy, use the AVX optimized version,
		// otherwise the overhead of the function call tends to be greater than
		// looping 2 or 3 times over 8 bytes.
		if n >= 32 && asm.equalFoldAVX2 != nil {
			if asm.equalFoldAVX2((*byte)(p), (*byte)(q), n) == 0 {
				return false
			}
			k := (n / 16) * 16
			p = unsafe.Pointer(uintptr(p) + k)
			q = unsafe.Pointer(uintptr(q) + k)
			n -= k
		}
		for n >= 8 {
			c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 0))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 0))]
			c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 1))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 1))]
			c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 2))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 2))]
			c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 3))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 3))]
			c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 4))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 4))]
			c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 5))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 5))]
			c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 6))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 6))]
			c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 7))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 7))]

			if c != 0 {
				return false
			}

			p = unsafe.Pointer(uintptr(p) + 8)
			q = unsafe.Pointer(uintptr(q) + 8)
			n -= 8
		}
	}

	switch n {
	case 7:
		c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 6))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 6))]
		fallthrough
	case 6:
		c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 5))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 5))]
		fallthrough
	case 5:
		c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 4))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 4))]
		fallthrough
	case 4:
		c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 3))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 3))]
		fallthrough
	case 3:
		c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 2))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 2))]
		fallthrough
	case 2:
		c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 1))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 1))]
		fallthrough
	case 1:
		c |= lower[*(*uint8)(unsafe.Pointer(uintptr(p) + 0))] ^ lower[*(*uint8)(unsafe.Pointer(uintptr(q) + 0))]
	}

	return c == 0

}

func HasPrefixFoldString(s, prefix string) bool {
	return len(s) >= len(prefix) && EqualFoldString(s[:len(prefix)], prefix)
}

func HasSuffixFoldString(s, prefix string) bool {
	return len(s) >= len(prefix) && EqualFoldString(s[len(s)-len(prefix):], prefix)
}
