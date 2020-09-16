package json

import (
	"reflect"
	"unsafe"
)

func appendInt(b []byte, n int64) []byte {
	return formatInteger(b, uint64(n), n < 0)
}

func appendUint(b []byte, n uint64) []byte {
	return formatInteger(b, n, false)
}

const intLookup = "00010203040506070809" +
	"10111213141516171819" +
	"20212223242526272829" +
	"30313233343536373839" +
	"40414243444546474849" +
	"50515253545556575859" +
	"60616263646566676869" +
	"70717273747576777879" +
	"80818283848586878889" +
	"90919293949596979899"

var u16Lookup = stringToU16(intLookup)

func stringToU16(s string) []uint16 {
	return *(*[]uint16)(unsafe.Pointer(&reflect.SliceHeader{
		Data: ((*reflect.StringHeader)(unsafe.Pointer(&s))).Data,
		Len:  len(s) / 2,
		Cap:  len(s) / 2,
	}))
}

func formatInteger(b []byte, n uint64, negative bool) []byte {
	if !negative {
		if n < 10 {
			return append(b, byte(n+'0'))
		} else if n < 100 {
			u := u16Lookup[n]
			return append(b, byte(u), byte(u >> 8))
		}
	} else {
		n = -n
	}

	var buf [22]byte
	i := len(buf) / 2

	u := *(*[]uint16)(cast(buf[:], 2))

	for n >= 100 {
		j := n % 100
		n /= 100
		i--
		u[i] = u16Lookup[j]
	}

	i--
	u[i] = u16Lookup[n]

	i *= 2
	if n < 10 {
		i++
	}
	if negative {
		i--
		buf[i] = '-'
	}

	return append(b, buf[i:]...)
}

func cast(b []byte, size int) unsafe.Pointer {
	return unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(*(*unsafe.Pointer)(unsafe.Pointer(&b))),
		Len:  len(b) / size,
		Cap:  len(b) / size,
	})
}
