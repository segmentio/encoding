package thrift

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
)

type CompactProtocol struct{}

func (p *CompactProtocol) NewReader(r *bufio.Reader) Reader {
	return p.NewCompactReader(r)
}

func (p *CompactProtocol) NewWriter(w *bufio.Writer) Writer {
	return p.NewCompactWriter(w)
}

func (p *CompactProtocol) NewCompactReader(r *bufio.Reader) *CompactReader {
	return &CompactReader{binary: BinaryReader{r: r}}
}

func (p *CompactProtocol) NewCompactWriter(w *bufio.Writer) *CompactWriter {
	return &CompactWriter{binary: BinaryWriter{w: w}}
}

type CompactReader struct {
	binary    BinaryReader
	lastField Field
}

func (r *CompactReader) Read(b []byte) (int, error) {
	return r.binary.Read(b)
}

func (r *CompactReader) ReadBool() (bool, error) {
	switch r.lastField.Type {
	case TRUE:
		r.lastField = Field{}
		return true, nil
	case FALSE:
		r.lastField = Field{}
		return false, nil
	default:
		return r.binary.ReadBool()
	}
}

func (r *CompactReader) ReadInt8() (int8, error) {
	return r.binary.ReadInt8()
}

func (r *CompactReader) ReadInt16() (int16, error) {
	v, err := r.readVarint("int16", math.MinInt16, math.MaxInt16)
	return int16(v), err
}

func (r *CompactReader) ReadInt32() (int32, error) {
	v, err := r.readVarint("int32", math.MinInt32, math.MaxInt32)
	return int32(v), err
}

func (r *CompactReader) ReadInt64() (int64, error) {
	return r.readVarint("int64", math.MinInt64, math.MaxInt64)
}

func (r *CompactReader) ReadFloat64() (float64, error) {
	return r.binary.ReadFloat64()
}

func (r *CompactReader) ReadBytes() ([]byte, error) {
	n, err := r.ReadLength()
	if err != nil {
		return nil, err
	}
	b, err := r.binary.read(n)
	return copyBytes(b), err
}

func (r *CompactReader) ReadString() (string, error) {
	n, err := r.ReadLength()
	if err != nil {
		return "", err
	}
	b, err := r.binary.read(n)
	return string(b), err
}

func (r *CompactReader) ReadLength() (int, error) {
	n, err := r.readVarint("length", 0, math.MaxInt32)
	if err != nil {
		return 0, err
	}
	if n < 0 || n > math.MaxInt32 {
		return 0, fmt.Errorf("length out of range: %d", n)
	}
	return int(n), nil
}

func (r *CompactReader) ReadMessage() (Message, error) {
	m := Message{}

	b0, err := r.binary.readByte()
	if err != nil {
		return m, dontExpectEOF(err)
	}
	if b0 != 0x82 {
		return m, fmt.Errorf("invalid protocol id found when reading thrift message: %#x", b0)
	}

	b1, err := r.binary.readByte()
	if err != nil {
		return m, dontExpectEOF(err)
	}

	m.Type = MessageType(b1) & 0x7
	m.SeqID, err = r.ReadInt32()
	if err != nil {
		return m, err
	}
	m.Name, err = r.ReadString()
	return m, err
}

func (r *CompactReader) ReadField() (Field, error) {
	f := Field{}
	defer func() { r.lastField = f }()

	b, err := r.binary.readByte()
	if err != nil {
		return f, dontExpectEOF(err)
	}

	if b == 0 { // stop field
		return f, nil
	}

	if (b >> 4) != 0 {
		f = Field{ID: int16(b>>4) + r.lastField.ID, Type: Type(b & 0xF)}
	} else {
		i, err := r.ReadInt16()
		if err != nil {
			return f, err
		}
		f = Field{ID: i, Type: Type(b)}
	}

	return f, nil
}

func (r *CompactReader) ReadList() (List, error) {
	b, err := r.binary.readByte()
	if err != nil {
		return List{}, dontExpectEOF(err)
	}
	if (b >> 4) != 0xF {
		return List{Size: int32(b >> 4), Type: Type(b & 0xF)}, nil
	}
	n, err := r.ReadInt32()
	if err != nil {
		return List{}, err
	}
	return List{Size: n, Type: Type(b & 0xF)}, nil
}

