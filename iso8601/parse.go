package iso8601

import (
	"encoding/binary"
	"time"
	"unsafe"
)

// Parse parses an ISO8601 string, e.g. "2021-03-25T21:36:12Z".
func Parse(input string) (time.Time, error) {
	switch len(input) {
	case 20:
		b := unsafeStringToBytes(input)
		_ = b[19] // compiler hint

		t1 := binary.LittleEndian.Uint64(b)
		t2 := binary.LittleEndian.Uint64(b[8:])
		t3 := binary.LittleEndian.Uint32(b[16:])

		// Check for valid "YYYY-MM-DDTHH:MM:SSZ" separators by masking input
		// with "    -  -  T  :  :  Z". If separators are all valid, replace
		// them with a '0' (0x30) byte and check all bytes are now numeric.
		// If it's not input we can handle on the fast path, pass it off to
		// time.Parse().
		if (t1&sepMask1u64) != sepMask1u64 ||
			(t2&sepMask2u64) != sepMask2u64 ||
			(t3&sepMask3u32) != sepMask3u32 ||
			nonNumeric64(t1 & ^sepMask1u64 | sepZero1u64) != 0 ||
			nonNumeric64(t2 & ^sepMask2u64 | sepZero2u64) != 0 ||
			nonNumeric32(t3 & ^sepMask3u32 | sepZero3u32) != 0 {

			goto fallback
		}

		// TODO: there's probably a faster way to extract the integers, e.g. see
		//  https://kholdstare.github.io/technical/2020/05/26/faster-integer-parsing.html
		year := uint32(b[0]-'0')*1000 + uint32(b[1]-'0')*100 + uint32(b[2]-'0')*10 + uint32(b[3]-'0')
		month := uint32(b[5]-'0')*10 + uint32(b[6]-'0')
		day := uint32(b[8]-'0')*10 + uint32(b[9]-'0')
		hour := uint32(b[11]-'0')*10 + uint32(b[12]-'0')
		minute := uint32(b[14]-'0')*10 + uint32(b[15]-'0')
		second := uint32(b[17]-'0')*10 + uint32(b[18]-'0')

		// From https://blog.reverberate.org/2020/05/12/optimizing-date-algorithms.html.
		monthAdjusted := month - 3
		var carry uint32
		if monthAdjusted > month {
			carry = 1
		}
		var adjust uint32
		if carry == 1 {
			adjust = 12
		}
		yearAdjusted := year + 4800 - carry
		monthDays := ((monthAdjusted+adjust)*62719 + 769) / 2048
		leapDays := yearAdjusted/4 - yearAdjusted/100 + yearAdjusted/400
		daysSinceEpoch := yearAdjusted*365 + leapDays + monthDays + (day - 1) - 2472632

		return time.Unix(int64(daysSinceEpoch)*86400+int64(hour*3600+minute*60+second), 0), nil
	}

fallback:
	return time.Parse(time.RFC3339Nano, input)
}

const (
	sepMask1u64 = uint64(0x2d00002d00000000) // YYYY-MM-
	sepMask2u64 = uint64(0x00003a0000540000) // DDTHH:MM
	sepMask3u32 = uint32(0x5a00003a)         // :SSZ

	sepZero1u64 = uint64(0x3000003000000000)
	sepZero2u64 = uint64(0x0000300000300000)
	sepZero3u32 = uint32(0x30000030)
)

func nonNumeric64(u uint64) uint64 {
	// Derived from https://graphics.stanford.edu/~seander/bithacks.html#HasLessInWord.
	// Subtract '0' (0x30) from each byte so that the MSB is set in each byte
	// if there's a byte less than '0' (0x30). Add 0x46 (0x7F-'9') so that the
	// MSB is set if there's a byte greater than '9' (0x39). To handle overflow
	// when adding 0x46, include the MSB from the input bytes in the final mask.
	// Remove all but the MSBs and then you're left with a mask where each
	// non-numeric byte from the input has its MSB set in the output.
	return ((u - 0x3030303030303030) | (u + 0x4646464646464646) | u) & 0x8080808080808080
}

func nonNumeric32(u uint32) uint32 {
	return ((u - 0x30303030) | (u + 0x46464646) | u) & 0x80808080
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
