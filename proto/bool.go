package proto

import (
	"unsafe"
)

var boolCodec = codec{
	wire:   varint,
	size:   sizeOfBool,
	encode: encodeBool,
	decode: decodeBool,
}

func sizeOfBool(p unsafe.Pointer, flags flags) int {
	if p != nil && *(*bool)(p) || flags.has(wantzero) {
		return 1
	}
	return 0
}

func encodeBool(b []byte, p unsafe.Pointer, flags flags) (int, error) {
	if p != nil && *(*bool)(p) || flags.has(wantzero) {
		return encodeVarint(b, 1)
	}
	return 0, nil
}

func decodeBool(b []byte, p unsafe.Pointer, _ flags) (int, error) {
	v, n, err := decodeVarint(b)
	*(*bool)(p) = v != 0
	return n, err
}
