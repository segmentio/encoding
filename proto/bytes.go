package proto

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

var bytesCodec = codec{
	wire:   varlen,
	size:   sizeOfBytes,
	encode: encodeBytes,
	decode: decodeBytes,
}

func sizeOfBytes(p unsafe.Pointer, flags flags) int {
	if p != nil {
		if v := *(*[]byte)(p); v != nil || flags.has(wantzero) {
			return sizeOfVarlen(len(v))
		}
	}
	return 0
}

func encodeBytes(b []byte, p unsafe.Pointer, flags flags) (int, error) {
	if p != nil {
		if v := *(*[]byte)(p); v != nil || flags.has(wantzero) {
			n, err := encodeVarint(b, uint64(len(v)))
			if err != nil {
				return n, err
			}
			c := copy(b[n:], v)
			n += c
			if c < len(v) {
				err = io.ErrShortBuffer
			}
			return n, err
		}
	}
	return 0, nil
}

func decodeBytes(b []byte, p unsafe.Pointer, _ flags) (int, error) {
	v, n, err := decodeVarlen(b)
	pb := (*[]byte)(p)
	if *pb == nil {
		*pb = make([]byte, 0, len(v))
	}
	*pb = append((*pb)[:0], v...)
	return n, err
}

func makeBytes(p unsafe.Pointer, n int) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(p),
		Len:  n,
		Cap:  n,
	}))
}

func isZeroBytes(b []byte) bool {
	for i := range b {
		if b[i] != 0 {
			return false
		}
	}
	return true
}

func byteArrayCodecOf(t reflect.Type, seen map[reflect.Type]*codec) *codec {
	n := t.Len()
	c := &codec{
		wire:   varlen,
		size:   byteArraySizeFuncOf(n),
		encode: byteArrayEncodeFuncOf(n),
		decode: byteArrayDecodeFuncOf(n),
	}
	seen[t] = c
	return c
}

func byteArraySizeFuncOf(n int) sizeFunc {
	size := sizeOfVarlen(n)
	return func(p unsafe.Pointer, flags flags) int {
		if p != nil && (flags.has(wantzero) || !isZeroBytes(makeBytes(p, n))) {
			return size
		}
		return 0
	}
}

func byteArrayEncodeFuncOf(n int) encodeFunc {
	return func(b []byte, p unsafe.Pointer, flags flags) (int, error) {
		if p != nil {
			if v := makeBytes(p, n); flags.has(wantzero) || !isZeroBytes(v) {
				return encodeBytes(b, unsafe.Pointer(&v), noflags)
			}
		}
		return 0, nil
	}
}

func byteArrayDecodeFuncOf(n int) decodeFunc {
	return func(b []byte, p unsafe.Pointer, _ flags) (int, error) {
		v, r, err := decodeVarlen(b)
		if err == nil {
			if copy(makeBytes(p, n), v) != n {
				err = fmt.Errorf("cannot decode byte sequence of size %d into byte array of size %d", len(v), n)
			}
		}
		return r, err
	}
}
