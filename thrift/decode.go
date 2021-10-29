package thrift

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"sync/atomic"
)

// Unmarshal deserializes the thrift data from b to v using to the protocol p.
//
// The function errors if the data in b does not match the type of v.
//
// The function panics if v cannot be converted to a thrift representation.
func Unmarshal(p Protocol, b []byte, v interface{}) error {
	br := bytes.NewReader(b)
	pr := p.NewReader(br)

	if err := NewDecoder(pr).Decode(v); err != nil {
		return err
	}

	if n := br.Len(); n != 0 {
		return fmt.Errorf("unexpected trailing bytes at the end of thrift input: %d", n)
	}

	return nil
}

type Decoder struct {
	r Reader
}

func NewDecoder(r Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) Decode(v interface{}) error {
	t := reflect.TypeOf(v)
	p := reflect.ValueOf(v)

	if t.Kind() != reflect.Ptr {
		panic("thrift.(*Decoder).Decode: expected pointer type but got " + t.String())
	}

	t = t.Elem()
	p = p.Elem()

	cache, _ := decoderCache.Load().(map[typeID]decodeFunc)
	decode, _ := cache[makeTypeID(t)]

	if decode == nil {
		decode = decodeFuncOf(t, make(decodeFuncCache))

		newCache := make(map[typeID]decodeFunc, len(cache)+1)
		newCache[makeTypeID(t)] = decode
		for k, v := range cache {
			newCache[k] = v
		}

		decoderCache.Store(newCache)
	}

	return decode(d.r, p)
}

func (d *Decoder) Reset(r Reader) {
	d.r = r
}

var decoderCache atomic.Value // map[typeID]decodeFunc

type decodeFunc func(Reader, reflect.Value) error

type decodeFuncCache map[reflect.Type]decodeFunc

func decodeFuncOf(t reflect.Type, seen decodeFuncCache) decodeFunc {
	f := seen[t]
	if f != nil {
		return f
	}
	switch t.Kind() {
	case reflect.Bool:
		f = decodeBool
	case reflect.Int8:
		f = decodeInt8
	case reflect.Int16:
		f = decodeInt16
	case reflect.Int32:
		f = decodeInt32
	case reflect.Int64, reflect.Int:
		f = decodeInt64
	case reflect.Uint8:
		f = decodeUint8
	case reflect.Uint16:
		f = decodeUint16
	case reflect.Uint32:
		f = decodeUint32
	case reflect.Uint64, reflect.Uint, reflect.Uintptr:
		f = decodeUint64
	case reflect.Float32, reflect.Float64:
		f = decodeFloat64
	case reflect.String:
		f = decodeString
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 { // []byte
			f = decodeBytes
		} else {
			f = decodeFuncSliceOf(t, seen)
		}
	case reflect.Map:
		f = decodeFuncMapOf(t, seen)
	case reflect.Struct:
		f = decodeFuncStructOf(t, seen)
	case reflect.Ptr:
		f = decodeFuncPtrOf(t, seen)
	default:
		panic("type cannot be decoded in thrift: " + t.String())
	}
	seen[t] = f
	return f
}

func decodeBool(r Reader, v reflect.Value) error {
	b, err := r.ReadBool()
	if err != nil {
		return err
	}
	v.SetBool(b)
	return nil
}

func decodeInt8(r Reader, v reflect.Value) error {
	i, err := r.ReadInt8()
	if err != nil {
		return err
	}
	v.SetInt(int64(i))
	return nil
}

func decodeInt16(r Reader, v reflect.Value) error {
	i, err := r.ReadInt16()
	if err != nil {
		return err
	}
	v.SetInt(int64(i))
	return nil
}

func decodeInt32(r Reader, v reflect.Value) error {
	i, err := r.ReadInt32()
	if err != nil {
		return err
	}
	v.SetInt(int64(i))
	return nil
}

func decodeInt64(r Reader, v reflect.Value) error {
	i, err := r.ReadInt64()
	if err != nil {
		return err
	}
	v.SetInt(int64(i))
	return nil
}

