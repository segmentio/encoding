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
