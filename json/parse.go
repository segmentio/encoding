package json

import (
	"bytes"
	"math"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/segmentio/encoding/ascii"
)

// All spaces characters defined in the json specification.
const (
	sp = ' '
	ht = '\t'
	nl = '\n'
	cr = '\r'
	np = '\f'
	bs = '\b'
)

const (
	escape = '\\'
	quote  = '"'
)

// Constants used for the optimized string scanning functions. The values are
// masks applied to test if a word contains a byte we're looking for.
const (
	escapeMask = uint64(escape<<56 | escape<<48 | escape<<40 | escape<<32 |
		escape<<24 | escape<<16 | escape<<8 | escape)

	quoteMask = uint64(quote<<56 | quote<<48 | quote<<40 | quote<<32 |
		quote<<24 | quote<<16 | quote<<8 | quote)
)

func skipSpaces(b []byte) []byte {
	i := 0
skipLoop:
	for _, c := range b {
		switch c {
		case sp, ht, nl, cr, np, bs:
			i++
		default:
			break skipLoop
		}
	}
	return b[i:]
}

// parseInt parses a decimanl representation of an int64 from b.
//
// The function is equivalent to calling strconv.ParseInt(string(b), 10, 64) but
// it prevents Go from making a memory allocation for converting a byte slice to
// a string (escape analysis fails due to the error returned by strconv.ParseInt).
//
// Because it only works with base 10 the function is also significantly faster
// than strconv.ParseInt.
func parseInt(b []byte) (int64, []byte, error) {
	var value int64
	var count int

	if len(b) == 0 {
		return 0, b, syntaxError(b, "cannot decode integer from an empty input")
	}

	if b[0] == '-' {
		const max = math.MinInt64
		const lim = max / 10

		for _, d := range b[1:] {
			if !(d >= '0' && d <= '9') {
				if count == 0 {
					return 0, b, syntaxError(b, "missing digits after negative sign")
				}
				break
			}

			if value < lim {
				return 0, b, syntaxError(b, "integer value out of range")
			}

			value *= 10
			x := int64(d - '0')

			if value < (max + x) {
				return 0, b, syntaxError(b, "integer value out of range")
			}

			value -= x
			count++
		}

		count++
	} else {
		const max = math.MaxInt64
		const lim = max / 10

		for _, d := range b {
			if !(d >= '0' && d <= '9') {
				if count == 0 {
					return 0, b, syntaxError(b, "expected digit but found '%c'", d)
				}
				break
			}
			x := int64(d - '0')

			if value > lim {
				return 0, b, syntaxError(b, "integer value out of range")
			}

			if value *= 10; value > (max - x) {
				return 0, b, syntaxError(b, "integer value out of range")
			}

			value += x
			count++
		}
	}

	return value, b[count:], nil
}

// parseUint is like parseInt but for unsigned integers.
func parseUint(b []byte) (uint64, []byte, error) {
	const max = math.MaxUint64
	const lim = max / 10

	var value uint64
	var count int

	if len(b) == 0 {
		return 0, b, syntaxError(b, "cannot decode integer value from an empty input")
	}

	for _, d := range b {
		if !(d >= '0' && d <= '9') {
			if count == 0 {
				return 0, b, syntaxError(b, "expected digit but found '%c'", d)
			}
			break
		}
		x := uint64(d - '0')

		if value > lim {
			return 0, b, syntaxError(b, "integer value out of range")
		}

		if value *= 10; value > (max - x) {
			return 0, b, syntaxError(b, "integer value out of range")
		}

		value += x
		count++
	}

	return value, b[count:], nil
}

