package engine

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/inflame-ue/gocar/internal/glob"
	"github.com/inflame-ue/gocar/internal/task"
)

func (e *Engine) runTimestamp(ctx context.Context, name string, t task.Task, spec task.Tasks) error {
	stale, err := e.stale(name, t, spec)
	if err != nil {
		return err
	}
	if !stale {
		return nil
	}

	return e.Runner.RunTask(ctx, name, t)
}

func (e *Engine) stale(name string, t task.Task, spec task.Tasks) (bool, error) {
	if len(t.Outputs) == 0 {
		return true, nil
	}

	newest, err := e.newestInput(name, t)
	if err != nil {
		return false, err
	}

	depNewest, missing, err := e.newestDepOutput(name, t, spec)
	if err != nil {
		return false, err
	}
	if missing {
		return true, nil
	}
	if depNewest.After(newest) {
		newest = depNewest
	}

	for _, output := range t.Outputs {
		path := filepath.Join(e.Dir, filepath.FromSlash(output))

		info, err := os.Stat(path)
		if errors.Is(err, fs.ErrNotExist) {
			return true, nil
		}
		if err != nil {
			return false, fmt.Errorf("task %q, output %q: %w", name, output, err)
		}
		if !info.Mode().IsRegular() {
			return true, nil
		}
		if info.ModTime().Before(newest) {
			return true, nil
		}
	}

	return false, nil
}

func (e *Engine) newestInput(name string, t task.Task) (time.Time, error) {
	var newest time.Time

	for _, pattern := range t.Inputs {
		matches, err := glob.Expand(e.Dir, pattern)
		if err != nil {
			return time.Time{}, fmt.Errorf("task %q, input %q: %w", name, pattern, err)
		}

		for _, rel := range matches {
			info, err := os.Stat(filepath.Join(e.Dir, filepath.FromSlash(rel)))
			if err != nil {
				return time.Time{}, fmt.Errorf("task %q, input %q: %w", name, rel, err)
			}
			if info.ModTime().After(newest) {
				newest = info.ModTime()
			}
		}
	}

	return newest, nil
}

func (e *Engine) newestDepOutput(name string, t task.Task, spec task.Tasks) (time.Time, bool, error) {
	var newest time.Time

	for _, dep := range t.Deps {
		d, ok := spec[dep]
		if !ok {
			return time.Time{}, false, fmt.Errorf("task %q, dependency %q: not found", name, dep)
		}

		for _, output := range d.Outputs {
			path := filepath.Join(e.Dir, filepath.FromSlash(output))

			info, err := os.Stat(path)
			if errors.Is(err, fs.ErrNotExist) {
				return time.Time{}, true, nil
			}
			if err != nil {
				return time.Time{}, false, fmt.Errorf("task %q, dependency %q output %q: %w",
					name, dep, output, err)
			}
			if info.ModTime().After(newest) {
				newest = info.ModTime()
			}
		}
	}

	return newest, false, nil
}
