package fingerprint

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/inflame-ue/gocar/internal/task"
)

func checkInputErr(t *testing.T, got error, wantErr *InputError) {
	t.Helper()

	if got == nil && wantErr != nil {
		t.Fatal("expected an err, got no err instead")
	}

	if got != nil && wantErr == nil {
		t.Fatalf("expected no err, got %v instead", got)
	}

	if got == nil && wantErr == nil {
		return
	}

	var ie *InputError

	if ok := errors.As(got, &ie); !ok {
		t.Errorf("expected err of type %T, got %T instead", wantErr, got)
	}

	if ie.Task != wantErr.Task {
		t.Errorf("expected task %s, got %s instead", wantErr.Task, ie.Task)
	}

	if ie.Path != wantErr.Path {
		t.Errorf("expected path %s, got %s instead", wantErr.Path, ie.Path)
	}
}

func mustKeys(t *testing.T, dir string, spec map[string]task.Task, order []string) map[string]string {
	t.Helper()

	fp := NewFingerprinter(dir)
	keys, err := fp.Keys(spec, order)
	if err != nil {
		t.Fatal(err)
	}

	return keys
}

func TestKeysMTime(t *testing.T) {
	testDir := t.TempDir()

	file, err := os.Create(filepath.Join(testDir, "main.go"))
	if err != nil {
		t.Fatalf("file creation failed: %v", err)
	}
	defer file.Close()

	spec := map[string]task.Task{
		"build": {
			Cmd: "go build .",
			Inputs: []string{file.Name()},
		},
	}
	order := []string{"build"}
	
	oldKeys := mustKeys(t, testDir, spec, order)
	os.Chtimes(file.Name(), time.Time{}, time.Now())
	newKeys := mustKeys(t, testDir, spec,  order)

	if !maps.Equal(oldKeys, newKeys) {
		t.Error("hash for build should remain unchanged regardless of modification time")
	}
}

func TestKeysFileContent(t *testing.T) {
	testDir := t.TempDir()

	file, err := os.Create(filepath.Join(testDir, "main.go"))
	if err != nil {
		t.Fatalf("file creation failed: %v", err)
	}
	defer file.Close()

	_, err = file.Write([]byte("package main\n"))
	if err != nil {
		t.Fatalf("file write failed: %v", err)
	}

	spec := map[string]task.Task{
		"build": {
			Cmd: "go build .",
			Inputs: []string{file.Name()},
		},
	}
	order := []string{"build"}
	
	oldKeys := mustKeys(t, testDir, spec, order)
	_, err = file.Write([]byte("import \"testing\"\n"))
	if err != nil {
		t.Fatalf("file write failed: %v", err)
	}
	newKeys := mustKeys(t, testDir, spec,  order)

	if maps.Equal(oldKeys, newKeys) {
		t.Error("hash for build did not change, it should because the content of the input file changed")
	}
}