func decodeUint8(r Reader, v reflect.Value) error {
	u, err := r.ReadInt8()
	if err != nil {
		return err
	}
	v.SetUint(uint64(u))
	return nil
}

func decodeUint16(r Reader, v reflect.Value) error {
	u, err := r.ReadInt16()
	if err != nil {
		return err
	}
	v.SetUint(uint64(u))
	return nil
}

func decodeUint32(r Reader, v reflect.Value) error {
	u, err := r.ReadInt32()
	if err != nil {
		return err
	}
	v.SetUint(uint64(u))
	return nil
}

func decodeUint64(r Reader, v reflect.Value) error {
	u, err := r.ReadInt64()
	if err != nil {
		return err
	}
	v.SetUint(uint64(u))
	return nil
}

func decodeFloat64(r Reader, v reflect.Value) error {
	f, err := r.ReadFloat64()
	if err != nil {
		return err
	}
	v.SetFloat(f)
	return nil
}

func decodeString(r Reader, v reflect.Value) error {
	s, err := r.ReadString()
	if err != nil {
		return err
	}
	v.SetString(s)
	return nil
}

func decodeBytes(r Reader, v reflect.Value) error {
	b, err := r.ReadBytes()
	if err != nil {
		return err
	}
	v.SetBytes(b)
	return nil
}

func decodeFuncSliceOf(t reflect.Type, seen decodeFuncCache) decodeFunc {
	elem := t.Elem()
	typ := TypeOf(elem)
	dec := decodeFuncOf(elem, seen)

	return func(r Reader, v reflect.Value) error {
		l, err := r.ReadList()
		if err != nil {
			return fmt.Errorf("decoding thrift list header: %w", err)
		}

		// TODO: implement type conversions?
		if typ != l.Type {
			return fmt.Errorf("element type mismatch in decoded thrift list of length %d: want %s but got %s", l.Size, typ, l.Type)
		}

		v.Set(reflect.MakeSlice(t, int(l.Size), int(l.Size)))

		for i := 0; i < int(l.Size); i++ {
			if err := dec(r, v.Index(i)); err != nil {
				return fmt.Errorf("decoding thrift list element of type %s at index %d: %w", l.Type, i, err)
			}
		}

		return nil
	}
}

func decodeFuncMapOf(t reflect.Type, seen decodeFuncCache) decodeFunc {
	key, elem := t.Key(), t.Elem()
	if elem.Size() == 0 { // map[?]struct{}
		return decodeFuncMapAsSetOf(t, seen)
	}

	mapType := reflect.MapOf(key, elem)
	keyZero := reflect.Zero(key)
	elemZero := reflect.Zero(elem)
	keyType := TypeOf(key)
	elemType := TypeOf(elem)
	decodeKey := decodeFuncOf(key, seen)
	decodeElem := decodeFuncOf(elem, seen)

	return func(r Reader, v reflect.Value) error {
		m, err := r.ReadMap()
		if err != nil {
			return fmt.Errorf("decoding thrift map header: %w", err)
		}

		v.Set(reflect.MakeMapWithSize(mapType, int(m.Size)))

		if m.Size == 0 { // empty map
			return nil
		}

		// TODO: implement type conversions?
		if keyType != m.Key {
			return fmt.Errorf("key type mismatch in decoded thrift map of length %d: want %s but got %s", m.Size, keyType, m.Key)
		}
		if elemType != m.Value {
			return fmt.Errorf("value type mismatch in decoded thrift map of length %d: want %s but got %s", m.Size, elemType, m.Value)
		}

		tmpKey := reflect.New(key).Elem()
		tmpElem := reflect.New(elem).Elem()

		for i := 0; i < int(m.Size); i++ {
			if err := decodeKey(r, tmpKey); err != nil {
				return fmt.Errorf("decoding thrift map key of type %s at index %d: %w", m.Key, i, err)
			}
			if err := decodeElem(r, tmpElem); err != nil {
				return fmt.Errorf("decoding thrift map value of type %s at index %d: %w", m.Value, i, err)
			}
			v.SetMapIndex(tmpKey, tmpElem)
			tmpKey.Set(keyZero)
			tmpElem.Set(elemZero)
		}

		return nil
	}
}

