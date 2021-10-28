package thrift

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

type BinaryProtocol struct {
	NonStrict bool
}

func (p *BinaryProtocol) NewReader(r *bufio.Reader) Reader {
	return p.NewBinaryReader(r)
}

func (p *BinaryProtocol) NewWriter(w *bufio.Writer) Writer {
	return p.NewBinaryWriter(w)
}

func (p *BinaryProtocol) NewBinaryReader(r *bufio.Reader) *BinaryReader {
	return &BinaryReader{r: r}
}

func (p *BinaryProtocol) NewBinaryWriter(w *bufio.Writer) *BinaryWriter {
	return &BinaryWriter{p: p, w: w}
}

type BinaryReader struct {
	r *bufio.Reader
	b bytes.Buffer
	l io.LimitedReader
}

func (r *BinaryReader) Read(b []byte) (int, error) {
	return r.r.Read(b)
}

func (r *BinaryReader) ReadBool() (bool, error) {
	v, err := r.readByte()
	return v != 0, err
}

func (r *BinaryReader) ReadInt8() (int8, error) {
	b, err := r.readByte()
	return int8(b), err
}

func (r *BinaryReader) ReadInt16() (int16, error) {
	b, err := r.r.Peek(2)
	if len(b) < 2 {
		return 0, dontExpectEOF(err)
	}
	v := int16(binary.BigEndian.Uint16(b))
	r.r.Discard(2)
	return v, nil
}

func (r *BinaryReader) ReadInt32() (int32, error) {
	b, err := r.r.Peek(4)
	if len(b) < 4 {
		return 0, dontExpectEOF(err)
	}
	v := int32(binary.BigEndian.Uint32(b))
	r.r.Discard(4)
	return v, nil
}

func (r *BinaryReader) ReadInt64() (int64, error) {
	b, err := r.r.Peek(8)
	if len(b) < 8 {
		return 0, dontExpectEOF(err)
	}
	v := int64(binary.BigEndian.Uint64(b))
	r.r.Discard(8)
	return v, nil
}

func (r *BinaryReader) ReadFloat64() (float64, error) {
	b, err := r.r.Peek(8)
	if len(b) < 8 {
		return 0, dontExpectEOF(err)
	}
	v := math.Float64frombits(binary.BigEndian.Uint64(b))
	r.r.Discard(8)
	return v, nil
}

func (r *BinaryReader) ReadBytes() ([]byte, error) {
	n, err := r.ReadLength()
	if err != nil {
		return nil, err
	}
	b, err := r.read(n)
	return copyBytes(b), err
}

func (r *BinaryReader) ReadString() (string, error) {
	n, err := r.ReadLength()
	if err != nil {
		return "", err
	}
	b, err := r.read(n)
	return string(b), err
}

func (r *BinaryReader) ReadLength() (int, error) {
	b, err := r.r.Peek(4)
	if len(b) < 4 {
		return 0, dontExpectEOF(err)
	}
	n := binary.BigEndian.Uint32(b)
	if n > math.MaxInt32 {
		return 0, fmt.Errorf("length out of range: %d", n)
	}
	r.r.Discard(4)
	return int(n), nil
}

func (r *BinaryReader) ReadMessage() (Message, error) {
	m := Message{}

	b, err := r.r.Peek(4)
	if len(b) < 4 {
		return m, dontExpectEOF(err)
	}

	if (b[0] >> 7) == 0 { // non-strict
		if m.Name, err = r.ReadString(); err != nil {
			return m, err
		}
		t, err := r.ReadInt8()
		if err != nil {
			return m, err
		}
		m.Type = MessageType(t & 0x7)
	} else {
		m.Type = MessageType(b[3] & 0x7)
		r.r.Discard(4)

		if m.Name, err = r.ReadString(); err != nil {
			return m, err
		}
	}

	m.SeqID, err = r.ReadInt32()
	return m, err
}

func (r *BinaryReader) ReadField() (Field, error) {
	t, err := r.ReadInt8()
	if err != nil {
		return Field{}, err
	}
	i, err := r.ReadInt16()
	if err != nil {
		return Field{}, err
	}
	return Field{ID: i, Type: Type(t)}, nil
}

func (r *BinaryReader) ReadList() (List, error) {
	t, err := r.ReadInt8()
	if err != nil {
		return List{}, err
	}
	n, err := r.ReadInt32()
	if err != nil {
		return List{}, err
	}
	return List{Size: n, Type: Type(t)}, nil
}

