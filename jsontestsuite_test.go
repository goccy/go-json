package json_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

func jsonTestSuiteFiles(t *testing.T, suite string) []string {
	t.Helper()

	files, err := filepath.Glob(filepath.Join("testdata", "jsontestsuite", suite, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatalf("no JSONTestSuite files found in %s", suite)
	}
	return files
}

func TestJSONTestSuiteParsing(t *testing.T) {
	for _, file := range jsonTestSuiteFiles(t, "test_parsing") {
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}

			var value any
			err = json.Unmarshal(data, &value)
			switch {
			case strings.HasPrefix(name, "y_"):
				if err != nil {
					t.Fatalf("Unmarshal rejected valid JSON: %v", err)
				}
			case strings.HasPrefix(name, "n_"):
				if err == nil {
					t.Fatal("Unmarshal accepted invalid JSON")
				}
			case strings.HasPrefix(name, "i_"):
				// The suite leaves acceptance of these files to the parser.
				if err == nil {
					encoded, err := json.Marshal(value)
					if err != nil {
						t.Fatalf("Marshal accepted input could not be encoded: %v", err)
					}
					if !json.Valid(encoded) {
						t.Fatal("Marshal returned invalid JSON")
					}
				}
			default:
				t.Fatalf("unrecognized JSONTestSuite filename: %s", name)
			}
		})
	}
}

func TestJSONTestSuiteTransform(t *testing.T) {
	for _, file := range jsonTestSuiteFiles(t, "test_transform") {
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}

			var value any
			if err := json.Unmarshal(data, &value); err != nil {
				// Transform vectors have no required parse result or expected output.
				return
			}
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatalf("Marshal parsed value: %v", err)
			}
			if !json.Valid(encoded) {
				t.Fatal("Marshal returned invalid JSON")
			}
		})
	}
}
