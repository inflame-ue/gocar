package engine

import (
	"context"

	"github.com/inflame-ue/gocar/internal/fingerprint"
	"github.com/inflame-ue/gocar/internal/graph"
	"github.com/inflame-ue/gocar/internal/task"
)

type Executor interface {
	RunTask(ctx context.Context, name string, t task.Task) error
}

type Cache interface {
	Get(key, dir string) (bool, error)
	Put(key string, outputs []string, dir string) error
}

type Engine struct {
	Dir    string
	Runner Executor
	Cache  Cache
}

func NewEngine(dir string, runner Executor, cache Cache) *Engine {
	return &Engine{
		Dir:    dir,
		Runner: runner,
		Cache:  cache,
	}
}

// Build is the orchestration method that builds the DAG, computes
// the topological order, and handles the execution loop for each task.
func (e *Engine) Build(ctx context.Context, spec map[string]task.Task) error {
	g, err := graph.Build(spec)
	if err != nil {
		return err
	}

	order, err := g.Kahn()
	if err != nil {
		return err
	}

	fp := fingerprint.NewFingerprinter(e.Dir)
	keys := make(map[string]string)
	for _, name := range order {
		t := spec[name]

		// always run the task, if it does not have outputs
		// or no cache is specified for the build system.
		// rationale: the task with no outptus has side-effects
		if len(t.Outputs) == 0 || e.Cache == nil {
			if err := e.Runner.RunTask(ctx, name, t); err != nil {
				return err
			}
			continue
		}

		key, err := fp.Key(name, t, keys)
		if err != nil {
			return err
		}
		keys[name] = key

		hit, err := e.Cache.Get(key, e.Dir)
		if err != nil {
			return err
		}

		// we hit the cache, there is no reason to execute the task
		if hit {
			continue
		}

		// otherwise run the task and put the key --> outputs to e.Cache
		if err := e.Runner.RunTask(ctx, name, t); err != nil {
			return err
		}

		if err := e.Cache.Put(key, t.Outputs, e.Dir); err != nil {
			return err
		}
	}

	return nil
}
