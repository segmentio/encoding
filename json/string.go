package json

import (
	"reflect"
	"unsafe"
)

const (
	lsb = 0x0101010101010101
	msb = 0x8080808080808080
)

func expand(b byte) uint64 {
	return lsb * uint64(b)
}

func below(n uint64, b byte) uint64 {
	return n - expand(b)
}

func contains(n uint64, b byte) uint64 {
	return (n ^ expand(b)) - lsb
}

func simpleString(s string, escapeHTML bool) bool {
	chunks := stringToUint64(s)
	for _, n := range chunks {
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

func stringToUint64(s string) []uint64 {
	return *(*[]uint64)(unsafe.Pointer(&reflect.SliceHeader{
		Data: ((*reflect.StringHeader)(unsafe.Pointer(&s))).Data,
		Len:  len(s) / 8,
		Cap:  len(s) / 8,
	}))
}
