package json

import (
	stdjson "encoding/json"
	"strings"
	"testing"
)

func BenchmarkEscapeIndex4KB(b *testing.B) {
	benchmarkEscapeIndex(b, strings.Repeat("!foobar!", 512), false)
}

func BenchmarkEscapeIndex4KBEscapeHTML(b *testing.B) {
	benchmarkEscapeIndex(b, strings.Repeat("!foobar!", 512), true)
}

func BenchmarkEscapeIndex1(b *testing.B) {
	benchmarkEscapeIndex(b, "1", false)
}

func BenchmarkEscapeIndex1EscapeHTML(b *testing.B) {
	benchmarkEscapeIndex(b, "1", true)
}

func BenchmarkEscapeIndex7(b *testing.B) {
	benchmarkEscapeIndex(b, "1234567", false)
}

func BenchmarkEscapeIndex7EscapeHTML(b *testing.B) {
	benchmarkEscapeIndex(b, "1234567", true)
}

func benchmarkEscapeIndex(b *testing.B, s string, escapeHTML bool) {
	b.ResetTimer()
	for range b.N {
		escapeIndex(s, escapeHTML)
	}
	b.SetBytes(int64(len(s)))
}

// reference: index of the first byte that needs escaping, or -1
func refEscapeIndex(s string, escapeHTML bool) int {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c > 0x7f || c == '"' || c == '\\' ||
			(escapeHTML && (c == '<' || c == '>' || c == '&')) {
			return i
		}
	}
	return -1
}

func TestEscapeIndexBigEndian(t *testing.T) {
	inputs := []string{
		`C:\Users\admin`,                 // backslash at byte 2
		"key\tvalue",                     // tab at byte 3
		"\nfoo_bar_",                     // newline at byte 0
		`ab"cdefg`,                       // quote at byte 2
		"line1\nline2",                   // newline at byte 5
		`say "hello"`,                    // quote at byte 4
		"1234567\nabc",                   // last byte of first chunk
		"12345678\nabc",                  // first byte of second chunk
		"-----BEGIN CERTIFICATE-----\nABCD\n-----END CERTIFICATE-----\n",
		"<tag>&more",                       // HTML characters
		"plain ascii with no escapes here", // no escaping needed
		strings.Repeat("a", 33) + `"`,      // quote far into the string
	}

	for _, s := range inputs {
		for _, html := range []bool{false, true} {
			if got, want := escapeIndex(s, html), refEscapeIndex(s, html); got != want {
				t.Errorf("escapeIndex(%q, %v) = %d, want %d", s, html, got, want)
			}
		}

		b, err := Marshal(struct {
			V string `json:"v"`
		}{s})
		if err != nil {
			t.Errorf("Marshal(%q): %v", s, err)
			continue
		}
		var out struct {
			V string `json:"v"`
		}
		if err := stdjson.Unmarshal(b, &out); err != nil {
			t.Errorf("%q produced invalid JSON %s: %v", s, b, err)
			continue
		}
		if out.V != s {
			t.Errorf("%q round-trips to %q", s, out.V)
		}
	}
}
