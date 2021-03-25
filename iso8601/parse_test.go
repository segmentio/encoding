package iso8601

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	for _, input := range []string{
		// Fast path
		"1987-12-16T23:45:12Z",
		"2006-01-02T15:04:05Z",
		"2020-02-29T23:59:59Z", // leap year
		"0000-01-01T00:00:00Z",
		"9999-12-31T23:59:59Z",

		// Slow path
		"2006-01-02T15:04:05+00:00",
		"2006-01-02T15:04:05.123Z",
		"2006-01-02T15:04:05.123+00:00",
		"2006-01-02T15:04:05.123456Z",
		"2006-01-02T15:04:05.123456789Z",

		// FIXME:
		// "2021-02-29T00:00:00Z", // not a leap year
		// "2021-04-31T00:00:00Z", // 30 days in month
		// "2021-06-31T00:00:00Z", // 30 days in month
		// "2021-09-31T00:00:00Z", // 30 days in month
		// "2021-11-31T00:00:00Z", // 30 days in month
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
