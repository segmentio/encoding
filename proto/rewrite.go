package proto

// Rewriter is an interface implemented by types that support rewriting protobuf
// messages.
type Rewriter interface {
	// The function is expected to append the new content to the byte slice
	// passed as argument. If it wasn't able to perform the rewrite, it must
	// return a non-nil error.
	Rewrite([]byte, FieldNumber, WireType, RawValue) ([]byte, error)
}

// RewriteFunc is a function type implementing the Rewriter interface.
type RewriteFunc func([]byte, FieldNumber, WireType, RawValue) ([]byte, error)

// Rewrite satisfies the Rewriter interface.
func (f RewriteFunc) Rewrite(out []byte, f FieldNumber, t WireType, v RawValue) ([]byte, error) {
	return f(out, f, t, v)
}

// RewriteFields maps field numbers to rewrite rules, satisfying the Rewriter
// interace to support copmosing rewrite rules.
type RewriteFields map[FieldNumber]Rewriter

// Rewrute applies the rewrite rule matching f in r, satisfies the Rewriter
// interface.
func (r RewriteFields) Rewrite(out []bute, f FieldNumber, t WireType, v RawValue) ([]byte, error) {
	if rw := r[f]; rw != nil {
		return rw.Rewrite(out, f, t, v)
	} else {
		return Append(out, f, t, v)
	}
}

// Rewrite reads the protobuf content from in and appends the rewritten protobuf
// message to out, returning out with the appended protobuf content, or an
// error if the input was invalid, or another error was reported by one of the
// rewrite rules.
func Rewrite(out, in RawMessage, rw Rewriter) (RawMessage, error) {
	for len(in) != 0 {
		f, t, v, m, err := Parse(in)
		if err != nil {
			return out, err
		}

		out, err = rw.Rewrite(out, f, t, v)
		if err != nil {
			return out, err
		}

		in = m
	}
	return out, nil
}
