package thrift

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"sort"
	"sync/atomic"
)

// Marshal serializes v into a thrift representation according to the the
// protocol p.
//
// The function panics if v cannot be converted to a thrift representation.
func Marshal(p Protocol, v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(p.NewWriter(buf))
	err := enc.Encode(v)
	return buf.Bytes(), err
}

type Encoder struct {
	w Writer
}

func NewEncoder(w Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(v interface{}) error {
	t := reflect.TypeOf(v)
	cache, _ := encoderCache.Load().(map[typeID]encodeFunc)
	encode, _ := cache[makeTypeID(t)]

	if encode == nil {
		encode = encodeFuncOf(t, make(encodeFuncCache))

		newCache := make(map[typeID]encodeFunc, len(cache)+1)
		newCache[makeTypeID(t)] = encode
		for k, v := range cache {
			newCache[k] = v
		}

		encoderCache.Store(newCache)
	}

	return encode(e.w, reflect.ValueOf(v), noflags)
}

func (e *Encoder) Reset(w Writer) {
	e.w = w
}

var encoderCache atomic.Value // map[typeID]encodeFunc

type encodeFunc func(Writer, reflect.Value, flags) error

type encodeFuncCache map[reflect.Type]encodeFunc

func encodeFuncOf(t reflect.Type, seen encodeFuncCache) encodeFunc {
	f := seen[t]
	if f != nil {
		return f
	}
	switch t.Kind() {
	case reflect.Bool:
		f = encodeBool
	case reflect.Int8:
		f = encodeInt8
	case reflect.Int16:
		f = encodeInt16
	case reflect.Int32:
		f = encodeInt32
	case reflect.Int64, reflect.Int:
		f = encodeInt64
	case reflect.Uint8:
		f = encodeUint8
	case reflect.Uint16:
		f = encodeUint16
	case reflect.Uint32:
		f = encodeUint32
	case reflect.Uint64, reflect.Uint, reflect.Uintptr:
		f = encodeUint64
	case reflect.Float32, reflect.Float64:
		f = encodeFloat64
	case reflect.String:
		f = encodeString
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			f = encodeBytes
		} else {
			f = encodeFuncSliceOf(t, seen)
		}
	case reflect.Map:
		f = encodeFuncMapOf(t, seen)
	case reflect.Struct:
		f = encodeFuncStructOf(t, seen)
	case reflect.Ptr:
		f = encodeFuncPtrOf(t, seen)
	default:
		panic("type cannot be encoded in thrift: " + t.String())
	}
	seen[t] = f
	return f
}

func encodeBool(w Writer, v reflect.Value, _ flags) error {
	return w.WriteBool(v.Bool())
}

func encodeInt8(w Writer, v reflect.Value, _ flags) error {
	return w.WriteInt8(int8(v.Int()))
}

func encodeInt16(w Writer, v reflect.Value, _ flags) error {
	return w.WriteInt16(int16(v.Int()))
}

func encodeInt32(w Writer, v reflect.Value, _ flags) error {
	return w.WriteInt32(int32(v.Int()))
}

func encodeInt64(w Writer, v reflect.Value, _ flags) error {
	return w.WriteInt64(v.Int())
}

func encodeUint8(w Writer, v reflect.Value, _ flags) error {
	return w.WriteInt8(int8(v.Uint()))
}

func encodeUint16(w Writer, v reflect.Value, _ flags) error {
	return w.WriteInt16(int16(v.Uint()))
}

func encodeUint32(w Writer, v reflect.Value, _ flags) error {
	return w.WriteInt32(int32(v.Uint()))
}

func encodeUint64(w Writer, v reflect.Value, _ flags) error {
	return w.WriteInt64(int64(v.Uint()))
}

func encodeFloat64(w Writer, v reflect.Value, _ flags) error {
	return w.WriteFloat64(v.Float())
}

func encodeString(w Writer, v reflect.Value, _ flags) error {
	return w.WriteString(v.String())
}

func encodeBytes(w Writer, v reflect.Value, _ flags) error {
	return w.WriteBytes(v.Bytes())
}

func encodeFuncSliceOf(t reflect.Type, seen encodeFuncCache) encodeFunc {
	elem := t.Elem()
	typ := TypeOf(elem)
	enc := encodeFuncOf(elem, seen)

	return func(w Writer, v reflect.Value, _ flags) error {
		n := v.Len()
		if n > math.MaxInt32 {
			return fmt.Errorf("slice length is too large to be represented in thrift: %d > max(int32)", n)
		}

		err := w.WriteList(List{
			Size: int32(n),
			Type: typ,
		})
		if err != nil {
			return err
		}

		for i := 0; i < n; i++ {
			if err := enc(w, v.Index(i), noflags); err != nil {
				return err
			}
		}

		return nil
	}
}

