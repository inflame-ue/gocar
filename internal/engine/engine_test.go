package engine

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/inflame-ue/gocar/internal/task"
	"github.com/inflame-ue/gocar/internal/testutil"
)

type fakeExecutor struct {
	mu    sync.Mutex
	ran   []string
	onRun func(ctx context.Context, name string, t task.Task) error
}

func (fe *fakeExecutor) RunTask(ctx context.Context, name string, t task.Task) error {
	fe.mu.Lock()
	fe.ran = append(fe.ran, name)
	fe.mu.Unlock()

	if fe.onRun != nil {
		return fe.onRun(ctx, name, t)
	}
	return nil
}

func (fe *fakeExecutor) ranNames() []string {
	fe.mu.Lock()
	defer fe.mu.Unlock()
	return slices.Clone(fe.ran)
}

type fakeCache struct {
	mu   sync.Mutex
	hit  bool
	gets []string
	puts []string
}

func (fc *fakeCache) Get(key, dir string) (bool, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.gets = append(fc.gets, key)
	return fc.hit, nil
}

func (fc *fakeCache) Put(key string, outputs []string, dir string) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.puts = append(fc.puts, outputs...)
	return nil
}

func (fc *fakeCache) getKeys() []string {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return slices.Clone(fc.gets)
}

func (fc *fakeCache) putOutputs() []string {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return slices.Clone(fc.puts)
}

func testSpec() task.Tasks {
	return task.Tasks{
		"A": {
			Cmd:     "echo 'a' > a.txt",
			Outputs: []string{"a.txt"},
		},
		"B": {
			Cmd:     "echo 'b' > b.txt",
			Outputs: []string{"b.txt"},
		},
		"C": {
			Cmd:     "cat a.txt b.txt > c.txt",
			Outputs: []string{"c.txt"},
			Deps:    []string{"A", "B"},
		},
		"D": {
			Cmd:  "wc -l a.txt",
			Deps: []string{"A"},
		},
	}
}

func assertRunsBeforeDeps(t *testing.T, ran []string, spec task.Tasks) {
	t.Helper()

	indices := make(map[string][]int)

	for name, task := range spec {
		for _, dep := range task.Deps {
			idx := slices.Index(ran, dep)
			if idx == -1 {
				t.Fatalf("task %q did not run, even though it should have", dep)
			}
			indices[name] = append(indices[name], idx)
		}
	}

	for name, idxs := range indices {
		for _, idx := range idxs {
			if slices.Index(ran, name) < idx {
				t.Errorf("task %q ran before dependency %q", name, ran[idx])
			}
		}
	}
}

func assertRanExactly(t *testing.T, got, want []string) {
	t.Helper()

	got = slices.Clone(got)
	want = slices.Clone(want)

	seen := make(map[string]int, len(got))
	for _, name := range got {
		seen[name]++
	}
	for name, n := range seen {
		if n > 1 {
			t.Errorf("task %q executed %d times", name, n)
		}
	}

	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("expected %v to be executed, got %v instead", want, got)
	}
}

func TestBuild_NilCacheRunsEverything(t *testing.T) {
	dir, spec := t.TempDir(), testSpec()
	ex := &fakeExecutor{}

	engine := NewEngine(dir, ex, nil)
	testutil.CheckNoErr(t, engine.Build(context.Background(), spec))

	assertRanExactly(t, ex.ranNames(), spec.Names())
	assertRunsBeforeDeps(t, ex.ranNames(), spec)
}

func TestBuild_CacheHitSkipExecution(t *testing.T) {
	dir, spec := t.TempDir(), testSpec()
	ex := &fakeExecutor{}
	cache := &fakeCache{hit: true}

	engine := NewEngine(dir, ex, cache)
	testutil.CheckNoErr(t, engine.Build(context.Background(), spec))

	if !slices.Equal(ex.ranNames(), []string{"D"}) {
		t.Errorf("expected only the no-outputs task D to run, got %v", ex.ranNames())
	}

	if got := cache.putOutputs(); len(got) != 0 {
		t.Errorf("expected 0 puts on cache hit, got %v", got)
	}

	if got := len(cache.getKeys()); got != 3 {
		t.Errorf("expected 3 cache Get calls (A, B, C), got %d", got)
	}
}

