package json

import (
	stdjson "encoding/json"
	"reflect"
	"testing"
)

func TestUnmarshalNumberString(t *testing.T) {
	for _, input := range []string{
		`151`, `null`, `"151"`, `"-0"`, `"1.25"`, `"1e+1000"`, `"\u0031\u0035\u0031"`,
		`""`, `"01"`, `"+1"`, `"1."`, `"1e"`, `"1 2"`, `"1 "`, `" 1"`, `"null"`, `"NaN"`,
	} {
		t.Run(input, func(t *testing.T) {
			for _, flags := range []ParseFlags{0, DontCopyNumber, ZeroCopy} {
				want, got := Number("7"), Number("7")
				wantErr := stdjson.Unmarshal([]byte(input), &want)
				_, gotErr := Parse([]byte(input), &got, flags)
				if (wantErr == nil) != (gotErr == nil) {
					t.Fatalf("flags=%v: error=%v, standard library error=%v", flags, gotErr, wantErr)
				}
				if gotErr == nil && got != want {
					t.Fatalf("flags=%v: got %q, want %q", flags, got, want)
				}
			}
		})
	}
}

func TestUnmarshalNumberStringContainers(t *testing.T) {
	for _, test := range []struct {
		input string
		typ   reflect.Type
	}{
		{`{"id":"151","next":2}`, reflect.TypeOf(struct {
			ID   Number `json:"id"`
			Next int    `json:"next"`
		}{})},
		{`["151",2,null,"1.5"]`, reflect.TypeOf([]Number{})},
		{`{"id":"151"}`, reflect.TypeOf(map[string]Number{})},
		{`{"id":"151"}`, reflect.TypeOf(struct {
			ID Number `json:"id,string"`
		}{})},
	} {
		t.Run(test.input+test.typ.String(), func(t *testing.T) {
			want, got := reflect.New(test.typ).Interface(), reflect.New(test.typ).Interface()
			if err := stdjson.Unmarshal([]byte(test.input), want); err != nil {
				t.Fatal(err)
			}
			if err := Unmarshal([]byte(test.input), got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v, want %#v", got, want)
			}
		})
	}
}
