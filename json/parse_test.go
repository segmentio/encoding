package json

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func TestParseString(t *testing.T) {
	tests := []struct {
		in  string
		out string
		ext string
		err string
	}{
		{`""`, `""`, ``, ``},
		{`"1234567890"`, `"1234567890"`, ``, ``},
		{`"Hello World!"`, `"Hello World!"`, ``, ``},
		{`"Hello\"World!"`, `"Hello\"World!"`, ``, ``},
		{`"\\"`, `"\\"`, ``, ``},
		{`"\u0061\u0062\u0063"`, `"\u0061\u0062\u0063"`, ``, ``},
		{`"\u0"`, ``, ``, `json: unicode code point must have at least 4 characters: 0"`},
	}

	d := decoder{}
	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			out, ext, _, err := d.parseString([]byte(test.in))

			if test.err == "" {
				if err != nil {
					t.Errorf("%s => %s", test.in, err)
					return
				}
			} else {
				if s := err.Error(); s != test.err {
					t.Error("invalid error")
					t.Logf("expected: %s", test.err)
					t.Logf("found:    %s", s)
				}
			}

			if s := string(out); s != test.out {
				t.Error("invalid output")
				t.Logf("expected: %s", test.out)
				t.Logf("found:    %s", s)
			}

			if s := string(ext); s != test.ext {
				t.Error("invalid extra bytes")
				t.Logf("expected: %s", test.ext)
				t.Logf("found:    %s", s)
			}
		})
	}
}

