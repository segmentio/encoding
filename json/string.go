package json

import (
	"reflect"
	"unsafe"
)

const (
	lsb = 0x0101010101010101
	msb = 0x8080808080808080
)

// simpleString checks whether `s` is made up of chars in the range [0x20, 0x7F]
// and doesn't include a double quote or backslash. If the escapeHTML mode is
// enabled, `s` cannot contain <, > or & either.
func simpleString(s string, escapeHTML bool) bool {
	chunks := stringToUint64(s)
	for _, n := range chunks {
		// combine masks before checking for the MSB of each byte. We include
		// `n` in the mask to check whether any of the *input* byte MSBs were
		// set (i.e. the byte was outside the ASCII range).
		mask := n | below(n, 0x20) | contains(n, '"') | contains(n, '\\')
		if escapeHTML {
			mask |= contains(n, '<') | contains(n, '>') | contains(n, '&')
		}
		if (mask & msb) != 0 {
			return false
		}
	}

	for i := len(chunks) * 8; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c > 0x7f || c == '"' || c == '\\' || (escapeHTML && (c == '<' || c == '>' || c == '&')) {
			return false
		}
	}

	return true
}

// expand puts the specified byte into each of the 8 bytes of a uint64.
func expand(b byte) uint64 {
	return lsb * uint64(b)
}

// below return a mask that can be used to determine if any of the bytes
// in `n` are below `b`. If a byte's MSB is set in the mask then that byte was
// below `b`.
//
// The result is only valid if `b` and each byte in `n` is below 0x80.
func below(n uint64, b byte) uint64 {
	return n - expand(b)
}

// contains returns a mask that can be used to determine if any of the
// bytes in `n` are equal to `b`. If a byte's MSB is set in the mask then
// that byte is equal to `b`.
//
// The result is only valid if `b` and each byte in `n` is below 0x80.
func contains(n uint64, b byte) uint64 {
	return (n ^ expand(b)) - lsb
}

func stringToUint64(s string) []uint64 {
	return *(*[]uint64)(unsafe.Pointer(&reflect.SliceHeader{
		Data: ((*reflect.StringHeader)(unsafe.Pointer(&s))).Data,
		Len:  len(s) / 8,
		Cap:  len(s) / 8,
	}))
}
