package proto

import (
	"encoding/binary"
	"io"
	"unsafe"
)

// EncodeTag encodes a pair of field number and wire type into a protobuf tag.
func EncodeTag(f FieldNumber, t WireType) uint64 {
	return uint64(f)<<3 | uint64(t)
}

// EncodeZigZag returns v as a zig-zag encoded value.
func EncodeZigZag(v int64) uint64 {
	return encodeZigZag64(v)
}

func encodeZigZag64(v int64) uint64 {
	return (uint64(v) << 1) ^ uint64(v>>63)
}

func encodeZigZag32(v int32) uint32 {
	return (uint32(v) << 1) ^ uint32(v>>31)
}

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
	return encodeVarint(b, encodeZigZag64(v))
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