func decodeFuncMapAsSetOf(t reflect.Type, seen decodeFuncCache) decodeFunc {
	key, elem := t.Key(), t.Elem()
	keyZero := reflect.Zero(key)
	elemZero := reflect.Zero(elem)
	typ := TypeOf(key)
	dec := decodeFuncOf(key, seen)

	return func(r Reader, v reflect.Value) error {
		s, err := r.ReadSet()
		if err != nil {
			return fmt.Errorf("decoding thrift set header: %w", err)
		}

		v.Set(reflect.MakeMapWithSize(t, int(s.Size)))

		if s.Size == 0 {
			return nil
		}

		// TODO: implement type conversions?
		if typ != s.Type {
			return fmt.Errorf("element type mismatch in decoded thrift set of length %d: want %s but got %s", s.Size, typ, s.Type)
		}

		tmp := reflect.New(key).Elem()

		for i := 0; i < int(s.Size); i++ {
			if err := dec(r, tmp); err != nil {
				return fmt.Errorf("decoding thrift set element of type %s at index %d: %w", s.Type, i, err)
			}
			v.SetMapIndex(tmp, elemZero)
			tmp.Set(keyZero)
		}

		return nil
	}
}

type structDecoder struct {
	fields []structDecoderField
	minID  int16
	zero   reflect.Value
}

func (dec *structDecoder) decode(r Reader, v reflect.Value) error {
	v.Set(dec.zero)
	return readStruct(r, func(r Reader, f Field) error {
		i := int(f.ID) - int(dec.minID)
		if i < 0 || i >= len(dec.fields) || dec.fields[i].decode == nil {
			return skipField(r, f)
		}
		field := &dec.fields[i]

		// TODO: implement type conversions?
		if f.Type != field.typ {
			return fmt.Errorf("value type mismatch in decoded thrift struct field %d: want %s but got %s", f.ID, field.typ, f.Type)
		}

		x := v
		for _, i := range field.index {
			if x.Kind() == reflect.Ptr {
				x = x.Elem()
			}
			if x = x.Field(i); x.Kind() == reflect.Ptr {
				if x.IsNil() {
					x.Set(reflect.New(x.Type().Elem()))
				}
			}
		}

		return field.decode(r, x)
	})
}

type structDecoderField struct {
	index  []int
	id     int16
	typ    Type
	decode decodeFunc
}

func decodeFuncStructOf(t reflect.Type, seen decodeFuncCache) decodeFunc {
	dec := &structDecoder{
		zero: reflect.Zero(t),
	}
	decode := dec.decode
	seen[t] = decode

	fields := make([]structDecoderField, 0, t.NumField())
	forEachStructField(t, nil, func(f structField) {
		fields = append(fields, structDecoderField{
			index:  f.index,
			id:     f.id,
			typ:    TypeOf(f.typ),
			decode: decodeFuncStructFieldOf(f, seen),
		})
	})

	minID := int16(0)
	maxID := int16(0)

	for _, f := range fields {
		if f.id < minID || minID == 0 {
			minID = f.id
		}
		if f.id > maxID {
			maxID = f.id
		}
	}

	dec.fields = make([]structDecoderField, (maxID-minID)+1)
	dec.minID = minID

	for _, f := range fields {
		i := f.id - minID
		if dec.fields[i].decode != nil {
			panic(fmt.Errorf("thrift struct field id %d is present multiple times", f.id))
		}
		dec.fields[i] = f
	}

	return decode
}

func decodeFuncStructFieldOf(f structField, seen decodeFuncCache) decodeFunc {
	if f.enum {
		switch f.typ.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return decodeInt32
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return decodeUint32
		}
	}
	return decodeFuncOf(f.typ, seen)
}

func decodeFuncPtrOf(t reflect.Type, seen decodeFuncCache) decodeFunc {
	elem := t.Elem()
	decode := decodeFuncOf(t.Elem(), seen)
	return func(r Reader, v reflect.Value) error {
		if v.IsNil() {
			v.Set(reflect.New(elem))
		}
		return decode(r, v.Elem())
	}
}

