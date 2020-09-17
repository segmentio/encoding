package json

import (
	"strings"
	"testing"
)

func BenchmarkSimpleString4KB(b *testing.B) {
	benchmarkSimpleString(b, strings.Repeat("!foobar!", 512), false)
}

func BenchmarkSimpleString4KBEscapeHTML(b *testing.B) {
	benchmarkSimpleString(b, strings.Repeat("!foobar!", 512), true)
}

func BenchmarkSimpleString1(b *testing.B) {
	benchmarkSimpleString(b, "1", false)
}

func BenchmarkSimpleString1EscapeHTML(b *testing.B) {
	benchmarkSimpleString(b, "1", true)
}

func BenchmarkSimpleString7(b *testing.B) {
	benchmarkSimpleString(b, "1234567", false)
}

func BenchmarkSimpleString7EscapeHTML(b *testing.B) {
	benchmarkSimpleString(b, "1234567", true)
}

func benchmarkSimpleString(b *testing.B, s string, escapeHTML bool) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		simpleString(s, escapeHTML)
	}
	b.SetBytes(int64(len(s)))
}
