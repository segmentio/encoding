package json

import "sync"

// Tokenizer is an iterator-style type which can be used to progressively parse
// through a json input.
//
// Tokenizing json is useful to build highly efficient parsing operations, for
// example when doing tranformations on-the-fly where as the program reads the
// input and produces the transformed json to an output buffer.
//
// Here is a common pattern to use a tokenizer:
//
//	for t := json.NewTokenizer(b); t.Next(); {
//		switch t.Delim {
//		case '{':
//			...
//		case '}':
//			...
//		case '[':
//			...
//		case ']':
//			...
//		case ':':
//			...
//		case ',':
//			...
//		}
//
//		switch {
//		case t.Value.String():
//			...
//		case t.Value.Null():
//			...
//		case t.Value.True():
//			...
//		case t.Value.False():
//			...
//		case t.Value.Number():
//			...
//		}
//	}
//
type Tokenizer struct {
	// When the tokenizer is positioned on a json delimiter this field is not
	// zero. In this case the possible values are '{', '}', '[', ']', ':', and
	// ','.
	Delim Delim

	// This field contains the raw json token that the tokenizer is pointing at.
	// When Delim is not zero, this field is a single-element byte slice
	// continaing the delimiter value. Otherwise, this field holds values like
	// null, true, false, numbers, or quoted strings.
	Value RawValue

	// When the tokenizer has encountered invalid content this field is not nil.
	Err error

	// When the value is in an array or an object, this field contains the depth
	// at which it was found.
	Depth int

	// When the value is in an array or an object, this field contains the
	// position at which it was found.
	Index int

	// This field is true when the value is the key of an object.
	IsKey bool

	// Tells whether the next value read from the tokenizer is a key.
	isKey bool

	// json input for the tokenizer, pointing at data right after the last token
	// that was parsed.
	json []byte

	// Stack used to track entering and leaving arrays, objects, and keys.
	stack *stack

	// Decoder used for parsing.
	decoder decoder
}

// NewTokenizer constructs a new Tokenizer which reads its json input from b.
func NewTokenizer(b []byte) *Tokenizer {
	return &Tokenizer{
		json:    b,
		decoder: decoder{flags: internalParseFlags(b)},
	}
}

// Reset erases the state of t and re-initializes it with the json input from b.
func (t *Tokenizer) Reset(b []byte) {
	if t.stack != nil {
		releaseStack(t.stack)
	}
	// This code is similar to:
	//
	//	*t = Tokenizer{json: b}
	//
	// However, it does not compile down to an invocation of duff-copy.
	t.Delim = 0
	t.Value = nil
	t.Err = nil
	t.Depth = 0
	t.Index = 0
	t.IsKey = false
	t.isKey = false
	t.json = b
	t.stack = nil
	t.decoder = decoder{flags: internalParseFlags(b)}
}

// Next returns a new tokenizer pointing at the next token, or the zero-value of
// Tokenizer if the end of the json input has been reached.
//
// If the tokenizer encounters malformed json while reading the input the method
// sets t.Err to an error describing the issue, and returns false. Once an error
// has been encountered, the tokenizer will always fail until its input is
// cleared by a call to its Reset method.
func (t *Tokenizer) Next() bool {
	if t.Err != nil {
		return false
	}

	// Inlined code of the skipSpaces function, this give a ~15% speed boost.
	i := 0
skipLoop:
	for _, c := range t.json {
		switch c {
		case sp, ht, nl, cr:
			i++
		default:
			break skipLoop
		}
	}

	if i > 0 {
		t.json = t.json[i:]
	}

	if len(t.json) == 0 {
		t.Reset(nil)
		return false
	}

	switch t.json[0] {
	case '"':
		t.Delim = 0
		t.Value, t.json, t.Err = t.decoder.parseString(t.json)
	case 'n':
		t.Delim = 0
		t.Value, t.json, t.Err = t.decoder.parseNull(t.json)
	case 't':
		t.Delim = 0
		t.Value, t.json, t.Err = t.decoder.parseTrue(t.json)
	case 'f':
		t.Delim = 0
		t.Value, t.json, t.Err = t.decoder.parseFalse(t.json)
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		t.Delim = 0
		t.Value, t.json, t.Err = t.decoder.parseNumber(t.json)
	case '{', '}', '[', ']', ':', ',':
		t.Delim, t.Value, t.json = Delim(t.json[0]), t.json[:1], t.json[1:]
	default:
		t.Delim = 0
		t.Value, t.json, t.Err = t.json[:1], t.json[1:], syntaxError(t.json, "expected token but found '%c'", t.json[0])
	}

	t.Depth = t.depth()
	t.Index = t.index()

	if t.Delim == 0 {
		t.IsKey = t.isKey
	} else {
		t.IsKey = false

		switch t.Delim {
		case '{':
			t.isKey = true
			t.push(inObject)
		case '[':
			t.push(inArray)
		case '}':
			t.Err = t.pop(inObject)
			t.Depth--
			t.Index = t.index()
		case ']':
			t.Err = t.pop(inArray)
			t.Depth--
			t.Index = t.index()
		case ':':
			t.isKey = false
		case ',':
			if t.stack == nil || len(t.stack.state) == 0 {
				t.Err = syntaxError(t.json, "found unexpected comma")
				return false
			}
			if t.stack.is(inObject) {
				t.isKey = true
			}
			t.stack.state[len(t.stack.state)-1].len++
		}
	}

	return (t.Delim != 0 || len(t.Value) != 0) && t.Err == nil
}

