package iso8601

import (
	"encoding/binary"
	"time"
	"unsafe"
)

// Parse parses an ISO8601 string, e.g. "2021-03-25T21:36:12Z".
func Parse(input string) (time.Time, error) {
	switch b := unsafeStringToBytes(input); len(b) {
	case 20:
		t1 := binary.LittleEndian.Uint64(b)
		t2 := binary.LittleEndian.Uint64(b[8:])
		t3 := uint64(binary.LittleEndian.Uint32(b[16:]))

		// Check for valid "YYYY-MM-DDTHH:MM:SSZ" separators by masking input
		// with "    -  -  T  :  :  Z". If separators are all valid, replace
		// them with a '0' (0x30) byte and check all bytes are now numeric.
		// If it's not input we can handle on the fast path, pass it off to
		// time.Parse().
		if !match(t1, mask1) || !match(t2, mask2) || !match(t3, mask3) {
			goto fallback
		}
		t1 ^= replace1
		t2 ^= replace2
		t3 ^= replace3
		if (nonNumeric(t1) | nonNumeric(t2) | nonNumeric(t3)) != 0 {
			goto fallback
		}

		t1 -= zero
		t2 -= zero
		t3 -= zero
		year := (t1&0xF)*1000 + (t1>>8&0xF)*100 + (t1>>16&0xF)*10 + (t1 >> 24 & 0xF)
		month := (t1>>40&0xF)*10 + (t1 >> 48 & 0xF)
		day := (t2&0xF)*10 + (t2 >> 8 & 0xF)
		hour := (t2>>24&0xF)*10 + (t2 >> 32 & 0xF)
		minute := (t2>>48&0xF)*10 + (t2 >> 56)
		second := (t3>>8&0xF)*10 + (t3 >> 16)

		if month > 12 || day > 31 || hour >= 24 || minute >= 60 || second >= 60 {
			goto fallback
		} else if day == 31 {
			switch month {
			case 4, 6, 9, 11:
				goto fallback
			}
		} else if month == 2 && day == 29 && !isLeapYear(year) {
			goto fallback
		}

		unixSeconds := int64(daysSinceEpoch(year, month, day))*86400 + int64(hour*3600+minute*60+second)
		return time.Unix(unixSeconds, 0), nil
	}

fallback:
	return time.Parse(time.RFC3339Nano, input)
}

const (
	mask1 = 0x2d00002d00000000 // YYYY-MM-
	mask2 = 0x00003a0000540000 // DDTHH:MM
	mask3 = 0x000000005a00003a // :SSZ____

	// Generate masks that replace the separators with a numeric byte.
	// The input must have valid separators. XOR with the separator bytes
	// to zero them out and then XOR with 0x30 to replace them with '0'.
	replace1 = mask1 ^ 0x3000003000000000
	replace2 = mask2 ^ 0x0000300000300000
	replace3 = mask3 ^ 0x3030303030000030

	zero = 0x3030303030303030
	msb  = 0x8080808080808080
)

func match(u, mask uint64) bool {
	return (u & mask) == mask
}

func nonNumeric(u uint64) uint64 {
	// Derived from https://graphics.stanford.edu/~seander/bithacks.html#HasLessInWord.
	// Subtract '0' (0x30) from each byte so that the MSB is set in each byte
	// if there's a byte less than '0' (0x30). Add 0x46 (0x7F-'9') so that the
	// MSB is set if there's a byte greater than '9' (0x39). To handle overflow
	// when adding 0x46, include the MSB from the input bytes in the final mask.
	// Remove all but the MSBs and then you're left with a mask where each
	// non-numeric byte from the input has its MSB set in the output.
	return ((u - zero) | (u + 0x4646464646464646) | u) & msb
}

func daysSinceEpoch(year, month, day uint64) uint64 {
	// Derived from https://blog.reverberate.org/2020/05/12/optimizing-date-algorithms.html.
	monthAdjusted := month - 3
	var carry uint64
	if monthAdjusted > month {
		carry = 1
	}
	var adjust uint64
	if carry == 1 {
		adjust = 12
	}
	yearAdjusted := year + 4800 - carry
	monthDays := ((monthAdjusted+adjust)*62719 + 769) / 2048
	leapDays := yearAdjusted/4 - yearAdjusted/100 + yearAdjusted/400
	return yearAdjusted*365 + leapDays + monthDays + (day - 1) - 2472632
}

func isLeapYear(y uint64) bool {
	return (y%4) == 0 && ((y%100) != 0 || (y%400) == 0)
}

func unsafeStringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&sliceHeader{
		Data: *(*unsafe.Pointer)(unsafe.Pointer(&s)),
		Len:  len(s),
		Cap:  len(s),
	}))
}

// sliceHeader is like reflect.SliceHeader but the Data field is a
// unsafe.Pointer instead of being a uintptr to avoid invalid
// conversions from uintptr to unsafe.Pointer.
type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}
