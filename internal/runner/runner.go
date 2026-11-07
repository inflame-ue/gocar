package runner

import (
	"context"
	"io"
	"os"

	"github.com/inflame-ue/gocar/internal/task"
)

type Runner struct {
	Stdout io.Writer
	Stderr io.Writer
	Dir    string
	cmder  commander
}

func NewRunner(stdout io.Writer, stderr io.Writer, dir string) *Runner {
	return &Runner{
		Stdout: stdout,
		Stderr: stderr,
		Dir:    dir,
		cmder:  &shellCommander{},
	}
}

// Run provides a common use-case for the Runner.RunTask method, where
// stdout is os.Stdout and stderr is os.Stderr.
func Run(ctx context.Context, name string, t task.Task, dir string) error {
	return NewRunner(os.Stdout, os.Stderr, dir).RunTask(ctx, name, t)
}

// RunTask performs a given task through the commander interface.
// Retursn a TaskError in case of an execution error.
func (r *Runner) RunTask(ctx context.Context, name string, t task.Task) error {
	err := r.cmder.run(ctx, t.Cmd, r.Dir, r.Stdout, r.Stderr)

	if err != nil {
		return &TaskError{
			Task: name,
			Err:  err,
		}
	}

	return nil
}
