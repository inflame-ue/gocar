package cache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/inflame-ue/gocar/internal/testutil"
)

const key = "test-key"

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
	testutil.CheckNoErr(t, err)

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
	testutil.CheckNoErr(t, err)

	outputs = []string{writeFile(t, dir, files[1].rel, files[1].content)}
	err = c.Put(key, outputs, dir)
	testutil.CheckNoErr(t, err)

	assertFile(t, filepath.Join(root, key, files[0].rel), files[0].content)
}

func TestPutMissingSource(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := t.TempDir()
	c := NewL1(root)

	err := c.Put(key, []string{"missing.go"}, dir)

	wantErr := &IOError{Path: filepath.Join(dir, "missing.go")}
	tErr := testutil.CheckErrAs(t, err, wantErr)
	if tErr.Path != wantErr.Path {
		t.Errorf("expected path %q in error, got %q instead", wantErr.Path, tErr.Path)
	}

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

	wantErr := &PathError{Path: "../escaping.go"}
	tErr := testutil.CheckErrAs(t, err, wantErr)
	if tErr.Path != wantErr.Path {
		t.Errorf("expectedp path %q in error, got %q instead", wantErr.Path, tErr.Path)
	}
}

func TestPutGetRoundTrip(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := t.TempDir()
	c := NewL1(root)

	file := struct{ rel, content string }{"src/main.go", "package main"}

	var outputs = []string{writeFile(t, dir, file.rel, file.content)}

	err := c.Put(key, outputs, dir)
	testutil.CheckNoErr(t, err)

	path := filepath.Join(dir, file.rel)
	err = os.RemoveAll(filepath.Dir(path))
	if err != nil {
		t.Fatalf("remove %q: %v", filepath.Dir(path), err)
	}

	ok, err := c.Get(key, dir)
	testutil.CheckNoErr(t, err)

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
	testutil.CheckNoErr(t, err)

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
	testutil.CheckNoErr(t, err)

	ok, err := c.Get(key, dir)
	testutil.CheckNoErr(t, err)

	if !ok {
		t.Errorf("expected the key %q to hit", key)
	}
}
