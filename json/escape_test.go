package json

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"math/bits"
	"strings"
	"testing"
)

func shouldEscape(b byte, escapeHTML bool) bool {
	if b < 0x20 || b > 0x7f || b == '"' || b == '\\' {
		return true
	}
	if escapeHTML && (b == '<' || b == '>' || b == '&') {
		return true
	}
	return false
}

func expectedEscapeIndex(s string, escapeHTML bool) int {
	for i := 0; i < len(s); i++ {
		if shouldEscape(s[i], escapeHTML) {
			return i
		}
	}
	return -1
}

func TestEscapeIndexAllBytes(t *testing.T) {
	for escapeHTML := range 2 {
		htmlMode := escapeHTML == 1
		for b := range 256 {
			by := byte(b)
			expected := shouldEscape(by, htmlMode)

			// Test single-char string
			s1 := string([]byte{by})
			got1 := escapeIndex(s1, htmlMode)
			if expected && got1 != 0 {
				t.Errorf("byte 0x%02x (html=%v) single char: expected 0, got %d", by, htmlMode, got1)
			} else if !expected && got1 != -1 {
				t.Errorf("byte 0x%02x (html=%v) single char: expected -1, got %d", by, htmlMode, got1)
			}

			// Test 8-byte chunk with the byte at index 3
			buf := []byte("aaaa.aaa")
			buf[3] = by
			s8 := string(buf)
			got8 := escapeIndex(s8, htmlMode)
			if expected && got8 != 3 {
				t.Errorf("byte 0x%02x (html=%v) 8-byte chunk index 3: expected 3, got %d", by, htmlMode, got8)
			} else if !expected && got8 != -1 {
				t.Errorf("byte 0x%02x (html=%v) 8-byte chunk index 3: expected -1, got %d", by, htmlMode, got8)
			}
		}
	}
}

func TestEscapeIndexChunkPositions(t *testing.T) {
	specials := []byte{'\\', '"', '\n', '\r', '\t', '\x00', '\x1f', 0x80, 0xff}

	for _, sp := range specials {
		// Test escape character at every position in chunk 0 (0..7)
		for pos := 0; pos < 8; pos++ {
			b := []byte("01234567")
			b[pos] = sp
			s := string(b)
			got := escapeIndex(s, false)
			if got != pos {
				t.Errorf("special 0x%02x in chunk 0 at pos %d: expected %d, got %d", sp, pos, pos, got)
			}
		}

		// Test escape character at every position in chunk 1 (8..15)
		for pos := 8; pos < 16; pos++ {
			b := []byte("0123456789abcdef")
			b[pos] = sp
			s := string(b)
			got := escapeIndex(s, false)
			if got != pos {
				t.Errorf("special 0x%02x in chunk 1 at pos %d: expected %d, got %d", sp, pos, pos, got)
			}
		}

		// Test escape character at every position in chunk 2 (16..23)
		for pos := 16; pos < 24; pos++ {
			b := []byte("0123456789abcdef01234567")
			b[pos] = sp
			s := string(b)
			got := escapeIndex(s, false)
			if got != pos {
				t.Errorf("special 0x%02x in chunk 2 at pos %d: expected %d, got %d", sp, pos, pos, got)
			}
		}

		// Test escape character in tail (e.g. string length 11, pos 8, 9, 10)
		for pos := 8; pos < 11; pos++ {
			b := []byte("0123456789a")
			b[pos] = sp
			s := string(b)
			got := escapeIndex(s, false)
			if got != pos {
				t.Errorf("special 0x%02x in tail at pos %d: expected %d, got %d", sp, pos, pos, got)
			}
		}
	}
}

func TestEscapeIndexEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		escapeHTML bool
		expected   int
	}{
		{name: "empty", input: "", escapeHTML: false, expected: -1},
		{name: "clean 7 bytes", input: "1234567", escapeHTML: false, expected: -1},
		{name: "clean 8 bytes", input: "12345678", escapeHTML: false, expected: -1},
		{name: "clean 16 bytes", input: "12345678abcdefgh", escapeHTML: false, expected: -1},
		{name: "all spaces 8 bytes", input: "        ", escapeHTML: false, expected: -1},
		{name: "all spaces 16 bytes", input: "                ", escapeHTML: false, expected: -1},
		{name: "spaces then newline at index 7", input: "       \n", escapeHTML: false, expected: 7},
		{name: "spaces then newline at index 15", input: "               \n", escapeHTML: false, expected: 15},
		{name: "spaces then quote at index 7", input: "       \"", escapeHTML: false, expected: 7},
		{name: "spaces then backslash at index 7", input: "       \\", escapeHTML: false, expected: 7},
		{name: "newline then spaces", input: "\n       ", escapeHTML: false, expected: 0},
		{name: "multiple escapes in chunk", input: "a\"b\\cdef", escapeHTML: false, expected: 1},
		{name: "multiple escapes across chunks", input: "1234567\n12\"45678", escapeHTML: false, expected: 7},
		{name: "html lt enabled", input: "abcdefg<", escapeHTML: true, expected: 7},
		{name: "html lt disabled", input: "abcdefg<", escapeHTML: false, expected: -1},
		{name: "html gt enabled", input: "12345678>abcdefg", escapeHTML: true, expected: 8},
		{name: "html gt disabled", input: "12345678>abcdefg", escapeHTML: false, expected: -1},
		{name: "html amp enabled", input: "12&45678", escapeHTML: true, expected: 2},
		{name: "html amp disabled", input: "12&45678", escapeHTML: false, expected: -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := escapeIndex(tc.input, tc.escapeHTML)
			if got != tc.expected {
				t.Errorf("escapeIndex(%q, %v): expected %d, got %d", tc.input, tc.escapeHTML, tc.expected, got)
			}
		})
	}
}

