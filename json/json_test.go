package json

import (
	"bytes"
	"compress/gzip"
	"encoding"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	marshal    func([]byte, interface{}) ([]byte, error)
	unmarshal  func([]byte, interface{}) error
	escapeHTML bool
)

func TestMain(m *testing.M) {
	var pkg string
	flag.StringVar(&pkg, "package", ".", "The name of the package to test (encoding/json, or default to this package)")
	flag.BoolVar(&escapeHTML, "escapehtml", false, "Whether to enable HTML escaping or not")
	flag.Parse()

	switch pkg {
	case "encoding/json":
		buf := &buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(escapeHTML)

		marshal = func(b []byte, v interface{}) ([]byte, error) {
			buf.data = b
			err := enc.Encode(v)
			return buf.data, err
		}

		unmarshal = json.Unmarshal

	default:
		flags := AppendFlags(0)
		if escapeHTML {
			flags |= EscapeHTML
		}

		marshal = func(b []byte, v interface{}) ([]byte, error) {
			return Append(b, v, flags)
		}

		unmarshal = func(b []byte, v interface{}) error {
			_, err := Parse(b, v, ZeroCopy)
			return err
		}
	}

	os.Exit(m.Run())
}

type point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type tree struct {
	Value string
	Left  *tree
	Right *tree
}

var testValues = [...]interface{}{
	// constants
	nil,
	false,
	true,

	// int
	int(0),
	int(1),
	int(42),
	int(-1),
	int(-42),
	int8(math.MaxInt8),
	int8(math.MinInt8),
	int16(math.MaxInt16),
	int16(math.MinInt16),
	int32(math.MaxInt32),
	int32(math.MinInt32),
	int64(math.MaxInt64),
	int64(math.MinInt64),

	// uint
	uint(0),
	uint(1),
	uintptr(0),
	uintptr(1),
	uint8(math.MaxUint8),
	uint16(math.MaxUint16),
	uint32(math.MaxUint32),
	uint64(math.MaxUint64),

	// float
	float32(0),
	float32(0.5),
	float32(math.SmallestNonzeroFloat32),
	float32(math.MaxFloat32),
	float64(0),
	float64(0.5),
	float64(math.SmallestNonzeroFloat64),
	float64(math.MaxFloat64),

	// number
	json.Number("0"),
	json.Number("1234567890"),
	json.Number("-0.5"),
	json.Number("-1e+2"),

	// string
	"",
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
	strings.Repeat("A", 32),
	strings.Repeat("A", 250),
	strings.Repeat("A", 1020),

	// bytes
	[]byte(""),
	[]byte("Hello World!"),
	bytes.Repeat([]byte("A"), 250),
	bytes.Repeat([]byte("A"), 1020),

	// time
	time.Unix(0, 0).In(time.UTC),
	time.Unix(1, 42).In(time.UTC),
	time.Unix(17179869184, 999999999).In(time.UTC),
	time.Date(2016, 12, 20, 0, 20, 1, 0, time.UTC),

	// array
	[...]int{},
	[...]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},

	// slice
	[]int{},
	[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	makeSlice(250),
	makeSlice(1020),
	[]string{"A", "B", "C"},
	[]interface{}{nil, true, false, 0.5, "Hello World!"},

	// map
	makeMapStringBool(0),
	makeMapStringBool(15),
	makeMapStringBool(1020),
	makeMapStringInterface(0),
	makeMapStringInterface(15),
	makeMapStringInterface(1020),
	map[int]bool{1: false, 42: true},
	map[textValue]bool{{1, 2}: true, {3, 4}: false},
	map[string]*point{
		"A": {1, 2},
		"B": {3, 4},
		"C": {5, 6},
	},
	map[string]RawMessage{
		"A": RawMessage(`{}`),
		"B": RawMessage(`null`),
		"C": RawMessage(`42`),
	},

	// struct
	struct{}{},
	struct{ A int }{42},
	struct{ A, B, C int }{1, 2, 3},
	struct {
		A int
		T time.Time
		S string
	}{42, time.Date(2016, 12, 20, 0, 20, 1, 0, time.UTC), "Hello World!"},
	// These types are interesting because they fit in a pointer so the compiler
	// puts their value directly into the pointer field of the interface{} that
	// is passed to Marshal.
	struct{ X *int }{},
	struct{ X *int }{new(int)},
	struct{ X **int }{},
	// Struct types with more than one pointer, those exercise the regular
	// pointer handling with code that dereferences the fields.
	struct{ X, Y *int }{},
	struct{ X, Y *int }{new(int), new(int)},
	struct {
		A string                 `json:"name"`
		B string                 `json:"-"`
		C string                 `json:",omitempty"`
		D map[string]interface{} `json:",string"`
		e string
	}{A: "Luke", D: map[string]interface{}{"answer": float64(42)}},
	struct{ point }{point{1, 2}},
	tree{
		Value: "T",
		Left:  &tree{Value: "L"},
		Right: &tree{Value: "R", Left: &tree{Value: "R-L"}},
	},

	// pointer
	(*string)(nil),
	new(int),

	// json.Marshaler/json.Unmarshaler
	jsonValue{},
	jsonValue{1, 2},

	// encoding.TextMarshaler/encoding.TextUnmarshaler
	textValue{},
	textValue{1, 2},

	// json.RawMessage
	json.RawMessage(`{
	"answer": 42,
	"hello": "world"
}`),

	// fixtures
	loadTestdata(filepath.Join(runtime.GOROOT(), "src/encoding/json/testdata/code.json.gz")),
}