func (t *Tokenizer) depth() int {
	if t.stack == nil {
		return 0
	}
	return t.stack.depth()
}

func (t *Tokenizer) index() int {
	if t.stack == nil {
		return 0
	}
	return t.stack.index()
}

func (t *Tokenizer) push(typ scope) {
	if t.stack == nil {
		t.stack = acquireStack()
	}
	t.stack.push(typ)
}

func (t *Tokenizer) pop(expect scope) error {
	if t.stack == nil || !t.stack.pop(expect) {
		return syntaxError(t.json, "found unexpected character while tokenizing json input")
	}
	return nil
}

// RawValue represents a raw json value, it is intended to carry null, true,
// false, number, and string values only.
type RawValue []byte

// String returns true if v contains a string value.
func (v RawValue) String() bool { return len(v) != 0 && v[0] == '"' }

// Null returns true if v contains a null value.
func (v RawValue) Null() bool { return len(v) != 0 && v[0] == 'n' }

// True returns true if v contains a true value.
func (v RawValue) True() bool { return len(v) != 0 && v[0] == 't' }

// False returns true if v contains a false value.
func (v RawValue) False() bool { return len(v) != 0 && v[0] == 'f' }

// Number returns true if v contains a number value.
func (v RawValue) Number() bool {
	if len(v) != 0 {
		switch v[0] {
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return true
		}
	}
	return false
}

// AppendUnquote writes the unquoted version of the string value in v into b.
func (v RawValue) AppendUnquote(b []byte) []byte {
	d := decoder{}
	s, r, new, err := d.parseStringUnquote(v, b)
	if err != nil {
		panic(err)
	}
	if len(r) != 0 {
		panic(syntaxError(r, "unexpected trailing tokens after json value"))
	}
	if new {
		b = s
	} else {
		b = append(b, s...)
	}
	return b
}

// Unquote returns the unquoted version of the string value in v.
func (v RawValue) Unquote() []byte {
	return v.AppendUnquote(nil)
}

type scope int

const (
	inArray scope = iota
	inObject
)

type state struct {
	typ scope
	len int
}

type stack struct {
	state []state
}

func (s *stack) push(typ scope) {
	s.state = append(s.state, state{typ: typ, len: 1})
}

func (s *stack) pop(expect scope) bool {
	i := len(s.state) - 1

	if i < 0 {
		return false
	}

	if found := s.state[i]; expect != found.typ {
		return false
	}

	s.state = s.state[:i]
	return true
}

func (s *stack) is(typ scope) bool {
	return len(s.state) != 0 && s.state[len(s.state)-1].typ == typ
}

func (s *stack) depth() int {
	return len(s.state)
}

func (s *stack) index() int {
	if len(s.state) == 0 {
		return 0
	}
	return s.state[len(s.state)-1].len - 1
}

func acquireStack() *stack {
	s, _ := stackPool.Get().(*stack)
	if s == nil {
		s = &stack{state: make([]state, 0, 4)}
	} else {
		s.state = s.state[:0]
	}
	return s
}

func releaseStack(s *stack) {
	stackPool.Put(s)
}

var (
	stackPool sync.Pool // *stack
)
