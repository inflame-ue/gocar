package runner

import (
	"context"
	"errors"
	"fmt"
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
			return errors.New("task not in spec") // placeholder, will make a separate error type later
		}

		// run as a shell to not deal with manual arg determination and stuffs
		cmd := exec.CommandContext(ctx, "sh", "-c", t.Cmd)
		cmd.Stdout = r.Stdout
		cmd.Stderr = r.Stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("cmd.Run failed: %v", err) // also a placeholder for now
		}
	}

	return nil
}