// parseUintHex parses a hexadecimanl representation of a uint64 from b.
//
// The function is equivalent to calling strconv.ParseUint(string(b), 16, 64) but
// it prevents Go from making a memory allocation for converting a byte slice to
// a string (escape analysis fails due to the error returned by strconv.ParseUint).
//
// Because it only works with base 16 the function is also significantly faster
// than strconv.ParseUint.
func parseUintHex(b []byte) (uint64, []byte, error) {
	const max = math.MaxUint64
	const lim = max / 0x10

	var value uint64
	var count int

	if len(b) == 0 {
		return 0, b, syntaxError(b, "cannot decode hexadecimal value from an empty input")
	}

parseLoop:
	for i, d := range b {
		var x uint64

		switch {
		case d >= '0' && d <= '9':
			x = uint64(d - '0')

		case d >= 'A' && d <= 'F':
			x = uint64(d-'A') + 0xA

		case d >= 'a' && d <= 'f':
			x = uint64(d-'a') + 0xA

		default:
			if i == 0 {
				return 0, b, syntaxError(b, "expected hexadecimal digit but found '%c'", d)
			}
			break parseLoop
		}

		if value > lim {
			return 0, b, syntaxError(b, "hexadecimal value out of range")
		}

		if value *= 0x10; value > (max - x) {
			return 0, b, syntaxError(b, "hexadecimal value out of range")
		}

		value += x
		count++
	}

	return value, b[count:], nil
}

func parseNull(b []byte) ([]byte, []byte, error) {
	if hasNullPrefix(b) {
		return b[:4], b[4:], nil
	}
	return nil, b, syntaxError(b, "expected 'null' but found invalid token")
}

func parseTrue(b []byte) ([]byte, []byte, error) {
	if hasTruePrefix(b) {
		return b[:4], b[4:], nil
	}
	return nil, b, syntaxError(b, "expected 'true' but found invalid token")
}

func parseFalse(b []byte) ([]byte, []byte, error) {
	if hasFalsePrefix(b) {
		return b[:5], b[5:], nil
	}
	return nil, b, syntaxError(b, "expected 'false' but found invalid token")
}

func parseNumber(b []byte) (v, r []byte, err error) {
	r = b

	if len(r) == 0 {
		err = syntaxError(b, "expected number but found no data")
		return
	}

	// sign
	if r[0] == '-' {
		r = r[1:]
	}

	if len(r) == 0 {
		err = syntaxError(b, "missing number value after sign")
		return
	}

	// integer part
	leadingZero := false
	integerLength := 0

	for i := 0; len(r) != 0; i++ {
		c := r[0]

		if i == 0 && c == '0' {
			leadingZero = true
		}

		if !('0' <= c && c <= '9') {
			if i == 0 {
				err = syntaxError(b, "expected digit but found '%c'", c)
				return
			}
			break
		}

		r = r[1:]
		integerLength++
	}

	if leadingZero && integerLength > 1 {
		err = syntaxError(b, "unexpected leading zero in number")
		return
	}

	// decimal part
	if len(r) != 0 && r[0] == '.' {
		decimalLength := 0
		r = r[1:]

		for i := 0; len(r) != 0; i++ {
			if c := r[0]; !('0' <= c && c <= '9') {
				if i == 0 {
					err = syntaxError(b, "expected digit but found '%c'", c)
					return
				}
				break
			}
			r = r[1:]
			decimalLength++
		}

		if decimalLength == 0 {
			err = syntaxError(b, "expected decimal part after '.'")
			return
		}
	}

	// exponent part
	if len(r) != 0 && (r[0] == 'e' || r[0] == 'E') {
		r = r[1:]

		if len(r) != 0 {
			if c := r[0]; c == '+' || c == '-' {
				r = r[1:]
			}
		}

		if len(r) == 0 {
			err = syntaxError(b, "missing exponent in number")
			return
		}

		for i := 0; len(r) != 0; i++ {
			if c := r[0]; !('0' <= c && c <= '9') {
				if i == 0 {
					err = syntaxError(b, "expected digit but found '%c'", c)
					return
				}
				break
			}
			r = r[1:]
		}
	}

	v = b[:len(b)-len(r)]
	return
}

