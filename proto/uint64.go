package proto

import "unsafe"

var uint64Codec = codec{
	wire:   fixed64,
	size:   sizeOfUint64,
	encode: encodeUint64,
	decode: decodeUint64,
}

func sizeOfUint64(p unsafe.Pointer, flags flags) int {
	if p != nil {
		if v := *(*uint64)(p); v != 0 || flags.has(wantzero) {
			return 8
		}
	}
	return 0
}

func encodeUint64(b []byte, p unsafe.Pointer, flags flags) (int, error) {
	if p != nil {
		if v := *(*uint64)(p); v != 0 || flags.has(wantzero) {
			return encodeFixed64(b, v)
		}
	}
	return 0, nil
}

func decodeUint64(b []byte, p unsafe.Pointer, _ flags) (int, error) {
	v, n, err := decodeFixed64(b)
	*(*uint64)(p) = uint64(v)
	return n, err
}
