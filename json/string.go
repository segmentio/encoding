package json

func simpleString(s string, escapeHTML bool) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c > 0x7f || c == '"' || c == '\\' || (escapeHTML && (c == '<' || c == '>' || c == '&')) {
			return false
		}
	}

	return true
}