func parseUnicode(b []byte) (rune, int, error) {
	if len(b) < 4 {
		return 0, 0, syntaxError(b, "unicode code point must have at least 4 characters")
	}

	u, r, err := parseUintHex(b[:4])
	if err != nil {
		return 0, 0, syntaxError(b, "parsing unicode code point: %s", err)
	}

	if len(r) != 0 {
		return 0, 0, syntaxError(b, "invalid unicode code point")
	}

	return rune(u), 4, nil
}

func parseStringFast(b []byte) ([]byte, []byte, bool, error) {
	if len(b) < 2 || b[0] != '"' {
		return nil, b, false, syntaxError(b, "expected '\"' at the beginning of a string value")
	}

	i := bytes.IndexByte(b[1:], '"')
	if i >= 0 && i < len(b) {
		if i++; bytes.IndexByte(b[1:i], '\\') < 0 && ascii.Valid(b[1:i]) {
			return b[:i+1], b[i+1:], false, nil
		}
	}

	offset := 1

	for offset < len(b) {
		i := bytes.IndexByte(b[offset:], '"')
		if i < 0 {
			break
		}
		i += offset
		j := bytes.IndexByte(b[offset:i], '\\')
		if j < 0 {
			return b[:i+1], b[i+1:], true, nil
		}
		offset += j + 1
		if offset < len(b) && b[offset] == '\\' || b[offset] == '"' {
			offset++ // skip escaped sequence
		}
	}

	return nil, b, false, syntaxError(b, "missing '\"' at the end of a string value")
}

func parseString(b []byte) ([]byte, []byte, error) {
	s, b, _, err := parseStringFast(b)
	return s, b, err
}

func parseStringUnquote(b []byte, r []byte) ([]byte, []byte, bool, error) {
	s, b, escaped, err := parseStringFast(b)
	if err != nil {
		return s, b, false, err
	}

	s = s[1 : len(s)-1] // trim the quotes

	if !escaped {
		return s, b, false, nil
	}

	if r == nil {
		r = make([]byte, 0, len(s))
	}

	for len(s) != 0 {
		i := bytes.IndexByte(s, '\\')

		if i < 0 {
			r = appendCoerceInvalidUTF8(r, s)
			break
		}

		r = appendCoerceInvalidUTF8(r, s[:i])
		s = s[i+1:]

		c := s[0]
		switch c {
		case '"', '\\', '/':
			// simple escaped character
		case 'n':
			c = '\n'

		case 'r':
			c = '\r'

		case 't':
			c = '\t'

		case 'b':
			c = '\b'

		case 'f':
			c = '\f'

		case 'u':
			s = s[1:]

			r1, n1, err := parseUnicode(s)
			if err != nil {
				return r, b, true, err
			}
			s = s[n1:]

			if utf16.IsSurrogate(r1) {
				if !hasPrefix(s, `\u`) {
					r1 = unicode.ReplacementChar
				} else {
					r2, n2, err := parseUnicode(s[2:])
					if err != nil {
						return r, b, true, err
					}
					if r1 = utf16.DecodeRune(r1, r2); r1 != unicode.ReplacementChar {
						s = s[2+n2:]
					}
				}
			}

			r = appendRune(r, r1)
			continue

		default: // not sure what this escape sequence is
			r = append(r, '\\')
			continue
		}

		r = append(r, c)
		s = s[1:]
	}

	return r, b, true, nil
}

func appendRune(b []byte, r rune) []byte {
	n := len(b)
	b = append(b, 0, 0, 0, 0)
	return b[:n+utf8.EncodeRune(b[n:], r)]
}

func appendCoerceInvalidUTF8(b []byte, s []byte) []byte {
	c := [4]byte{}

	for _, r := range string(s) {
		b = append(b, c[:utf8.EncodeRune(c[:], r)]...)
	}

	return b
}