func readBinary(r Reader, f func(io.Reader) error) error {
	n, err := r.ReadLength()
	if err != nil {
		return fmt.Errorf("decoding thrift binary value length: %w", err)
	}
	if err := f(io.LimitReader(r.Reader(), int64(n))); err != nil {
		return fmt.Errorf("decoding thrift binary value of length %d: %w", n, err)
	}
	return nil
}

func readList(r Reader, f func(Reader, Type) error) error {
	l, err := r.ReadList()
	if err != nil {
		return fmt.Errorf("decoding thrift list header: %w", err)
	}

	for i := 0; i < int(l.Size); i++ {
		if err := f(r, l.Type); err != nil {
			return fmt.Errorf("decoding thrift list element of type %s at index %d: %w", l.Type, i, err)
		}
	}

	return nil
}

func readSet(r Reader, f func(Reader, Type) error) error {
	s, err := r.ReadSet()
	if err != nil {
		return fmt.Errorf("decoding thrift set header: %w", err)
	}

	for i := 0; i < int(s.Size); i++ {
		if err := f(r, s.Type); err != nil {
			return fmt.Errorf("decoding thrift set element of type %s at index %d: %w", s.Type, i, err)
		}
	}

	return nil
}

func readMap(r Reader, f func(Reader, Type, Type) error) error {
	m, err := r.ReadMap()
	if err != nil {
		return err
	}

	for i := 0; i < int(m.Size); i++ {
		if err := f(r, m.Key, m.Value); err != nil {
			return fmt.Errorf("decoding thrift map entry at index %d: %w", i, err)
		}
	}

	return nil
}

func readStruct(r Reader, f func(Reader, Field) error) error {
	for {
		e, err := r.ReadField()
		if err != nil {
			return err
		}

		if e.ID == 0 && e.Type == 0 {
			return nil
		}

		if err := f(r, e); err != nil {
			return fmt.Errorf("decoding thrift struct field id %d of type %s: %w", e.ID, e.Type, err)
		}
	}
}

func skip(r Reader, t Type) error {
	var err error
	switch t {
	case TRUE, FALSE:
		_, err = r.ReadBool()
	case I8:
		_, err = r.ReadInt8()
	case I16:
		_, err = r.ReadInt16()
	case I32:
		_, err = r.ReadInt32()
	case I64:
		_, err = r.ReadInt64()
	case DOUBLE:
		_, err = r.ReadFloat64()
	case BINARY:
		err = skipBinary(r)
	case LIST:
		err = skipList(r)
	case SET:
		err = skipSet(r)
	case MAP:
		err = skipMap(r)
	case STRUCT:
		err = skipStruct(r)
	default:
		return fmt.Errorf("skipping unsupported thrift type %d", t)
	}
	return err
}

func skipField(r Reader, f Field) error {
	if err := skip(r, f.Type); err != nil {
		return fmt.Errorf("skipping thrift field id %d of type %s: %w", f.ID, f.Type, err)
	}
	return nil
}

func skipBinary(r Reader) error {
	n, err := r.ReadLength()
	if err != nil {
		return err
	}
	switch x := r.Reader().(type) {
	case *bufio.Reader:
		_, err = x.Discard(int(n))
	default:
		_, err = io.CopyN(ioutil.Discard, x, int64(n))
	}
	return err
}

func skipList(r Reader) error {
	return readList(r, skip)
}

func skipSet(r Reader) error {
	return readSet(r, skip)
}

func skipMap(r Reader) error {
	return readMap(r, func(r Reader, k, v Type) error {
		if err := skip(r, k); err != nil {
			return fmt.Errorf("skipping thrift map key of type %s: %w", v, err)
		}
		if err := skip(r, v); err != nil {
			return fmt.Errorf("skipping thrift map value of type %s: %w", v, err)
		}
		return nil
	})
}

func skipStruct(r Reader) error {
	return readStruct(r, skipField)
}