func encodeFuncMapOf(t reflect.Type, seen encodeFuncCache) encodeFunc {
	key, elem := t.Key(), t.Elem()
	if elem.Size() == 0 { // map[?]struct{}
		return encodeFuncMapAsSetOf(t, seen)
	}

	keyType := TypeOf(key)
	elemType := TypeOf(elem)
	encodeKey := encodeFuncOf(key, seen)
	encodeElem := encodeFuncOf(elem, seen)

	return func(w Writer, v reflect.Value, _ flags) error {
		n := v.Len()
		if n > math.MaxInt32 {
			return fmt.Errorf("map length is too large to be represented in thrift: %d > max(int32)", n)
		}

		err := w.WriteMap(Map{
			Size:  int32(n),
			Key:   keyType,
			Value: elemType,
		})
		if err != nil {
			return err
		}
		if n == 0 { // empty map
			return nil
		}

		for i, iter := 0, v.MapRange(); iter.Next(); i++ {
			if err := encodeKey(w, iter.Key(), noflags); err != nil {
				return err
			}
			if err := encodeElem(w, iter.Value(), noflags); err != nil {
				return err
			}
		}

		return nil
	}
}

func encodeFuncMapAsSetOf(t reflect.Type, seen encodeFuncCache) encodeFunc {
	key := t.Key()
	typ := TypeOf(key)
	enc := encodeFuncOf(key, seen)

	return func(w Writer, v reflect.Value, _ flags) error {
		n := v.Len()
		if n > math.MaxInt32 {
			return fmt.Errorf("map length is too large to be represented in thrift: %d > max(int32)", n)
		}

		err := w.WriteSet(Set{
			Size: int32(n),
			Type: typ,
		})
		if err != nil {
			return err
		}
		if n == 0 { // empty map
			return nil
		}

		for i, iter := 0, v.MapRange(); iter.Next(); i++ {
			if err := enc(w, iter.Key(), noflags); err != nil {
				return err
			}
		}

		return nil
	}
}

type structEncoder struct {
	fields []structEncoderField
	union  bool
}

func (enc *structEncoder) encode(w Writer, v reflect.Value, flags flags) error {
	numFields := 0

encodeFields:
	for _, f := range enc.fields {
		x := v
		for _, i := range f.index {
			if x.Kind() == reflect.Ptr {
				x = x.Elem()
			}
			if x = x.Field(i); x.Kind() == reflect.Ptr {
				if x.IsNil() {
					continue encodeFields
				}
			}
		}

		if !f.flags.have(required) && isZero(x) {
			continue encodeFields
		}

		field := Field{
			ID:   f.id,
			Type: f.typ,
		}
		if f.typ == BOOL && x.Bool() == true {
			field.Type = TRUE
		}

		if err := w.WriteField(field); err != nil {
			return err
		}

		if err := f.encode(w, x, f.flags); err != nil {
			return err
		}

		numFields++
	}

	if err := w.WriteField(Field{}); err != nil {
		return err
	}

	if numFields > 1 && enc.union {
		return fmt.Errorf("thrift union had more than one field with a non-zero value (%d)", numFields)
	}

	return nil
}

func (enc *structEncoder) String() string {
	if enc.union {
		return "union"
	}
	return "struct"
}

type structEncoderField struct {
	index  []int
	id     int16
	flags  flags
	typ    Type
	encode encodeFunc
}

func encodeFuncStructOf(t reflect.Type, seen encodeFuncCache) encodeFunc {
	enc := &structEncoder{
		fields: make([]structEncoderField, 0, t.NumField()),
	}
	encode := enc.encode
	seen[t] = encode

	forEachStructField(t, nil, func(f structField) {
		if f.flags.have(union) {
			enc.union = true
		} else {
			enc.fields = append(enc.fields, structEncoderField{
				index:  f.index,
				id:     f.id,
				flags:  f.flags,
				typ:    TypeOf(f.typ),
				encode: encodeFuncStructFieldOf(f, seen),
			})
		}
	})

	sort.SliceStable(enc.fields, func(i, j int) bool {
		return enc.fields[i].id < enc.fields[j].id
	})

	for i := len(enc.fields) - 1; i > 0; i-- {
		if enc.fields[i-1].id == enc.fields[i].id {
			panic(fmt.Errorf("thrift struct field id %d is present multiple times", enc.fields[i].id))
		}
	}

	return encode
}

func encodeFuncStructFieldOf(f structField, seen encodeFuncCache) encodeFunc {
	if f.flags.have(enum) {
		switch f.typ.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return encodeInt32
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return encodeUint32
		}
	}
	return encodeFuncOf(f.typ, seen)
}

func encodeFuncPtrOf(t reflect.Type, seen encodeFuncCache) encodeFunc {
	typ := t.Elem()
	enc := encodeFuncOf(typ, seen)
	zero := reflect.Zero(typ)

	return func(w Writer, v reflect.Value, f flags) error {
		if v.IsNil() {
			v = zero
		}
		return enc(w, v, f)
	}
}

func isZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Ptr:
		return v.IsNil()
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	default:
		return v.IsZero()
	}
}
