package engine

import (
	"context"
	"fmt"

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

// NewEngine returns a new [Engine] that builds the given spec in dir.
//
// dir is the project root and must match the directory the runner is
// configured with; otherwise artifacts are written to an unexpected location.
//
// runner executes each task's command. If runner is nil, [Engine.Build]
// panics as soon as a task needs to run (cache-only paths do not).
//
// cache stores and retrieves task outputs, local (L1) or remote (L2).
// If cache is nil, every build is treated as a cold one.
func NewEngine(dir string, runner Executor, cache Cache) *Engine {
	return &Engine{
		Dir:    dir,
		Runner: runner,
		Cache:  cache,
	}
}

// Build is the orchestration method that builds the DAG, computes
// the topological order, and handles the execution loop for each task.
func (e *Engine) Build(ctx context.Context, spec task.Tasks) error {
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
		if err := ctx.Err(); err != nil {
			return err
		}

		key, err := e.runOne(ctx, name, spec[name], keys, fp)
		if err != nil {
			return err
		}
		keys[name] = key
	}

	return nil
}

// runOne processes a single task. It returns the task's key
// so that dependents can compute their own keys.
func (e *Engine) runOne(ctx context.Context, name string, t task.Task,
	depKeys map[string]string, fp *fingerprint.Fingerprinter) (string, error) {

	// with no cache there is nothing to fingerprint or restore,
	// so just execute the task.
	if e.Cache == nil {
		return "", e.Runner.RunTask(ctx, name, t)
	}

	key, err := fp.Key(name, t, depKeys)
	if err != nil {
		return "", err
	}

	// always run the task, if it does not have outputs
	// rationale: the task with no outputs has side-effects
	if len(t.Outputs) == 0 {
		if err := e.Runner.RunTask(ctx, name, t); err != nil {
			return "", err
		}
		fp.Clear()
		return key, nil
	}

	hit, err := e.Cache.Get(key, e.Dir)
	if err != nil {
		return "", fmt.Errorf("task %q, cache get: %w", name, err)
	}

	// we hit the cache, there is no reason to execute the task
	if hit {
		fp.Forget(t.Outputs...)
		return key, nil
	}

	// otherwise run the task and put the key --> outputs to e.Cache
	if err := e.Runner.RunTask(ctx, name, t); err != nil {
		return "", err
	}

	fp.Forget(t.Outputs...)
	if err := e.Cache.Put(key, t.Outputs, e.Dir); err != nil {
		return "", fmt.Errorf("task %q, cache put: %w", name, err)
	}

	return key, nil
}
