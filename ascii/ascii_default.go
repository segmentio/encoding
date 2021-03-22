// +build !amd64

package ascii

func validAVX2(s *byte, n uintptr) int { return valid16(s, n) }

func validPrintAVX2(s *byte, n uintptr) int { return validPrint16(s, n) }
