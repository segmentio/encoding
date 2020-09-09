package proto

import "unsafe"

var int32Codec = codec{
	wire:   fixed32,
	size:   sizeOfInt32,
	encode: encodeInt32,
	decode: decodeInt32,
}

func sizeOfInt32(p unsafe.Pointer, flags flags) int {
	if p != nil {
		if v := *(*int32)(p); v != 0 || flags.has(wantzero) {
			return 4
		}
	}
	return 0
}

func encodeInt32(b []byte, p unsafe.Pointer, flags flags) (int, error) {
	if p != nil {
		if v := *(*int32)(p); v != 0 || flags.has(wantzero) {
			return encodeFixed32(b, encodeZigZag32(v))
		}
	}
	return 0, nil
}

func decodeInt32(b []byte, p unsafe.Pointer, _ flags) (int, error) {
	v, n, err := decodeFixed32(b)
	*(*int32)(p) = decodeZigZag32(v)
	return n, err
}
