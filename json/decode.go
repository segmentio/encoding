package json

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

func (d decoder) decodeNull(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}
	return b, unmarshalTypeError(b, nullType)
}

func (d decoder) decodeBool(b []byte, p unsafe.Pointer) ([]byte, error) {
	switch {
	case hasPrefix(b, "true"):
		*(*bool)(p) = true
		return b[4:], nil

	case hasPrefix(b, "false"):
		*(*bool)(p) = false
		return b[5:], nil

	case hasPrefix(b, "null"):
		return b[4:], nil

	default:
		return b, unmarshalTypeError(b, boolType)
	}
}

func (d decoder) decodeInt(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	v, r, err := parseInt(b)
	if err != nil {
		return r, unmarshalTypeError(b, intType)
	}

	*(*int)(p) = int(v)
	return r, nil
}

func (d decoder) decodeInt8(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	v, r, err := parseInt(b)
	if err != nil {
		return r, unmarshalTypeError(b, int8Type)
	}

	if v < math.MinInt8 || v > math.MaxInt8 {
		return r, unmarshalOverflow(b[:len(b)-len(r)], int8Type)
	}

	*(*int8)(p) = int8(v)
	return r, nil
}

func (d decoder) decodeInt16(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	v, r, err := parseInt(b)
	if err != nil {
		return r, unmarshalTypeError(b, int16Type)
	}

	if v < math.MinInt16 || v > math.MaxInt16 {
		return r, unmarshalOverflow(b[:len(b)-len(r)], int16Type)
	}

	*(*int16)(p) = int16(v)
	return r, nil
}

func (d decoder) decodeInt32(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	v, r, err := parseInt(b)
	if err != nil {
		return r, unmarshalTypeError(b, int32Type)
	}

	if v < math.MinInt32 || v > math.MaxInt32 {
		return r, unmarshalOverflow(b[:len(b)-len(r)], int32Type)
	}

	*(*int32)(p) = int32(v)
	return r, nil
}

func (d decoder) decodeInt64(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	v, r, err := parseInt(b)
	if err != nil {
		return r, unmarshalTypeError(b, int64Type)
	}

	*(*int64)(p) = v
	return r, nil
}

func (d decoder) decodeUint(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	v, r, err := parseUint(b)
	if err != nil {
		return r, unmarshalTypeError(b, uintType)
	}

	*(*uint)(p) = uint(v)
	return r, nil
}

func (d decoder) decodeUintptr(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	v, r, err := parseUint(b)
	if err != nil {
		return r, unmarshalTypeError(b, uintType)
	}

	*(*uintptr)(p) = uintptr(v)
	return r, nil
}

func (d decoder) decodeUint8(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	v, r, err := parseUint(b)
	if err != nil {
		return r, unmarshalTypeError(b, uint8Type)
	}

	if v > math.MaxUint8 {
		return r, unmarshalOverflow(b[:len(b)-len(r)], uint8Type)
	}

	*(*uint8)(p) = uint8(v)
	return r, nil
}

func (d decoder) decodeUint16(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	v, r, err := parseUint(b)
	if err != nil {
		return r, unmarshalTypeError(b, uint16Type)
	}

	if v > math.MaxUint16 {
		return r, unmarshalOverflow(b[:len(b)-len(r)], uint16Type)
	}

	*(*uint16)(p) = uint16(v)
	return r, nil
}

func (d decoder) decodeUint32(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	v, r, err := parseUint(b)
	if err != nil {
		return r, unmarshalTypeError(b, uint32Type)
	}

	if v > math.MaxUint32 {
		return r, unmarshalOverflow(b[:len(b)-len(r)], uint32Type)
	}

	*(*uint32)(p) = uint32(v)
	return r, nil
}

func (d decoder) decodeUint64(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	v, r, err := parseUint(b)
	if err != nil {
		return r, unmarshalTypeError(b, uint64Type)
	}

	*(*uint64)(p) = v
	return r, nil
}

func (d decoder) decodeFloat32(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	v, r, err := parseNumber(b)
	if err != nil {
		return r, unmarshalTypeError(b, float32Type)
	}

	f, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&v)), 32)
	if err != nil {
		return r, unmarshalTypeError(b, float32Type)
	}

	*(*float32)(p) = float32(f)
	return r, nil
}

