package engine

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	"github.com/inflame-ue/gocar/internal/cache"
	"github.com/inflame-ue/gocar/internal/runner"
	"github.com/inflame-ue/gocar/internal/task"
	"github.com/inflame-ue/gocar/internal/testutil"
)

func canonicalSpec() task.Tasks {
	return task.Tasks{
		"mkout": {
			Cmd: "mkdir -p out",
		},
		"alpha": {
			Cmd:     "echo 'alpha\n' > out/a.txt",
			Outputs: []string{"out/a.txt"},
			Deps:    []string{"mkout"},
		},
		"beta": {
			Cmd:     "echo 'beta\n' > out/b.txt",
			Outputs: []string{"out/b.txt"},
			Deps:    []string{"mkout"},
		},
		"merge": {
			Cmd:     "cat out/a.txt out/b.txt > out/all.txt",
			Inputs:  []string{"out/a.txt", "out/b.txt"},
			Outputs: []string{"out/all.txt"},
			Deps:    []string{"alpha", "beta"},
		},
		"check": {
			Cmd:    "wc -l out/all.txt",
			Inputs: []string{"out/all.txt"},
			Deps:   []string{"merge"},
		},
	}
}

type countingExecutor struct {
	Executor
	mu    sync.Mutex
	calls []string
}

func (ce *countingExecutor) RunTask(ctx context.Context, name string, t task.Task) error {
	ce.mu.Lock()
	ce.calls = append(ce.calls, name)
	ce.mu.Unlock()
	return ce.Executor.RunTask(ctx, name, t)
}

func (ce *countingExecutor) names() []string {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	return slices.Clone(ce.calls)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}
	return string(data)
}

func TestBuild_Integration(t *testing.T) {
	dir, root := t.TempDir(), t.TempDir()
	spec := canonicalSpec()

	real := runner.NewRunner(io.Discard, io.Discard, dir)
	ca := cache.NewL1(root)
	ctx := context.Background()

	// first a cold build, everything must execute
	cold := &countingExecutor{Executor: real}
	engine := NewEngine(dir, cold, ca)
	testutil.CheckNoErr(t, engine.Build(ctx, spec))

	assertRanExactly(t, cold.names(), []string{"alpha", "beta", "check", "merge", "mkout"})

	// store file content and remove from temp dir
	// so that a warm build can restore from cache instead
	a := readFile(t, filepath.Join(dir, "out", "a.txt"))
	b := readFile(t, filepath.Join(dir, "out", "b.txt"))
	all := readFile(t, filepath.Join(dir, "out", "all.txt"))
	for _, name := range []string{"out/a.txt", "out/b.txt", "out/all.txt"} {
		testutil.CheckNoErr(t, os.Remove(filepath.Join(dir, name)))
	}

	repeat := &countingExecutor{Executor: real}
	engine = NewEngine(dir, repeat, ca)
	testutil.CheckNoErr(t, engine.Build(ctx, spec))

	assertRanExactly(t, repeat.names(), []string{"check", "mkout"})

	if got := readFile(t, filepath.Join(dir, "out", "a.txt")); got != a {
		t.Errorf("restored a.txt = %q, want %q", got, a)
	}
	if got := readFile(t, filepath.Join(dir, "out", "b.txt")); got != b {
		t.Errorf("restored b.txt = %q, want %q", got, b)
	}
	if got := readFile(t, filepath.Join(dir, "out", "all.txt")); got != all {
		t.Errorf("restored all.txt = %q, want %q", got, all)
	}
}
