package proto

import (
	"math"
	"testing"
)

func TestAppendVarint(t *testing.T) {
	m := AppendVarint(nil, 1, 42)

	f, w, v, r, err := Parse(m)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 0 {
		t.Fatal("unexpected trailing bytes:", r)
	}
	if w != Varint {
		t.Fatal("unexpected wire type:", t)
	}
	if f != 1 {
		t.Fatal("unexpected field number:", f)
	}
	if u := v.Varint(); u != 42 {
		t.Fatal("value mismatch, want 42 but got", u)
	}
}

func TestAppendVarlen(t *testing.T) {
	m := AppendVarlen(nil, 1, []byte("Hello World!"))

	f, w, v, r, err := Parse(m)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 0 {
		t.Fatal("unexpected trailing bytes:", r)
	}
	if w != Varlen {
		t.Fatal("unexpected wire type:", t)
	}
	if f != 1 {
		t.Fatal("unexpected field number:", f)
	}
	if string(v) != "Hello World!" {
		t.Fatalf("value mismatch, want \"Hello World!\" but got %q", v)
	}
}

func TestAppendFixed32(t *testing.T) {
	m := AppendFixed32(nil, 1, 42)

	f, w, v, r, err := Parse(m)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 0 {
		t.Fatal("unexpected trailing bytes:", r)
	}
	if w != Fixed32 {
		t.Fatal("unexpected wire type:", t)
	}
	if f != 1 {
		t.Fatal("unexpected field number:", f)
	}
	if u := v.Fixed32(); u != 42 {
		t.Fatal("value mismatch, want 42 but got", u)
	}
}

func TestAppendFixed64(t *testing.T) {
	m := AppendFixed64(nil, 1, 42)

	f, w, v, r, err := Parse(m)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 0 {
		t.Fatal("unexpected trailing bytes:", r)
	}
	if w != Fixed64 {
		t.Fatal("unexpected wire type:", t)
	}
	if f != 1 {
		t.Fatal("unexpected field number:", f)
	}
	if u := v.Fixed64(); u != 42 {
		t.Fatal("value mismatch, want 42 but got", u)
	}
}

func TestDecodeFromAppend(t *testing.T) {
	m := RawMessage(nil)
	m = AppendVarint(m, 1, EncodeZigZag(-1))
	m = AppendVarlen(m, 2, []byte("Hello World!"))
	m = AppendFixed32(m, 3, math.Float32bits(42.0))
	m = AppendFixed64(m, 4, math.Float64bits(1234.0))

	type M struct {
		I   int
		S   string
		F32 float32
		F64 float64
	}

	x := M{}

	if err := Unmarshal(m, &x); err != nil {
		t.Fatal(err)
	}
	if x.I != -1 {
		t.Errorf("x.I=%d", x.I)
	}
	if x.S != "Hello World!" {
		t.Errorf("x.S=%q", x.S)
	}
	if x.F32 != 42 {
		t.Errorf("x.F32=%g", x.F32)
	}
	if x.F64 != 1234 {
		t.Errorf("x.F64=%g", x.F64)
	}
}