func (d decoder) decodeFloat64(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	v, r, err := parseNumber(b)
	if err != nil {
		return r, unmarshalTypeError(b, float64Type)
	}

	f, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&v)), 64)
	if err != nil {
		return r, unmarshalTypeError(b, float64Type)
	}

	*(*float64)(p) = f
	return r, nil
}

func (d decoder) decodeNumber(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	v, r, err := parseNumber(b)
	if err != nil {
		return r, unmarshalTypeError(b, numberType)
	}

	if (d.flags & DontCopyNumber) != 0 {
		*(*Number)(p) = *(*Number)(unsafe.Pointer(&v))
	} else {
		*(*Number)(p) = Number(v)
	}

	return r, nil
}

func (d decoder) decodeString(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	s, r, new, err := parseStringUnquote(b, nil)
	if err != nil {
		return b, unmarshalTypeError(r, stringType)
	}

	if new || (d.flags&DontCopyString) != 0 {
		*(*string)(p) = *(*string)(unsafe.Pointer(&s))
	} else {
		*(*string)(p) = string(s)
	}

	return r, nil
}

func (d decoder) decodeFromString(b []byte, p unsafe.Pointer, decode decodeFunc) ([]byte, error) {
	if hasPrefix(b, "null") {
		return decode(d, b, p)
	}

	v, b, _, err := parseStringUnquote(b, nil)
	if err != nil {
		return b, unmarshalTypeError(v, stringType)
	}

	if v, err = decode(d, v, p); err != nil {
		return b, err
	}

	if v = skipSpaces(v); len(v) != 0 {
		return b, syntaxError(v, "unexpected trailing tokens after string value")
	}

	return b, nil
}

func (d decoder) decodeBytes(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	if len(b) < 2 || b[0] != '"' {
		// Go 1.7- behavior: bytes slices may be decoded from array of integers.
		if len(b) >= 1 && b[0] == '[' {
			return d.decodeSlice(b, p, 1, bytesType, decoder.decodeUint8)
		}
		return b, unmarshalTypeError(b, bytesType)
	}

	// The input string contains escaped sequences, we need to parse it before
	// decoding it to match the encoding/json package behvaior.
	src, r, _, err := parseStringUnquote(b, nil)
	if err != nil {
		return b, unmarshalTypeError(b, bytesType)
	}

	dst := make([]byte, base64.StdEncoding.DecodedLen(len(src)))

	n, err := base64.StdEncoding.Decode(dst, src)
	if err != nil {
		return b, unmarshalTypeError(b, bytesType)
	}

	*(*[]byte)(p) = dst[:n]
	return r, nil
}

func (d decoder) decodeDuration(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	// in order to inter-operate with the stdlib, we must be able to interpret
	// durations passed as integer values.  there's some discussion about being
	// flexible on how durations are formatted, but for the time being, it's
	// been punted to go2 at the earliest: https://github.com/golang/go/issues/4712
	if len(b) > 0 && b[0] != '"' {
		v, r, err := parseInt(b)
		if err != nil {
			return r, unmarshalTypeError(b, int32Type)
		}

		if v < math.MinInt64 || v > math.MaxInt64 {
			return r, unmarshalOverflow(b[:len(b)-len(r)], int32Type)
		}

		*(*time.Duration)(p) = time.Duration(v)
		return r, nil
	}

	if len(b) < 2 || b[0] != '"' {
		return b, unmarshalTypeError(b, durationType)
	}

	i := bytes.IndexByte(b[1:], '"') + 1
	if i <= 0 {
		return b, unmarshalTypeError(b, durationType)
	}

	s := b[1:i] // trim quotes

	v, err := time.ParseDuration(*(*string)(unsafe.Pointer(&s)))
	if err != nil {
		return b, unmarshalTypeError(b, durationType)
	}

	*(*time.Duration)(p) = v
	return b[i+1:], nil
}

func (d decoder) decodeTime(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	if len(b) < 2 || b[0] != '"' {
		return b, unmarshalTypeError(b, durationType)
	}

	i := bytes.IndexByte(b[1:], '"') + 1
	if i <= 0 {
		return b, unmarshalTypeError(b, durationType)
	}

	s := b[1:i] // trim quotes

	v, err := time.Parse(time.RFC3339Nano, *(*string)(unsafe.Pointer(&s)))
	if err != nil {
		return b, unmarshalTypeError(b, durationType)
	}

	*(*time.Time)(p) = v
	return b[i+1:], nil
}

