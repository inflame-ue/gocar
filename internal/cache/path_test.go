package cache

import (
	"errors"
	"testing"
)

func checkPathError(t *testing.T, got error, wantErr *PathError) {
	t.Helper()

	if got == nil && wantErr == nil {
		return
	}

	if got != nil && wantErr == nil {
		t.Fatalf("expected no err, got %v", got)
	}

	if got == nil && wantErr != nil {
		t.Fatalf("expected an err, got no err")
	}

	var pe *PathError
	if !errors.As(got, &pe) {
		t.Fatalf("expected err type %T, got %T instead", wantErr, got)
	}

	if pe.Path != wantErr.Path {
		t.Errorf("expected path %q, got %q instead", wantErr.Path, pe.Path)
	}
}

func TestSafeStore(t *testing.T) {
	testDir := t.TempDir()

	tests := map[string]struct {
		path     string
		wantPath string
		wantErr  *PathError
	}{
		"normal path": {
			path:     "src/main.go",
			wantPath: "src/main.go",
		},
		"valid ..": {
			path:     "src/../main.go",
			wantPath: "main.go",
		},
		"leading ..": {
			path: "../main.go",
			wantErr: &PathError{
				Path: "../main.go",
			},
		},
		"absolute path": {
			path: "/abs/main.go",
			wantErr: &PathError{
				Path: "/abs/main.go",
			},
		},
		"empty path": {
			path: "",
			wantErr: &PathError{
				Path: "",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path, err := safeStore(testDir, tc.path)
			checkPathError(t, err, tc.wantErr)

			if path != tc.wantPath {
				t.Errorf("expected path %q, got %q instead", tc.path, path)
			}
		})
	}
}