func (r *BinaryReader) ReadMap() (Map, error) {
	k, err := r.readByte()
	if err != nil {
		return Map{}, dontExpectEOF(err)
	}
	v, err := r.readByte()
	if err != nil {
		return Map{}, dontExpectEOF(err)
	}
	n, err := r.ReadInt32()
	if err != nil {
		return Map{}, err
	}
	return Map{Size: n, Key: Type(k), Value: Type(v)}, nil
}

func (r *BinaryReader) readByte() (byte, error) {
	return r.r.ReadByte()
}

func (r *BinaryReader) read(n int) ([]byte, error) {
	r.b.Reset()

	if b, _ := r.r.Peek(n); len(b) == n {
		r.b.Write(b)
		r.r.Discard(n)
		return r.b.Bytes(), nil
	}

	r.l.R = r.r
	r.l.N = int64(n)
	_, err := r.b.ReadFrom(&r.l)
	return r.b.Bytes(), err
}

type BinaryWriter struct {
	p *BinaryProtocol
	w *bufio.Writer
	b [8]byte
}

func (w *BinaryWriter) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

func (w *BinaryWriter) WriteBool(v bool) error {
	var b byte
	if v {
		b = 1
	}
	return w.w.WriteByte(b)
}

func (w *BinaryWriter) WriteInt8(v int8) error {
	return w.w.WriteByte(byte(v))
}

func (w *BinaryWriter) WriteInt16(v int16) error {
	binary.BigEndian.PutUint16(w.b[:2], uint16(v))
	return w.write(w.b[:2])
}

func (w *BinaryWriter) WriteInt32(v int32) error {
	binary.BigEndian.PutUint32(w.b[:4], uint32(v))
	return w.write(w.b[:4])
}

func (w *BinaryWriter) WriteInt64(v int64) error {
	binary.BigEndian.PutUint64(w.b[:8], uint64(v))
	return w.write(w.b[:8])
}

func (w *BinaryWriter) WriteFloat64(v float64) error {
	binary.BigEndian.PutUint64(w.b[:8], math.Float64bits(v))
	return w.write(w.b[:8])
}

func (w *BinaryWriter) WriteBytes(v []byte) error {
	if err := w.WriteLength(len(v)); err != nil {
		return err
	}
	return w.write(v)
}

func (w *BinaryWriter) WriteString(v string) error {
	if err := w.WriteLength(len(v)); err != nil {
		return err
	}
	return w.writeString(v)
}

func (w *BinaryWriter) WriteLength(n int) error {
	if n < 0 {
		return fmt.Errorf("negative length cannot be encoded in thrift: %d", n)
	}
	if n > math.MaxInt32 {
		return fmt.Errorf("length is too large to be encoded in thrift: %d", n)
	}
	return w.WriteInt32(int32(n))
}

func (w *BinaryWriter) WriteMessage(m Message) error {
	if w.p.NonStrict {
		if err := w.WriteString(m.Name); err != nil {
			return err
		}
		if err := w.writeByte(byte(m.Type)); err != nil {
			return err
		}
	} else {
		w.b[0] = 1 << 7
		w.b[1] = 0
		w.b[2] = 0
		w.b[3] = byte(m.Type) & 0x7
		binary.BigEndian.PutUint32(w.b[4:], uint32(len(m.Name)))

		if err := w.write(w.b[:8]); err != nil {
			return err
		}
		if err := w.writeString(m.Name); err != nil {
			return err
		}
	}
	return w.WriteInt32(m.SeqID)
}

func (w *BinaryWriter) WriteField(f Field) error {
	if err := w.writeByte(byte(f.Type)); err != nil {
		return err
	}
	return w.WriteInt16(f.ID)
}

func (w *BinaryWriter) WriteList(l List) error {
	if err := w.writeByte(byte(l.Type)); err != nil {
		return err
	}
	return w.WriteInt32(l.Size)
}

func (w *BinaryWriter) WriteMap(m Map) error {
	if err := w.writeByte(byte(m.Key)); err != nil {
		return err
	}
	if err := w.writeByte(byte(m.Value)); err != nil {
		return err
	}
	return w.WriteInt32(m.Size)
}

func (w *BinaryWriter) write(b []byte) error {
	_, err := w.w.Write(b)
	return err
}

func (w *BinaryWriter) writeString(s string) error {
	_, err := w.w.WriteString(s)
	return err
}

func (w *BinaryWriter) writeByte(b byte) error {
	return w.w.WriteByte(b)
}

func dontExpectEOF(err error) error {
	switch err {
	case nil:
		return nil
	case io.EOF:
		return io.ErrUnexpectedEOF
	default:
		return err
	}
}

func copyBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

var (
	_ Protocol = (*BinaryProtocol)(nil)
	_ Reader   = (*BinaryReader)(nil)
	_ Writer   = (*BinaryWriter)(nil)
)
