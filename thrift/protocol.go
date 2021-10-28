package thrift

import (
	"bufio"
	"io"
)

type Protocol interface {
	NewReader(r *bufio.Reader) Reader
	NewWriter(w *bufio.Writer) Writer
}

type Reader interface {
	io.Reader
	ReadBool() (bool, error)
	ReadInt8() (int8, error)
	ReadInt16() (int16, error)
	ReadInt32() (int32, error)
	ReadInt64() (int64, error)
	ReadFloat64() (float64, error)
	ReadBytes() ([]byte, error)
	ReadString() (string, error)
	ReadLength() (int, error)
	ReadMessage() (Message, error)
	ReadField() (Field, error)
	ReadList() (List, error)
	ReadMap() (Map, error)
}

type Writer interface {
	io.Writer
	WriteBool(bool) error
	WriteInt8(int8) error
	WriteInt16(int16) error
	WriteInt32(int32) error
	WriteInt64(int64) error
	WriteFloat64(float64) error
	WriteBytes([]byte) error
	WriteString(string) error
	WriteLength(int) error
	WriteMessage(Message) error
	WriteField(Field) error
	WriteList(List) error
	WriteMap(Map) error
}