var durationTestValues = []interface{}{
	// duration
	time.Nanosecond,
	time.Microsecond,
	time.Millisecond,
	time.Second,
	time.Minute,
	time.Hour,

	// struct with duration
	struct{ D1, D2 time.Duration }{time.Millisecond, time.Hour},
}

func makeSlice(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}
	return s
}

func makeMapStringBool(n int) map[string]bool {
	m := make(map[string]bool, n)
	for i := 0; i != n; i++ {
		m[strconv.Itoa(i)] = true
	}
	return m
}

func makeMapStringInterface(n int) map[string]interface{} {
	m := make(map[string]interface{}, n)
	for i := 0; i != n; i++ {
		m[strconv.Itoa(i)] = nil
	}
	return m
}

func testName(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

type codeResponse2 struct {
	Tree     *codeNode2 `json:"tree"`
	Username string     `json:"username"`
}

type codeNode2 struct {
	Name     string      `json:"name"`
	Kids     []*codeNode `json:"kids"`
	CLWeight float64     `json:"cl_weight"`
	Touches  int         `json:"touches"`
	MinT     int64       `json:"min_t"`
	MaxT     int64       `json:"max_t"`
	MeanT    int64       `json:"mean_t"`
}

func loadTestdata(path string) interface{} {
	f, err := os.Open(path)
	if err != nil {
		return err.Error()
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return err.Error()
	}
	defer r.Close()

	testdata := new(codeResponse2)
	if err := json.NewDecoder(r).Decode(testdata); err != nil {
		return err.Error()
	}
	return testdata
}

func TestCodec(t *testing.T) {
	for _, v1 := range testValues {
		t.Run(testName(v1), func(t *testing.T) {
			v2 := newValue(v1)

			a, err := json.MarshalIndent(v1, "", "\t")
			if err != nil {
				t.Error(err)
				return
			}
			a = append(a, '\n')

			buf := &bytes.Buffer{}
			enc := NewEncoder(buf)
			enc.SetIndent("", "\t")

			if err := enc.Encode(v1); err != nil {
				t.Error(err)
				return
			}
			b := buf.Bytes()

			if !Valid(b) {
				t.Error("invalid JSON representation")
			}

			if !bytes.Equal(a, b) {
				t.Error("JSON representations mismatch")
				t.Log("expected:", string(a))
				t.Log("found:   ", string(b))
			}

			dec := NewDecoder(bytes.NewBuffer(b))

			if err := dec.Decode(v2.Interface()); err != nil {
				t.Errorf("%T: %v", err, err)
				return
			}

			x1 := v1
			x2 := v2.Elem().Interface()

			if !reflect.DeepEqual(x1, x2) {
				t.Error("values mismatch")
				t.Logf("expected: %#v", x1)
				t.Logf("found:    %#v", x2)
			}

			if b, err := ioutil.ReadAll(dec.Buffered()); err != nil {
				t.Error(err)
			} else if len(b) != 0 {
				t.Errorf("leftover trailing bytes in the decoder: %q", b)
			}
		})
	}
}

// TestCodecDuration isolates testing of time.Duration.  The stdlib un/marshals
// this type as integers whereas this library un/marshals formatted string
// values.  Therefore, plugging durations into TestCodec would cause fail since
// it checks equality on the marshaled strings from the two libraries.
func TestCodecDuration(t *testing.T) {
	for _, v1 := range durationTestValues {
		t.Run(testName(v1), func(t *testing.T) {
			v2 := newValue(v1)

			// encode using stdlib. (will be an int)
			std, err := json.MarshalIndent(v1, "", "\t")
			if err != nil {
				t.Error(err)
				return
			}
			std = append(std, '\n')

			// decode using our decoder. (reads int to duration)
			dec := NewDecoder(bytes.NewBuffer([]byte(std)))

			if err := dec.Decode(v2.Interface()); err != nil {
				t.Errorf("%T: %v", err, err)
				return
			}

			x1 := v1
			x2 := v2.Elem().Interface()

			if !reflect.DeepEqual(x1, x2) {
				t.Error("values mismatch")
				t.Logf("expected: %#v", x1)
				t.Logf("found:    %#v", x2)
			}

			// encoding using our encoder. (writes duration as string)
			buf := &bytes.Buffer{}
			enc := NewEncoder(buf)
			enc.SetIndent("", "\t")

			if err := enc.Encode(v1); err != nil {
				t.Error(err)
				return
			}
			b := buf.Bytes()

			if !Valid(b) {
				t.Error("invalid JSON representation")
			}

			if reflect.DeepEqual(std, b) {
				t.Error("encoded durations should not match stdlib")
				t.Logf("got: %s", b)
			}

			// decode using our decoder. (reads string to duration)
			dec = NewDecoder(bytes.NewBuffer([]byte(std)))

			if err := dec.Decode(v2.Interface()); err != nil {
				t.Errorf("%T: %v", err, err)
				return
			}

			x1 = v1
			x2 = v2.Elem().Interface()

			if !reflect.DeepEqual(x1, x2) {
				t.Error("values mismatch")
				t.Logf("expected: %#v", x1)
				t.Logf("found:    %#v", x2)
			}
		})
	}
}

func newValue(model interface{}) reflect.Value {
	if model == nil {
		return reflect.New(reflect.TypeOf(&model).Elem())
	}
	return reflect.New(reflect.TypeOf(model))
}

func BenchmarkMarshal(b *testing.B) {
	j := make([]byte, 0, 128*1024)

	for _, v := range testValues {
		b.Run(testName(v), func(b *testing.B) {
			if marshal == nil {
				return
			}

			for i := 0; i != b.N; i++ {
				j, _ = marshal(j[:0], v)
			}

			b.SetBytes(int64(len(j)))
		})
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	for _, v := range testValues {
		b.Run(testName(v), func(b *testing.B) {
			if unmarshal == nil {
				return
			}

			x := v
			if d, ok := x.(time.Duration); ok {
				x = duration(d)
			}

			j, _ := json.Marshal(x)
			x = newValue(v).Interface()

			for i := 0; i != b.N; i++ {
				unmarshal(j, x)
			}

			b.SetBytes(int64(len(j)))
		})
	}
}

type buffer struct{ data []byte }

func (buf *buffer) Write(b []byte) (int, error) {
	buf.data = append(buf.data, b...)
	return len(b), nil
}

func (buf *buffer) WriteString(s string) (int, error) {
	buf.data = append(buf.data, s...)
	return len(s), nil
}

type jsonValue struct {
	x int32
	y int32
}

func (v jsonValue) MarshalJSON() ([]byte, error) {
	return Marshal([2]int32{v.x, v.y})
}

func (v *jsonValue) UnmarshalJSON(b []byte) error {
	var a [2]int32
	err := Unmarshal(b, &a)
	v.x = a[0]
	v.y = a[1]
	return err
}

type textValue struct {
	x int32
	y int32
}

func (v textValue) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("(%d,%d)", v.x, v.y)), nil
}