func (d decoder) decodeArray(b []byte, p unsafe.Pointer, n int, size uintptr, t reflect.Type, decode decodeFunc) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}

	if len(b) < 2 || b[0] != '[' {
		return b, unmarshalTypeError(b, t)
	}
	b = b[1:]

	var err error
	for i := 0; i < n; i++ {
		b = skipSpaces(b)

		if i != 0 {
			if len(b) == 0 {
				return b, syntaxError(b, "unexpected EOF after array element")
			}
			switch b[0] {
			case ',':
				b = skipSpaces(b[1:])
			case ']':
				return b[1:], nil
			default:
				return b, syntaxError(b, "expected ',' after array element but found '%c'", b[0])
			}
		}

		b, err = decode(d, b, unsafe.Pointer(uintptr(p)+(uintptr(i)*size)))
		if err != nil {
			if e, ok := err.(*UnmarshalTypeError); ok {
				e.Struct = t.String() + e.Struct
				e.Field = strconv.Itoa(i) + "." + e.Field
			}
			return b, err
		}
	}

	// The encoding/json package ignores extra elements found when decoding into
	// array types (which have a fixed size).
	for {
		b = skipSpaces(b)

		if len(b) == 0 {
			return b, syntaxError(b, "missing closing ']' in array value")
		}

		switch b[0] {
		case ',':
			b = skipSpaces(b[1:])
		case ']':
			return b[1:], nil
		}

		_, b, err = parseValue(b)
		if err != nil {
			return b, err
		}
	}
}

func (d decoder) decodeSlice(b []byte, p unsafe.Pointer, size uintptr, t reflect.Type, decode decodeFunc) ([]byte, error) {
	if hasPrefix(b, "null") {
		*(*slice)(p) = slice{}
		return b[4:], nil
	}

	if len(b) < 2 || b[0] != '[' {
		// Go 1.7- behavior: fallback to decoding as a []byte if the element
		// type is byte; allow conversions from JSON strings even tho the
		// underlying type implemented unmarshaler interfaces.
		if t.Elem().Kind() == reflect.Uint8 {
			return d.decodeBytes(b, p)
		}
		return b, unmarshalTypeError(b, t)
	}
	b = b[1:]

	s := (*slice)(p)
	s.len = 0

	var err error
	for {
		b = skipSpaces(b)

		if len(b) != 0 && b[0] == ']' {
			if s.data == nil {
				s.data = unsafe.Pointer(&empty)
			}
			return b[1:], nil
		}

		if s.len != 0 {
			if len(b) == 0 {
				return b, syntaxError(b, "unexpected EOF after array element")
			}
			if b[0] != ',' {
				return b, syntaxError(b, "expected ',' after array element but found '%c'", b[0])
			}
			b = skipSpaces(b[1:])
		}

		if s.len == s.cap {
			c := s.cap

			if c == 0 {
				c = 10
			} else {
				c *= 2
			}

			v := reflect.MakeSlice(t, s.len, c)
			reflect.Copy(v, reflect.NewAt(t, p).Elem())

			*s = *(*slice)((*iface)(unsafe.Pointer(&v)).ptr)
		}

		b, err = decode(d, b, unsafe.Pointer(uintptr(s.data)+(uintptr(s.len)*size)))
		if err != nil {
			if e, ok := err.(*UnmarshalTypeError); ok {
				e.Struct = t.String() + e.Struct
				e.Field = strconv.Itoa(s.len) + "." + e.Field
			}
			return b, err
		}

		s.len++
	}
}

