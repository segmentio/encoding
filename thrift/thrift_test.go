package thrift_test

import (
	"bytes"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/segmentio/encoding/thrift"
)

var marshalTestValues = [...]struct {
	scenario string
	values   []interface{}
}{
	{
		scenario: "bool",
		values:   []interface{}{false, true},
	},

	{
		scenario: "int",
		values: []interface{}{
			int(0),
			int(-1),
			int(1),
		},
	},

	{
		scenario: "int8",
		values: []interface{}{
			int8(0),
			int8(-1),
			int8(1),
			int8(math.MinInt8),
			int8(math.MaxInt8),
		},
	},

	{
		scenario: "int16",
		values: []interface{}{
			int16(0),
			int16(-1),
			int16(1),
			int16(math.MinInt16),
			int16(math.MaxInt16),
		},
	},

	{
		scenario: "int32",
		values: []interface{}{
			int32(0),
			int32(-1),
			int32(1),
			int32(math.MinInt32),
			int32(math.MaxInt32),
		},
	},

	{
		scenario: "int64",
		values: []interface{}{
			int64(0),
			int64(-1),
			int64(1),
			int64(math.MinInt64),
			int64(math.MaxInt64),
		},
	},

	{
		scenario: "uint",
		values: []interface{}{
			uint(0),
			uint(1),
		},
	},

	{
		scenario: "uint8",
		values: []interface{}{
			uint8(0),
			uint8(1),
			uint8(math.MaxUint8),
		},
	},

	{
		scenario: "uint16",
		values: []interface{}{
			uint16(0),
			uint16(1),
			uint16(math.MaxUint16),
		},
	},

	{
		scenario: "uint32",
		values: []interface{}{
			uint32(0),
			uint32(1),
			uint32(math.MaxUint32),
		},
	},

	{
		scenario: "uint64",
		values: []interface{}{
			uint64(0),
			uint64(1),
			uint64(math.MaxUint64),
		},
	},

	{
		scenario: "uintptr",
		values: []interface{}{
			uintptr(0),
			uintptr(1),
		},
	},

	{
		scenario: "string",
		values: []interface{}{
			"",
			"A",
			"1234567890",
			strings.Repeat("qwertyuiop", 100),
		},
	},

	{
		scenario: "[]byte",
		values: []interface{}{
			[]byte(""),
			[]byte("A"),
			[]byte("1234567890"),
			bytes.Repeat([]byte("qwertyuiop"), 100),
		},
	},

	{
		scenario: "[]string",
		values: []interface{}{
			[]string{},
			[]string{"A"},
			[]string{"hello", "world", "!!!"},
			[]string{"0", "1", "3", "4", "5", "6", "7", "8", "9"},
		},
	},

	{
		scenario: "map[string]int",
		values: []interface{}{
			map[string]int{},
			map[string]int{"A": 1},
			map[string]int{"hello": 1, "world": 2, "answer": 42},
		},
	},

	{
		scenario: "map[int64]struct{}",
		values: []interface{}{
			map[int64]struct{}{},
			map[int64]struct{}{0: {}, 1: {}, 2: {}},
		},
	},

	{
		scenario: "[]map[string]struct{}",
		values: []interface{}{
			[]map[string]struct{}{},
			[]map[string]struct{}{{}, {"A": {}, "B": {}, "C": {}}},
		},
	},

	{
		scenario: "struct{}",
		values:   []interface{}{struct{}{}},
	},

	{
		scenario: "Point2D",
		values: []interface{}{
			Point2D{},
			Point2D{X: 1},
			Point2D{Y: 2},
			Point2D{X: 3, Y: 4},
		},
	},

	{
		scenario: "RecursiveStruct",
		values: []interface{}{
			RecursiveStruct{},
			RecursiveStruct{Value: "hello"},
			RecursiveStruct{Value: "hello", Next: &RecursiveStruct{}},
			RecursiveStruct{Value: "hello", Next: &RecursiveStruct{Value: "world"}},
		},
	},

	{
		scenario: "StructWithEnum",
		values: []interface{}{
			StructWithEnum{},
			StructWithEnum{Enum: 1},
			StructWithEnum{Enum: 2},
		},
	},
}

type Point2D struct {
	X float64 `thrift:"1"`
	Y float64 `thrift:"2"`
}

type RecursiveStruct struct {
	Value string           `thrift:"1"`
	Next  *RecursiveStruct `thrift:"2"`
}

type StructWithEnum struct {
	Enum int8 `thrift:"1,enum"`
}

func TestMarshalUnmarshal(t *testing.T) {
	for _, p := range protocols {
		t.Run(p.name, func(t *testing.T) { testMarshalUnmarshal(t, p.proto) })
	}
}

func testMarshalUnmarshal(t *testing.T, p thrift.Protocol) {
	for _, test := range marshalTestValues {
		t.Run(test.scenario, func(t *testing.T) {
			for _, value := range test.values {
				b, err := thrift.Marshal(p, value)
				if err != nil {
					t.Fatal("marshal:", err)
				}

				v := reflect.New(reflect.TypeOf(value))
				if err := thrift.Unmarshal(p, b, v.Interface()); err != nil {
					t.Fatal("unmarshal:", err)
				}

				if result := v.Elem().Interface(); !reflect.DeepEqual(value, result) {
					t.Errorf("value mismatch:\nwant: %#v\ngot:  %#v", value, result)
				}
			}
		})
	}
}

func BenchmarkMarshal(b *testing.B) {
	for _, p := range protocols {
		b.Run(p.name, func(b *testing.B) { benchmarkMarshal(b, p.proto) })
	}
}

type BenchmarkEncodeType struct {
	Name     string               `thrift:"1"`
	Question string               `thrift:"2"`
	Answer   string               `thrift:"3"`
	Sub      *BenchmarkEncodeType `thrift:"4"`
}

func benchmarkMarshal(b *testing.B, p thrift.Protocol) {
	buf := new(bytes.Buffer)
	enc := thrift.NewEncoder(p.NewWriter(buf))
	val := &BenchmarkEncodeType{
		Name:     "Luke",
		Question: "How are you?",
		Answer:   "42",
		Sub: &BenchmarkEncodeType{
			Name:     "Leia",
			Question: "?",
			Answer:   "whatever",
		},
	}

	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc.Encode(val)
	}

	b.SetBytes(int64(buf.Len()))
}

func BenchmarkUnmarshal(b *testing.B) {
	for _, p := range protocols {
		b.Run(p.name, func(b *testing.B) { benchmarkUnmarshal(b, p.proto) })
	}
}

type BenchmarkDecodeType struct {
	Name     string               `thrift:"1"`
	Question string               `thrift:"2"`
	Answer   string               `thrift:"3"`
	Sub      *BenchmarkDecodeType `thrift:"4"`
}

func benchmarkUnmarshal(b *testing.B, p thrift.Protocol) {
	buf, _ := thrift.Marshal(p, &BenchmarkDecodeType{
		Name:     "Luke",
		Question: "How are you?",
		Answer:   "42",
		Sub: &BenchmarkDecodeType{
			Name:     "Leia",
			Question: "?",
			Answer:   "whatever",
		},
	})

	rb := bytes.NewReader(nil)
	dec := thrift.NewDecoder(p.NewReader(rb))
	val := &BenchmarkDecodeType{}

	for i := 0; i < b.N; i++ {
		rb.Reset(buf)
		dec.Decode(val)
	}

	b.SetBytes(int64(len(buf)))
}
