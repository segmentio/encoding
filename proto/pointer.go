package proto

import (
	"reflect"
	"unsafe"
)

func pointerCodecOf(t reflect.Type, seen map[reflect.Type]*codec) *codec {
	p := new(codec)
	seen[t] = p
	c := codecOf(t.Elem(), seen)
	p.wire = c.wire
	p.size = pointerSizeFuncOf(t, c.size)
	p.encode = pointerEncodeFuncOf(t, c.encode)
	p.decode = pointerDecodeFuncOf(t, c.decode)
	return p
}

func pointerSizeFuncOf(t reflect.Type, size sizeFunc) sizeFunc {
	return func(p unsafe.Pointer, flags flags) int {
		if p != nil {
			if !flags.has(inline) {
				p = *(*unsafe.Pointer)(p)
			}
			return size(p, flags.without(inline).with(wantzero))
		}
		return 0
	}
}

func pointerEncodeFuncOf(t reflect.Type, encode encodeFunc) encodeFunc {
	return func(b []byte, p unsafe.Pointer, flags flags) (int, error) {
		if p != nil {
			if !flags.has(inline) {
				p = *(*unsafe.Pointer)(p)
			}
			return encode(b, p, flags.without(inline).with(wantzero))
		}
		return 0, nil
	}
}

func pointerDecodeFuncOf(t reflect.Type, decode decodeFunc) decodeFunc {
	t = t.Elem()
	return func(b []byte, p unsafe.Pointer, _ flags) (int, error) {
		v := (*unsafe.Pointer)(p)
		if *v == nil {
			*v = unsafe.Pointer(reflect.New(t).Pointer())
		}
		return decode(b, *v, noflags)
	}
}
