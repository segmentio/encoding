package json

import (
	"math"
	"strconv"
	"testing"
)

func benchStd(b *testing.B, n int64) {
	var buf [20]byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strconv.AppendInt(buf[:0], n, 10)
	}
}

func benchNew(b *testing.B, n int64) {
	var buf [20]byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		appendInt(buf[:0], n)
	}
}

func BenchmarkAppendIntStd1(b *testing.B) {
	benchStd(b, 1)
}

func BenchmarkAppendInt1(b *testing.B) {
	benchNew(b, 1)
}

func BenchmarkAppendIntStdMinI64(b *testing.B) {
	benchStd(b, math.MinInt64)
}

func BenchmarkAppendIntMinI64(b *testing.B) {
	benchNew(b, math.MinInt64)
}
