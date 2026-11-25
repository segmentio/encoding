package json_test

import (
	"strings"
	"testing"
	"time"

	"github.com/segmentio/encoding/json"
)

type durationErrStruct struct {
	D time.Duration
}

func TestDecodeDurationErrorType(t *testing.T) {
	var s durationErrStruct
	err := json.Unmarshal([]byte(`{"D":true}`), &s)
	if err == nil {
		t.Fatal("expected an error decoding a bool into time.Duration")
	}
	if !strings.Contains(err.Error(), "time.Duration") {
		t.Fatalf("error = %q, want it to mention time.Duration", err)
	}
}

func TestDecodeDurationOverflow(t *testing.T) {
	var s durationErrStruct
	// math.MaxInt64 + 1, which parseInt must reject.
	err := json.Unmarshal([]byte(`{"D":9223372036854775808}`), &s)
	if err == nil {
		t.Fatal("expected an error decoding an out-of-range duration")
	}
	if !strings.Contains(err.Error(), "time.Duration") {
		t.Fatalf("error = %q, want it to mention time.Duration", err)
	}
}
