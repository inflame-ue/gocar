package engine

import (
	"context"
	"fmt"
	"io"
	"runtime"

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

type Rebuilder int

const (
	RebuilderHash Rebuilder = iota
	RebuilderMTime
)

type Engine struct {
	Dir    string
	Runner Executor

	// Set Cache to nil to force cold builds every time.
	Cache Cache

	Jobs      int
	Rebuilder Rebuilder

	// Set Log to nil to supress reporting.
	Log     io.Writer
	Verbose bool
	report  reportState
}

// NewEngine returns a new [Engine] that builds the given spec in dir.
//
// dir is the project root and must match the directory the runner is
// configured with; otherwise artifacts are written to an unexpected location.
//
// runner executes each task's command. If runner is nil, [Engine.Build]
// panics as soon as a task needs to run (cache-only paths do not).
//
// cache stores and retrieves task outputs in the local cache.
// If cache is nil, every build is treated as a cold one.
func NewEngine(dir string, runner Executor, cache Cache) *Engine {
	return &Engine{
		Dir:    dir,
		Runner: runner,
		Cache:  cache,
	}
}

func (e *Engine) jobs() int {
	if e.Jobs <= 0 {
		return runtime.NumCPU()
	}
	return e.Jobs
}

// Build is the orchestration method that builds the DAG, computes
// the topological order, and handles the execution loop for each task.
func (e *Engine) Build(ctx context.Context, spec task.Tasks) (err error) {
	e.report.reset(len(spec))
	defer func() { e.report.finish(e.Log, err) }()

	g, err := graph.Build(spec)
	if err != nil {
		return err
	}

	if cycle, ok := g.Cycle(); ok {
		return &graph.CycleError{Path: cycle}
	}

	return e.run(ctx, spec, g)
}

// runOne processes a single task. It returns the task's key
// so that dependents can compute their own keys, plus the action outcome.
func (e *Engine) runOne(ctx context.Context, name string, t task.Task,
	depKeys map[string]string, spec task.Tasks, fp *fingerprint.Fingerprinter) (string, outcome, error) {

	if e.Rebuilder == RebuilderMTime {
		ran, err := e.runTimestamp(ctx, name, t, spec)
		if err != nil {
			return "", outcomeExecuted, err
		}
		if !ran {
			return "", outcomeUpToDate, nil
		}
		return "", outcomeExecuted, nil
	}

	// with no cache there is nothing to fingerprint or restore,
	// so just execute the task.
	if e.Cache == nil {
		return "", outcomeExecuted, e.Runner.RunTask(ctx, name, t)
	}

	key, err := fp.Key(name, t, depKeys)
	if err != nil {
		return "", outcomeExecuted, err
	}

	// always run the task, if it does not have outputs
	// rationale: the task with no outputs has side-effects
	if len(t.Outputs) == 0 {
		if err := e.Runner.RunTask(ctx, name, t); err != nil {
			return "", outcomeExecuted, err
		}
		fp.Clear()
		return key, outcomeExecuted, nil
	}

	hit, err := e.Cache.Get(key, e.Dir)
	if err != nil {
		return "", outcomeExecuted, fmt.Errorf("task %q, cache get: %w", name, err)
	}

	// we hit the cache, there is no reason to execute the task
	if hit {
		fp.Forget(t.Outputs...)
		return key, outcomeCached, nil
	}

	// otherwise run the task and put the key --> outputs to e.Cache
	if err := e.Runner.RunTask(ctx, name, t); err != nil {
		return "", outcomeExecuted, err
	}

	fp.Forget(t.Outputs...)
	if err := e.Cache.Put(key, t.Outputs, e.Dir); err != nil {
		return "", outcomeExecuted, fmt.Errorf("task %q, cache put: %w", name, err)
	}

	return key, outcomeExecuted, nil
}
