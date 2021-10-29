package thrift

import (
	"fmt"
	"reflect"
)

type Message struct {
	Type  MessageType
	Name  string
	SeqID int32
}

type MessageType int8

const (
	Call MessageType = iota
	Reply
	Exception
	Oneway
)

func (m MessageType) String() string {
	switch m {
	case Call:
		return "Call"
	case Reply:
		return "Reply"
	case Exception:
		return "Exception"
	case Oneway:
		return "Oneway"
	default:
		return "?"
	}
}

type Field struct {
	ID   int16
	Type Type
}

func (f Field) String() string {
	return fmt.Sprintf("FIELD<%s>(%d)", f.Type, f.ID)
}

type Type int8

const (
	TRUE Type = iota + 1
	FALSE
	I8
	I16
	I32
	I64
	DOUBLE
	BINARY
	LIST
	SET
	MAP
	STRUCT
	BOOL = FALSE
)

func (t Type) String() string {
	switch t {
	case TRUE:
		return "TRUE"
	case BOOL:
		return "BOOL"
	case I8:
		return "I8"
	case I16:
		return "I16"
	case I32:
		return "I32"
	case I64:
		return "I64"
	case DOUBLE:
		return "DOUBLE"
	case BINARY:
		return "BINARY"
	case LIST:
		return "LIST"
	case SET:
		return "SET"
	case MAP:
		return "MAP"
	case STRUCT:
		return "STRUCT"
	default:
		return "?"
	}
}

type List struct {
	Size int32
	Type Type
}

func (l List) String() string {
	return fmt.Sprintf("LIST<%s>(%d)", l.Type, l.Size)
}

type Set List

func (s Set) String() string {
	return fmt.Sprintf("SET<%s>(%d)", s.Type, s.Size)
}

type Map struct {
	Size  int32
	Key   Type
	Value Type
}

func (m Map) String() string {
	return fmt.Sprintf("MAP<%s,%s>(%d)", m.Key, m.Value, m.Size)
}

func TypeOf(t reflect.Type) Type {
	switch t.Kind() {
	case reflect.Bool:
		return BOOL
	case reflect.Int8, reflect.Uint8:
		return I8
	case reflect.Int16, reflect.Uint16:
		return I16
	case reflect.Int32, reflect.Uint32:
		return I32
	case reflect.Int64, reflect.Uint64, reflect.Int, reflect.Uint, reflect.Uintptr:
		return I64
	case reflect.Float32, reflect.Float64:
		return DOUBLE
	case reflect.String:
		return BINARY
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 { // []byte
			return BINARY
		} else {
			return LIST
		}
	case reflect.Map:
		if t.Elem().Size() == 0 {
			return SET
		} else {
			return MAP
		}
	case reflect.Struct:
		return STRUCT
	case reflect.Ptr:
		return TypeOf(t.Elem())
	default:
		panic("type cannot be represented in thrift: " + t.String())
	}
}
