package cache

import (
	"testing"

	"github.com/inflame-ue/gocar/internal/testutil"
)

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

			tErr := testutil.CheckErrAs(t, err, tc.wantErr)
			if tc.wantErr != nil && tErr.Path != tc.wantErr.Path {
				t.Errorf("expected path %q, got %q instead", tc.wantErr.Path, tErr.Path)
			}

			if path != tc.wantPath {
				t.Errorf("expected path %q, got %q instead", tc.path, path)
			}
		})
	}
}