func (r *CompactReader) ReadMap() (Map, error) {
	n, err := r.ReadInt32()
	if err != nil {
		return Map{}, err
	}
	if n == 0 { // empty map
		return Map{}, nil
	}
	b, err := r.binary.readByte()
	if err != nil {
		return Map{}, dontExpectEOF(err)
	}
	return Map{Size: n, Key: Type(b >> 4), Value: Type(b & 0xF)}, nil
}

func (r *CompactReader) readVarint(typ string, min, max int64) (int64, error) {
	v, err := binary.ReadVarint(r.binary.r)
	if err == nil {
		if v < min || v > max {
			err = fmt.Errorf("%s varint out of range: %d not in [%d;%d]", typ, v, min, max)
		}
	}
	return v, err
}

type CompactWriter struct {
	binary    BinaryWriter
	varint    [binary.MaxVarintLen64]byte
	lastField Field
}

func (w *CompactWriter) Write(b []byte) (int, error) {
	return w.binary.Write(b)
}

func (w *CompactWriter) WriteBool(v bool) error {
	switch w.lastField.Type {
	case TRUE, FALSE:
		return nil // the value is already encoded in the type
	default:
		return w.binary.WriteBool(v)
	}
}

func (w *CompactWriter) WriteInt8(v int8) error {
	return w.binary.WriteInt8(v)
}

func (w *CompactWriter) WriteInt16(v int16) error {
	return w.writeVarint(int64(v))
}

func (w *CompactWriter) WriteInt32(v int32) error {
	return w.writeVarint(int64(v))
}

func (w *CompactWriter) WriteInt64(v int64) error {
	return w.writeVarint(v)
}

func (w *CompactWriter) WriteFloat64(v float64) error {
	return w.binary.WriteFloat64(v)
}

func (w *CompactWriter) WriteBytes(v []byte) error {
	if err := w.WriteLength(len(v)); err != nil {
		return err
	}
	return w.binary.write(v)
}

func (w *CompactWriter) WriteString(v string) error {
	if err := w.WriteLength(len(v)); err != nil {
		return err
	}
	return w.binary.writeString(v)
}

func (w *CompactWriter) WriteLength(n int) error {
	if n < 0 {
		return fmt.Errorf("negative length cannot be encoded in thrift: %d", n)
	}
	if n > math.MaxInt32 {
		return fmt.Errorf("length is too large to be encoded in thrift: %d", n)
	}
	return w.writeVarint(int64(n))
}

func (w *CompactWriter) WriteMessage(m Message) error {
	if err := w.binary.writeByte(0x82); err != nil {
		return err
	}
	if err := w.binary.writeByte(byte(m.Type)); err != nil {
		return err
	}
	if err := w.WriteInt32(m.SeqID); err != nil {
		return err
	}
	return w.WriteString(m.Name)
}

func (w *CompactWriter) WriteField(f Field) error {
	defer func() { w.lastField = f }()
	if f.ID == 0 && f.Type == 0 { // stop field
		return w.binary.writeByte(0)
	}
	if f.ID > w.lastField.ID && (f.ID-w.lastField.ID) <= 15 { // delta encoding
		return w.binary.writeByte(byte((f.ID-w.lastField.ID)<<4) | byte(f.Type))
	}
	if err := w.WriteInt8(int8(f.Type)); err != nil {
		return err
	}
	return w.WriteInt16(f.ID)
}

func (w *CompactWriter) WriteList(l List) error {
	if l.Size <= 15 {
		return w.binary.writeByte(byte(l.Size<<4) | byte(l.Type))
	}
	if err := w.binary.writeByte(0xF0 | byte(l.Type)); err != nil {
		return err
	}
	return w.WriteInt32(l.Size)
}

func (w *CompactWriter) WriteMap(m Map) error {
	if err := w.WriteInt32(m.Size); err != nil || m.Size == 0 {
		return err
	}
	return w.binary.writeByte((byte(m.Key) << 4) | byte(m.Value))
}

func (w *CompactWriter) writeVarint(v int64) error {
	n := binary.PutVarint(w.varint[:], v)
	return w.binary.write(w.varint[:n])
}

var (
	_ Protocol = (*CompactProtocol)(nil)
	_ Reader   = (*CompactReader)(nil)
	_ Writer   = (*CompactWriter)(nil)
)
