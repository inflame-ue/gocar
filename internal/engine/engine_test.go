package engine

import (
	"context"
	"testing"

	"github.com/inflame-ue/gocar/internal/task"
)

func spec() map[string]task.Task {
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
			Inputs:  []string{"a.txt", "b.txt"},
			Outputs: []string{"c.txt"},
			Deps:    []string{"A", "B"},
		},
		"D": {
			Cmd:    "wc -l a.txt",
			Inputs: []string{"a.txt"},
			Deps:   []string{"A"},
		},
	}
}

type fakeExecutor struct {
	ran []string
}

func (fe *fakeExecutor) RunTask(ctx context.Context, name string, t task.Task) error {
	return nil
}

type fakeCache struct {
}

func (fc *fakeCache) Get(key, dir string) (bool, error) {
	return true, nil
}

func (fc *fakeCache) Put(key string, outputs []string, dir string) error {
	return nil
}

func TestEngine_NilCacheRunsEverything(t *testing.T) {
	dir := t.TempDir()
	_ = spec()
	_ = NewEngine(dir, &fakeExecutor{ran: []string{}}, &fakeCache{})
}
