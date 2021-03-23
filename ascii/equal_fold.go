package ascii

// EqualFold is a version of bytes.EqualFold designed to work on ASCII input
// instead of UTF-8.
//
// When the program has guarantees that the input is composed of ASCII
// characters only, it allows for greater optimizations.
func EqualFold(a, b []byte) bool {
	return EqualFoldString(unsafeString(a), unsafeString(b))
}

func HasPrefixFold(s, prefix []byte) bool {
	return len(s) >= len(prefix) && EqualFold(s, prefix)
}

func HasSuffixFold(s, prefix []byte) bool {
	return len(s) >= len(prefix) && EqualFold(s[len(s)-len(prefix):], prefix)
}

// EqualFoldString is a version of strings.EqualFold designed to work on ASCII
// input instead of UTF-8.
//
// When the program has guarantees that the input is composed of ASCII
// characters only, it allows for greater optimizations.
func EqualFoldString(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		c := a[i]
		d := b[i]

		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}

		if 'A' <= d && d <= 'Z' {
			d += 'a' - 'A'
		}

		if c != d {
			return false
		}
	}

	return true
}

func HasPrefixFoldString(s, prefix string) bool {
	return len(s) >= len(prefix) && EqualFoldString(s[:len(prefix)], prefix)
}

func HasSuffixFoldString(s, prefix string) bool {
	return len(s) >= len(prefix) && EqualFoldString(s[len(s)-len(prefix):], prefix)
}