func (v *textValue) UnmarshalText(b []byte) error {
	_, err := fmt.Sscanf(string(b), "(%d,%d)", &v.x, &v.y)
	return err
}

type duration time.Duration

func (d duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Duration(d).String() + `"`), nil
}

func (d *duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	x, err := time.ParseDuration(s)
	*d = duration(x)
	return err
}

var (
	_ json.Marshaler = jsonValue{}
	_ json.Marshaler = duration(0)

	_ encoding.TextMarshaler = textValue{}

	_ json.Unmarshaler = (*jsonValue)(nil)
	_ json.Unmarshaler = (*duration)(nil)

	_ encoding.TextUnmarshaler = (*textValue)(nil)
)

func TestDecodeStructFieldCaseInsensitive(t *testing.T) {
	b := []byte(`{ "type": "changed" }`)
	s := struct {
		Type string
	}{"unchanged"}

	if err := json.Unmarshal(b, &s); err != nil {
		t.Error(err)
	}

	if s.Type != "changed" {
		t.Error("s.Type: expected to be changed but found", s.Type)
	}
}

func TestDecodeLines(t *testing.T) {
	tests := []struct {
		desc        string
		reader      io.Reader
		expectCount int
	}{

		// simple

		{
			desc:        "bare object",
			reader:      strings.NewReader("{\"Good\":true}"),
			expectCount: 1,
		},
		{
			desc:        "multiple objects on one line",
			reader:      strings.NewReader("{\"Good\":true}{\"Good\":true}\n"),
			expectCount: 2,
		},
		{
			desc:        "object spanning multiple lines",
			reader:      strings.NewReader("{\n\"Good\":true\n}\n"),
			expectCount: 1,
		},

		// whitespace handling

		{
			desc:        "trailing newline",
			reader:      strings.NewReader("{\"Good\":true}\n{\"Good\":true}\n"),
			expectCount: 2,
		},
		{
			desc:        "multiple trailing newlines",
			reader:      strings.NewReader("{\"Good\":true}\n{\"Good\":true}\n\n"),
			expectCount: 2,
		},
		{
			desc:        "blank lines",
			reader:      strings.NewReader("{\"Good\":true}\n\n{\"Good\":true}"),
			expectCount: 2,
		},
		{
			desc:        "no trailing newline",
			reader:      strings.NewReader("{\"Good\":true}\n{\"Good\":true}"),
			expectCount: 2,
		},
		{
			desc:        "leading whitespace",
			reader:      strings.NewReader("  {\"Good\":true}\n\t{\"Good\":true}"),
			expectCount: 2,
		},

		// multiple reads

		{
			desc: "one object, multiple reads",
			reader: io.MultiReader(
				strings.NewReader("{"),
				strings.NewReader("\"Good\": true"),
				strings.NewReader("}\n"),
			),
			expectCount: 1,
		},

		// EOF reads

		{
			desc:        "one object + EOF",
			reader:      &eofReader{"{\"Good\":true}\n"},
			expectCount: 1,
		},
		{
			desc:        "leading whitespace + EOF",
			reader:      &eofReader{"\n{\"Good\":true}\n"},
			expectCount: 1,
		},
		{
			desc:        "multiple objects + EOF",
			reader:      &eofReader{"{\"Good\":true}\n{\"Good\":true}\n"},
			expectCount: 2,
		},
		{
			desc: "one object + multiple reads + EOF",
			reader: io.MultiReader(
				strings.NewReader("{"),
				strings.NewReader("  \"Good\": true"),
				&eofReader{"}\n"},
			),
			expectCount: 1,
		},
		{
			desc: "multiple objects + multiple reads + EOF",
			reader: io.MultiReader(
				strings.NewReader("{"),
				strings.NewReader("  \"Good\": true}{\"Good\": true}"),
				&eofReader{"\n"},
			),
			expectCount: 2,
		},

		{
			// the 2nd object should be discarded, as 42 cannot be cast to bool
			desc:        "unmarshal error while decoding",
			reader:      strings.NewReader("{\"Good\":true}\n{\"Good\":42}\n{\"Good\":true}\n"),
			expectCount: 2,
		},
		{
			// the 2nd object should be discarded, as 42 cannot be cast to bool
			desc:        "unmarshal error while decoding last object",
			reader:      strings.NewReader("{\"Good\":true}\n{\"Good\":42}\n"),
			expectCount: 1,
		},
	}

	type obj struct {
		Good bool
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			d := NewDecoder(test.reader)
			var count int
			var err error
			for {
				var o obj
				err = d.Decode(&o)
				if err != nil {
					if err == io.EOF {
						break
					}

					switch err.(type) {
					case *json.SyntaxError, *json.UnmarshalTypeError, *json.UnmarshalFieldError:
						t.Log("unmarshal error", err)
						continue
					}

					t.Error("decode error", err)
					break
				}
				if !o.Good {
					t.Errorf("object was not unmarshaled correctly: %#v", o)
				}
				count++
			}

			if err != nil && err != io.EOF {
				t.Error(err)
			}

			if count != test.expectCount {
				t.Errorf("expected %d objects, got %d", test.expectCount, count)
			}
		})
	}
}

