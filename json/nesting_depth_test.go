package json

import (
	"bytes"
	"strings"
	"testing"
)

// These tests cover the nesting-depth guard added to the decoder. Without it,
// deeply nested input recurses through parseValue/parseArray/parseObject until
// the goroutine stack is exhausted, which is a fatal (unrecoverable) crash in Go
// rather than a returned error. The standard library's encoding/json applies the
// same maxNestingDepth limit.

func nestedArray(depth int) []byte {
	b := make([]byte, 0, depth*2)
	b = append(b, bytes.Repeat([]byte("["), depth)...)
	b = append(b, bytes.Repeat([]byte("]"), depth)...)
	return b
}

func nestedObject(depth int) []byte {
	// {"a":{"a":{...:null...}}}
	var b bytes.Buffer
	for i := 0; i < depth; i++ {
		b.WriteString(`{"a":`)
	}
	b.WriteString("null")
	for i := 0; i < depth; i++ {
		b.WriteByte('}')
	}
	return b.Bytes()
}

func TestUnmarshalRejectsExcessiveArrayNesting(t *testing.T) {
	var v interface{}
	err := Unmarshal(nestedArray(maxNestingDepth+1), &v)
	if err == nil {
		t.Fatal("expected error for input nested beyond maxNestingDepth, got nil")
	}
	if !strings.Contains(err.Error(), "depth") {
		t.Fatalf("expected a max-depth error, got: %v", err)
	}
}

func TestUnmarshalRejectsExcessiveObjectNesting(t *testing.T) {
	var v interface{}
	err := Unmarshal(nestedObject(maxNestingDepth+1), &v)
	if err == nil {
		t.Fatal("expected error for object nested beyond maxNestingDepth, got nil")
	}
	if !strings.Contains(err.Error(), "depth") {
		t.Fatalf("expected a max-depth error, got: %v", err)
	}
}

// Nesting that reaches a deeply nested value inside an otherwise-skipped struct
// field must also be bounded (the value is consumed via parseValue).
func TestUnmarshalRejectsExcessiveNestingInSkippedField(t *testing.T) {
	payload := append([]byte(`{"unknown":`), nestedArray(maxNestingDepth+1)...)
	payload = append(payload, '}')
	var v struct {
		Known string `json:"known"`
	}
	if err := Unmarshal(payload, &v); err == nil {
		t.Fatal("expected error for deeply nested value in skipped field, got nil")
	}
}

func TestUnmarshalAcceptsNestingWithinLimit(t *testing.T) {
	var v interface{}
	if err := Unmarshal(nestedArray(maxNestingDepth-1), &v); err != nil {
		t.Fatalf("input within the depth limit should decode, got: %v", err)
	}
}

// A wide but shallow document (depth 2) must not be affected by the depth guard:
// the limit counts nesting level, not the number of sibling elements.
func TestUnmarshalAllowsWideShallowArray(t *testing.T) {
	var b bytes.Buffer
	b.WriteByte('[')
	for i := 0; i < 100000; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('0')
	}
	b.WriteByte(']')

	var v []int
	if err := Unmarshal(b.Bytes(), &v); err != nil {
		t.Fatalf("wide shallow array should decode, got: %v", err)
	}
	if len(v) != 100000 {
		t.Fatalf("expected 100000 elements, got %d", len(v))
	}
}
