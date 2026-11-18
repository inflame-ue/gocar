package engine

import (
	"context"

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

func (e *Engine) stale(name string, t task.Task, spec task.Tasks) (bool,  error) {
	return false, nil
}