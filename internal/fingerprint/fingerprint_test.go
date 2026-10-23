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

func getKeys(t *testing.T, dir string, spec map[string]task.Task, order []string) map[string]string {
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

	_, err = file.Write([]byte("package main\n"))
	if err != nil {
		t.Fatalf("file write failed: %v", err)
	}

	spec := map[string]task.Task{
		"build": {
			Cmd:    "go build .",
			Inputs: []string{"main.go"},
		},
	}
	order := []string{"build"}

	oldKeys := getKeys(t, testDir, spec, order)
	os.Chtimes(file.Name(), time.Time{}, time.Now())
	newKeys := getKeys(t, testDir, spec, order)

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
			Cmd:    "go build .",
			Inputs: []string{"main.go"},
		},
	}
	order := []string{"build"}

	oldKeys := getKeys(t, testDir, spec, order)
	_, err = file.Write([]byte("import \"testing\"\n"))
	if err != nil {
		t.Fatalf("file write failed: %v", err)
	}
	newKeys := getKeys(t, testDir, spec, order)

	if maps.Equal(oldKeys, newKeys) {
		t.Error("hash for build did not change, it should because the content of the input file changed")
	}
}

func TestKeysMerkleCascade(t *testing.T) {
	testDir := makeTree(t)

	err := os.WriteFile(filepath.Join(testDir, "src", "main.go"), []byte("package main\n"), 0o644)
	if err != nil {
		t.Fatalf("file write failed: %v", err)
	}

	spec := map[string]task.Task{
		"fmt": {
			Cmd:    "go fmt main.go",
			Inputs: []string{"src/**/*.go"},
		},
		"vet": {
			Cmd:  "go vet .",
			Deps: []string{"fmt"},
		},
		"build": {
			Cmd:  "go build .",
			Deps: []string{"fmt", "vet"},
		},
		"lint": {
			Cmd: "golangci lint",
		},
	}
	order := []string{"lint", "fmt", "vet", "build"}

	oldKeys := getKeys(t, testDir, spec, order)
	err = os.WriteFile(filepath.Join(testDir, "src", "main.go"), []byte("import testing\n"), 0o644)
	if err != nil {
		t.Fatalf("file write failed: %v", err)
	}
	newKeys := getKeys(t, testDir, spec, order)

	if oldKeys["vet"] == newKeys["vet"] || oldKeys["build"] == newKeys["build"] {
		t.Error("per merkle cascade, change in a dependency should have changed the dependants")
	}

	if oldKeys["lint"] != newKeys["lint"] {
		t.Error("the task unrelated to the merkle chain should remain untouched")
	}
}

func TestKeysCmdChange(t *testing.T) {
	testDir := t.TempDir()

	spec := map[string]task.Task{
		"build": {
			Cmd: "go build .",
		},
	}
	order := []string{"build"}

	oldKeys := getKeys(t, testDir, spec, order)
	changedSpec := map[string]task.Task{
		"build": {
			Cmd: "go build main.go",
		},
	}
	newKeys := getKeys(t, testDir, changedSpec, order)

	if maps.Equal(oldKeys, newKeys) {
		t.Error("hash for build should change when the command changes")
	}
}

func TestKeysMissingFile(t *testing.T) {
	testDir := t.TempDir()

	spec := map[string]task.Task{
		"build": {
			Cmd:    "go build main.go",
			Inputs: []string{filepath.Join(testDir, "main.go")},
		},
	}
	order := []string{"build"}
	wantErr := &InputError{
		Task: "build",
		Path: spec["build"].Inputs[0],
	}

	_, err := NewFingerprinter(testDir).Keys(spec, order)
	checkInputErr(t, err, wantErr)
}

func TestKeysIdenticalFileContent(t *testing.T) {
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

	otherFile, err := os.Create(filepath.Join(testDir, "other.go"))
	if err != nil {
		t.Fatalf("file creation failed: %v", err)
	}
	defer file.Close()

	_, err = otherFile.Write([]byte("package main\n"))
	if err != nil {
		t.Fatalf("file write failed: %v", err)
	}

	spec := map[string]task.Task{
		"build": {
			Cmd:    "go build main.go",
			Inputs: []string{"main.go"},
		},
		"otherBuild": {
			Cmd:    "go build other.go",
			Inputs: []string{"other.go"},
		},
	}
	order := []string{"build", "otherBuild"}

	keys := getKeys(t, testDir, spec, order)

	if keys["build"] == keys["otherBuild"] {
		t.Error("hash for build and run should not be the same, since it should depend on filepath as well as one content")
	}
}

func TestKeysNewFile(t *testing.T) {
	testDir := makeTree(t)

	spec := map[string]task.Task{
		"build": {
			Cmd:    "go build .",
			Inputs: []string{"src/**/*.go"},
		},
	}
	order := []string{"build"}

	oldKeys := getKeys(t, testDir, spec, order)
	file, err := os.Create(filepath.Join(testDir, "src/main.py"))
	if err != nil {
		t.Fatalf("could not create file \"src/main.py\": %v", err)
	}
	file.Close()
	newKeys := getKeys(t, testDir, spec, order)
	if !maps.Equal(oldKeys, newKeys) {
		t.Error("hash for build should not change after a file that is not matched by glob expansion is introduced into the cwd")
	}

	file, err = os.Create(filepath.Join(testDir, "src/additional.go"))
	if err != nil {
		t.Fatalf("could not create file \"src/additional.go\": %v", err)
	}
	file.Close()
	newKeys = getKeys(t, testDir, spec, order)
	if maps.Equal(oldKeys, newKeys) {
		t.Error("hash for build should chain after a new file is added to the directory under glob expansion")
	}
}
