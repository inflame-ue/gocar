package runner

import (
	"context"
	"errors"
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

// Run provides a common use-case for the Runner.Run method, where
// stdout is os.Stdout and stderr is os.Stderr.
func Run(ctx context.Context, spec map[string]task.Task, order []string, dir string) error {
	return NewRunner(os.Stdout, os.Stderr, dir).Run(ctx, spec, order)
}

// Runner.Run orchestrates by using the order to extract tasks from the spec
// and then delegates actual command execution.
func (r *Runner) Run(ctx context.Context, spec map[string]task.Task, order []string) error {
	for _, name := range order {
		t, ok := spec[name]

		if !ok {
			return &TaskError{
				Task: name,
				Err:  errors.New("task name does not exist in spec"),
			}
		}

		// delegate actual execution of the command
		err := r.cmder.run(ctx, t.Cmd, r.Dir, r.Stdout, r.Stderr)
		if err != nil {
			return &TaskError{
				Task: name,
				Err:  err,
			}
		}
	}

	return nil
}