// eofReader is a simple io.Reader that reads its full contents _and_ returns
// and EOF in the first call. Subsequent Read calls only return EOF.
type eofReader struct {
	s string
}

func (r *eofReader) Read(p []byte) (n int, err error) {
	n = copy(p, r.s)
	r.s = r.s[n:]
	if r.s == "" {
		err = io.EOF
	}
	return
}

func TestDontMatchCaseIncensitiveStructFields(t *testing.T) {
	b := []byte(`{ "type": "changed" }`)
	s := struct {
		Type string
	}{"unchanged"}

	if _, err := Parse(b, &s, DontMatchCaseInsensitiveStructFields); err != nil {
		t.Error(err)
	}

	if s.Type != "unchanged" {
		t.Error("s.Type: expected to be unchanged but found", s.Type)
	}
}

func TestMarshalFuzzBugs(t *testing.T) {
	tests := []struct {
		value  interface{}
		output string
	}{
		{ // html sequences are escaped even in RawMessage
			value: struct {
				P RawMessage
			}{P: RawMessage(`"<"`)},
			output: "{\"P\":\"\\u003c\"}",
		},
		{ // raw message output is compacted
			value: struct {
				P RawMessage
			}{P: RawMessage(`{"" :{}}`)},
			output: "{\"P\":{\"\":{}}}",
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			b, err := Marshal(test.value)
			if err != nil {
				t.Fatal(err)
			}

			if string(b) != test.output {
				t.Error("values mismatch")
				t.Logf("expected: %#v", test.output)
				t.Logf("found:    %#v", string(b))
			}
		})
	}
}

