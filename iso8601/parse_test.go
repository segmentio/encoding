package iso8601

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	for _, input := range []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.123Z",
		"2006-01-02T15:04:05.123456Z",
		"2006-01-02T15:04:05.123456789Z",
	} {
		t.Run(input, func(t *testing.T) {
			expect, err := time.Parse(time.RFC3339Nano, input)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := Parse(input)
			if err != nil {
				t.Error(err)
			} else if !actual.Equal(expect) {
				t.Errorf("unexpected time: %v vs expected %v", actual, expect)
			}
		})
	}
}

func BenchmarkParseSeconds(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse("2006-01-02T15:04:05Z")
	}
}

func BenchmarkParseMilliseconds(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse("2006-01-02T15:04:05.123Z")
	}
}

func BenchmarkParseMicroseconds(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse("2006-01-02T15:04:05.123456Z")
	}
}

func BenchmarkParseNanoseconds(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse("2006-01-02T15:04:05.123456789Z")
	}
}
