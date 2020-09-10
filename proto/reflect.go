package proto

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// Kind is an enumeration representing the various data types supported by the
// protobuf language.
type Kind int

const (
	Bool Kind = iota
	Int32
	Int64
	Sint32
	Sint64
	Uint32
	Uint64
	Fix32
	Fix64
	Sfix32
	Sfix64
	Float
	Double
	String
	Bytes
	Map
	Struct
)

// Type is an interface similar to reflect.Type. Values implementing this
// interface represent high level protobuf types.
type Type interface {
	String() string

	Kind() Kind

	Key() Type

	Elem() Type

	WireType() WireType

	NumField() int

	FieldByIndex(int) Field

	FieldByName(string) Field

	FieldByNumber(FieldNumber) Field
}

// TypeOf returns the protobuf type used to represent the go value v.
//
// The function uses the following table to map Go types to Protobuf:
//
//	Go      | Protobuf
//	--------+---------
//	bool    | bool
//	int     | sint64
//	int32   | sint32
//	int64   | sint64
//	uint    | uint64
//	uint32  | fixed32
//	uint64  | fixed64
//	float32 | float
//	float64 | double
//	string  | string
//	[]byte  | bytes
//	map     | map
//	struct  | message
//
// Pointer types are also supported and automatically dereferenced.
//
// Types are comparable values.
func TypeOf(v interface{}) Type {
	t := reflect.TypeOf(v)

	typesMutex.RLock()
	r, ok := typesCache[t]
	typesMutex.RUnlock()

	if ok {
		return r
	}

	r = typeOf(reflect.TypeOf(v))

	typesMutex.Lock()
	defer typesMutex.Unlock()

	if x, ok := typesCache[t]; ok {
		r = x
	} else {
		typesCache[t] = r
	}

	return r
}

func typeOf(t reflect.Type) Type {
	switch t.Kind() {
	case reflect.Bool:
		return &primitiveTypes[Bool]
	case reflect.Int:
		return &primitiveTypes[Sint64]
	case reflect.Int32:
		return &primitiveTypes[Sint32]
	case reflect.Int64:
		return &primitiveTypes[Sint64]
	case reflect.Uint:
		return &primitiveTypes[Uint64]
	case reflect.Uint32:
		return &primitiveTypes[Fix32]
	case reflect.Uint64:
		return &primitiveTypes[Fix64]
	case reflect.Float32:
		return &primitiveTypes[Float]
	case reflect.Float64:
		return &primitiveTypes[Double]
	case reflect.String:
		return &primitiveTypes[String]
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return &primitiveTypes[Bytes]
		}
	case reflect.Map:
		return mapTypeOf(t)
	case reflect.Struct:
		return structTypeOf(t)
	case reflect.Ptr:
		return typeOf(t.Elem())
	}
	panic(fmt.Errorf("cannot construct protobuf type from go value of type %s", t))
}

var (
	typesMutex sync.RWMutex
	typesCache = map[reflect.Type]Type{}
)

type Field struct {
	Index    int
	Number   FieldNumber
	Name     string
	Type     Type
	Repeated bool
}

type primitiveType struct {
	name string
	kind Kind
	wire WireType
}

func (t *primitiveType) String() string {
	return t.name
}

func (t *primitiveType) Kind() Kind {
	return t.kind
}

func (t *primitiveType) Key() Type {
	panic(fmt.Errorf("proto.Type.Key: called on unsupported type: %s", t.name))
}

func (t *primitiveType) Elem() Type {
	panic(fmt.Errorf("proto.Type.Elem: called on unsupported type: %s", t.name))
}

func (t *primitiveType) WireType() WireType {
	return t.wire
}

func (t *primitiveType) NumField() int {
	return 0
}

func (t *primitiveType) FieldByIndex(int) Field {
	panic(fmt.Errorf("proto.Type.FieldByIndex: called on unsupported type: %s", t.name))
}

func (t *primitiveType) FieldByName(string) Field {
	panic(fmt.Errorf("proto.Type.FieldByName: called on unsupported type: %s", t.name))
}

