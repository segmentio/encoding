package proto

import "unsafe"

var int64Codec = codec{
	wire:   fixed64,
	size:   sizeOfInt64,
	encode: encodeInt64,
	decode: decodeInt64,
}

func sizeOfInt64(p unsafe.Pointer, flags flags) int {
	if p != nil {
		if v := *(*int64)(p); v != 0 || flags.has(wantzero) {
			return 8
		}
	}
	return 0
}

func encodeInt64(b []byte, p unsafe.Pointer, flags flags) (int, error) {
	if p != nil {
		if v := *(*int64)(p); v != 0 || flags.has(wantzero) {
			return encodeFixed64(b, encodeZigZag64(v))
		}
	}
	return 0, nil
}

func decodeInt64(b []byte, p unsafe.Pointer, _ flags) (int, error) {
	v, n, err := decodeFixed64(b)
	*(*int64)(p) = decodeZigZag64(v)
	return n, err
}
