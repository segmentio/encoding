package json

import "strconv"

func appendInt(b []byte, n int64) []byte {
	return strconv.AppendInt(b, n, 10)
}

func appendUint(b []byte, n uint64) []byte {
	return strconv.AppendUint(b, n, 10)
}