func TestBuild_CacheMissTriggersPut(t *testing.T) {
	dir, spec := t.TempDir(), testSpec()
	ex := &fakeExecutor{}
	cache := &fakeCache{}

	engine := NewEngine(dir, ex, cache)
	testutil.CheckNoErr(t, engine.Build(context.Background(), spec))

	assertRanExactly(t, ex.ranNames(), spec.Names())
	assertRunsBeforeDeps(t, ex.ranNames(), spec)

	puts := cache.putOutputs()
	slices.Sort(puts)
	if !slices.Equal(puts, []string{"a.txt", "b.txt", "c.txt"}) {
		t.Errorf("expected [a.txt b.txt c.txt] to be cached, got %v instead", puts)
	}
}

func TestBuild_ExecutorFailureStopsBuild(t *testing.T) {
	dir := t.TempDir()
	spec := task.Tasks{
		"A": {Cmd: "a", Outputs: []string{"a.txt"}},
		"B": {Cmd: "b", Deps: []string{"A"}, Outputs: []string{"b.txt"}},
		"C": {Cmd: "c", Deps: []string{"B"}, Outputs: []string{"c.txt"}},
	}
	errBoom := errors.New("boom")

	ex := &fakeExecutor{onRun: func(ctx context.Context, name string, t task.Task) error {
		if name == "B" {
			return errBoom
		}
		return nil
	}}
	cache := &fakeCache{}

	engine := NewEngine(dir, ex, cache)
	err := engine.Build(context.Background(), spec)
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected executor error, got %v", err)
	}

	if !slices.Equal(ex.ranNames(), []string{"A", "B"}) {
		t.Errorf("expected A and B to run before failure, got %v", ex.ranNames())
	}
	if got := cache.putOutputs(); len(got) != 1 {
		t.Errorf("expected 1 put (A) before failure, got %v", got)
	}
}

func TestBuild_ContextCancelledBeforeBuild(t *testing.T) {
	dir, spec := t.TempDir(), testSpec()
	ctx, cancel := context.WithCancel(context.Background())
	ex := &fakeExecutor{}
	cache := &fakeCache{hit: true}

	cancel()

	engine := NewEngine(dir, ex, cache)
	err := engine.Build(ctx, spec)
	if err == nil || !errors.Is(err, ctx.Err()) {
		t.Errorf("expected Build to fail with context cancellation, got %v", err)
	}
	if got := ex.ranNames(); len(got) != 0 {
		t.Errorf("expected no tasks to run, got %v", got)
	}
}

func TestBuild_ContextCancelledDuringBuild(t *testing.T) {
	dir := t.TempDir()
	spec := task.Tasks{
		"A": {Cmd: "a", Outputs: []string{"a.txt"}},
		"B": {Cmd: "b", Deps: []string{"A"}, Outputs: []string{"b.txt"}},
		"C": {Cmd: "c", Deps: []string{"B"}, Outputs: []string{"c.txt"}},
	}
	ctx, cancel := context.WithCancel(context.Background())

	ex := &fakeExecutor{onRun: func(ctx context.Context, name string, t task.Task) error {
		if name == "A" {
			cancel()
		}
		return nil
	}}

	engine := NewEngine(dir, ex, nil)
	err := engine.Build(ctx, spec)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Errorf("expected Build to fail with context cancellation, got %v", err)
	}
	if !slices.Equal(ex.ranNames(), []string{"A"}) {
		t.Errorf("expected only A to run before cancellation, got %v", ex.ranNames())
	}
}