// simulateBigEndianChunkLayout converts a string so that each 8-byte chunk
// has its bytes laid out in simulated big-endian word order when loaded
// by a uint64 cast on a little-endian host.
//
// On a big-endian system, the byte at chunk offset k is stored at the most
// significant byte (bits 63-8k..56-8k). On a little-endian host, uint64
// loads place byte j of memory into bits 8j..8j+7.
// Therefore, reversing the bytes within an 8-byte chunk causes a little-endian
// uint64 load to produce the exact 64-bit integer that a big-endian host would
// load from the original string.
func simulateBigEndianWordLayout(s string) string {
	b := []byte(s)
	chunks := len(b) / 8
	for i := 0; i < chunks; i++ {
		start := i * 8
		for j := 0; j < 4; j++ {
			b[start+j], b[start+7-j] = b[start+7-j], b[start+j]
		}
	}
	return string(b)
}

func TestEscapeIndexSimulatedEndianLayout(t *testing.T) {
	origEndianness := endianness
	defer func() {
		endianness = origEndianness
	}()

	testCases := []struct {
		name       string
		original   string
		escapeHTML bool
	}{
		{name: "clean single chunk", original: "abcdefgh", escapeHTML: false},
		{name: "clean two chunks", original: "abcdefghijklmnop", escapeHTML: false},
		{name: "clean with tail", original: "abcdefghijklmnopq", escapeHTML: false},
		{name: "issue 170 backslash in chunk 0", original: `C:\Users\admin`, escapeHTML: false},
		{name: "issue 170 tab", original: "key\tvalue", escapeHTML: false},
		{name: "issue 170 newline at start", original: "\nfoo_bar_", escapeHTML: false},
		{name: "issue 170 quote in chunk 0", original: `ab"cdefg`, escapeHTML: false},
		{name: "backslash at chunk 0 pos 0", original: `\bcdefgh`, escapeHTML: false},
		{name: "backslash at chunk 0 pos 1", original: `a\cdefgh`, escapeHTML: false},
		{name: "backslash at chunk 0 pos 2", original: `ab\defgh`, escapeHTML: false},
		{name: "backslash at chunk 0 pos 3", original: `abc\efgh`, escapeHTML: false},
		{name: "backslash at chunk 0 pos 4", original: `abcd\fgh`, escapeHTML: false},
		{name: "backslash at chunk 0 pos 5", original: `abcde\gh`, escapeHTML: false},
		{name: "backslash at chunk 0 pos 6", original: `abcdef\h`, escapeHTML: false},
		{name: "backslash at chunk 0 pos 7", original: `abcdefg\`, escapeHTML: false},
		{name: "backslash at chunk 1 pos 0", original: `abcdefgh\jklmnop`, escapeHTML: false},
		{name: "backslash at chunk 1 pos 3", original: `abcdefghijk\mnop`, escapeHTML: false},
		{name: "backslash at chunk 1 pos 7", original: `abcdefghijklmno\`, escapeHTML: false},
		{name: "spaces then newline", original: "       \n", escapeHTML: false},
		{name: "spaces then newline chunk 1", original: "        " + "       \n", escapeHTML: false},
		{name: "control char at start", original: "\x01bcdefgh", escapeHTML: false},
		{name: "html tags enabled", original: "abc<efgh", escapeHTML: true},
		{name: "html tags disabled", original: "abc<efgh", escapeHTML: false},
		{name: "html amp enabled", original: "abcdefg&", escapeHTML: true},
		{name: "html amp disabled", original: "abcdefg&", escapeHTML: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expected := expectedEscapeIndex(tc.original, tc.escapeHTML)

			// 1. Verify under native Little-Endian mode (endianness = 0)
			endianness = 0
			gotLE := escapeIndex(tc.original, tc.escapeHTML)
			if gotLE != expected {
				t.Fatalf("LE mode: escapeIndex(%q, %v): expected %d, got %d", tc.original, tc.escapeHTML, expected, gotLE)
			}

			// 2. Verify under simulated Big-Endian mode (endianness = 1)
			// We feed the simulated big-endian layout into escapeIndex.
			// The simulated string has its 8-byte words reversed so that on this LE host,
			// uint64 memory loads match the big-endian register layout.
			endianness = 1
			simulatedBE := simulateBigEndianWordLayout(tc.original)
			gotBE := escapeIndex(simulatedBE, tc.escapeHTML)
			if gotBE != expected {
				t.Fatalf("BE mode: escapeIndex(simulated %q, %v): expected %d, got %d", tc.original, tc.escapeHTML, expected, gotBE)
			}
		})
	}
}

func TestEscapeIndexWordBitPatterns(t *testing.T) {
	// Directly validate the bit arithmetic on 64-bit words for both byte orders.
	// For every byte position k in 0..7 and various special characters:
	specials := []byte{'\\', '"', '\n', '\r', '\t', '<', '>', '&', 0x1f, 0x80}

	for _, sp := range specials {
		for k := 0; k < 8; k++ {
			// Construct an 8-byte chunk with safe ASCII 'A' and the special byte at index k.
			var b [8]byte
			for i := range b {
				b[i] = 'A'
			}
			b[k] = sp

			// Little-endian uint64 value: byte k is at bits 8k..8k+7
			var nLE uint64
			for i := 0; i < 8; i++ {
				nLE |= uint64(b[i]) << (8 * i)
			}

			// Big-endian uint64 value: byte k is at bits (7-k)*8..(7-k)*8+7
			var nBE uint64
			for i := 0; i < 8; i++ {
				nBE |= uint64(b[i]) << (8 * (7 - i))
			}

			// Verify that reversing the BE word restores the LE word
			if bits.ReverseBytes64(nBE) != nLE {
				t.Fatalf("ReverseBytes64(nBE) != nLE for special 0x%02x at pos %d", sp, k)
			}

			// Verify that below/contains mask on reversed word yields exact index k
			mask := nLE | below(nLE, 0x20) | contains(nLE, '"') | contains(nLE, '\\')
			mask |= contains(nLE, '<') | contains(nLE, '>') | contains(nLE, '&')
			if (mask & msb) == 0 {
				t.Fatalf("expected mask MSB to be set for special 0x%02x at pos %d", sp, k)
			}
			idx := bits.TrailingZeros64(mask&msb) / 8
			if idx != k {
				t.Fatalf("expected index %d for special 0x%02x, got %d", k, sp, idx)
			}
		}
	}
}

func TestEscapeIndexIssue170Reproducer(t *testing.T) {
	cases := []string{
		`C:\Users\admin`, // backslash near the start
		"key\tvalue",     // tab
		"\nfoo_bar_",     // newline at the very start
		`ab"cdefg`,       // quote near the start
		`"quoted"`,
		`C:\Program Files\App\bin`,
		"multi\nline\tstring\"with\\escapes",
	}

	for _, in := range cases {
		type W struct {
			V string `json:"v"`
		}
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		enc.SetAppendNewline(false)
		if err := enc.Encode(W{V: in}); err != nil {
			t.Errorf("%q: encode error: %v", in, err)
			continue
		}

		var got W
		if err := stdjson.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Errorf("%q: got invalid JSON: %v\n  raw: %s", in, err, buf.Bytes())
			continue
		}
		if got.V != in {
			t.Errorf("%q: came back as %q", in, got.V)
		}
	}
}

