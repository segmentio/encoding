//go:generate go run equal_fold_asm.go -out equal_fold_amd64.s -stubs equal_fold_amd64.go
package ascii

import (
	"unsafe"
)

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
		const mask = 0xDFDFDFDFDFDFDFDF

		if (*(*uint64)(p) & mask) != (*(*uint64)(q) & mask) {
			return false
		}

		p = unsafe.Pointer(uintptr(p) + 8)
		q = unsafe.Pointer(uintptr(q) + 8)
		n -= 8
	}

	if n >= 4 {
		const mask = 0xDFDFDFDF

		if (*(*uint32)(p) & mask) != (*(*uint32)(q) & mask) {
			return false
		}

		p = unsafe.Pointer(uintptr(p) + 4)
		q = unsafe.Pointer(uintptr(q) + 4)
		n -= 4
	}

	switch n {
	case 3:
		x := uint32(*(*uint16)(p)) | uint32(*(*uint8)(unsafe.Pointer(uintptr(p) + 2)))
		y := uint32(*(*uint16)(q)) | uint32(*(*uint8)(unsafe.Pointer(uintptr(q) + 2)))
		return (x & 0xDFDFDF) == (y & 0xDFDFDF)
	case 2:
		return (*(*uint16)(p) & 0xDFDF) == (*(*uint16)(q) & 0xDFDF)
	case 1:
		return (*(*uint8)(p) & 0xDF) == (*(*uint8)(q) & 0xDF)
	default:
		return true
	}
}

func HasPrefixFoldString(s, prefix string) bool {
	return len(s) >= len(prefix) && EqualFoldString(s[:len(prefix)], prefix)
}

func HasSuffixFoldString(s, prefix string) bool {
	return len(s) >= len(prefix) && EqualFoldString(s[len(s)-len(prefix):], prefix)
}