func (t *primitiveType) FieldByNumber(FieldNumber) Field {
	panic(fmt.Errorf("proto.Type.FieldByNumber: called on unsupported type: %s", t.name))
}

var primitiveTypes = [...]primitiveType{
	{name: "bool", kind: Bool, wire: Varint},
	{name: "int32", kind: Int32, wire: Varint},
	{name: "int64", kind: Int64, wire: Varint},
	{name: "sint32", kind: Sint32, wire: Varint},
	{name: "sint64", kind: Sint64, wire: Varint},
	{name: "uint32", kind: Uint32, wire: Varint},
	{name: "uint64", kind: Uint64, wire: Varint},
	{name: "fixed32", kind: Fix32, wire: Fixed32},
	{name: "fixed64", kind: Fix64, wire: Fixed64},
	{name: "sfixed32", kind: Sfix32, wire: Fixed32},
	{name: "sfixed64", kind: Sfix64, wire: Fixed64},
	{name: "float", kind: Float, wire: Fixed32},
	{name: "double", kind: Double, wire: Fixed64},
	{name: "string", kind: String, wire: Varlen},
	{name: "bytes", kind: Bytes, wire: Varlen},
}

func mapTypeOf(t reflect.Type) *mapType {
	return &mapType{
		key:  typeOf(t.Key()),
		elem: typeOf(t.Elem()),
	}
}

type mapType struct {
	key  Type
	elem Type
}

func (t *mapType) String() string {
	return fmt.Sprintf("map<%s, %s>", t.key, t.elem)
}

func (t *mapType) Kind() Kind {
	return Map
}

func (t *mapType) Key() Type {
	return t.key
}

func (t *mapType) Elem() Type {
	return t.elem
}

func (t *mapType) WireType() WireType {
	return Varlen
}

func (t *mapType) NumField() int {
	return 0
}

func (t *mapType) FieldByIndex(int) Field {
	panic(fmt.Errorf("proto.Type.FieldByIndex: called on unsupported type: %s", t))
}

func (t *mapType) FieldByName(string) Field {
	panic(fmt.Errorf("proto.Type.FieldByName: called on unsupported type: %s", t))
}

func (t *mapType) FieldByNumber(FieldNumber) Field {
	panic(fmt.Errorf("proto.Type.FieldByNumber: called on unsupported type: %s", t))
}

func structTypeOf(t reflect.Type) *structType {
	st := &structType{
		name:           t.Name(),
		fieldsByName:   make(map[string]int),
		fieldsByNumber: make(map[FieldNumber]int),
	}

	fieldNumber := FieldNumber(0)
	taggedFields := FieldNumber(0)

	for i, n := 0, t.NumField(); i < n; i++ {
		f := t.Field(i)

		if f.PkgPath != "" {
			continue // unexported
		}

		repeated := false
		if f.Type.Kind() == reflect.Slice && f.Type.Elem().Kind() != reflect.Uint8 {
			repeated = true
			f.Type = f.Type.Elem() // for typeOf
		}

		fieldName := f.Name
		fieldType := typeOf(f.Type)

		if tag, ok := f.Tag.Lookup("protobuf"); ok {
			if fieldNumber != taggedFields {
				panic("conflicting use of struct tag and naked fields")
			}
			t, err := parseStructTag(tag)
			if err != nil {
				panic(err)
			}
			fieldName = t.name
			fieldNumber = t.fieldNumber
			repeated = t.repeated
			taggedFields = t.fieldNumber
		} else if fieldNumber == 0 && len(st.fieldsByIndex) != 0 {
			panic("conflicting use of struct tag and naked fields")
		} else {
			fieldNumber++
		}

		index := len(st.fieldsByIndex)
		st.fieldsByIndex = append(st.fieldsByIndex, Field{
			Index:    index,
			Number:   fieldNumber,
			Name:     fieldName,
			Type:     fieldType,
			Repeated: repeated,
		})
		st.fieldsByName[fieldName] = index
		st.fieldsByNumber[fieldNumber] = index
	}

	return st
}