func (d decoder) decodeMap(b []byte, p unsafe.Pointer, t, kt, vt reflect.Type, kz, vz reflect.Value, decodeKey, decodeValue decodeFunc) ([]byte, error) {
	if hasPrefix(b, "null") {
		*(*unsafe.Pointer)(p) = nil
		return b[4:], nil
	}

	if len(b) < 2 || b[0] != '{' {
		return b, unmarshalTypeError(b, t)
	}
	b = b[1:]

	i := 0
	m := reflect.NewAt(t, p).Elem()

	k := reflect.New(kt).Elem()
	v := reflect.New(vt).Elem()

	kptr := (*iface)(unsafe.Pointer(&k)).ptr
	vptr := (*iface)(unsafe.Pointer(&v)).ptr

	if m.IsNil() {
		m = reflect.MakeMap(t)
	}

	var err error
	for {
		k.Set(kz)
		v.Set(vz)
		b = skipSpaces(b)

		if len(b) != 0 && b[0] == '}' {
			*(*unsafe.Pointer)(p) = unsafe.Pointer(m.Pointer())
			return b[1:], nil
		}

		if i != 0 {
			if len(b) == 0 {
				return b, syntaxError(b, "unexpected EOF after object field value")
			}
			if b[0] != ',' {
				return b, syntaxError(b, "expected ',' after object field value but found '%c'", b[0])
			}
			b = skipSpaces(b[1:])
		}

		if b, err = decodeKey(d, b, kptr); err != nil {
			return b, err
		}
		b = skipSpaces(b)

		if len(b) == 0 {
			return b, syntaxError(b, "unexpected EOF after object field key")
		}
		if b[0] != ':' {
			return b, syntaxError(b, "expected ':' after object field key but found '%c'", b[0])
		}
		b = skipSpaces(b[1:])

		if b, err = decodeValue(d, b, vptr); err != nil {
			if e, ok := err.(*UnmarshalTypeError); ok {
				e.Struct = "map[" + kt.String() + "]" + vt.String() + "{" + e.Struct + "}"
				e.Field = fmt.Sprint(k.Interface()) + "." + e.Field
			}
			return b, err
		}

		m.SetMapIndex(k, v)
		i++
	}
}

func (d decoder) decodeMapStringInterface(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		*(*unsafe.Pointer)(p) = nil
		return b[4:], nil
	}

	if len(b) < 2 || b[0] != '{' {
		return b, unmarshalTypeError(b, mapStringInterfaceType)
	}
	b = b[1:]

	i := 0
	m := *(*map[string]interface{})(p)

	if m == nil {
		m = make(map[string]interface{}, 64)
	}

	var err error
	var key string
	var val interface{}

	for {
		key = ""
		val = nil

		b = skipSpaces(b)

		if len(b) != 0 && b[0] == '}' {
			*(*unsafe.Pointer)(p) = *(*unsafe.Pointer)(unsafe.Pointer(&m))
			return b[1:], nil
		}

		if i != 0 {
			if len(b) == 0 {
				return b, syntaxError(b, "unexpected EOF after object field value")
			}
			if b[0] != ',' {
				return b, syntaxError(b, "expected ',' after object field value but found '%c'", b[0])
			}
			b = skipSpaces(b[1:])
		}

		b, err = d.decodeString(b, unsafe.Pointer(&key))
		if err != nil {
			return b, err
		}
		b = skipSpaces(b)

		if len(b) == 0 {
			return b, syntaxError(b, "unexpected EOF after object field key")
		}
		if b[0] != ':' {
			return b, syntaxError(b, "expected ':' after object field key but found '%c'", b[0])
		}
		b = skipSpaces(b[1:])

		b, err = d.decodeInterface(b, unsafe.Pointer(&val))
		if err != nil {
			if e, ok := err.(*UnmarshalTypeError); ok {
				e.Struct = mapStringInterfaceType.String() + e.Struct
				e.Field = key + "." + e.Field
			}
			return b, err
		}

		m[key] = val
		i++
	}
}

func (d decoder) decodeMapStringRawMessage(b []byte, p unsafe.Pointer) ([]byte, error) {
	if hasPrefix(b, "null") {
		*(*unsafe.Pointer)(p) = nil
		return b[4:], nil
	}

	if len(b) < 2 || b[0] != '{' {
		return b, unmarshalTypeError(b, mapStringInterfaceType)
	}
	b = b[1:]

	i := 0
	m := *(*map[string]RawMessage)(p)

	if m == nil {
		m = make(map[string]RawMessage, 64)
	}

	var err error
	var key string
	var val RawMessage

	for {
		key = ""
		val = nil

		b = skipSpaces(b)

		if len(b) != 0 && b[0] == '}' {
			*(*unsafe.Pointer)(p) = *(*unsafe.Pointer)(unsafe.Pointer(&m))
			return b[1:], nil
		}

		if i != 0 {
			if len(b) == 0 {
				return b, syntaxError(b, "unexpected EOF after object field value")
			}
			if b[0] != ',' {
				return b, syntaxError(b, "expected ',' after object field value but found '%c'", b[0])
			}
			b = skipSpaces(b[1:])
		}

		b, err = d.decodeString(b, unsafe.Pointer(&key))
		if err != nil {
			return b, err
		}
		b = skipSpaces(b)

		if len(b) == 0 {
			return b, syntaxError(b, "unexpected EOF after object field key")
		}
		if b[0] != ':' {
			return b, syntaxError(b, "expected ':' after object field key but found '%c'", b[0])
		}
		b = skipSpaces(b[1:])

		b, err = d.decodeRawMessage(b, unsafe.Pointer(&val))
		if err != nil {
			if e, ok := err.(*UnmarshalTypeError); ok {
				e.Struct = mapStringRawMessageType.String() + e.Struct
				e.Field = key + "." + e.Field
			}
			return b, err
		}

		m[key] = val
		i++
	}
}

