package proto

import (
	"reflect"
	"testing"
)

func TestRewrite(t *testing.T) {
	type message struct {
		A int
		B float32
		C float64
		D string
	}

	tests := []struct {
		scenario string
		in       message
		out      message
		rw       Rewriter
	}{
		{
			scenario: "identity",
			in:       message{A: 42},
			out:      message{A: 42},
			rw:       RewriteFields(nil),
		},

		{
			scenario: "rewrite field 1",
			in:       message{A: 21},
			out:      message{A: 42},
			rw: RewriteFields{
				1: FieldNumber(1).Int(42),
			},
		},

		{
			scenario: "rewrite field 2",
			in:       message{A: 21, B: 0.125},
			out:      message{A: 21, B: -1},
			rw: RewriteFields{
				2: FieldNumber(2).Float32(-1),
			},
		},

		{
			scenario: "rewrite field 3",
			in:       message{A: 21, B: 0.125, C: 0.0},
			out:      message{A: 21, B: 0.125, C: 1.0},
			rw: RewriteFields{
				3: FieldNumber(3).Float64(+1),
			},
		},

		{
			scenario: "rewrite field 4",
			in:       message{A: 21, B: 0.125, C: 1.0, D: "A"},
			out:      message{A: 21, B: 0.125, C: 1.0, D: "Hello World!"},
			rw: RewriteFields{
				4: FieldNumber(4).String("Hello World!"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			b, err := Marshal(test.in)
			if err != nil {
				t.Fatal(err)
			}

			b, err = test.rw.Rewrite(nil, b)
			if err != nil {
				t.Fatal(err)
			}

			m := message{}
			if err := Unmarshal(b, &m); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(m, test.out) {
				t.Errorf("messages mismatch:\nwant: %+v\ngot:  %+v", test.out, m)
			}
		})
	}
}

func TestParseRewriteTemplate(t *testing.T) {
	type submessage struct {
		Question string `protobuf:"bytes,1,opt,name=question,proto3"`
		Answer   string `protobuf:"bytes,2,opt,name=answer,proto3"`
	}

	type message struct {
		Field1 bool `protobuf:"varint,1,opt,name=field_1,proto3"`

		Field2 int   `protobuf:"varint,2,opt,name=field_2,proto3"`
		Field3 int32 `protobuf:"varint,3,opt,name=field_3,proto3"`
		Field4 int64 `protobuf:"varint,4,opt,name=field_4,proto3"`

		Field5 uint   `protobuf:"varint,5,opt,name=field_5,proto3"`
		Field6 uint32 `protobuf:"varint,6,opt,name=field_6,proto3"`
		Field7 uint64 `protobuf:"varint,7,opt,name=field_7,proto3"`

		Field8 int32 `protobuf:"zigzag32,8,opt,name=field_8,proto3"`
		Field9 int64 `protobuf:"zigzag64,9,opt,name=field_9,proto3"`

		Field10 float32 `protobuf:"fixed32,10,opt,name=field_10,proto3"`
		Field11 float64 `protobuf:"fixed64,11,opt,name=field_11,proto3"`

		Field12 string `protobuf:"bytes,12,opt,name=field_12,proto3"`
		Field13 []byte `protobuf:"bytes,13,opt,name=field_13,proto3"`

		Subfield    *submessage  `protobuf:"bytes,99,opt,name=subfield,proto3"`
		Submessages []submessage `protobuf:"bytes,100,rep,name=submessages,proto3"`
	}

	original := &message{
		Field1: false,

		Field2: -1,
		Field3: -2,
		Field4: -3,

		Field5: 1,
		Field6: 2,
		Field7: 3,

		Field8: -10,
		Field9: -11,

		Field10: 1.0,
		Field11: 2.0,

		Field12: "field 12",
		Field13: nil,

		Subfield: &submessage{
			Question: "How are you?",
			Answer:   "Good!",
		},

		Submessages: []submessage{
			{Question: "Q1?", Answer: "A1"},
			{Question: "Q2?", Answer: "A2"},
			{Question: "Q3?", Answer: "A3"},
		},
	}

	expected := &message{
		Field1: true,

		Field2: 2,
		Field3: 3,
		Field4: 4,

		Field5: 10,
		Field6: 11,
		Field7: 12,

		Field8: -21,
		Field9: -42,

		Field10: 0.25,
		Field11: 0.5,

		Field12: "Hello!",
		Field13: []byte("World!"),

		Subfield: &submessage{
			Question: "How are you?",
			Answer:   "Good!",
		},

		Submessages: []submessage{
			{Question: "Q1?", Answer: "A1"},
			{Question: "Q2?", Answer: "A2"},
			{Question: "Q3?", Answer: "Hello World!"},
		},
	}

	b1, err := Marshal(*original)
	if err != nil {
		t.Fatal(err)
	}

	rw, err := ParseRewriteTemplate(TypeOf(original), []byte(`{
  "field_1": true,

  "field_2": 2,
  "field_3": 3,
  "field_4": 4,

  "field_5": 10,
  "field_6": 11,
  "field_7": 12,

  "field_8": -21,
  "field_9": -42,

  "field_10": 0.25,
  "field_11": 0.5,

  "field_12": "Hello!",
  "field_13": "V29ybGQh",

  "subfield": {
    "question": "How are you?"
  },

  "submessages": [
    {
      "question": "Q1?",
      "answer": "A1"
    },
    {
      "question": "Q2?",
      "answer": "A2"
    },
    {
      "question": "Q3?",
      "answer": "Hello World!"
    }
  ]
}`))
	if err != nil {
		t.Fatal(err)
	}

	b2, err := rw.Rewrite(nil, b1)
	if err != nil {
		t.Fatal(err)
	}

	found := &message{}
	if err := Unmarshal(b2, &found); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, found) {
		t.Error("messages mismatch after rewrite")
		t.Logf("want:\n%+v", expected)
		t.Logf("got:\n%+v", found)
	}
}

func BenchmarkRewrite(b *testing.B) {
	type message struct {
		A int
		B float32
		C float64
		D string
	}

	in := message{A: 21, B: 0.125, D: "A"}
	rw := RewriteFields{
		1: FieldNumber(1).Int(42),
		2: FieldNumber(2).Float32(-1),
		3: FieldNumber(3).Float64(+1),
		4: FieldNumber(4).String("Hello World!"),
	}

	p, err := Marshal(in)
	if err != nil {
		b.Fatal(err)
	}

	out := make([]byte, 0, 2*cap(p))

	for i := 0; i < b.N; i++ {
		out, err = rw.Rewrite(out[:0], p)
	}
}
