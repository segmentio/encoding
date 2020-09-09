package proto

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Message is an interface implemented by types that supported being encoded to
// and decoded from protobuf.
type Message interface {
	// Size is the size of the protobuf representation (in bytes).
	Size() int

	// Marshal writes the message to the byte slice passed as argument.
	Marshal([]byte) error

	// Unmarshal reads the message from the byte slice passed as argument.
	Unmarshal([]byte) error
}

// RawMessage represents a raw protobuf-encoded message.
type RawMessage []byte

// Size satisfies the Message interface.
func (m RawMessage) Size() int { return len(m) }

// Marshal satisfies the Message interface.
func (m RawMessage) Marshal(b []byte) error {
	copy(b, m)
	return nil
}

// Unmarshal satisfies the Message interface.
func (m *RawMessage) Unmarshal(b []byte) error {
	*m = make([]byte, len(b))
	copy(*m, b)
	return nil
}

// Rewrite satisfies the Rewriter interface.
func (m RawMessage) Rewrite(out, _ []byte) ([]byte, error) {
	return append(out, m...), nil
}

// FieldNumber represents a protobuf field number.
type FieldNumber int

func (f FieldNumber) Bool(v bool) RawMessage {
	var x uint64
	if v {
		x = 1
	}
	return AppendVarint(nil, f, x)
}

func (f FieldNumber) Int(v int) RawMessage {
	return AppendVarint(nil, f, encodeZigZag64(int64(v)))
}

func (f FieldNumber) Int32(v int32) RawMessage {
	return AppendFixed32(nil, f, encodeZigZag32(v))
}

func (f FieldNumber) Int64(v int64) RawMessage {
	return AppendFixed64(nil, f, encodeZigZag64(v))
}

func (f FieldNumber) Uint(v uint) RawMessage {
	return AppendVarint(nil, f, uint64(v))
}

func (f FieldNumber) Uint32(v uint32) RawMessage {
	return AppendFixed32(nil, f, v)
}

func (f FieldNumber) Uint64(v uint64) RawMessage {
	return AppendFixed64(nil, f, v)
}

func (f FieldNumber) Float32(v float32) RawMessage {
	return AppendFixed32(nil, f, math.Float32bits(v))
}

func (f FieldNumber) Float64(v float64) RawMessage {
	return AppendFixed64(nil, f, math.Float64bits(v))
}

func (f FieldNumber) String(v string) RawMessage {
	return AppendVarlen(nil, f, []byte(v))
}

func (f FieldNumber) Bytes(v []byte) RawMessage {
	return AppendVarlen(nil, f, v)
}

// Value constructs a RawMessage for field number f from v.
func (f FieldNumber) Value(v interface{}) RawMessage {
	switch x := v.(type) {
	case bool:
		return f.Bool(x)
	case int:
		return f.Int(x)
	case int32:
		return f.Int32(x)
	case int64:
		return f.Int64(x)
	case uint:
		return f.Uint(x)
	case uint32:
		return f.Uint32(x)
	case uint64:
		return f.Uint64(x)
	case float32:
		return f.Float32(x)
	case float64:
		return f.Float64(x)
	case string:
		return f.String(x)
	case []byte:
		return f.Bytes(x)
	default:
		panic("cannot rewrite value of unsupported type")
	}
}

// The WireType enumeration represents the different protobuf wire types.
type WireType int

const (
	Varint  WireType = 0
	Fixed64 WireType = 1
	Varlen  WireType = 2
	Fixed32 WireType = 5
)

func (wt WireType) String() string {
	return wireType(wt).String()
}

func Append(m RawMessage, f FieldNumber, t WireType, v []byte) RawMessage {
	b := [24]byte{}
	n, _ := encodeVarint(b[:], EncodeTag(f, t))
	if t == Varlen {
		n1, _ := encodeVarint(b[n:], uint64(len(v)))
		n += n1
	}
	m = append(m, b[:n]...)
	m = append(m, v...)
	return m
}

func AppendVarint(m RawMessage, f FieldNumber, v uint64) RawMessage {
	b := [12]byte{}
	n, _ := encodeVarint(b[:], v)
	return Append(m, f, Varint, b[:n])
}

func AppendVarlen(m RawMessage, f FieldNumber, v []byte) RawMessage {
	return Append(m, f, Varlen, v)
}

func AppendFixed32(m RawMessage, f FieldNumber, v uint32) RawMessage {
	b := [4]byte{}
	binary.LittleEndian.PutUint32(b[:], v)
	return Append(m, f, Fixed32, b[:])
}

func AppendFixed64(m RawMessage, f FieldNumber, v uint64) RawMessage {
	b := [8]byte{}
	binary.LittleEndian.PutUint64(b[:], v)
	return Append(m, f, Fixed64, b[:])
}

func Parse(m RawMessage) (FieldNumber, WireType, RawValue, RawMessage, error) {
	tag, n, err := decodeVarint(m)
	if err != nil {
		return 0, 0, nil, m, fmt.Errorf("decoding protobuf field number: %w", err)
	}
	m = m[n:]
	f, t := DecodeTag(tag)

	switch t {
	case Varint:
		_, n, err := decodeVarint(m)
		if err != nil {
			return f, t, nil, m, fmt.Errorf("decoding varint field %d: %w", f, err)
		}
		return f, t, RawValue(m[:n]), m[n:], nil

	case Varlen:
		l, n, err := decodeVarint(m)
		if err != nil {
			return f, t, nil, m, fmt.Errorf("decoding varlen field %d: %w", f, err)
		}
		return f, t, RawValue(m[n : n+int(l)]), m[n+int(l):], nil

	case Fixed32:
		if len(m) < 4 {
			return f, t, nil, m, fmt.Errorf("decoding fixed32 field %d: %w", f, io.ErrUnexpectedEOF)
		}
		return f, t, RawValue(m[:4]), m[4:], nil

	case Fixed64:
		if len(m) < 8 {
			return f, t, nil, m, fmt.Errorf("decoding fixed64 field %d: %w", f, io.ErrUnexpectedEOF)
		}
		return f, t, RawValue(m[:8]), m[8:], nil

	default:
		return f, t, nil, m, fmt.Errorf("invalid wire type: %d", t)
	}
}

// RawValue represents a single protobuf value.
//
// RawValue instances are returned by Parse and share the backing array of the
// RawMessage that they were decoded from.
type RawValue []byte

// Varint decodes v as a varint.
//
// The content of v will always be a valid varint if v was returned by a call to
// Parse and the associated wire type was Varint. In other cases, the behavior
// of Varint is undefined.
func (v RawValue) Varint() uint64 {
	u, _, _ := decodeVarint(v)
	return u
}

// Fixed32 decodes v as a fixed32.
//
// The content of v will always be a valid fixed32 if v was returned by a call
// to Parse and the associated wire type was Fixed32. In other cases, the
// behavior of Fixed32 is undefined.
func (v RawValue) Fixed32() uint32 {
	return binary.LittleEndian.Uint32(v)
}

// Fixed64 decodes v as a fixed64.
//
// The content of v will always be a valid fixed64 if v was returned by a call
// to Parse and the associated wire type was Fixed64. In other cases, the
// behavior of Fixed64 is undefined.
func (v RawValue) Fixed64() uint64 {
	return binary.LittleEndian.Uint64(v)
}

var (
	_ Message  = &RawMessage{}
	_ Rewriter = RawMessage{}
)