func (d decoder) decodeStruct(b []byte, p unsafe.Pointer, st *structType) ([]byte, error) {
	if hasPrefix(b, "null") {
		return b[4:], nil
	}
	if len(b) < 2 || b[0] != '{' {
		return b, unmarshalTypeError(b, st.typ)
	}
	b = b[1:]

	var err error
	var k []byte
	var i int

	// memory buffer used to convert short field names to lowercase
	var buf [64]byte
	var key []byte

	for {
		b = skipSpaces(b)

		if len(b) != 0 && b[0] == '}' {
			return b[1:], nil
		}

		if i != 0 {
			if len(b) == 0 {
				return b, syntaxError(b, "unexpected EOF after object field value")
			}
			if b[0] != ',' {
				return b, syntaxError(b, "expected ',' after object field value but found '%c'", b[0])
			}
			b = skipSpaces(b[1:])
		}
		i++

		k, b, _, err = parseStringUnquote(b, nil)
		if err != nil {
			return b, syntaxError(b, "parsing object field key: %s", err)
		}
		b = skipSpaces(b)

		if len(b) == 0 {
			return b, syntaxError(b, "unexpected EOF after object field key")
		}
		if b[0] != ':' {
			return b, syntaxError(b, "expected ':' after object field key but found '%c'", b[0])
		}
		b = skipSpaces(b[1:])

		f := st.fieldsIndex[string(k)]

		if f == nil && (d.flags&DontMatchCaseInsensitiveStructFields) == 0 {
			key = appendToLower(buf[:0], k)
			f = st.ficaseIndex[string(key)]
		}

		if f == nil {
			if (d.flags & DisallowUnknownFields) != 0 {
				return b, fmt.Errorf("json: unknown field %q", k)
			}
			if _, b, err = parseValue(b); err != nil {
				return b, err
			}
			continue
		}

		if b, err = f.codec.decode(d, b, unsafe.Pointer(uintptr(p)+f.offset)); err != nil {
			if e, ok := err.(*UnmarshalTypeError); ok {
				e.Struct = st.typ.String() + e.Struct
				e.Field = string(k) + "." + e.Field
			}
			return b, err
		}
	}
}

func (d decoder) decodeEmbeddedStructPointer(b []byte, p unsafe.Pointer, t reflect.Type, unexported bool, offset uintptr, decode decodeFunc) ([]byte, error) {
	v := *(*unsafe.Pointer)(p)

	if v == nil {
		if unexported {
			return b, fmt.Errorf("json: cannot set embedded pointer to unexported struct: %s", t)
		}
		v = unsafe.Pointer(reflect.New(t).Pointer())
		*(*unsafe.Pointer)(p) = v
	}

	return decode(d, b, unsafe.Pointer(uintptr(v)+offset))
}

func (d decoder) decodePointer(b []byte, p unsafe.Pointer, t reflect.Type, decode decodeFunc) ([]byte, error) {
	if hasPrefix(b, "null") {
		pp := *(*unsafe.Pointer)(p)
		if pp != nil && t.Kind() == reflect.Ptr {
			return decode(d, b, pp)
		}
		*(*unsafe.Pointer)(p) = nil
		return b[4:], nil
	}

	v := *(*unsafe.Pointer)(p)
	if v == nil {
		v = unsafe.Pointer(reflect.New(t).Pointer())
		*(*unsafe.Pointer)(p) = v
	}

	return decode(d, b, v)
}

