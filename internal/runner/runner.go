package runner

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/inflame-ue/gocar/internal/task"
)

type Runner struct {
	Stdout io.Writer
	Stderr io.Writer
}

func NewRunner(stdout io.Writer, stderrr io.Writer) *Runner {
	return &Runner{
		Stdout: stdout,
		Stderr: stderrr,
	}
}

func Run(ctx context.Context, spec map[string]task.Task, order []string) error {
	return (&Runner{Stdout: os.Stdout, Stderr: os.Stderr}).Run(ctx, spec, order)
}

func (r *Runner) Run(ctx context.Context, spec map[string]task.Task, order []string) error {
	for _, name := range order {
		t, ok := spec[name]

		if !ok {
			return &TaskError{
				Task: name,
				Err:  errors.New("task name does not exist in spec"),
			}
		}

		// run as a shell to not deal with manual arg determination and stuffs
		cmd := exec.CommandContext(ctx, "sh", "-c", t.Cmd)
		cmd.Stdout = r.Stdout
		cmd.Stderr = r.Stderr

		err := cmd.Run()
		if err != nil {
			return &TaskError{
				Task: name,
				Err:  err,
			}
		}
	}

	return nil
}