func parseObject(b []byte) ([]byte, []byte, error) {
	if len(b) < 2 || b[0] != '{' {
		return nil, b, syntaxError(b, "expected '{' at the beginning of an object value")
	}

	var err error
	var a = b
	var n = len(b)
	var i = 0

	b = b[1:]
	for {
		b = skipSpaces(b)

		if len(b) == 0 {
			return nil, b, syntaxError(b, "cannot decode object from empty input")
		}

		if b[0] == '}' {
			j := (n - len(b)) + 1
			return a[:j], a[j:], nil
		}

		if i != 0 {
			if len(b) == 0 {
				return nil, b, syntaxError(b, "unexpected EOF after object field value")
			}
			if b[0] != ',' {
				return nil, b, syntaxError(b, "expected ',' after object field value but found '%c'", b[0])
			}
			b = skipSpaces(b[1:])
		}

		_, b, err = parseString(b)
		if err != nil {
			return nil, b, err
		}
		b = skipSpaces(b)

		if len(b) == 0 {
			return nil, b, syntaxError(b, "unexpected EOF after object field key")
		}
		if b[0] != ':' {
			return nil, b, syntaxError(b, "expected ':' after object field key but found '%c'", b[0])
		}
		b = skipSpaces(b[1:])

		_, b, err = parseValue(b)
		if err != nil {
			return nil, b, err
		}

		i++
	}
}

func parseArray(b []byte) ([]byte, []byte, error) {
	if len(b) < 2 || b[0] != '[' {
		return nil, b, syntaxError(b, "expected '[' at the beginning of array value")
	}

	var err error
	var a = b
	var n = len(b)
	var i = 0

	b = b[1:]
	for {
		b = skipSpaces(b)

		if len(b) == 0 {
			return nil, b, syntaxError(b, "missing closing ']' after array value")
		}

		if b[0] == ']' {
			j := (n - len(b)) + 1
			return a[:j], a[j:], nil
		}

		if i != 0 {
			if len(b) == 0 {
				return nil, b, syntaxError(b, "unexpected EOF after array element")
			}
			if b[0] != ',' {
				return nil, b, syntaxError(b, "expected ',' after array element but found '%c'", b[0])
			}
			b = skipSpaces(b[1:])
		}

		_, b, err = parseValue(b)
		if err != nil {
			return nil, b, err
		}

		i++
	}
}

func parseValue(b []byte) ([]byte, []byte, error) {
	if len(b) != 0 {
		switch b[0] {
		case '{':
			return parseObject(b)
		case '[':
			return parseArray(b)
		case '"':
			return parseString(b)
		case 'n':
			return parseNull(b)
		case 't':
			return parseTrue(b)
		case 'f':
			return parseFalse(b)
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return parseNumber(b)
		default:
			return nil, b, syntaxError(b, "expected token but found '%c'", b[0])
		}
	}
	return nil, b, syntaxError(b, "expected json but found no data")
}

func hasNullPrefix(b []byte) bool {
	return len(b) >= 4 && string(b[:4]) == "null"
}

func hasTruePrefix(b []byte) bool {
	return len(b) >= 4 && string(b[:4]) == "true"
}

func hasFalsePrefix(b []byte) bool {
	return len(b) >= 5 && string(b[:5]) == "false"
}

func hasPrefix(b []byte, s string) bool {
	return len(b) >= len(s) && s == string(b[:len(s)])
}

func appendToLower(b, s []byte) []byte {
	if ascii.Valid(s) { // fast path for ascii strings
		i := 0

		for j := range s {
			c := s[j]

			if 'A' <= c && c <= 'Z' {
				b = append(b, s[i:j]...)
				b = append(b, c+('a'-'A'))
				i = j + 1
			}
		}

		return append(b, s[i:]...)
	}

	for _, r := range string(s) {
		b = appendRune(b, foldRune(r))
	}

	return b
}

func foldRune(r rune) rune {
	if r = unicode.SimpleFold(r); 'A' <= r && r <= 'Z' {
		r = r + ('a' - 'A')
	}
	return r
}
