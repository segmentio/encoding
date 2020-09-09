package proto

// Rewriter is an interface implemented by types that support rewriting protobuf
// messages.
type Rewriter interface {
	// The function is expected to append the new content to the byte slice
	// passed as argument. If it wasn't able to perform the rewrite, it must
	// return a non-nil error.
	Rewrite(out, in []byte) ([]byte, error)
}

// RewriteFunc is a function type implementing the Rewriter interface.
type RewriteFunc func([]byte, []byte) ([]byte, error)

// Rewrite satisfies the Rewriter interface.
func (r RewriteFunc) Rewrite(out, in []byte) ([]byte, error) {
	return r(out, in)
}

// RewriteFields maps field numbers to rewrite rules, satisfying the Rewriter
// interace to support copmosing rewrite rules.
type RewriteFields map[FieldNumber]Rewriter

// Rewrute applies the rewrite rule matching f in r, satisfies the Rewriter
// interface.
func (r RewriteFields) Rewrite(out, in []byte) ([]byte, error) {
	buffer := [4]uint64{}
	fields := fieldset(buffer[:])

	for f := range r {
		i := int(f) - 1
		n := fields.len()

		if i >= 0 {
			if i > n {
				fields = growFields(fields, i+1)
			}
			fields.set(i)
		}
	}

	for len(in) != 0 {
		f, t, v, m, err := Parse(in)
		if err != nil {
			return out, err
		}

		if rw := r[f]; rw != nil && fields.has(int(f)-1) {
			if out, err = rw.Rewrite(out, v); err != nil {
				return out, err
			}
		} else {
			out = Append(out, f, t, v)
		}

		fields.unset(int(f) - 1)
		in = m
	}

	for i := range fields {
		if fields.has(i) {
			b, err := r[FieldNumber(i+1)].Rewrite(out, nil)
			if err != nil {
				return b, err
			}
			out = b
		}
	}

	return out, nil
}

type fieldset []uint64

func growFields(f fieldset, n int) fieldset {
	g := makeFieldset(n)
	copy(g, f)
	return g
}

func makeFieldset(n int) fieldset {
	if (n % 64) != 0 {
		n = (n + 1) / 64
	} else {
		n /= 64
	}
	return make(fieldset, n)
}

func (f fieldset) len() int {
	return len(f) * 64
}

func (f fieldset) has(i int) bool {
	x, y := f.index(i)
	return ((f[x] >> y) & 1) != 0
}

func (f fieldset) set(i int) {
	x, y := f.index(i)
	f[x] |= 1 << y
}

func (f fieldset) unset(i int) {
	x, y := f.index(i)
	f[x] &= ^(1 << y)
}

func (f fieldset) index(i int) (int, int) {
	return i / 64, i % 64
}
