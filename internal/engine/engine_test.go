package engine

import (
	"context"
	"maps"
	"slices"
	"testing"

	"github.com/inflame-ue/gocar/internal/task"
)

type fakeExecutor struct {
	ran []string
}

func (fe *fakeExecutor) RunTask(ctx context.Context, name string, t task.Task) error {
	fe.ran = append(fe.ran, name)
	return nil
}

type put struct {
	key     string
	outputs []string
}

type fakeCache struct {
	hit  bool
	puts []put
}

func (fc *fakeCache) Get(key, dir string) (bool, error) {
	return fc.hit, nil
}

func (fc *fakeCache) Put(key string, outputs []string, dir string) error {
	fc.puts = append(fc.puts, put{key: key, outputs: outputs})
	return nil
}

func testSpec() map[string]task.Task {
	return map[string]task.Task{
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

func assertRunsBeforeDeps(t *testing.T, ran []string, spec map[string]task.Task) {
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

func TestBuild_NilCacheRunsEverything(t *testing.T) {
	dir, spec := t.TempDir(), testSpec()
	ctx := context.Background()
	ex := &fakeExecutor{ran: []string{}}

	engine := NewEngine(dir, ex, nil)
	if err := engine.Build(ctx, spec); err != nil {
		t.Fatal(err)
	}

	got, want := slices.Clone(ex.ran), slices.Sorted(maps.Keys(spec))
	assertRunsBeforeDeps(t, got, spec)

	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("expected %v to be executed, got %v instead", want, got)
	}
}

func TestBuild_CacheHitSkipExecution(t *testing.T) {
	dir, spec := t.TempDir(), testSpec()
	ctx := context.Background()
	ex := &fakeExecutor{ran: []string{}}
	cache := &fakeCache{hit: true, puts: []put{}}

	engine := NewEngine(dir, ex, cache)
	if err := engine.Build(ctx, spec); err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(ex.ran, []string{"D"}) {
		t.Errorf("expected only the no-outputs task D to run, instead %v ran", ex.ran)
	}

	if len(cache.puts) != 0 {
		t.Errorf("expected 0 puts to happen on cache hit, got %d instead", len(cache.puts))
	}
}
