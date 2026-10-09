package runner

import (
	"context"
	"io"
	"os/exec"
)

type commander interface {
	run(ctx context.Context, cmd, dir string, stdout, stderr io.Writer) error
}

type shellCommander struct{}

func (sc *shellCommander) run(ctx context.Context, cmd, dir string, stdout, stderr io.Writer) error {
	c := exec.CommandContext(ctx, "sh", "-c", cmd)
	c.Stdout, c.Stderr, c.Dir = stdout, stderr, dir
	return c.Run()
}
