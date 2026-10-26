package cache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const key = "test-key"

func checkIOError(t *testing.T, got error, wantErr *IOError) {
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

	var ioe *IOError
	if !errors.As(got, &ioe) {
		t.Fatalf("expected err type %T, got %T instead", wantErr, got)
	}

	if ioe.Path != wantErr.Path {
		t.Errorf("expected path %q, got %q instead", wantErr.Path, ioe.Path)
	}
}

func writeFile(t *testing.T, dir, rel, content string) string {
	t.Helper()

	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	return rel
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}

	if string(got) != want {
		t.Errorf("%s = %q, want %q", path, got, want)
	}
}

func TestPutZeroOutputs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := t.TempDir()

	c := NewL1(root)
	err := c.Put(key, []string{}, dir)
	checkIOError(t, err, nil)

	path := filepath.Join(root, key)
	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatalf("expected directory %q to exist", path)
	}

	if len(entries) != 0 {
		t.Errorf("expected zero entries in %q, got %d", path, len(entries))
	}
}

