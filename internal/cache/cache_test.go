package cache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const key = "test-key"

// TODO: the issue of repeating error check helpers is prelevant enough
// to introduce a generic that will be utilized across all of the internal packages.
// This has to be done after the cache functionality is implemented, so as to not
// pollute the implementation branch.
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

func TestPutMultiOutputMirror(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := t.TempDir()
	c := NewL1(root)

	files := []struct{ rel, content string }{
		{"src/main.go", "package main"},
		{"other/main.go", "package other"},
	}

	var outputs []string
	for _, f := range files {
		outputs = append(outputs, writeFile(t, dir, f.rel, f.content))
	}

	err := c.Put(key, outputs, dir)
	checkIOError(t, err, nil)

	for _, f := range files {
		assertFile(t, filepath.Join(root, key, f.rel), f.content)
	}
}

func TestPutDuplicate(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := t.TempDir()
	c := NewL1(root)

	files := []struct{ rel, content string }{
		{"src/main.go", "package main"},
		{"src/main.go", "package other"},
	}

	var outputs []string
	outputs = []string{writeFile(t, dir, files[0].rel, files[0].content)}
	err := c.Put(key, outputs, dir)
	checkIOError(t, err, nil)

	outputs = []string{writeFile(t, dir, files[1].rel, files[1].content)}
	err = c.Put(key, outputs, dir)
	checkIOError(t, err, nil)

	assertFile(t, filepath.Join(root, key, files[0].rel), files[0].content)
}

func TestPutMissingSource(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := t.TempDir()
	c := NewL1(root)

	err := c.Put(key, []string{"missing.go"}, dir)
	checkIOError(t, err, &IOError{Path: filepath.Join(dir, "missing.go")})

	if _, err := os.Stat(filepath.Join(root, key)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("cache key dir should not exist after failed Put")
	}
}

func TestPutEscapingOutput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := t.TempDir()
	c := NewL1(root)

	err := c.Put(key, []string{"../escaping.go"}, dir)
	checkPathError(t, err, &PathError{Path: "../escaping.go"})
}

func TestPutGetRoundTrip(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := t.TempDir()
	c := NewL1(root)

	file := struct{ rel, content string }{"src/main.go", "package main"}

	var outputs = []string{writeFile(t, dir, file.rel, file.content)}

	err := c.Put(key, outputs, dir)
	checkIOError(t, err, nil)

	path := filepath.Join(dir, file.rel)
	err = os.RemoveAll(filepath.Dir(path))
	if err != nil {
		t.Fatalf("remove %q: %v", filepath.Dir(path), err)
	}

	ok, err := c.Get(key, dir)
	checkIOError(t, err, nil)

	if !ok {
		t.Fatalf("expected the key %q to hit", key)
	}

	assertFile(t, path, file.content)
}

func TestGetMiss(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := t.TempDir()
	c := NewL1(root)

	ok, err := c.Get(key, dir)
	checkIOError(t, err, nil)

	if ok {
		t.Errorf("expected the key %q to miss", key)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %q: %v", dir, err)
	}

	if len(entries) != 0 {
		t.Error("copied files to source dir, even though cache was not hit")
	}
}

func TestPutGetEmptySentinel(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := t.TempDir()

	c := NewL1(root)
	err := c.Put(key, []string{}, dir)
	checkIOError(t, err, nil)

	ok, err := c.Get(key, dir)
	checkIOError(t, err, nil)

	if !ok {
		t.Errorf("expected the key %q to hit", key)
	}
}
