package proto

import "unsafe"

var uint32Codec = codec{
	wire:   fixed32,
	size:   sizeOfUint32,
	encode: encodeUint32,
	decode: decodeUint32,
}

func sizeOfUint32(p unsafe.Pointer, flags flags) int {
	if p != nil {
		if v := *(*uint32)(p); v != 0 || flags.has(wantzero) {
			return 4
		}
	}
	return 0
}

func encodeUint32(b []byte, p unsafe.Pointer, flags flags) (int, error) {
	if p != nil {
		if v := *(*uint32)(p); v != 0 || flags.has(wantzero) {
			return encodeFixed32(b, v)
		}
	}
	return 0, nil
}

func decodeUint32(b []byte, p unsafe.Pointer, _ flags) (int, error) {
	v, n, err := decodeFixed32(b)
	*(*uint32)(p) = uint32(v)
	return n, err
}