func TestParseStringUnquote(t *testing.T) {
	tests := []struct {
		in  string
		out string
		ext string
	}{
		{`""`, ``, ``},
		{`"1234567890"`, `1234567890`, ``},
		{`"Hello World!"`, `Hello World!`, ``},
		{`"Hello\"World!"`, `Hello"World!`, ``},
		{`"\\"`, `\`, ``},
		{`"\u0061\u0062\u0063"`, `abc`, ``},
	}

	d := decoder{}
	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			out, ext, _, err := d.parseStringUnquote([]byte(test.in), nil)
			if err != nil {
				t.Errorf("%s => %s", test.in, err)
				return
			}

			if s := string(out); s != test.out {
				t.Error("invalid output")
				t.Logf("expected: %s", test.out)
				t.Logf("found:    %s", s)
			}

			if s := string(ext); s != test.ext {
				t.Error("invalid extra bytes")
				t.Logf("expected: %s", test.ext)
				t.Logf("found:    %s", s)
			}
		})
	}
}

func TestAppendToLower(t *testing.T) {
	tests := []string{
		"",
		"A",
		"a",
		"__segment_internal",
		"someFieldWithALongBName",
		"Hello World!",
		"Hello\"World!",
		"Hello\\World!",
		"Hello\nWorld!",
		"Hello\rWorld!",
		"Hello\tWorld!",
		"Hello\bWorld!",
		"Hello\fWorld!",
		"你好",
		"<",
		">",
		"&",
		"\u001944",
		"\u00c2e>",
		"\u00c2V?",
		"\u000e=8",
		"\u001944\u00c2e>\u00c2V?\u000e=8",
		"ir\u001bQJ\u007f\u0007y\u0015)",
	}

	for _, test := range tests {
		s1 := strings.ToLower(test)
		s2 := string(appendToLower(nil, []byte(test)))

		if s1 != s2 {
			t.Error("lowercase values mismatch")
			t.Log("expected:", s1)
			t.Log("found:   ", s2)
		}
	}
}

func BenchmarkParseString(b *testing.B) {
	s := []byte(`"__segment_internal"`)

	d := decoder{}
	for range b.N {
		d.parseString(s)
	}
}

func BenchmarkToLower(b *testing.B) {
	s := []byte("someFieldWithALongName")

	for range b.N {
		bytes.ToLower(s)
	}
}

func BenchmarkAppendToLower(b *testing.B) {
	a := []byte(nil)
	s := []byte("someFieldWithALongName")

	for range b.N {
		a = appendToLower(a[:0], s)
	}
}

var (
	benchmarkHasPrefixString = []byte("some random string")
	benchmarkHasPrefixResult = false
)

func BenchmarkHasPrefix(b *testing.B) {
	for range b.N {
		benchmarkHasPrefixResult = hasPrefix(benchmarkHasPrefixString, "null")
	}
}

func BenchmarkHasNullPrefix(b *testing.B) {
	for range b.N {
		benchmarkHasPrefixResult = hasNullPrefix(benchmarkHasPrefixString)
	}
}

func BenchmarkHasTruePrefix(b *testing.B) {
	for range b.N {
		benchmarkHasPrefixResult = hasTruePrefix(benchmarkHasPrefixString)
	}
}

func BenchmarkHasFalsePrefix(b *testing.B) {
	for range b.N {
		benchmarkHasPrefixResult = hasFalsePrefix(benchmarkHasPrefixString)
	}
}

func BenchmarkParseStringEscapeNone(b *testing.B) {
	j := []byte(`"` + strings.Repeat(`a`, 1000) + `"`)
	var s string
	b.SetBytes(int64(len(j)))

	for range b.N {
		if err := Unmarshal(j, &s); err != nil {
			b.Fatal(err)
		}
		s = ""
	}
}

func BenchmarkParseStringEscapeOne(b *testing.B) {
	j := []byte(`"` + strings.Repeat(`a`, 998) + `\n"`)
	var s string
	b.SetBytes(int64(len(j)))

	for range b.N {
		if err := Unmarshal(j, &s); err != nil {
			b.Fatal(err)
		}
		s = ""
	}
}

func BenchmarkParseStringEscapeAll(b *testing.B) {
	j := []byte(`"` + strings.Repeat(`\`, 1000) + `"`)
	var s string
	b.SetBytes(int64(len(j)))

	for range b.N {
		if err := Unmarshal(j, &s); err != nil {
			b.Fatal(err)
		}
		s = ""
	}
}

func TestParseIntOverflow(t *testing.T) {
	inputs := []string{
		"9223372036854775806",
		"9223372036854775807",
		"9223372036854775808",
		"9223372036854775809",
		"9223372036854775810",
		"10000000000000000000",
		"18446744073709551616",
		"21000000000000000000",
		"92233720368547758070",
		"-9223372036854775807",
		"-9223372036854775808",
		"-9223372036854775809",
	}

	d := decoder{}
	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			want, wantErr := strconv.ParseInt(in, 10, 64)
			got, _, err := d.parseInt([]byte(in), int64Type)

			if wantErr != nil {
				if err == nil {
					t.Errorf("expected an overflow error, got %d", got)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != want {
				t.Errorf("got %d, want %d", got, want)
			}
		})
	}
}

func TestParseUintOverflow(t *testing.T) {
	inputs := []string{
		"18446744073709551614",
		"18446744073709551615",
		"18446744073709551616",
		"18446744073709551617",
		"18446744073709551620",
		"20000000000000000000",
		"21000000000000000000",
		"99999999999999999999",
		"184467440737095516150",
	}

	d := decoder{}
	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			want, wantErr := strconv.ParseUint(in, 10, 64)
			got, _, err := d.parseUint([]byte(in), uint64Type)

			if wantErr != nil {
				if err == nil {
					t.Errorf("expected an overflow error, got %d", got)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != want {
				t.Errorf("got %d, want %d", got, want)
			}
		})
	}
}

func TestUnmarshalIntegerOverflow(t *testing.T) {
	inputs := []string{
		`{"u":18446744073709551616}`,
		`{"u":21000000000000000000}`,
		`{"i":9223372036854775808}`,
		`{"i":21000000000000000000}`,
	}

	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			var v struct {
				U uint64 `json:"u"`
				I int64  `json:"i"`
			}
			if err := Unmarshal([]byte(in), &v); err == nil {
				t.Errorf("expected an overflow error, got %+v", v)
			}
		})
	}
}
