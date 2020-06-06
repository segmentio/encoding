package proto

import (
	"encoding/binary"
	"io"
	"unsafe"
)

type encodeFunc = func([]byte, unsafe.Pointer, flags) (int, error)

func encodeVarint(b []byte, v uint64) (int, error) {
	i := 0
	for v >= 0x80 {
		if i >= len(b) {
			return i, io.ErrShortBuffer
		}
		b[i] = byte(v) | 0x80
		v >>= 7
		i++
	}
	if i >= len(b) {
		return i, io.ErrShortBuffer
	}
	b[i] = byte(v)
	return i + 1, nil
}

func encodeVarintZigZag(b []byte, v int64) (int, error) {
	return encodeVarint(b, uint64((v<<1)^(v>>63)))
}

func encodeFixed32(b []byte, v uint32) (int, error) {
	if len(b) < 4 {
		return 0, io.ErrShortBuffer
	}
	binary.LittleEndian.PutUint32(b, v)
	return 4, nil
}

func encodeFixed64(b []byte, v uint64) (int, error) {
	if len(b) < 8 {
		return 0, io.ErrShortBuffer
	}
	binary.LittleEndian.PutUint64(b, v)
	return 8, nil
}

func encodeTag(b []byte, f fieldNumber, t wireType) (int, error) {
	return encodeVarint(b, uint64(f)<<3|uint64(t))
}