func TestStringEncodingRoundTrip(t *testing.T) {
	// Exhaustively test string encoding against standard library encoding/json
	testStrings := []string{
		"",
		"a",
		"hello world",
		`"`,
		`\`,
		`""`,
		`\\`,
		`"hello"`,
		`c:\path\to\file`,
		"\x00\x01\x02\x03\x04\x05\x06\x07\x08\t\n\x0b\x0c\r\x0e\x0f",
		"\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f",
		"        \n",
		"       \n ",
		"      \n  ",
		"     \n   ",
		"    \n    ",
		"   \n     ",
		"  \n      ",
		" \n       ",
		"\n        ",
		strings.Repeat("a", 7) + "\"",
		strings.Repeat("a", 8) + "\"",
		strings.Repeat("a", 9) + "\"",
		strings.Repeat("a", 15) + "\\",
		strings.Repeat("a", 16) + "\\",
		strings.Repeat("a", 17) + "\\",
		"<script>alert('xss')&;</script>",
		"Unicode: \u0000 \u001f \u0020 \u007f \u0080 \u00ff \u4e16\u754c",
	}

	for i, s := range testStrings {
		t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			gotJSON, err := Marshal(s)
			if err != nil {
				t.Fatalf("Marshal(%q) error: %v", s, err)
			}

			wantJSON, err := stdjson.Marshal(s)
			if err != nil {
				t.Fatalf("stdjson.Marshal(%q) error: %v", s, err)
			}

			if !bytes.Equal(gotJSON, wantJSON) {
				t.Errorf("Marshal mismatch for %q:\n got:  %s\n want: %s", s, gotJSON, wantJSON)
			}

			var unmarshaled string
			if err := Unmarshal(gotJSON, &unmarshaled); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			if unmarshaled != s {
				t.Errorf("roundtrip mismatch: got %q, want %q", unmarshaled, s)
			}
		})
	}
}
