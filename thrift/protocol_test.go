package thrift_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/segmentio/encoding/thrift"
)

var protocolReadWriteTests = [...]struct {
	scenario string
	read     interface{}
	write    interface{}
	values   []interface{}
}{
	{
		scenario: "bool",
		read:     thrift.Reader.ReadBool,
		write:    thrift.Writer.WriteBool,
		values:   []interface{}{false, true},
	},

	{
		scenario: "int8",
		read:     thrift.Reader.ReadInt8,
		write:    thrift.Writer.WriteInt8,
		values:   []interface{}{int8(0), int8(1), int8(-1)},
	},

	{
		scenario: "int16",
		read:     thrift.Reader.ReadInt16,
		write:    thrift.Writer.WriteInt16,
		values:   []interface{}{int16(0), int16(1), int16(-1)},
	},

	{
		scenario: "int32",
		read:     thrift.Reader.ReadInt32,
		write:    thrift.Writer.WriteInt32,
		values:   []interface{}{int32(0), int32(1), int32(-1)},
	},

	{
		scenario: "int64",
		read:     thrift.Reader.ReadInt64,
		write:    thrift.Writer.WriteInt64,
		values:   []interface{}{int64(0), int64(1), int64(-1)},
	},

	{
		scenario: "float64",
		read:     thrift.Reader.ReadFloat64,
		write:    thrift.Writer.WriteFloat64,
		values:   []interface{}{float64(0), float64(1), float64(-1)},
	},

	{
		scenario: "bytes",
		read:     thrift.Reader.ReadBytes,
		write:    thrift.Writer.WriteBytes,
		values: []interface{}{
			[]byte(""),
			[]byte("A"),
			[]byte("1234567890"),
			bytes.Repeat([]byte("qwertyuiop"), 100),
		},
	},

	{
		scenario: "string",
		read:     thrift.Reader.ReadString,
		write:    thrift.Writer.WriteString,
		values: []interface{}{
			"",
			"A",
			"1234567890",
			strings.Repeat("qwertyuiop", 100),
		},
	},

	{
		scenario: "message",
		read:     thrift.Reader.ReadMessage,
		write:    thrift.Writer.WriteMessage,
		values: []interface{}{
			thrift.Message{},
			thrift.Message{Type: thrift.Call, Name: "Hello", SeqID: 10},
			thrift.Message{Type: thrift.Reply, Name: "World", SeqID: 11},
			thrift.Message{Type: thrift.Exception, Name: "Foo", SeqID: 40},
			thrift.Message{Type: thrift.Oneway, Name: "Bar", SeqID: 42},
		},
	},

	{
		scenario: "field",
		read:     thrift.Reader.ReadField,
		write:    thrift.Writer.WriteField,
		values: []interface{}{
			thrift.Field{ID: 101, Type: thrift.TRUE},
			thrift.Field{ID: 102, Type: thrift.FALSE},
			thrift.Field{ID: 103, Type: thrift.I8},
			thrift.Field{ID: 104, Type: thrift.I16},
			thrift.Field{ID: 105, Type: thrift.I32},
			thrift.Field{ID: 106, Type: thrift.I64},
			thrift.Field{ID: 107, Type: thrift.DOUBLE},
			thrift.Field{ID: 108, Type: thrift.BINARY},
			thrift.Field{ID: 109, Type: thrift.LIST},
			thrift.Field{ID: 110, Type: thrift.SET},
			thrift.Field{ID: 111, Type: thrift.MAP},
			thrift.Field{ID: 112, Type: thrift.STRUCT},
			thrift.Field{},
		},
	},

	{
		scenario: "list",
		read:     thrift.Reader.ReadList,
		write:    thrift.Writer.WriteList,
		values: []interface{}{
			thrift.List{},
			thrift.List{Size: 0, Type: thrift.BOOL},
			thrift.List{Size: 1, Type: thrift.I8},
			thrift.List{Size: 1000, Type: thrift.BINARY},
		},
	},

	{
		scenario: "map",
		read:     thrift.Reader.ReadMap,
		write:    thrift.Writer.WriteMap,
		values: []interface{}{
			thrift.Map{},
			thrift.Map{Size: 1, Key: thrift.BINARY, Value: thrift.MAP},
			thrift.Map{Size: 1000, Key: thrift.BINARY, Value: thrift.LIST},
		},
	},
}

var protocols = [...]struct {
	name  string
	proto thrift.Protocol
}{
	{
		name:  "binary(default)",
		proto: &thrift.BinaryProtocol{},
	},

	{
		name: "binary(non-strict)",
		proto: &thrift.BinaryProtocol{
			NonStrict: true,
		},
	},

	{
		name:  "compact",
		proto: &thrift.CompactProtocol{},
	},
}

func TestProtocols(t *testing.T) {
	for _, test := range protocols {
		t.Run(test.name, func(t *testing.T) { testProtocolReadWriteValues(t, test.proto) })
	}
}

func testProtocolReadWriteValues(t *testing.T, p thrift.Protocol) {
	for _, test := range protocolReadWriteTests {
		t.Run(test.scenario, func(t *testing.T) {
			b := new(bytes.Buffer)
			r := p.NewReader(b)
			w := p.NewWriter(b)

			for _, value := range test.values {
				ret := reflect.ValueOf(test.write).Call([]reflect.Value{
					reflect.ValueOf(w),
					reflect.ValueOf(value),
				})
				if err, _ := ret[0].Interface().(error); err != nil {
					t.Fatal("encoding:", err)
				}
			}

			for _, value := range test.values {
				ret := reflect.ValueOf(test.read).Call([]reflect.Value{
					reflect.ValueOf(r),
				})
				if err, _ := ret[1].Interface().(error); err != nil {
					t.Fatal("decoding:", err)
				}
				if res := ret[0].Interface(); !reflect.DeepEqual(value, res) {
					t.Errorf("value mismatch:\nwant: %#v\ngot:  %#v", value, res)
				}
			}

			if b.Len() != 0 {
				t.Errorf("unexpected trailing bytes: %d", b.Len())
			}
		})
	}
}