type structType struct {
	name           string
	fieldsByIndex  []Field
	fieldsByName   map[string]int
	fieldsByNumber map[FieldNumber]int
}

func (t *structType) String() string {
	return t.name
}

func (t *structType) Format(w fmt.State, v rune) {
	switch v {
	case 's':
		io.WriteString(w, t.name)
	case 'v':
		if t.name == "" {
			io.WriteString(w, "message {")
		} else {
			fmt.Fprintf(w, "message %s {", t.name)
		}

		for _, f := range t.fieldsByIndex {
			if f.Repeated {
				fmt.Fprintf(w, "\n  repeated %s %s = %d;", f.Type, f.Name, f.Number)
			} else {
				fmt.Fprintf(w, "\n  %s %s = %d;", f.Type, f.Name, f.Number)
			}
		}

		if len(t.fieldsByIndex) == 0 {
			io.WriteString(w, "}")
		} else {
			io.WriteString(w, "\n}")
		}
	}
}

func (t *structType) Kind() Kind {
	return Struct
}

func (t *structType) Key() Type {
	panic(fmt.Errorf("proto.Type.Key: called on unsupported type: %s", t.name))
}

func (t *structType) Elem() Type {
	panic(fmt.Errorf("proto.Type.Elem: called on unsupported type: %s", t.name))
}

func (t *structType) WireType() WireType {
	return Varlen
}

func (t *structType) NumField() int {
	return len(t.fieldsByIndex)
}

func (t *structType) FieldByIndex(index int) Field {
	if index >= 0 && index < len(t.fieldsByIndex) {
		return t.fieldsByIndex[index]
	}
	panic(fmt.Errorf("proto.Type.FieldByIndex: protobuf message field out of bounds: %d/%d", index, len(t.fieldsByIndex)))
}

func (t *structType) FieldByName(name string) Field {
	i, ok := t.fieldsByName[name]
	if ok {
		return t.fieldsByIndex[i]
	}
	panic(fmt.Errorf("proto.Type.FieldByName: protobuf message has not field named %q", name))
}

func (t *structType) FieldByNumber(number FieldNumber) Field {
	i, ok := t.fieldsByNumber[number]
	if ok {
		return t.fieldsByIndex[i]
	}
	panic(fmt.Errorf("proto.Type.FieldByNumber: protobuf message has no field number %d", number))
}

type structTag struct {
	name        string
	enum        string
	json        string
	version     int
	wireType    WireType
	fieldNumber FieldNumber
	extensions  map[string]string
	repeated    bool
}

func parseStructTag(tag string) (structTag, error) {
	t := structTag{
		version:    2,
		extensions: make(map[string]string),
	}

	for i, f := range splitFields(tag) {
		switch i {
		case 0:
			switch f {
			case "varint":
				t.wireType = Varint
			case "bytes":
				t.wireType = Varlen
			case "fixed32":
				t.wireType = Fixed32
			case "fixed64":
				t.wireType = Fixed64
			default:
				return t, fmt.Errorf("unsupported wire type in struct tag %q: %s", tag, f)
			}

		case 1:
			n, err := strconv.Atoi(f)
			if err != nil {
				return t, fmt.Errorf("unsupported field number in struct tag %q: %w", tag, err)
			}
			t.fieldNumber = FieldNumber(n)

		case 2:
			switch f {
			case "opt":
				// not sure what this is for
			case "rep":
				t.repeated = true
			default:
				return t, fmt.Errorf("unsupported field option in struct tag %q: %s", tag, f)
			}

		default:
			name, value := splitNameValue(f)
			switch name {
			case "name":
				t.name = value
			case "enum":
				t.enum = value
			case "json":
				t.json = value
			case "proto3":
				t.version = 3
			default:
				t.extensions[name] = value
			}
		}
	}

	return t, nil
}

func splitFields(s string) []string {
	return strings.Split(s, ",")
}

func splitNameValue(s string) (name, value string) {
	i := strings.IndexByte(s, '=')
	if i < 0 {
		return strings.TrimSpace(s), ""
	} else {
		return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:])
	}
}