func (d decoder) decodeInterface(b []byte, p unsafe.Pointer) ([]byte, error) {
	val := *(*interface{})(p)
	*(*interface{})(p) = nil

	if t := reflect.TypeOf(val); t != nil && t.Kind() == reflect.Ptr {
		if v := reflect.ValueOf(val); v.IsNil() || t.Elem().Kind() != reflect.Ptr {
			// If the destination is nil the only value that is OK to decode is
			// `null`, and the encoding/json package always nils the destination
			// interface value in this case.
			if hasPrefix(b, "null") {
				*(*interface{})(p) = nil
				return b[4:], nil
			}
		}

		b, err := Parse(b, val, d.flags)
		if err == nil {
			*(*interface{})(p) = val
		}
		return b, err
	}

	v, b, err := parseValue(b)
	if err != nil {
		return b, err
	}

	switch v[0] {
	case '{':
		m := make(map[string]interface{})
		v, err = d.decodeMapStringInterface(v, unsafe.Pointer(&m))
		val = m

	case '[':
		a := make([]interface{}, 0, 10)
		v, err = d.decodeSlice(v, unsafe.Pointer(&a), unsafe.Sizeof(a[0]), sliceInterfaceType, decoder.decodeInterface)
		val = a

	case '"':
		s := ""
		v, err = d.decodeString(v, unsafe.Pointer(&s))
		val = s

	case 'n':
		v, err = d.decodeNull(v, nil)
		val = nil

	case 't', 'f':
		x := false
		v, err = d.decodeBool(v, unsafe.Pointer(&x))
		val = x

	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if (d.flags & UseNumber) != 0 {
			n := Number("")
			v, err = d.decodeNumber(v, unsafe.Pointer(&n))
			val = n
		} else {
			f := 0.0
			v, err = d.decodeFloat64(v, unsafe.Pointer(&f))
			val = f
		}

	default:
		return b, syntaxError(v, "expected token but found '%c'", v[0])
	}

	if err != nil {
		return b, err
	}

	if v = skipSpaces(v); len(v) != 0 {
		return b, syntaxError(v, "unexpected trailing trailing tokens after json value")
	}

	*(*interface{})(p) = val
	return b, nil
}

func (d decoder) decodeUnmarshalTypeError(b []byte, p unsafe.Pointer, t reflect.Type) ([]byte, error) {
	v, b, err := parseValue(b)
	if err != nil {
		return b, err
	}
	return b, &UnmarshalTypeError{
		Value: string(v),
		Type:  t,
	}
}

func (d decoder) decodeRawMessage(b []byte, p unsafe.Pointer) ([]byte, error) {
	v, r, err := parseValue(b)
	if err != nil {
		return r, unmarshalTypeError(b, rawMessageType)
	}

	if (d.flags & DontCopyRawMessage) == 0 {
		v = append(make([]byte, 0, len(v)), v...)
	}

	*(*RawMessage)(p) = json.RawMessage(v)
	return r, err
}

func (d decoder) decodeJSONUnmarshaler(b []byte, p unsafe.Pointer, t reflect.Type, pointer bool) ([]byte, error) {
	v, b, err := parseValue(b)
	if err != nil {
		return b, err
	}

	if len(v) != 0 && v[0] == 'n' { // null
		return b, nil
	}

	u := reflect.NewAt(t, p)
	if !pointer {
		u = u.Elem()
		t = t.Elem()
	}
	if u.IsNil() {
		u.Set(reflect.New(t))
	}
	return b, u.Interface().(Unmarshaler).UnmarshalJSON(v)
}

func (d decoder) decodeTextUnmarshaler(b []byte, p unsafe.Pointer, t reflect.Type, pointer bool) ([]byte, error) {
	var value string

	switch b[0] {
	case 'n':
		_, b, err := parseNull(b)
		return b, err
	case '"':
		s, b, _, err := parseStringUnquote(b, nil)
		if err != nil {
			return b, err
		}
		u := reflect.NewAt(t, p)
		if !pointer {
			u = u.Elem()
			t = t.Elem()
		}
		if u.IsNil() {
			u.Set(reflect.New(t))
		}
		return b, u.Interface().(encoding.TextUnmarshaler).UnmarshalText(s)
	case '{':
		value = "object"
	case '[':
		value = "array"
	case 't':
		value = "true"
	case 'f':
		value = "false"
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		value = "number"
	}

	return b, &UnmarshalTypeError{Value: value, Type: reflect.PtrTo(t)}
}

var (
	// This is a placeholder used to consturct non-nil empty slices.
	empty struct{}
)
