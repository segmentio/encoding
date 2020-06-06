package proto

import (
	"unsafe"
)

var intCodec = codec{
	wire:   varint,
	size:   sizeOfInt,
	encode: encodeInt,
	decode: decodeInt,
}

func sizeOfInt(p unsafe.Pointer, flags flags) int {
	if p != nil {
		if v := *(*int)(p); v != 0 || flags.has(wantzero) {
			return sizeOfVarintZigZag(int64(v))
		}
	}
	return 0
}

func encodeInt(b []byte, p unsafe.Pointer, flags flags) (int, error) {
	if p != nil {
		if v := *(*int)(p); v != 0 || flags.has(wantzero) {
			return encodeVarintZigZag(b, int64(v))
		}
	}
	return 0, nil
}

func decodeInt(b []byte, p unsafe.Pointer, _ flags) (int, error) {
	v, n, err := decodeVarintZigZag(b)
	*(*int)(p) = int(v)
	return n, err
}
