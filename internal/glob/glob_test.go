package glob

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func checkErr(t *testing.T, got, want error) {
	t.Helper()
	if reflect.TypeOf(got) != reflect.TypeOf(want) {
		t.Fatalf("expected %T, got %T", want, got)
	}
}

func makeTree(t *testing.T) string {
	t.Helper()

	testDir := t.TempDir()
	files := []string{"main.go", "src/main.go", "src/pkg/util.go",
		"src/pkg/util_test.go", "src/cmd/cli/main.go"}

	for _, filename := range files {
		path := filepath.Join(testDir, filename)

		err := os.MkdirAll(filepath.Dir(path), 0o755)
		if err != nil {
			t.Fatalf("failed to mkdir: %v", err)
		}

		file, err := os.Create(path)
		if err != nil {
			t.Fatalf("could not create file %s: %v", filename, err)
		}
		file.Close()
	}

	return testDir
}

func TestExpand(t *testing.T) {
	dir := makeTree(t)
	tests := map[string]struct {
		pattern string
		want    []string
		wantErr error
	}{
		"literal": {
			pattern: "main.go",
			want:    []string{"main.go"},
		},
		"missing literal": {
			pattern: "missing.go",
			wantErr: &NoMatchError{
				Pattern: "missing.go",
			},
		},
		"* expansion": {
			pattern: "src/*.go",
			want:    []string{"src/main.go"},
		},
		"** expansion": {
			pattern: "src/**/*.go",
			want:    []string{"src/cmd/cli/main.go", "src/main.go", "src/pkg/util.go", "src/pkg/util_test.go"},
		},
		"* expansion no match": {
			pattern: "*.rs",
			wantErr: &NoMatchError{
				Pattern: "*.rs",
			},
		},
		"invalid pattern": {
			pattern: "../*.go",
			wantErr: &PatternError{
				Pattern: "../*.go",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			paths, err := Expand(dir, tc.pattern)
			checkErr(t, err, tc.wantErr)

			if !slices.Equal(paths, tc.want) {
				t.Errorf("expected paths %v, got %v instead", tc.want, paths)
			}
		})
	}
}

func TestSplitSegments(t *testing.T) {
	tests := map[string]struct {
		pattern      string
		wantSegments []string
		wantErr      error
	}{
		"valid pattern": {
			pattern:      "src/**/*.go",
			wantSegments: []string{"src", "**", "*.go"},
			wantErr:      nil,
		},
		"absolute path": {
			pattern:      "/home/user/src/**/*.go",
			wantSegments: nil,
			wantErr: &PatternError{
				Pattern: "/home/user/src/**/*.go",
			},
		},
		"double star inside segment": {
			pattern:      "s**rc/**/*.go",
			wantSegments: nil,
			wantErr: &PatternError{
				Pattern: "s**rc/**/*.go",
			},
		},
		"escaping pattern": {
			pattern:      "../*.go",
			wantSegments: nil,
			wantErr: &PatternError{
				Pattern: "../*.go",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			segments, err := splitSegments(tc.pattern)
			checkErr(t, err, tc.wantErr)

			if !slices.Equal(segments, tc.wantSegments) {
				t.Errorf("expected segments %v, got %v instead", tc.wantSegments, segments)
			}
		})
	}
}