func TestUnmarshalFuzzBugs(t *testing.T) {
	tests := []struct {
		input string
		value interface{}
	}{
		{ // non-UTF8 sequences must be converted to the utf8.RuneError character.
			input: "[\"00000\xef\"]",
			value: []interface{}{"00000�"},
		},
		{ // UTF16 surrogate followed by null character
			input: "[\"\\ud800\\u0000\"]",
			value: []interface{}{"�\x00"},
		},
		{ // UTF16 surrogate followed by ascii character
			input: "[\"\\uDF00\\u000e\"]",
			value: []interface{}{"�\x0e"},
		},
		{ // UTF16 surrogate followed by unicode character
			input: "[[\"\\uDF00\\u0800\"]]",
			value: []interface{}{[]interface{}{"�ࠀ"}},
		},
		{ // invalid UTF16 surrogate sequenced followed by a valid UTF16 surrogate sequence
			input: "[\"\\udf00\\udb00\\udf00\"]",
			value: []interface{}{"�\U000d0300"},
		},
		{ // decode single-element slice into []byte field
			input: "{\"f\":[0],\"0\":[0]}",
			value: struct {
				F []byte
			}{F: []byte{0}},
		},
		{ // decode multi-element slice into []byte field
			input: "{\"F\":[3,1,1,1,9,9]}",
			value: struct {
				F []byte
			}{F: []byte{3, 1, 1, 1, 9, 9}},
		},
		{ // decode string with escape sequence into []byte field
			input: "{\"F\":\"0p00\\r\"}",
			value: struct {
				F []byte
			}{F: []byte("ҝ4")},
		},
		{ // decode unicode code points which fold into ascii characters
			input: "{\"ſ\":\"8\"}",
			value: struct {
				S int `json:",string"`
			}{S: 8},
		},
		{ // decode unicode code points which don't fold into ascii characters
			input: "{\"İ\":\"\"}",
			value: struct {
				I map[string]string
			}{I: nil},
		},
		{ // override pointer-to-pointer field clears the inner pointer only
			input: "{\"o\":0,\"o\":null}",
			value: struct {
				O **int
			}{O: new(*int)},
		},
		{ // subsequent occurrences of a map field retain keys previously loaded
			input: "{\"i\":{\"\":null},\"i\":{}}",
			value: struct {
				I map[string]string
			}{I: map[string]string{"": ""}},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			ptr := reflect.New(reflect.TypeOf(test.value)).Interface()

			if err := Unmarshal([]byte(test.input), ptr); err != nil {
				t.Fatal(err)
			}

			if value := reflect.ValueOf(ptr).Elem().Interface(); !reflect.DeepEqual(test.value, value) {
				t.Error("values mismatch")
				t.Logf("expected: %#v", test.value)
				t.Logf("found:    %#v", value)
			}
		})
	}
}
