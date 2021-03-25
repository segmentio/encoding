package iso8601

import (
	"fmt"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	for _, input := range []string{
		// Fast path
		"1987-12-16T23:45:12Z",
		"2006-01-02T15:04:05Z",
		"2000-02-29T23:59:59Z", // leap year
		"2020-02-29T23:59:59Z", // leap year
		"0000-01-01T00:00:00Z",
		"9999-12-31T23:59:59Z",

		// Slow path
		"2006-01-02T15:04:05+00:00",
		"2006-01-02T15:04:05.123Z",
		"2006-01-02T15:04:05.123+00:00",
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

	// Check ~4M YYYY-MM-DD dates.
	for year := 0; year <= 9999; year++ {
		for month := 0; month <= 13; month++ {
			for day := 0; day <= 32; day++ {
				input := fmt.Sprintf("%04d-%02d-%02dT00:00:00Z", year, month, day)
				expect, expectErr := time.Parse(time.RFC3339Nano, input)
				actual, actualErr := Parse(input)
				if (expectErr != nil) != (actualErr != nil) {
					t.Errorf("unexpected error for %v: %v vs. %v expected", input, actualErr, expectErr)
				} else if !actual.Equal(expect) {
					t.Errorf("unexpected time for %v: %v vs. %v expected", input, actual, expect)
				}
			}
		}
	}

	// Check all ~1M HH:MM:SS times.
	for hour := 0; hour < 100; hour++ {
		for minute := 0; minute < 100; minute++ {
			for second := 0; second < 100; second++ {
				input := fmt.Sprintf("2000-01-01T%02d-%02d-%02dZ", hour, minute, second)
				expect, expectErr := time.Parse(time.RFC3339Nano, input)
				actual, actualErr := Parse(input)
				if (expectErr != nil) != (actualErr != nil) {
					t.Errorf("unexpected error for %v: %v vs. %v expected", input, actualErr, expectErr)
				} else if !actual.Equal(expect) {
					t.Errorf("unexpected time for %v: %v vs. %v expected", input, actual, expect)
				}
			}
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, input := range []string{
		"XXXXXXXXXXXXXXXXXXXX",
		"00000000000000000000",
		"1900-02-29T00:00:00Z", // 28 days in month (not a leap year)
		"2021-02-29T00:00:00Z", // 28 days in month (not a leap year)
		"2021-02-30T00:00:00Z", // 28 days in month
		"2021-02-31T00:00:00Z", // 28 days in month
		"2021-04-31T00:00:00Z", // 30 days in month
		"2021-06-31T00:00:00Z", // 30 days in month
		"2021-09-31T00:00:00Z", // 30 days in month
		"2021-11-31T00:00:00Z", // 30 days in month
		"2000-13-01T00:00:00Z", // invalid month
		"2000-00-01T00:00:00Z", // invalid month
		"2000-12-32T00:00:00Z", // invalid day
		"2000-12-00T00:00:00Z", // invalid day
		"2000-12-31T24:00:00Z", // invalid hour
		"2000-12-31T23:60:00Z", // invalid minute
		"2000-12-31T23:59:60Z", // invalid second
		"1999-01-01 23:45:00Z", // missing T separator
		"1999 01 01T23:45:00Z", // missing date separators
		"1999-01-01T23 45 00Z", // missing time separators
		"1999-01-01T23:45:00 ", // missing timezone
	} {
		t.Run(input, func(t *testing.T) {
			ts, err := time.Parse(time.RFC3339Nano, input)
			if err == nil {
				t.Fatalf("expected time.Parse('%s') error, got %v", input, ts)
			}
			ts, actualErr := Parse(input)
			if actualErr == nil {
				t.Fatalf("expected Parse('%s') error %v, got %v", input, err, ts)
			} else if err.Error() != actualErr.Error() {
				t.Fatalf("expected Parse('%s') error %v, got %v", input, err, actualErr)
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
