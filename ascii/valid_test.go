package ascii

import (
	"strings"
	"testing"
)

var testStrings = [...]string{
	"",
	"hello",
	"Hello World!",
	"Hello\"World!",
	"Hello\\World!",
	"Hello\nWorld!",
	"Hello\rWorld!",
	"Hello\tWorld!",
	"Hello\bWorld!",
	"Hello\fWorld!",
	"你好",
	"\x80",
	"\xFF",
	"some kind of long string with only ascii characters.",
	"some kind of long string with a non-ascii character at the end.\xff",
	strings.Repeat("1234567890", 1000),
}

func TestValid(t *testing.T) {
	for _, test := range testStrings {
		t.Run(limit(test), func(t *testing.T) {
			expect := true

			for i := range test {
				if test[i] > 0x7f {
					expect = false
					break
				}
			}

			if valid := Valid([]byte(test)); expect != valid {
				t.Errorf("expected %t but got %t", expect, valid)
			}
		})
	}
}

func BenchmarkValid(b *testing.B) {
	for _, test := range testStrings {
		b.Run(limit(test), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ValidString(test)
			}
			b.SetBytes(int64(len(test)))
		})
	}
}

func limit(s string) string {
	if len(s) > 17 {
		return s[:17] + "..."
	}
	return s
}
