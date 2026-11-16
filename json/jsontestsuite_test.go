package json

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSONTestSuiteParsing(t *testing.T) {
	files := jsonTestSuiteFiles(t, "test_parsing")
	if got, want := len(files), 318; got != want {
		t.Fatalf("found %d test_parsing cases, want %d", got, want)
	}

	for _, file := range files {
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}

			var value interface{}
			err = Unmarshal(data, &value)
			switch {
			case strings.HasPrefix(name, "y_"):
				if err != nil {
					t.Errorf("Unmarshal() error = %v, want nil", err)
				}
			case strings.HasPrefix(name, "n_"):
				if err == nil {
					t.Error("Unmarshal() error = nil, want an error")
				}
			case strings.HasPrefix(name, "i_"):
				// Both accepting and rejecting implementation-defined cases are valid.
			default:
				t.Fatalf("unexpected test_parsing case name %q", name)
			}
		})
	}
}

func TestJSONTestSuiteTransform(t *testing.T) {
	files := jsonTestSuiteFiles(t, "test_transform")
	if got, want := len(files), 22; got != want {
		t.Fatalf("found %d test_transform cases, want %d", got, want)
	}

	for _, file := range files {
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}

			var value interface{}
			if err := Unmarshal(data, &value); err != nil {
				// test_transform has no required parse outcome.
				return
			}

			encoded, err := Marshal(value)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if !Valid(encoded) {
				t.Errorf("Marshal() returned invalid JSON: %s", encoded)
			}
		})
	}
}

func jsonTestSuiteFiles(t *testing.T, suite string) []string {
	t.Helper()

	files, err := filepath.Glob(filepath.Join("testdata", "JSONTestSuite", suite, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	return files
}
